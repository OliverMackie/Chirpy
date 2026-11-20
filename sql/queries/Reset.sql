-- name: Reset :exec
TRUNCATE TABLE users, chirps, refresh_tokens RESTART IDENTITY CASCADE;