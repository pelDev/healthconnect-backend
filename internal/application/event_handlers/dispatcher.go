package eventhandlers

import (
	"context"
	"fmt"

	"github.com/pelDev/health-connect/internal/domain/events"
)

type EventDispatcher struct {
	handlers map[events.DomainEventType][]EventHandler
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[events.DomainEventType][]EventHandler),
	}
}

func (d *EventDispatcher) Register(handler EventHandler) {
	for _, name := range handler.SupportedEventNames() {
		d.handlers[name] = append(d.handlers[name], handler)
	}
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event events.DomainEvent) error {
	handlers, ok := d.handlers[event.GetEventType()]
	if !ok {
		return fmt.Errorf("no handlers registered for event: %s", event.GetEventType())
	}

	for _, h := range handlers {
		if err := h.HandleMessage(ctx, event); err != nil {
			return fmt.Errorf("handler failed for event %s: %w", event.GetEventType(), err)
		}
	}

	return nil
}

func (d *EventDispatcher) Handlers() map[events.DomainEventType][]EventHandler {
	return d.handlers
}
