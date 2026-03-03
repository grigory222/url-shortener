-- name: GetURLByShortCode :one
SELECT id, short_code, original_url, created_at
FROM urls
WHERE short_code = $1
LIMIT 1;

-- name: GetURLByOriginalURL :one
SELECT id, short_code, original_url, created_at
FROM urls
WHERE original_url = $1
LIMIT 1;

-- name: InsertURL :one
INSERT INTO urls (short_code, original_url)
VALUES ($1, $2)
RETURNING id, short_code, original_url, created_at;
