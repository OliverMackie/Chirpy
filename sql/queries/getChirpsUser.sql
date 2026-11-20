-- name: GetUserChirps :many
SELECT *
FROM chirps
WHERE chirps.user_id = $1
ORDER BY chirps.created_at ASC;