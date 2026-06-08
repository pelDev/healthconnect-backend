package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/events"
)

type AgentRequestType string

const (
	AgentRequestTypeReferToDoc AgentRequestType = "refer_to_doc"
	AgentRequestTypeEmergency  AgentRequestType = "emergency"
)

type AgentRequest struct {
	ID          uuid.UUID        `json:"id"`
	SessionID   uuid.UUID        `json:"session_id"`
	RequestType AgentRequestType `json:"request_type"`

	CreatedAt time.Time `json:"created_at"`

	AcceptedAt *time.Time      `json:"accepted_at"`
	AcceptedBy *uuid.UUID      `json:"accepted_by"`
	Metadata   json.RawMessage `json:"metadata"`

	domainEvents []events.DomainEvent
}

func NewAgentRequest(
	sessionID uuid.UUID,
	requestType AgentRequestType,
	metadata json.RawMessage,
) *AgentRequest {
	now := time.Now().UTC()

	request := &AgentRequest{
		ID:          uuid.New(),
		SessionID:   sessionID,
		RequestType: requestType,
		CreatedAt:   now,
		Metadata:    metadata,
		AcceptedAt:  nil,
		AcceptedBy:  nil,
	}

	request.addEvent(events.AgentReferDocRequestCreatedEvent{
		RequestID: request.ID,
		SessionID: request.SessionID,
		Timestamp: now,
	})

	return request
}

func (r *AgentRequest) IsPending() bool {
	return r.AcceptedAt == nil
}

func (r *AgentRequest) addEvent(event events.DomainEvent) {
	r.domainEvents = append(r.domainEvents, event)
}

func (r *AgentRequest) GetEvents() []events.DomainEvent {
	events := r.domainEvents
	r.domainEvents = nil
	return events
}
