package ports

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/events"
)

type Subscriber <-chan events.DomainEvent

type EventPublisher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}

type EventSubscriber interface {
	Subscribe(
		ctx context.Context,
		eventName events.DomainEventType,
		buffer int,
	) (Subscriber, error)
}

type EventBus interface {
	EventPublisher
	EventSubscriber
}
