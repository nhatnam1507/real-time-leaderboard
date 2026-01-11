-- Remove dev seed data
-- This will delete all test users and their leaderboard entries

-- Delete leaderboard entries for seed users
DELETE FROM leaderboard 
WHERE user_id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000006',
    '00000000-0000-0000-0000-000000000007',
    '00000000-0000-0000-0000-000000000008',
    '00000000-0000-0000-0000-000000000009',
    '00000000-0000-0000-0000-000000000010'
);

-- Delete seed users
DELETE FROM users 
WHERE username IN ('alice', 'bob', 'charlie', 'dave', 'eve', 'frank', 'grace', 'henry', 'ivy', 'jack');
