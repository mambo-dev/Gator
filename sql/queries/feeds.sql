-- name: CreateFeed :one
INSERT INTO feeds (id, name, created_at, updated_at, url, user_id) 
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) 
RETURNING *;

-- name: GetFeeds :many
SELECT 
    feeds.id,
    feeds.name AS feed_name,
    feeds.url,
    users.id,
    users.name AS user_name
FROM 
    feeds
JOIN 
    users ON feeds.user_id = users.id;


-- name: GetFeed :one
SELECT 
    feeds.id,
    feeds.name AS feed_name,
    feeds.url,
    users.id,
    users.name AS user_name
FROM 
    feeds
JOIN 
    users ON feeds.user_id = users.id
WHERE feeds.url = $1;


-- name: MarkFeedFetched :exec
UPDATE feeds 
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;


-- name: GetNextFeedToFetch :one
SELECT id, last_fetched_at , url  FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;