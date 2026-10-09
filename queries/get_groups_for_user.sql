SELECT * FROM groups gp WHERE gp.id IN (
    SELECT gm.group_id FROM group_members gm WHERE gm.is_team_member = FALSE AND gm.user_id = 'test1'
);