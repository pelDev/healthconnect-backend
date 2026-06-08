-- name: CreateAgentRequest :exec
INSERT INTO agent_requests (
    id, session_id, request_type, created_at, accepted_at, accepted_by, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetAgentRequestByID :one
SELECT * FROM agent_requests WHERE id = $1;