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
