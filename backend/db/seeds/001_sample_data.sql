-- HyperMotion League - Seed Data
-- Example data for development and testing

BEGIN;

-- =============================================================================
-- USERS (password hash is bcrypt of "password123")
-- =============================================================================
INSERT INTO users (username, email, password_hash, display_name, avatar_url) VALUES
    ('javier',   'javier@example.com',   '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Javier',   NULL),
    ('carlos',   'carlos@example.com',   '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Carlos',   NULL),
    ('maria',    'maria@example.com',    '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Maria',    NULL),
    ('pedro',    'pedro@example.com',    '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Pedro',    NULL),
    ('laura',    'laura@example.com',    '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Laura',    NULL),
    ('admin',    'admin@example.com',    '$2a$10$xJwL5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5TZ5TZ5TZuK5v5Jz5TZ5T', 'Admin',    NULL);

-- =============================================================================
-- LEAGUES
-- =============================================================================
INSERT INTO leagues (name, invite_code, max_members, budget_per_user, market_close_time, created_by) VALUES
    ('La Liga Fantasy',       'LALIGA2026',  8,  100000000, '18:00:00', 1),
    ('Premier League Heroes', 'PLHEROES',   10, 150000000, '19:00:00', 2),
    ('Champions Elite',       'UCL2026',    12, 200000000, '20:00:00', 1);

-- =============================================================================
-- LEAGUE MEMBERS
-- =============================================================================
INSERT INTO league_members (league_id, user_id, role, budget) VALUES
    -- La Liga Fantasy
    (1, 1, 'owner',    100000000),
    (1, 2, 'member',   100000000),
    (1, 3, 'member',   100000000),
    (1, 4, 'member',   100000000),
    -- Premier League Heroes
    (2, 2, 'owner',    150000000),
    (2, 1, 'member',   150000000),
    (2, 5, 'member',   150000000),
    -- Champions Elite
    (3, 1, 'owner',    200000000),
    (3, 3, 'admin',    200000000),
    (3, 4, 'member',   200000000),
    (3, 5, 'member',   200000000),
    (3, 6, 'member',   200000000);

-- =============================================================================
-- PLAYERS (real-world inspired fantasy players)
-- =============================================================================
INSERT INTO players (first_name, last_name, position, team_name, market_value) VALUES
    -- Goalkeepers
    ('Thibaut',      'Courtois',      'GK',  'Real Madrid',       15000000),
    ('Marc-Andre',   'ter Stegen',    'GK',  'Barcelona',         12000000),
    ('Alisson',      'Becker',        'GK',  'Liverpool',         14000000),
    ('Ederson',      'Moraes',        'GK',  'Manchester City',   13000000),

    -- Defenders
    ('Jules',        'Kounde',        'DEF', 'Barcelona',         18000000),
    ('Eder',         'Militao',       'DEF', 'Real Madrid',       20000000),
    ('Virgil',       'van Dijk',      'DEF', 'Liverpool',         22000000),
    ('Ruben',        'Dias',          'DEF', 'Manchester City',   21000000),
    ('William',      'Saliba',        'DEF', 'Arsenal',           19000000),
    ('Achraf',       'Hakimi',        'DEF', 'PSG',               17000000),

    -- Midfielders
    ('Jude',         'Bellingham',    'MID', 'Real Madrid',       35000000),
    ('Pedri',        'Gonzalez',      'MID', 'Barcelona',         30000000),
    ('Kevin',        'De Bruyne',     'MID', 'Manchester City',   28000000),
    ('Martin',       'Odegaard',      'MID', 'Arsenal',           27000000),
    ('Bruno',        'Fernandes',     'MID', 'Manchester United', 25000000),
    ('Luka',         'Modric',        'MID', 'Real Madrid',       15000000),
    ('Gavi',         'Paez',          'MID', 'Barcelona',         22000000),

    -- Forwards
    ('Kylian',       'Mbappe',        'FWD', 'Real Madrid',       45000000),
    ('Erling',       'Haaland',       'FWD', 'Manchester City',   42000000),
    ('Robert',       'Lewandowski',   'FWD', 'Barcelona',         20000000),
    ('Mohamed',      'Salah',         'FWD', 'Liverpool',         28000000),
    ('Vinicius',     'Junior',        'FWD', 'Real Madrid',       40000000),
    ('Bukayo',       'Saka',          'FWD', 'Arsenal',           30000000),
    ('Lamine',       'Yamal',         'FWD', 'Barcelona',         25000000);

-- =============================================================================
-- MATCHDAYS (for La Liga Fantasy league)
-- =============================================================================
INSERT INTO matchdays (league_id, number, name, start_date, end_date, is_current) VALUES
    (1, 1,  'Matchday 1',  '2026-01-10 00:00:00+00', '2026-01-12 23:59:59+00', FALSE),
    (1, 2,  'Matchday 2',  '2026-01-17 00:00:00+00', '2026-01-19 23:59:59+00', FALSE),
    (1, 3,  'Matchday 3',  '2026-01-24 00:00:00+00', '2026-01-26 23:59:59+00', TRUE),
    (1, 4,  'Matchday 4',  '2026-01-31 00:00:00+00', '2026-02-02 23:59:59+00', FALSE),

    -- Premier League Heroes
    (2, 1,  'Matchday 1',  '2026-01-10 00:00:00+00', '2026-01-12 23:59:59+00', FALSE),
    (2, 2,  'Matchday 2',  '2026-01-17 00:00:00+00', '2026-01-19 23:59:59+00', TRUE);

-- =============================================================================
-- PLAYER POINTS (example data for matchdays 1-3 in La Liga Fantasy)
-- =============================================================================
INSERT INTO player_points (player_id, matchday_id, points, goals, assists, minutes_played, yellow_cards, red_cards, clean_sheet) VALUES
    -- Matchday 1 (id=1)
    (11, 1, 12, 1, 1, 90, 0, 0, FALSE), -- Bellingham
    (18, 1, 15, 2, 0, 90, 0, 0, FALSE), -- Mbappe
    (22, 1,  8, 1, 0, 85, 1, 0, FALSE), -- Vinicius
    (1,  1,  6, 0, 0, 90, 0, 0, TRUE),  -- Courtois (clean sheet)
    (5,  1,  5, 0, 0, 90, 0, 0, TRUE),  -- Kounde (clean sheet)

    -- Matchday 2 (id=2)
    (11, 2, 10, 1, 0, 90, 0, 0, FALSE), -- Bellingham
    (22, 2, 18, 2, 1, 90, 0, 0, FALSE), -- Vinicius
    (18, 2,  5, 0, 1, 90, 0, 0, FALSE), -- Mbappe
    (12, 2,  7, 0, 2, 90, 0, 0, FALSE), -- Pedri

    -- Matchday 3 (id=3) - current
    (18, 3, 20, 3, 0, 90, 0, 0, FALSE), -- Mbappe
    (11, 3,  8, 1, 0, 90, 0, 0, FALSE), -- Bellingham
    (22, 3,  6, 1, 0, 75, 1, 0, FALSE), -- Vinicius
    (1,  3,  8, 0, 0, 90, 0, 0, TRUE),  -- Courtois (clean sheet);

-- =============================================================================
-- TEAM PLAYERS (ownership in La Liga Fantasy)
-- =============================================================================
INSERT INTO team_players (league_id, user_id, player_id, purchase_price) VALUES
    -- Javier's squad (La Liga Fantasy)
    (1, 1, 11, 30000000),  -- Bellingham
    (1, 1, 18, 40000000),  -- Mbappe
    (1, 1,  1, 12000000),  -- Courtois
    (1, 1,  5, 15000000),  -- Kounde
    (1, 1, 12, 25000000),  -- Pedri

    -- Carlos's squad (La Liga Fantasy)
    (1, 2, 22, 35000000),  -- Vinicius
    (1, 2,  6, 18000000),  -- Militao
    (1, 2, 16, 12000000),  -- Modric
    (1, 2, 20, 15000000),  -- Lewandowski

    -- Maria's squad (La Liga Fantasy)
    (1, 3, 24, 20000000),  -- Yamal
    (1, 3, 17, 18000000),  -- Gavi
    (1, 3,  2, 10000000);  -- ter Stegen

-- =============================================================================
-- LINEUPS (Javier's lineup for matchday 3 in La Liga Fantasy)
-- =============================================================================
INSERT INTO lineups (league_id, user_id, matchday_id, total_points) VALUES
    (1, 1, 1, 46),  -- Javier MD1
    (1, 1, 2, 40),  -- Javier MD2
    (1, 1, 3, 36),  -- Javier MD3 (current)
    (1, 2, 1, 30),  -- Carlos MD1
    (1, 2, 2, 25);  -- Carlos MD2

-- =============================================================================
-- LINEUP PLAYERS (Javier's MD3 lineup)
-- =============================================================================
INSERT INTO lineup_players (lineup_id, player_id, position, is_starter, points) VALUES
    -- Javier's lineup ID 3 (MD3)
    (3,  1, 'GK',  TRUE,  8),  -- Courtois
    (3,  5, 'DEF', TRUE,  0),  -- Kounde
    (3, 11, 'MID', TRUE,  8),  -- Bellingham
    (3, 12, 'MID', TRUE,  0),  -- Pedri
    (3, 18, 'FWD', TRUE, 20),  -- Mbappe

    -- Javier's lineup ID 1 (MD1)
    (1,  1, 'GK',  TRUE,  6),  -- Courtois
    (1,  5, 'DEF', TRUE,  5),  -- Kounde
    (1, 11, 'MID', TRUE, 12),  -- Bellingham
    (1, 12, 'MID', TRUE,  0),  -- Pedri
    (1, 18, 'FWD', TRUE, 15);  -- Mbappe

-- =============================================================================
-- MARKET LISTINGS (active listings in La Liga Fantasy)
-- =============================================================================
INSERT INTO market_listings (league_id, player_id, base_price, seller_id, status, listed_at, expires_at) VALUES
    (1,  3, 10000000, NULL, 'active', '2026-03-15 10:00:00+00', '2026-03-19 18:00:00+00'), -- Alisson (global pool)
    (1,  7, 18000000, NULL, 'active', '2026-03-15 10:00:00+00', '2026-03-19 18:00:00+00'), -- van Dijk
    (1, 13, 25000000, 2,    'active', '2026-03-16 14:00:00+00', '2026-03-20 18:00:00+00'), -- De Bruyne (Carlos selling)
    (1, 19, 38000000, NULL, 'active', '2026-03-17 08:00:00+00', '2026-03-21 18:00:00+00'), -- Haaland (global pool)
    (1, 23, 25000000, NULL, 'active', '2026-03-18 12:00:00+00', '2026-03-22 18:00:00+00'); -- Saka (global pool)

-- =============================================================================
-- BIDS (active bids on market listings)
-- =============================================================================
INSERT INTO bids (listing_id, user_id, amount, status) VALUES
    -- Bids on Alisson (listing 1)
    (1, 1, 11000000, 'active'),
    (1, 3, 12000000, 'active'),
    (1, 4, 10500000, 'active'),

    -- Bids on De Bruyne (listing 3)
    (3, 1, 26000000, 'active'),
    (3, 3, 28000000, 'active'),

    -- Bids on Haaland (listing 4)
    (4, 1, 40000000, 'active'),
    (4, 2, 39000000, 'active'),
    (4, 4, 42000000, 'active');

COMMIT;
