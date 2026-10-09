SELECT p.id, p.name, p.short_name, p.description, p.team_id, p.created_by, p.leader_id FROM projects p WHERE 
    p.created_by = $1 OR p.leader_id = $1 OR EXISTS (
        SELECT 1 FROM group_members gm WHERE gm.group_id = p.team_id AND gm.is_team_member = TRUE AND (
            gm.user_id = $1 OR EXISTS (
                SELECT 1 FROM group_members gm2 WHERE gm2.group_id = gm.member_group_id AND gm2.user_id = $1
            )
        )
    );
