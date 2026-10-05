-- Create roles table based on YouTrack API (Role entity)
CREATE TABLE IF NOT EXISTS roles (
  -- The identifier of the role.
  id          VARCHAR(64) PRIMARY KEY,
  -- The name of the role.
  name        VARCHAR(255) NOT NULL UNIQUE,
  -- The description of the role.
  description TEXT,
  -- The list of permissions granted by this role.
  permissions JSONB,
  -- Indicates whether this role is built-in to YouTrack.
  built_in    BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at  TIMESTAMPTZ,
  -- Application audit timestamp when this record was last updated.
  updated_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
CREATE INDEX IF NOT EXISTS idx_roles_built_in ON roles (built_in);

-- Default built-in roles, matching internal/services/roles_service/role.go BuildDefaultRoles().
-- Permissions are pre-resolved with the implied-permission closure, so they match
-- the values the service computes at runtime.

-- System Admin: full access across every scope (all catalog permissions).
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('SYSTEM_ADMIN', 'System Admin',
     'Full administrative access to the entire YouTrack installation and all projects and organizations.',
     ARRAY['ADMIN_READ_APP', 'ADMIN_UPDATE_APP', 'APPLY_COMMANDS_SILENTLY', 'CREATE_ARTICLE', 'CREATE_ARTICLE_COMMENT', 'CREATE_ATTACHMENT_ISSUE', 'CREATE_COMMENT', 'CREATE_ISSUE', 'CREATE_NOT_OWN_WORK_ITEM', 'CREATE_ORGANIZATION', 'CREATE_PROJECT', 'CREATE_USER', 'CREATE_WATCH_FOLDER', 'CREATE_WORK_ITEM', 'DELETE_ARTICLE', 'DELETE_ARTICLE_COMMENT', 'DELETE_ATTACHMENT_ISSUE', 'DELETE_COMMENT', 'DELETE_ISSUE', 'DELETE_NOT_OWN_COMMENT', 'DELETE_ORGANIZATION', 'DELETE_PROJECT', 'DELETE_USER', 'DELETE_WATCH_FOLDER', 'LINK_ISSUE', 'PRIVATE_READ_ISSUE', 'PRIVATE_UPDATE_ISSUE', 'READ_APP_CONTENT', 'READ_ARTICLE', 'READ_ARTICLE_COMMENT', 'READ_COMMENT', 'READ_HIDDEN_STUFF', 'READ_ISSUE', 'READ_ORGANIZATION', 'READ_PROJECT', 'READ_PROJECT_BASIC', 'READ_USER', 'READ_USER_BASIC', 'READ_WORK_ITEM', 'SHARE_WATCH_FOLDER', 'UPDATE_APP_CONTENT', 'UPDATE_ARTICLE', 'UPDATE_ARTICLE_COMMENT', 'UPDATE_ATTACHMENT_ISSUE', 'UPDATE_COMMENT', 'UPDATE_ISSUE', 'UPDATE_NOT_OWN_COMMENT', 'UPDATE_NOT_OWN_WORK_ITEM', 'UPDATE_ORGANIZATION', 'UPDATE_PROFILE', 'UPDATE_PROJECT', 'UPDATE_USER', 'UPDATE_WATCHERS', 'UPDATE_WATCH_FOLDER', 'UPDATE_WORK_ITEM', 'VIEW_VOTERS', 'VIEW_WATCHERS'],
     TRUE);

-- Project Admin: project settings, issues, comments, work items, articles.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('PROJECT_ADMIN', 'Project Admin',
     'Administrative access to project settings, issues, comments, work items, and articles.',
     ARRAY['APPLY_COMMANDS_SILENTLY', 'CREATE_ARTICLE', 'CREATE_ARTICLE_COMMENT', 'CREATE_ATTACHMENT_ISSUE', 'CREATE_COMMENT', 'CREATE_ISSUE', 'CREATE_NOT_OWN_WORK_ITEM', 'CREATE_WATCH_FOLDER', 'CREATE_WORK_ITEM', 'DELETE_ARTICLE', 'DELETE_ARTICLE_COMMENT', 'DELETE_ATTACHMENT_ISSUE', 'DELETE_COMMENT', 'DELETE_ISSUE', 'DELETE_NOT_OWN_COMMENT', 'DELETE_WATCH_FOLDER', 'LINK_ISSUE', 'PRIVATE_READ_ISSUE', 'PRIVATE_UPDATE_ISSUE', 'READ_APP_CONTENT', 'READ_ARTICLE', 'READ_ARTICLE_COMMENT', 'READ_COMMENT', 'READ_ISSUE', 'READ_PROJECT', 'READ_PROJECT_BASIC', 'READ_WORK_ITEM', 'SHARE_WATCH_FOLDER', 'UPDATE_APP_CONTENT', 'UPDATE_ARTICLE', 'UPDATE_ARTICLE_COMMENT', 'UPDATE_ATTACHMENT_ISSUE', 'UPDATE_COMMENT', 'UPDATE_ISSUE', 'UPDATE_NOT_OWN_COMMENT', 'UPDATE_NOT_OWN_WORK_ITEM', 'UPDATE_PROJECT', 'UPDATE_WATCHERS', 'UPDATE_WATCH_FOLDER', 'UPDATE_WORK_ITEM', 'VIEW_VOTERS', 'VIEW_WATCHERS'],
     FALSE);

-- Contributor: daily work on issues, articles, comments, and work items.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('CONTRIBUTOR', 'Contributor',
     'Standard role for project team members to create and edit issues, comments, work items, and articles.',
     ARRAY['CREATE_ARTICLE', 'CREATE_ARTICLE_COMMENT', 'CREATE_ATTACHMENT_ISSUE', 'CREATE_COMMENT', 'CREATE_ISSUE', 'CREATE_WATCH_FOLDER', 'CREATE_WORK_ITEM', 'DELETE_ATTACHMENT_ISSUE', 'DELETE_COMMENT', 'DELETE_ISSUE', 'DELETE_WATCH_FOLDER', 'LINK_ISSUE', 'PRIVATE_READ_ISSUE', 'PRIVATE_UPDATE_ISSUE', 'READ_ARTICLE', 'READ_ARTICLE_COMMENT', 'READ_COMMENT', 'READ_ISSUE', 'READ_PROJECT_BASIC', 'READ_WORK_ITEM', 'UPDATE_ATTACHMENT_ISSUE', 'UPDATE_COMMENT', 'UPDATE_ISSUE', 'UPDATE_WATCHERS', 'UPDATE_WATCH_FOLDER', 'UPDATE_WORK_ITEM', 'VIEW_VOTERS', 'VIEW_WATCHERS'],
     FALSE);

-- Observer: basic user profile access only.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('OBSERVER', 'Observer',
     'Basic access to view user profiles and update own profile.',
     ARRAY['READ_USER', 'READ_USER_BASIC', 'UPDATE_PROFILE'],
     FALSE);

-- User Manager: creating users globally.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('USER_MANAGER', 'User Manager',
     'Ability to create new user accounts in YouTrack.',
     ARRAY['CREATE_USER'],
     FALSE);

-- Project Creator: creating projects globally.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('PROJECT_CREATOR', 'Project Creator',
     'Ability to create new projects in YouTrack.',
     ARRAY['CREATE_PROJECT'],
     FALSE);

-- Developer: legacy role kept for upgraded installations; mirrors Contributor.
INSERT INTO roles (id, name, description, permissions, immutable) VALUES
    ('DEVELOPER', 'Developer',
     'Legacy role for project team members in upgraded installations.',
     ARRAY['CREATE_ARTICLE', 'CREATE_ARTICLE_COMMENT', 'CREATE_ATTACHMENT_ISSUE', 'CREATE_COMMENT', 'CREATE_ISSUE', 'CREATE_WATCH_FOLDER', 'CREATE_WORK_ITEM', 'DELETE_ATTACHMENT_ISSUE', 'DELETE_COMMENT', 'DELETE_ISSUE', 'DELETE_WATCH_FOLDER', 'LINK_ISSUE', 'PRIVATE_READ_ISSUE', 'PRIVATE_UPDATE_ISSUE', 'READ_ARTICLE', 'READ_ARTICLE_COMMENT', 'READ_COMMENT', 'READ_ISSUE', 'READ_PROJECT_BASIC', 'READ_WORK_ITEM', 'UPDATE_ATTACHMENT_ISSUE', 'UPDATE_COMMENT', 'UPDATE_ISSUE', 'UPDATE_WATCHERS', 'UPDATE_WATCH_FOLDER', 'UPDATE_WORK_ITEM', 'VIEW_VOTERS', 'VIEW_WATCHERS'],
     FALSE);