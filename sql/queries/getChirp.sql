-- name: GetChirp :one
SELECT *
FROM chirps
WHERE chirps.id = $1
ORDER BY chirps.created_at ASC;