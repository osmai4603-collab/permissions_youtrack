-- ===========================================================================
-- Assigned roles (AssignedRole entity)
-- ===========================================================================
-- Source: https://www.jetbrains.com/help/youtrack/devportal/api-entity-AssignedRole.html
-- "Represents a role assigned to a user or a group in a specific scope."
-- Available since YouTrack 2026.1.
--
-- This is the central junction table that answers: "who holds which role and
-- where?" YouTrack's /api/assignedRoles resource models this as a first-class
-- entity with three relationships:
--
--   role   → Role          (the role being granted)
--   scope  → AccessScope   (where: Global, Organization, or Project)
--   holder → User | Group  (to whom)
--
-- The scope is polymorphic in the API (GlobalScope has no extra fields,
-- OrganizationScope carries an organization reference, ProjectScope carries a
-- project reference). We flatten this into an enum discriminator + two nullable
-- FK columns, following the same discriminated-pair pattern used in
-- group_visibility.

-- ---------------------------------------------------------------------------
-- 1. Scope type enum
-- ---------------------------------------------------------------------------
-- Maps the three AccessScope subtypes from the API:
--   GlobalScope       → GLOBAL
--   OrganizationScope → ORGANIZATION
--   ProjectScope      → PROJECT
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'access_scope_type') THEN
    CREATE TYPE access_scope_type AS ENUM ('GLOBAL', 'ORGANIZATION', 'PROJECT');
  END IF;
END$$;

-- ---------------------------------------------------------------------------
-- 2. Main table
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS assigned_roles (
  -- The database ID of the assigned role. Read-only in the API.
  id              VARCHAR(64) PRIMARY KEY,

  -- ── role ──────────────────────────────────────────────────────────────
  -- The role that is assigned.
  -- Source: AssignedRole.role → Role entity.
  role_id         VARCHAR(64) NOT NULL REFERENCES roles (id) ON DELETE CASCADE,

  -- ── scope ─────────────────────────────────────────────────────────────
  -- The scope where the role is assigned.
  -- Source: AssignedRole.scope → AccessScope (GlobalScope | OrganizationScope | ProjectScope).
  scope_type      access_scope_type NOT NULL,

  -- For PROJECT scope: the project this assignment applies to.
  -- Source: ProjectScope.project → Project.
  -- NULL when scope_type is GLOBAL or ORGANIZATION.
  -- No FK constraint yet: no projects table exists in the schema. Will be
  -- added as an ALTER TABLE when the projects migration lands, mirroring the
  -- pattern used in group_customer_projects.project_id.
  project_id      VARCHAR(64),

  -- For ORGANIZATION scope: the organization this assignment applies to.
  -- Source: OrganizationScope.organization → Organization.
  -- NULL when scope_type is GLOBAL or PROJECT.
  -- No FK constraint yet: no organizations table exists in the schema.
  organization_id VARCHAR(64),

  -- ── holder ────────────────────────────────────────────────────────────
  -- The user or group that holds the role. Read-only in the API; once
  -- created, the holder cannot be changed—delete and re-create instead.
  -- Exactly one of user_id / group_id is set per row.
  -- Source: AssignedRole.holder → User | UserGroup.
  user_id         VARCHAR(64) REFERENCES users (id) ON DELETE CASCADE,
  group_id        VARCHAR(64) REFERENCES groups (id) ON DELETE CASCADE,

  -- ── audit ─────────────────────────────────────────────────────────────
  created_at      TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ,

  -- ── constraints ───────────────────────────────────────────────────────

  -- A holder is either a user or a group, never both and never neither.
  -- Mirrors the group_visibility.group_visibility_single_target pattern.
  CONSTRAINT assigned_roles_single_holder
    CHECK (num_nonnulls(user_id, group_id) = 1),

  -- GLOBAL scope must not reference a project or organization.
  CONSTRAINT assigned_roles_global_scope_clean
    CHECK (scope_type <> 'GLOBAL'
           OR (project_id IS NULL AND organization_id IS NULL)),

  -- PROJECT scope must carry a project_id and no organization_id.
  CONSTRAINT assigned_roles_project_scope_has_project
    CHECK (scope_type <> 'PROJECT'
           OR (project_id IS NOT NULL AND organization_id IS NULL)),

  -- ORGANIZATION scope must carry an organization_id and no project_id.
  CONSTRAINT assigned_roles_org_scope_has_org
    CHECK (scope_type <> 'ORGANIZATION'
           OR (organization_id IS NOT NULL AND project_id IS NULL)),

  -- Customer groups cannot hold roles. The docs are explicit:
  --   "Customer groups don't grant roles or permissions."
  --   "Roles can't be assigned to customer groups."
  -- This is enforced via trigger below because CHECK cannot subquery.

  -- Prevent the exact same assignment from being recorded twice.
  -- A role is uniquely assigned to a holder in a specific scope.
  CONSTRAINT assigned_roles_unique_assignment
    UNIQUE (role_id, scope_type,
            COALESCE(project_id, ''),
            COALESCE(organization_id, ''),
            COALESCE(user_id, ''),
            COALESCE(group_id, ''))
);

-- ---------------------------------------------------------------------------
-- 3. Indexes
-- ---------------------------------------------------------------------------

-- "Which roles does this group have?" — the primary group-role lookup.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_group_id
  ON assigned_roles (group_id) WHERE group_id IS NOT NULL;

-- "Which roles does this user have?" — the primary user-role lookup.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_user_id
  ON assigned_roles (user_id) WHERE user_id IS NOT NULL;

-- "Who has this role?" — reverse lookup from role to holders.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_role_id
  ON assigned_roles (role_id);

-- "All assignments in this project" — used when listing project team roles.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_project_id
  ON assigned_roles (project_id) WHERE project_id IS NOT NULL;

-- "All assignments in this organization" — used when listing org-level roles.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_organization_id
  ON assigned_roles (organization_id) WHERE organization_id IS NOT NULL;

-- Scope type filter — useful when separating global from scoped assignments.
CREATE INDEX IF NOT EXISTS idx_assigned_roles_scope_type
  ON assigned_roles (scope_type);

-- ---------------------------------------------------------------------------
-- 4. Customer-group guard trigger
-- ---------------------------------------------------------------------------
-- Customer groups cannot be assigned roles. This constraint requires reading
-- groups.is_customer_group, so it must be a trigger rather than a CHECK.
-- Follows the same pattern as group_customer_projects_require_customer_group.
CREATE OR REPLACE FUNCTION assigned_roles_reject_customer_group()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.group_id IS NOT NULL THEN
    IF (SELECT is_customer_group FROM groups WHERE id = NEW.group_id) THEN
      RAISE EXCEPTION
        'group % is a customer group and cannot be assigned roles',
        NEW.group_id;
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_assigned_roles_reject_customer_group ON assigned_roles;

CREATE TRIGGER trg_assigned_roles_reject_customer_group
  BEFORE INSERT OR UPDATE ON assigned_roles
  FOR EACH ROW EXECUTE FUNCTION assigned_roles_reject_customer_group();

-- ---------------------------------------------------------------------------
-- 5. Effective-roles view
-- ---------------------------------------------------------------------------
-- Resolves the complete set of roles a group holds, including roles inherited
-- from ancestor groups via the nesting tree. The docs state:
--   "The group also inherits roles that are assigned to a parent group."
-- This view walks UP the tree (child → parent → grandparent → …) to collect
-- all inherited assignments, with cycle detection.
CREATE OR REPLACE VIEW group_effective_roles AS
WITH RECURSIVE ancestors AS (
    -- Base: the group itself.
    SELECT g.id AS group_id, g.id AS current_id, ARRAY[g.id]::VARCHAR(64)[] AS path
    FROM groups g
  UNION ALL
    -- Recursive: walk to parent.
    SELECT a.group_id, g.parent_group_id, (a.path || g.parent_group_id)::VARCHAR(64)[] AS path
    FROM ancestors a
    JOIN groups g ON g.id = a.current_id
    WHERE g.parent_group_id IS NOT NULL
      AND NOT g.parent_group_id = ANY(a.path)
)
SELECT DISTINCT
  a.group_id,
  ar.role_id,
  ar.scope_type,
  ar.project_id,
  ar.organization_id,
  -- Indicates whether this assignment is directly on the group or inherited.
  CASE WHEN ar.group_id = a.group_id THEN FALSE ELSE TRUE END AS inherited
FROM ancestors a
JOIN assigned_roles ar ON ar.group_id = a.current_id;

-- ---------------------------------------------------------------------------
-- 6. Seed: default role assignment
-- ---------------------------------------------------------------------------
-- The docs state that Registered Users is assigned the Observer role in the
-- Global scope by default.
-- Source: https://www.jetbrains.com/help/youtrack/server/default-user-groups.html
--   "By default, this group is assigned the Observer role in the Global project."
INSERT INTO assigned_roles (id, role_id, scope_type, group_id)
VALUES ('REGISTERED_USERS_OBSERVER_GLOBAL', 'OBSERVER', 'GLOBAL', 'REGISTERED_USERS_GROUP')
ON CONFLICT (id) DO NOTHING;
