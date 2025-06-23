-- name: CreateUser :exec
INSERT INTO users (name) VALUES (?);

-- name: GetUserById :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM users WHERE name = ? LIMIT 1;

-- name: UpdateUser :exec
UPDATE users SET name = ? WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: GetAllUsers :many
SELECT * FROM users;