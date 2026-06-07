-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    vid UUID NOT NULL,
    reference VARCHAR(255),
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_sessions_vid ON sessions(vid);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_ended_at ON sessions(ended_at) WHERE ended_at IS NULL;
CREATE INDEX idx_sessions_vid_active ON sessions(vid, ended_at) WHERE ended_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_sessions_vid;
DROP INDEX IF EXISTS idx_sessions_created_at;
DROP INDEX IF EXISTS idx_sessions_ended_at;
DROP INDEX IF EXISTS idx_sessions_vid_active;
DROP TABLE IF EXISTS sessions;

-- +goose StatementEnd