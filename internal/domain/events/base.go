package events

import (
	"time"

	"github.com/google/uuid"
)

type DomainEventType string

const (
	DomainEventTypeReferDoctorTriggered DomainEventType = "refer_doctor_triggered"
)

type DomainEvent interface {
	GetEventType() DomainEventType
	GetAggregateID() uuid.UUID
	GetTimestamp() time.Time
}
