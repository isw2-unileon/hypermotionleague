-- =============================================================================
-- CHECKS.sql — verificación READ-ONLY del estado de migraciones y esquema
-- =============================================================================
--
-- Para correr A MANO en el editor SQL de Supabase, bloque por bloque, ANTES y
-- DESPUÉS de `make migrate`. Comprueba:
--   (1) qué hay registrado en schema_migrations (tracking del make migrate casero),
--   (2) si el esquema real tiene lo que las migraciones del Sprint 2 deberían
--       haber añadido (tabla teams, campos API-Football en players, columnas de
--       stats en player_points, constraints únicas).
--
-- ⚠️  SOLO LECTURA. Todo son SELECT / consultas a catálogos. No hay DDL ni
--     INSERT/UPDATE/DELETE. No modifica nada. (Las sentencias para *arreglar*
--     el tracking están al final, COMENTADAS, como referencia — no se ejecutan
--     salvo que tú las descomentes a conciencia.)
--
-- Contexto del renumerado (lo que esperamos tras el fix de migraciones):
--   001_initial_schema
--   002_add_auth_provider
--   003_fix_lineup_position_constraint
--   004_add_teams_and_apifootball_fields     (antes ya aplicada a mano)
--   005_extend_player_points_stats           (antes se llamaba 004_*)
--   006_unique_bid_per_listing_user          (antes se llamaba 005_*)
-- =============================================================================


-- =============================================================================
-- BLOQUE 1 — ¿Existe la tabla de tracking schema_migrations?
-- -----------------------------------------------------------------------------
-- Devuelve NULL si NO existe. Si es NULL, el make migrate nunca se ha corrido
-- contra esta BD: arrancaría desde 001 (y 001/003 NO son idempotentes →
-- petarían sobre un esquema ya montado). En ese caso, ve al BLOQUE 6 antes de
-- correr make migrate.
-- =============================================================================
SELECT to_regclass('public.schema_migrations') AS schema_migrations_table;


-- =============================================================================
-- BLOQUE 2 — Contenido actual de schema_migrations
-- -----------------------------------------------------------------------------
-- Lista lo que el tracking cree aplicado. (Si el BLOQUE 1 dio NULL, sáltate
-- este, fallaría porque la tabla no existe.)
-- =============================================================================
SELECT version, applied_at
FROM schema_migrations
ORDER BY version;


-- =============================================================================
-- BLOQUE 3 — ¿Qué versiones ESPERADAS están registradas y cuáles faltan?
-- -----------------------------------------------------------------------------
-- registered = TRUE  → make migrate la saltará.
-- registered = FALSE → make migrate intentará aplicarla. Con las migraciones ya
--                      idempotentes (004 add_teams, 005 extend, 006 bids) eso es
--                      seguro: será no-op si el objeto ya existe en el esquema.
-- Lo normal tras el rename: 004/005/006 saldrán como FALSE (nombres "nuevos")
-- aunque su efecto ya esté en la BD — es esperado y seguro.
-- =============================================================================
WITH expected(version) AS (
    VALUES
        ('001_initial_schema'),
        ('002_add_auth_provider'),
        ('003_fix_lineup_position_constraint'),
        ('004_add_teams_and_apifootball_fields'),
        ('005_extend_player_points_stats'),
        ('006_unique_bid_per_listing_user')
)
SELECT e.version,
       (sm.version IS NOT NULL) AS registered
FROM expected e
LEFT JOIN schema_migrations sm ON sm.version = e.version
ORDER BY e.version;


-- =============================================================================
-- BLOQUE 4 — Filas de tracking OBSOLETAS (nombres pre-rename)
-- -----------------------------------------------------------------------------
-- Si aparecen filas aquí, son los nombres ANTIGUOS (de antes del renumerado).
-- No rompen nada, pero son ruido. Puedes borrarlas con la sentencia comentada
-- del BLOQUE 7.
--   004_extend_player_points_stats   → ahora es 005_extend_player_points_stats
--   005_unique_bid_per_listing_user  → ahora es 006_unique_bid_per_listing_user
-- =============================================================================
SELECT version, applied_at
FROM schema_migrations
WHERE version IN ('004_extend_player_points_stats',
                  '005_unique_bid_per_listing_user');


-- =============================================================================
-- BLOQUE 5 — Estado REAL del esquema (lo que de verdad importa)
-- -----------------------------------------------------------------------------
-- Una sola pasada: cada fila es un check con present = TRUE/FALSE. Lo que esté
-- en FALSE es lo que make migrate (idempotente) creará al correr.
-- Usa to_regclass / EXISTS sobre catálogos, así que es seguro aunque 'teams'
-- todavía no exista.
-- =============================================================================
SELECT '004 teams: tabla teams'                       AS check,
       to_regclass('public.teams') IS NOT NULL        AS present
UNION ALL
SELECT '004 players.external_id',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='players' AND column_name='external_id')
UNION ALL
SELECT '004 players.team_id',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='players' AND column_name='team_id')
UNION ALL
SELECT '004 players.jersey_number',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='players' AND column_name='jersey_number')
UNION ALL
SELECT '004 players.photo_url',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='players' AND column_name='photo_url')
UNION ALL
SELECT '004 players.age',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='players' AND column_name='age')
UNION ALL
SELECT '004 constraint uq_players_external_id',
       EXISTS (SELECT 1 FROM pg_constraint
               WHERE conrelid='public.players'::regclass AND conname='uq_players_external_id')
UNION ALL
-- La 004 DROPEA uq_player_identity: aquí present=TRUE significa que SIGUE ahí
-- (la 004 no llegó a aplicarse). Lo deseado es FALSE.
SELECT '004 uq_player_identity DEBE estar ausente',
       EXISTS (SELECT 1 FROM pg_constraint
               WHERE conrelid='public.players'::regclass AND conname='uq_player_identity')
UNION ALL
SELECT '005 player_points.goals_conceded',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='player_points' AND column_name='goals_conceded')
UNION ALL
SELECT '005 player_points.pens_missed',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='player_points' AND column_name='pens_missed')
UNION ALL
SELECT '005 player_points.pens_saved',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='player_points' AND column_name='pens_saved')
UNION ALL
SELECT '005 player_points.saves',
       EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema='public' AND table_name='player_points' AND column_name='saves')
UNION ALL
SELECT '006 constraint uq_bid_listing_user',
       EXISTS (SELECT 1 FROM pg_constraint
               WHERE conrelid='public.bids'::regclass AND conname='uq_bid_listing_user')
ORDER BY check;


-- =============================================================================
-- BLOQUE 5b — Detalle de columnas y constraints (opcional, para inspección)
-- -----------------------------------------------------------------------------
-- Tipos/defaults exactos de las columnas nuevas. Lo esperado:
--   players.*        → external_id/team_id BIGINT, jersey_number/age INT,
--                      photo_url TEXT (todas NULLABLE)
--   player_points.*  → INT, NOT NULL, default 0
-- =============================================================================
SELECT table_name, column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema='public'
  AND ( (table_name='players'
         AND column_name IN ('external_id','team_id','jersey_number','photo_url','age'))
     OR (table_name='player_points'
         AND column_name IN ('goals_conceded','pens_missed','pens_saved','saves')) )
ORDER BY table_name, column_name;

-- Definición de las constraints únicas relevantes (esperado: UNIQUE (...)).
SELECT c.conname,
       t.relname AS table_name,
       pg_get_constraintdef(c.oid) AS definition
FROM pg_constraint c
JOIN pg_class t ON t.oid = c.conrelid
WHERE c.conname IN ('uq_players_external_id', 'uq_teams_external_id', 'uq_bid_listing_user')
ORDER BY t.relname, c.conname;

-- Conteos de referencia (solo si las tablas existen). Esperado aprox.:
-- teams ≈ 22, players ≈ 696 con external_id.
SELECT (SELECT COUNT(*) FROM teams)                              AS teams_count,
       (SELECT COUNT(*) FROM players)                            AS players_total,
       (SELECT COUNT(*) FROM players WHERE external_id IS NOT NULL) AS players_with_external_id;


-- =============================================================================
-- BLOQUE 6 — (REFERENCIA, COMENTADO) Sembrar tracking si schema_migrations
--            está vacía / no refleja 001-003 ya aplicadas
-- -----------------------------------------------------------------------------
-- Solo si el BLOQUE 1 dio NULL o el BLOQUE 3 muestra 001/002/003 como FALSE
-- pese a que el esquema base ya existe. Evita que make migrate intente
-- re-aplicar 001/003 (que NO son idempotentes). Descomenta a conciencia.
-- =============================================================================
-- CREATE TABLE IF NOT EXISTS schema_migrations (
--     version    TEXT PRIMARY KEY,
--     applied_at TIMESTAMPTZ DEFAULT NOW()
-- );
-- INSERT INTO schema_migrations(version) VALUES
--     ('001_initial_schema'),
--     ('002_add_auth_provider'),
--     ('003_fix_lineup_position_constraint')
-- ON CONFLICT (version) DO NOTHING;


-- =============================================================================
-- BLOQUE 7 — (REFERENCIA, COMENTADO) Limpiar filas de tracking obsoletas
-- -----------------------------------------------------------------------------
-- Solo si el BLOQUE 4 devolvió filas. Borra los nombres pre-rename. Opcional:
-- es cosmético, la idempotencia ya evita cualquier problema funcional.
-- =============================================================================
-- DELETE FROM schema_migrations
-- WHERE version IN ('004_extend_player_points_stats',
--                   '005_unique_bid_per_listing_user');
