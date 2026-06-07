-- +goose Up
-- +goose StatementBegin


CREATE TYPE agent_request_type_enum AS ENUM (
    'refer_to_doc',
    'emergency'
);

CREATE TABLE IF NOT EXISTS agent_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    request_type agent_request_type_enum NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE,
    accepted_by UUID,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    CONSTRAINT accepted_at_check CHECK (accepted_at IS NULL OR accepted_by IS NOT NULL),
    CONSTRAINT accepted_by_check CHECK (accepted_by IS NULL OR accepted_at IS NOT NULL)
);

CREATE INDEX idx_agent_requests_session_id ON agent_requests(session_id);
CREATE INDEX idx_agent_requests_created_at ON agent_requests(created_at);
CREATE INDEX idx_agent_requests_request_type ON agent_requests(request_type);
CREATE INDEX idx_agent_requests_accepted_at ON agent_requests(accepted_at) WHERE accepted_at IS NOT NULL;
CREATE INDEX idx_agent_requests_pending ON agent_requests(session_id) WHERE accepted_at IS NULL;
CREATE INDEX idx_agent_requests_session_pending ON agent_requests(session_id, request_type, accepted_at) WHERE accepted_at IS NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_agent_requests_session_id;
DROP INDEX IF EXISTS idx_agent_requests_created_at;
DROP INDEX IF EXISTS idx_agent_requests_request_type;
DROP INDEX IF EXISTS idx_agent_requests_accepted_at;
DROP INDEX IF EXISTS idx_agent_requests_pending;
DROP INDEX IF EXISTS idx_agent_requests_session_pending;
DROP TABLE IF EXISTS agent_requests;
DROP TYPE IF EXISTS agent_request_type_enum;

-- +goose StatementEnd
