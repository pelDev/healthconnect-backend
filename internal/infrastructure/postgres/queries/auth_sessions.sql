-- name: UpsertAuthSession :exec
INSERT INTO auth_sessions (
    id, user_id, created_at, expires_at, is_revoked
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetAuthSessionByID :one
SELECT * from auth_sessions WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAuthSessionByUserID :many
SELECT * from auth_sessions WHERE user_id = $1 AND deleted_at IS NULL;

-- name: DeleteAuthSessionByID :exec
UPDATE auth_sessions SET deleted_at = NOW() WHERE id = $1;