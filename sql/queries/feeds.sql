-- name: ResetFeeds :exec
DELETE FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds(name, url, user_id)
VALUES(
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE url = $1 LIMIT 1;

-- name: GetFeeds :many
SELECT * FROM feeds;