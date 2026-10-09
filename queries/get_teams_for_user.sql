SELECT gm.group_id FROM group_members gm WHERE gm.is_team_member = TRUE AND (
    CASE
        WHEN gm.user_id IS NOT NULL THEN gm.user_id = $1
        ELSE EXISTS (
            SELECT 1 FROM group_members gm2 WHERE gm2.group_id = gm.member_group_id AND gm2.user_id = $1 AND gm2.is_team_member = FALSE
        )
    END
);