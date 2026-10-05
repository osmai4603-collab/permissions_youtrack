
-- ===========================================================================
-- Helpdesk customer group projects
-- ===========================================================================
-- UserGroup.customerGroupProjects: the set of helpdesk projects where a group is
-- configured as a customer group, empty for every other group. This relation is
-- the counterpart to the is_customer_group flag on groups, which answers "is it
-- a customer group" but not "in which helpdesk projects".
-- Read-only in YouTrack and available only through the Workflow API since
-- 2026.2, so like is_customer_group it cannot be filled from a plain REST sync.
CREATE TABLE IF NOT EXISTS group_customer_projects (
  -- The group configured as a customer group in this project.
  group_id   VARCHAR(64) NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  -- The helpdesk project. No foreign key: this project has no projects table yet,
  -- so the reference is left unenforced until one exists.
  project_id VARCHAR(64) NOT NULL,
  -- True when new tickets from group members are shared with the group
  -- automatically. Project specific, so it lives on the relation and not on the
  -- group. Agents can still share individual tickets when this is false.
  auto_share BOOLEAN NOT NULL DEFAULT FALSE,
  -- Application audit timestamp when this record was created.
  created_at TIMESTAMPTZ,
  PRIMARY KEY (group_id, project_id)
);

-- Only customer groups can be attached to a helpdesk project catalog. This is a
-- trigger rather than a CHECK constraint because CHECK cannot contain the
-- subquery that has to read groups.is_customer_group.
CREATE OR REPLACE FUNCTION group_customer_projects_require_customer_group()
RETURNS TRIGGER AS $$
BEGIN
  IF NOT (SELECT is_customer_group FROM groups WHERE id = NEW.group_id) THEN
    RAISE EXCEPTION
      'group % is not a customer group and cannot be added to a helpdesk project',
      NEW.group_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Postgres has no CREATE TRIGGER IF NOT EXISTS, so drop first to keep this
-- migration re-runnable like the rest of the file.
DROP TRIGGER IF EXISTS trg_group_customer_projects_require_customer_group ON group_customer_projects;

CREATE TRIGGER trg_group_customer_projects_require_customer_group
  BEFORE INSERT OR UPDATE ON group_customer_projects
  FOR EACH ROW EXECUTE FUNCTION group_customer_projects_require_customer_group();

CREATE INDEX IF NOT EXISTS idx_group_customer_projects_project_id
  ON group_customer_projects (project_id);
