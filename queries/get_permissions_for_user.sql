SELECT rl.permissions FROM roles rl WHERE rl.id IN (
    SELECT ar.role_id FROM assigned_roles ar WHERE (
        CASE
            WHEN ar.user_id IS NOT NULL THEN ar.user_id = $1
            ELSE EXISTS (
                SELECT 1 FROM group_members gm WHERE gm.group_id = ar.group_id AND gm.user_id = $1
            )
        END
    )
);