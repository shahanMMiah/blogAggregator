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

-- name: GetUsers :many
Select name FROM users;

-- name: GetUserFromId :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

