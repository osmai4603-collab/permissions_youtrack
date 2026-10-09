
INSERT INTO users (id, login, email, full_name, name, avatar_url, ring_id, type,
                   guest, online, banned, hash_password)
VALUES
    ('test1', 'test1', 'test1@youtrack.local', 'Test One',   'test1',
     '/hub/api/rest/avatar/t1a00000-0000-0000-0000-000000000001?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test2', 'test2', 'test2@youtrack.local', 'Test Two',   'test2',
     '/hub/api/rest/avatar/t2b00000-0000-0000-0000-000000000002?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC'),
    ('test3', 'test3', 'test3@youtrack.local', 'Test Three', 'test3',
     '/hub/api/rest/avatar/t3c00000-0000-0000-0000-000000000003?s=48', '', 'STANDARD_USER',
     FALSE, TRUE, FALSE, '$2a$10$fZRjMe.6mx8/9x7zSqwNeed8j.zpvqWmSjVOpMK18OkqdxqKNnOBC')
ON CONFLICT (id) DO NOTHING;


INSERT INTO projects (id, name, short_name, created_by, leader_id)
VALUES
    ('proj-t1-a', 'Test1 Project Alpha', 'T1A', 'test1', 'test1'),
    ('proj-t1-b', 'Test1 Project Beta',  'T1B', 'test1', 'test1'),
    ('proj-t2-a', 'Test2 Project Alpha', 'T2A', 'test2', 'test2'),
    ('proj-t2-b', 'Test2 Project Beta',  'T2B', 'test2', 'test2')
ON CONFLICT (id) DO NOTHING;

INSERT INTO groups (id, name, group_type, description)
VALUES
    ('grp-t1-a', 'T1A Project Members',  'GROUP', 'Members of T1A'),
    ('grp-t1-b', 'T1B Project Members',  'GROUP', 'Members of T1B'),
    ('grp-t2-a', 'T2A Project Members',  'GROUP', 'Members of T2A'),
    ('grp-t2-b', 'T2B Project Members',  'GROUP', 'Members of T2B')
ON CONFLICT (id) DO NOTHING;

-- Add users to project-member groups.
INSERT INTO group_members (group_id, member_type, user_id, is_team_member)
VALUES
    ('grp-t1-a', 'USER', 'test1', FALSE),
    ('grp-t1-a', 'USER', 'test3', FALSE),
    ('grp-t1-b', 'USER', 'test2', FALSE),
    ('grp-t1-b', 'USER', 'test3', FALSE),
    ('grp-t2-a', 'USER', 'test1', FALSE),
    ('grp-t2-a', 'USER', 'test2', FALSE),
    ('grp-t2-b', 'USER', 'test2', FALSE),
    ('grp-t2-b', 'USER', 'test3', FALSE)
ON CONFLICT DO NOTHING;

-- Add each project-member group to its project's team.
INSERT INTO group_members (group_id, member_type, member_group_id, is_team_member)
SELECT p.team_id, 'GROUP', membership.group_id, TRUE
FROM (VALUES
    ('proj-t1-a', 'grp-t1-a'),
    ('proj-t1-b', 'grp-t1-b'),
    ('proj-t2-a', 'grp-t2-a'),
    ('proj-t2-b', 'grp-t2-b')
) AS membership(project_id, group_id)
JOIN projects p ON p.id = membership.project_id
WHERE p.team_id IS NOT NULL
ON CONFLICT DO NOTHING;
