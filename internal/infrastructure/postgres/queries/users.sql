-- name: UpsertUser :exec
INSERT INTO users (
    id, first_name, last_name, email, password, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserByID :one
SELECT * from users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * from users WHERE email = $1 AND deleted_at IS NULL;

-- name: DeleteUserByID :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;