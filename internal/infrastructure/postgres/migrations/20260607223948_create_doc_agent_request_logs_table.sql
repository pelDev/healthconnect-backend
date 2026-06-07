-- +goose Up
-- +goose StatementBegin

-- Create enum type for log actions
CREATE TYPE doc_agent_request_log_action_enum AS ENUM (
    'decline_request',
    'accept_request',
    'send_message',
    'send_prescription',
    'view_request',
    'escalate'
);

-- Create doc agent request logs table
CREATE TABLE IF NOT EXISTS doc_agent_request_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES agent_requests(id) ON DELETE CASCADE,
    doc_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action doc_agent_request_log_action_enum NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_doc_agent_request_logs_request_id ON doc_agent_request_logs(request_id);
CREATE INDEX idx_doc_agent_request_logs_doc_id ON doc_agent_request_logs(doc_id);
CREATE INDEX idx_doc_agent_request_logs_created_at ON doc_agent_request_logs(created_at);
CREATE INDEX idx_doc_agent_request_logs_action ON doc_agent_request_logs(action);
CREATE INDEX idx_doc_agent_request_logs_doc_actions ON doc_agent_request_logs(doc_id, created_at DESC);
CREATE INDEX idx_doc_agent_request_logs_request_actions ON doc_agent_request_logs(request_id, action, created_at);

-- Composite index for checking if a doctor has already accepted a request
CREATE UNIQUE INDEX idx_unique_accept_per_request_doc 
ON doc_agent_request_logs(request_id, doc_id, action) 
WHERE action = 'accept_request';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_doc_agent_request_logs_request_id;
DROP INDEX IF EXISTS idx_doc_agent_request_logs_doc_id;
DROP INDEX IF EXISTS idx_doc_agent_request_logs_created_at;
DROP INDEX IF EXISTS idx_doc_agent_request_logs_action;
DROP INDEX IF EXISTS idx_doc_agent_request_logs_doc_actions;
DROP INDEX IF EXISTS idx_doc_agent_request_logs_request_actions;
DROP INDEX IF EXISTS idx_unique_accept_per_request_doc;
DROP TABLE IF EXISTS doc_agent_request_logs;
DROP TYPE IF EXISTS doc_agent_request_log_action_enum;
-- +goose StatementEnd