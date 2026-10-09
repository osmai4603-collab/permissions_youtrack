SELECT * FROM roles WHERE id IN (
    SELECT ar.role_id FROM assigned_roles ar WHERE scope_type = 'PROJECT' AND (
        CASE
            WHEN ar.user_id IS NOT NULL THEN ar.user_id = $2
            ELSE EXISTS (
                SELECT 1 FROM group_members gm WHERE gm.group_id = ar.group_id AND (
                    CASE
                        WHEN gm.user_id IS NOT NULL THEN gm.user_id = $2
                        ELSE EXISTS (
                            SELECT 1 FROM group_members gm2 WHERE gm2.group_id = gm.member_group_id AND gm2.user_id = $2
                        )
                    END
                )
            )
        END
    )
);
