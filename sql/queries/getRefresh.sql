-- name: GetRefresh :one
SELECT *
FROM refresh_tokens
WHERE refresh_tokens.token = $1;