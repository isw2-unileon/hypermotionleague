UPDATE players p
SET market_value = GREATEST(
        500000,
        -- suelo mínimo: nadie vale menos de 500k
        ROUND(
            (
                -- 1. Valor base por posición (en euros)
                CASE
                    p.position
                    WHEN 'FWD' THEN 8000000
                    WHEN 'MID' THEN 7000000
                    WHEN 'DEF' THEN 5000000
                    WHEN 'GK' THEN 4000000
                    ELSE 4000000
                END -- 2. Modificador por edad (curva con pico 23-28)
                * CASE
                    WHEN p.age IS NULL THEN 0.80
                    WHEN p.age < 21 THEN 0.70
                    WHEN p.age BETWEEN 21 AND 22 THEN 0.85
                    WHEN p.age BETWEEN 23 AND 28 THEN 1.00
                    WHEN p.age BETWEEN 29 AND 31 THEN 0.85
                    WHEN p.age BETWEEN 32 AND 34 THEN 0.65
                    ELSE 0.45
                END -- 3. Modificador por equipo (tier según clasificación actual de Segunda)
                * CASE
                    t.name -- Tier 1 — ascenso directo (1º-2º) ×1.5
                    WHEN 'Racing Santander' THEN 1.5
                    WHEN 'Deportivo La Coruna' THEN 1.5 -- Tier 2 — playoff (3º-6º) ×1.3
                    WHEN 'Almeria' THEN 1.3
                    WHEN 'Malaga' THEN 1.3
                    WHEN 'Las Palmas' THEN 1.3
                    WHEN 'Castellón' THEN 1.3 -- Tier 3 — media-alta (7º-11º) ×1.1
                    WHEN 'Burgos' THEN 1.1
                    WHEN 'Eibar' THEN 1.1
                    WHEN 'Cordoba' THEN 1.1
                    WHEN 'Sporting Gijon' THEN 1.1
                    WHEN 'AD Ceuta FC' THEN 1.1 -- Tier 4 — media-baja (12º-18º) ×0.9
                    WHEN 'Albacete' THEN 0.9
                    WHEN 'FC Andorra' THEN 0.9
                    WHEN 'Granada CF' THEN 0.9
                    WHEN 'Real Sociedad II' THEN 0.9
                    WHEN 'Leganes' THEN 0.9
                    WHEN 'Valladolid' THEN 0.9
                    WHEN 'Cadiz' THEN 0.9 -- Tier 5 — descenso (19º-22º) ×0.75
                    WHEN 'Mirandes' THEN 0.75
                    WHEN 'Huesca' THEN 0.75
                    WHEN 'Cultural Leonesa' THEN 0.75
                    WHEN 'Zaragoza' THEN 0.75
                    ELSE 1.0 -- no debería entrar aquí ningún equipo
                END -- 4. Pizca de azar: ±10%
                * (0.90 + random() * 0.20)
            ) / 100000 -- redondeo a la centena de millar
        ) * 100000
    )
FROM teams t
WHERE p.team_id = t.id;