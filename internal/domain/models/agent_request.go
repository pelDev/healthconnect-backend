package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AgentRequestType string

const (
	AgentRequestTypeReferToDoc AgentRequestType = "refer_to_doc"
	AgentRequestTypeEmergency  AgentRequestType = "emergency"
)

type AgentRequest struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	RequestType AgentRequestType

	CreatedAt time.Time

	AcceptedAt *time.Time
	AcceptedBy *uuid.UUID

	Metadata json.RawMessage
}

func (r *AgentRequest) IsPending() bool {
	return r.AcceptedAt == nil
}
