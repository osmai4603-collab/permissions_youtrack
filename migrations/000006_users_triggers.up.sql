CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION createGroupMembersAfterInsertUser()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO group_members (group_id, member_type, user_id)
    VALUES ('ALL_USERS_GROUP', 'USER', NEW.id);
    INSERT INTO group_members (group_id, member_type, user_id)
    VALUES ('REGISTERED_USERS_GROUP', 'USER', NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS triggerAfterInsertUser ON users;

CREATE TRIGGER triggerAfterInsertUser
  AFTER INSERT ON users
  FOR EACH ROW EXECUTE FUNCTION createGroupMembersAfterInsertUser();

INSERT INTO users (
  login, email, full_name, name, type, hash_password, created_at, updated_at
) VALUES (
  'admin', 'admin@example.com', 'System Administrator', 'admin', 'AGENT',
  crypt('admin123', gen_salt('bf', 12)), NOW(), NOW()
)
ON CONFLICT (login) DO NOTHING;
