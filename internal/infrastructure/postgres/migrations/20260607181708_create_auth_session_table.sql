-- +goose Up
CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    is_revoked BOOLEAN,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ -- Soft delete
);

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;

