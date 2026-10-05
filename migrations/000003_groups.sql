-- Create groups table based on YouTrack User Groups API (UserGroup entity)
CREATE TABLE IF NOT EXISTS groups (
  -- The identifier of the user group.
  id               VARCHAR(64) PRIMARY KEY,
  -- The name of the group.
  name             VARCHAR(255) NOT NULL UNIQUE,
  -- The description of the group, shown to users who can view or update it.
  description      TEXT,
  -- The ID of the group this group is nested under. Every group in YouTrack is
  -- a nested group whose root is the All Users group, and a group inherits the
  -- roles assigned to its ancestors down the whole subtree. NULL only for the
  -- root and for customer groups, which the docs forbid from being nested.
  -- No ON DELETE action, so RESTRICT: a group that still has subgroups cannot
  -- be deleted until they are re-parented or removed. That is deliberate, since
  -- the docs make a replacement group part of the delete flow and
  -- ON DELETE CASCADE would silently take a whole subtree with it.
  parent_group_id  VARCHAR(64) REFERENCES groups (id),
  -- The ID of the group in Hub, used to match groups between YouTrack and Hub.
  ring_id          VARCHAR(64),
  -- The number of users in the group.
  users_count      BIGINT NOT NULL DEFAULT 0,
  -- The URL of the group logo.
  icon             TEXT,
  -- True if this group contains all users, otherwise false.
  all_users_group  BOOLEAN NOT NULL DEFAULT FALSE,
  -- True if this is the default Registered Users group. Distinguishes it from a
  -- user-created group that merely has Auto-join enabled, which is otherwise
  -- indistinguishable: YouTrack marks the group through the RegisteredUsersGroup
  -- entity subclass, not through a setting.
  registered_users_group BOOLEAN NOT NULL DEFAULT FALSE,
  -- True if this group is a helpdesk customer group, otherwise false.
  -- Read-only in YouTrack, exposed only through the Workflow API as
  -- UserGroup.isCustomerGroup. The REST resource /api/groups does not return it,
  -- so this column cannot be populated by a plain REST sync. See
  -- docs/research/youtrack_group_type_field_research.md.
  -- Available since YouTrack 2026.2. A boolean is sufficient rather than an enum
  -- because the docs fix the type at creation time and allow no conversion in
  -- either direction, so false means a user group with no third state.
  is_customer_group BOOLEAN NOT NULL DEFAULT FALSE,
  -- True if new accounts are added to this group automatically.
  auto_join        BOOLEAN NOT NULL DEFAULT FALSE,
  -- The email domain whose accounts auto-join this group at first login. NULL
  -- when Auto-join is not domain driven. This is the mechanism the docs describe
  -- for customer groups, which have no roles and rely on Auto-join alone.
  auto_join_domain VARCHAR(255),
  -- True if members of this group must enable two-factor authentication. Until
  -- they do, their granted access is reduced to Update Self rather than revoked.
  require_two_factor_authentication BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at       TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at       TIMESTAMPTZ,
  -- A group cannot be nested under itself.
  CONSTRAINT groups_parent_not_self CHECK (parent_group_id IS NULL OR parent_group_id <> id),
  -- Customer groups cannot be nested under other groups.
  CONSTRAINT groups_customer_not_nested CHECK (NOT is_customer_group OR parent_group_id IS NULL),
  -- The All Users and Registered Users markers are mutually exclusive, since a
  -- single group cannot be both subclasses of UserGroup.
  CONSTRAINT groups_system_markers_exclusive
    CHECK (NOT (all_users_group AND registered_users_group))
);

CREATE INDEX IF NOT EXISTS idx_groups_name ON groups (name);
CREATE INDEX IF NOT EXISTS idx_groups_ring_id ON groups (ring_id);
CREATE INDEX IF NOT EXISTS idx_groups_all_users_group ON groups (all_users_group);
CREATE INDEX IF NOT EXISTS idx_groups_registered_users_group ON groups (registered_users_group);
CREATE INDEX IF NOT EXISTS idx_groups_is_customer_group ON groups (is_customer_group);
CREATE INDEX IF NOT EXISTS idx_groups_auto_join ON groups (auto_join);
-- Serves the search filter "Has subgroups" and the inherited-role walk up the
-- nesting tree, both of which read children by parent.
CREATE INDEX IF NOT EXISTS idx_groups_parent_group_id ON groups (parent_group_id);

-- ===========================================================================
-- Group membership
-- ===========================================================================
-- groups.users_count is a denormalized count read-only in YouTrack, so it needs
-- a source to be reproducible from. This table holds that source: the direct
-- memberships only, mirroring the /api/groups/{groupID}/ownUsers resource.
-- Transitive membership is the union over the nesting tree and is never stored.
CREATE TABLE IF NOT EXISTS group_members (
  -- The group the user was added to directly.
  group_id   VARCHAR(64) NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  -- The user account.
  user_id    VARCHAR(64) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  -- Application audit timestamp when this record was created.
  created_at TIMESTAMPTZ,
  PRIMARY KEY (group_id, user_id)
);

-- Serves the reverse lookup "which groups is this user in", used when resolving
-- the roles a user inherits.
CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members (user_id);

-- Transitive membership, mirroring the /api/groups/{groupID}/users resource,
-- which the docs describe as including transitive users. The walk goes down the
-- nesting tree, not up: a group nested under another is a subset of it, since the
-- docs have the child inherit the parent's roles and that is only sound while
-- every member of the child is already a member of the parent. So the transitive
-- users of a group are its own members plus those of every group nested under it.
-- The path column guards against cycles in the nesting tree, so a malformed
-- parent_group_id chain cannot send the recursion into an infinite loop.
CREATE OR REPLACE VIEW group_members_transitive AS
WITH RECURSIVE descendants AS (
    SELECT g.id AS origin_id, g.id AS current_id, ARRAY[g.id]::VARCHAR(64)[] AS path
    FROM groups g
  UNION ALL
    SELECT d.origin_id, c.id, (d.path || c.id)::VARCHAR(64)[] AS path
    FROM descendants d
    JOIN groups parent ON parent.id = d.current_id
    JOIN groups c ON c.parent_group_id = parent.id
    WHERE NOT c.id = ANY (d.path)
)
SELECT DISTINCT d.origin_id AS group_id, gm.user_id
FROM descendants d
JOIN group_members gm ON gm.group_id = d.current_id;

-- ===========================================================================
-- Visible to / Updatable by
-- ===========================================================================
-- Starting with 2025.3 access to user groups is global rather than permission
-- based: a non-admin can read a group if they are in a group listed under
-- Visible to, and can update it if they are listed under Updatable by. Each
-- listing accepts a mix of user accounts and groups, so the target is a
-- discriminated pair rather than a single foreign key.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_visibility_setting') THEN
    CREATE TYPE group_visibility_setting AS ENUM ('VISIBLE_TO', 'UPDATABLE_BY');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS group_visibility (
  -- The group whose visibility is being restricted.
  group_id     VARCHAR(64) NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  -- Which of the two settings this row belongs to.
  setting      group_visibility_setting NOT NULL,
  -- The user account listed in the setting. Exactly one of user_id and
  -- target_group_id is set per row, enforced by the check constraint below.
  user_id      VARCHAR(64) REFERENCES users (id) ON DELETE CASCADE,
  -- The group listed in the setting. Set when the entry names a group rather
  -- than an individual account.
  target_group_id VARCHAR(64) REFERENCES groups (id) ON DELETE CASCADE,
  -- Application audit timestamp when this record was created.
  created_at   TIMESTAMPTZ,
  -- A setting accepts either an account or a group, never both and never
  -- neither.
  CONSTRAINT group_visibility_single_target
    CHECK (num_nonnulls(user_id, target_group_id) = 1)
);

-- A composite unique index rather than a table-level UNIQUE constraint, because
-- plain UNIQUE treats the unused NULL column as distinct and would let the same
-- target be listed twice for the same setting. COALESCE maps the unused column
-- to an empty string so both branches collapse into one key.
CREATE UNIQUE INDEX IF NOT EXISTS idx_group_visibility_unique
  ON group_visibility (group_id, setting, COALESCE(user_id, ''), COALESCE(target_group_id, ''));
-- Serves "can I read this group", the check the docs gate the groups list on.
CREATE INDEX IF NOT EXISTS idx_group_visibility_lookup
  ON group_visibility (setting, COALESCE(target_group_id, ''), COALESCE(user_id, ''));

-- Default groups, matching the two groups that every YouTrack installation ships with.
-- Source: https://www.jetbrains.com/help/youtrack/server/default-user-groups.html
-- Both are Auto-join enabled, and the docs state their names and the Auto-join
-- setting cannot be changed, nor can they be deleted.
-- Every group seeded in this file is a user group, so is_customer_group is omitted
-- from the INSERT columns and stays at its FALSE default.
-- The docs state that every group other than the All Users root is nested under
-- All Users, so parent_group_id is seeded to ALL_USERS_GROUP throughout.

-- All Users: every account is added here automatically, including the guest account.
-- The docs assign it no role in any project. ALL_USERS_GROUP is the ID already
-- declared in internal/services/roles_service/assignment.go as AllUsersGroupID.
INSERT INTO groups (id, name, all_users_group, auto_join) VALUES
    ('ALL_USERS_GROUP', 'All Users', TRUE, TRUE)
    ON CONFLICT (id) DO NOTHING;

-- Registered Users: every account created in YouTrack is added here automatically.
-- Assigned the Observer role in the Global project by default. The guest account
-- is the only account excluded from this group.
INSERT INTO groups (id, name, parent_group_id, all_users_group, registered_users_group, auto_join)
VALUES
    ('REGISTERED_USERS_GROUP', 'Registered Users', 'ALL_USERS_GROUP', FALSE, TRUE, TRUE)
    ON CONFLICT (id) DO NOTHING;

-- System-generated groups, created by the upgrade migrations that moved globally
-- scoped permissions out of the default roles without reducing anyone's access.
-- Source: https://www.jetbrains.com/help/youtrack/server/system-generated-groups.html
-- Unlike the groups above, these are conditional: the upgrade only creates each
-- one when it finds users who would otherwise lose access. A fresh installation
-- has none of them, so they are all recorded with auto_join = false and no members.

-- Holds the access to the optional Reports feature, replacing the retired
-- Read Report and Create Report permissions (2026.1).
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('REPORTS_FEATURE_GROUP', 'Reports Feature Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;

-- Holds the access to the optional Read Groups feature, preserving access for
-- users who could read groups under the legacy Read Group permission (2026.1).
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('READ_GROUPS_FEATURE_GROUP', 'Read Groups Feature Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;

-- Isolates the globally scoped Read User Details permission, which was removed
-- from the default Project Admin and Contributor roles (2026.1). Named
-- Read User Full Group on installations upgraded before 2026.2.
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('READ_USER_DETAILS_GROUP', 'Read User Details Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;

-- Isolates the globally scoped Create Project permission, which was removed from
-- the default Project Admin role. Assigned the Project Creator role by default.
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('CREATE_PROJECT_GROUP', 'Create Project Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;

-- Isolates the globally scoped Create User permission, which was removed from
-- the default Project Admin role. Assigned the User Manager role by default.
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('CREATE_USER_GROUP', 'Create User Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;

-- Isolates the globally scoped Read Organization permission, which was removed
-- from the default Project Admin and Contributor roles (2026.1).
INSERT INTO groups (id, name, parent_group_id, all_users_group, auto_join) VALUES
    ('READ_ORGANIZATION_GROUP', 'Read Organization Group', 'ALL_USERS_GROUP', FALSE, FALSE)
    ON CONFLICT (id) DO NOTHING;