--DECLARE
  --  teamID VARCHAR(64);

--SELECT team_id INTO teamID FROM projects WHERE project_id = $1;

SELECT * FROM roles WHERE id IN (
    SELECT role_id FROM assigned_roles WHERE scope_type = 'PROJECT' AND project_id = $1 AND (
        user_id = $2 OR
        group_id = $1 OR
        group_id IN (
            SELECT group_id FROM group_members WHERE user_id = $1
            SELECT member_group_id FROM group_members WHERE member_group_id IS NOT NULL AND group_id = $1
        )
    )
);