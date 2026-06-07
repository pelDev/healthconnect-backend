-- name: CreateSession :exec
INSERT INTO sessions (
    id, vid, reference, ended_at, created_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: GetSessionByReference :one
SELECT * FROM sessions WHERE reference = $1;

-- name: DeleteSessionByID :exec
DELETE FROM sessions WHERE id = $1; -- ASK: I think this should be a soft delete with a 30 days cleanup