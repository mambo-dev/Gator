-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, feed_id, user_id, created_at, updated_at) 
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    ) 
    RETURNING *
)
SELECT 
    inserted_feed_follow.*, 
    users.name AS user_name, 
    feeds.name AS feed_name
FROM inserted_feed_follow
INNER JOIN users on inserted_feed_follow.user_id = users.id
INNER JOIN feeds on inserted_feed_follow.feed_id = feeds.id;


-- name: GetFeedFollowsForUser :many

SELECT 
    feeds.name AS feed_name,
    users.name AS user_name
FROM feed_follows
INNER JOIN users on feed_follows.user_id = $1
INNER JOIN feeds on feed_follows.feed_id = feeds.id;


-- name: DeleteFeedFollowForUser :exec

DELETE FROM feed_follows
USING feeds
WHERE feed_follows.user_id = $1 
AND feeds.url = $2;
