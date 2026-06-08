package events

import (
	"time"

	"github.com/google/uuid"
)

type AgentReferDocRequestCreatedEvent struct {
	RequestID uuid.UUID `json:"request_id"`
	SessionID uuid.UUID `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e AgentReferDocRequestCreatedEvent) GetEventType() DomainEventType {
	return DomainEventTypeReferDoctorTriggered
}
func (e AgentReferDocRequestCreatedEvent) GetAggregateID() uuid.UUID { return e.SessionID }
func (e AgentReferDocRequestCreatedEvent) GetTimestamp() time.Time   { return e.Timestamp }

type AgentEmergencyAlertRequestCreatedEvent struct {
	RequestID uuid.UUID `json:"request_id"`
	SessionID uuid.UUID `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e AgentEmergencyAlertRequestCreatedEvent) GetEventType() DomainEventType {
	return DomainEventTypeEmergencyAlertTriggered
}
func (e AgentEmergencyAlertRequestCreatedEvent) GetAggregateID() uuid.UUID { return e.SessionID }
func (e AgentEmergencyAlertRequestCreatedEvent) GetTimestamp() time.Time   { return e.Timestamp }
