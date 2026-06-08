package events

import (
	"time"

	"github.com/google/uuid"
)

type DomainEventType string

const (
	DomainEventTypeReferDoctorTriggered     DomainEventType = "refer_doctor_triggered"
	DomainEventTypeReferDoctorTriggeredUser DomainEventType = "refer_doctor_triggered_user"
	DomainEventTypeEmergencyAlertTriggered  DomainEventType = "emergency_alert_triggered"
)

type DomainEvent interface {
	GetEventType() DomainEventType
	GetAggregateID() uuid.UUID
	GetTimestamp() time.Time
}
