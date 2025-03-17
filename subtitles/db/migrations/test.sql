WITH
    exists AS (SELECT FROM foo WHERE id='bar'),
    valid AS (SELECT NOW() < (NOW()) AS valid)
SELECT
    'not found' AS error, NULL as id, NULL as attr FROM foo WHERE NOT EXISTS (SELECT FROM exists)
UNION ALL (
    SELECT 'invalid' AS error, NULL as id, NULL as attr
    FROM valid WHERE NOT valid
) UNION ALL (
    SELECT 'ok' AS error, id, attr FROM foo WHERE id = 'bar' AND NOW() < NOW()
);