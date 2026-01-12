-- Dev seed data migration
-- This migration seeds test users and leaderboard scores for development only
-- All inserts are idempotent using ON CONFLICT clauses

-- Seed test users
-- Password for all users: password123
-- Bcrypt hash: $2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC
INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'alice', 'alice@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000002', 'bob', 'bob@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000003', 'charlie', 'charlie@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000004', 'dave', 'dave@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000005', 'eve', 'eve@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000006', 'frank', 'frank@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000007', 'grace', 'grace@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000008', 'henry', 'henry@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000009', 'ivy', 'ivy@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000010', 'jack', 'jack@example.com', '$2a$10$9DSodToLn2m.h3i1uYQocu//OKlkvjHk3rdtB463cKQ0bSKDupJwC', NOW(), NOW())
ON CONFLICT (username) DO NOTHING;

-- Seed leaderboard scores
-- Scores are varied to test leaderboard ranking
INSERT INTO leaderboard (id, user_id, score, created_at, updated_at)
VALUES 
    ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', 10000, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000002', 8500, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000103', '00000000-0000-0000-0000-000000000003', 7500, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000104', '00000000-0000-0000-0000-000000000004', 6000, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000105', '00000000-0000-0000-0000-000000000005', 5000, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000106', '00000000-0000-0000-0000-000000000006', 4000, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000107', '00000000-0000-0000-0000-000000000007', 3000, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000108', '00000000-0000-0000-0000-000000000008', 2500, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000109', '00000000-0000-0000-0000-000000000009', 1500, NOW(), NOW()),
    ('00000000-0000-0000-0000-000000000110', '00000000-0000-0000-0000-000000000010', 1000, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET 
    score = EXCLUDED.score,
    updated_at = EXCLUDED.updated_at;
