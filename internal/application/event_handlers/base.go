package eventhandlers

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/events"
)

type EventHandler interface {
	HandleMessage(ctx context.Context, event events.DomainEvent) error
	SupportedEventNames() []events.DomainEventType
}
