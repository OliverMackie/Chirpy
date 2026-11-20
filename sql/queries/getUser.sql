-- name: GetUser :one
SELECT *
FROM users
WHERE users.email = $1;