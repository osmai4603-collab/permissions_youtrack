DROP TRIGGER IF EXISTS triggerAfterInsertUser ON users;
DROP FUNCTION IF EXISTS createGroupMembersAfterInsertUser();
DELETE FROM users WHERE login = 'admin';
