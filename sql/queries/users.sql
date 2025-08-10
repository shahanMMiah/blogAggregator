-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
Select * FROM users
WHERE name = $1 LIMIT 1;

-- name: ResetUsers :exec
DELETE FROM users; 

-- name: ResetFeeds :exec
DELETE FROM feeds;

-- name: GetUsers :many
Select name FROM users;

-- name: CreateFeed :one
INSERT INTO feeds(name, url, user_id)
VALUES(
    $1,
    $2,
    $3
)
RETURNING *;