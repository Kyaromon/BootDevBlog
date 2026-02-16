-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT id, username, created_at, updated_at
FROM users
WHERE username = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT id, username, created_at, updated_at
FROM users
ORDER BY username;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeedsWithUser :many
SELECT
    f.id,
    f.created_at,
    f.updated_at,
    f.name,
    f.url,
    u.username AS user_name
FROM feeds f
JOIN users u ON f.user_id = u.id;

-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, created_at, updated_at, user_id, feed_id
)
SELECT
    ff.id,
    ff.created_at,
    ff.updated_at,
    f.name AS feed_name,
    u.username AS user_name
FROM inserted ff
JOIN feeds f ON ff.feed_id = f.id
JOIN users u ON ff.user_id = u.id;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT
    ff.id,
    ff.created_at,
    ff.updated_at,
    f.name AS feed_name,
    u.username AS user_name
FROM feed_follows ff
JOIN feeds f ON ff.feed_id = f.id
JOIN users u ON ff.user_id = u.id
WHERE ff.user_id = $1
ORDER BY f.name;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2, updated_at = $2
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT p.*
FROM posts p
JOIN feed_follows ff ON p.feed_id = ff.feed_id
JOIN users u ON ff.user_id = u.id
WHERE u.id = $1
ORDER BY p.published_at DESC
LIMIT $2;
