package inmem

import (
	"context"
	"errors"
	"sync"

	"github.com/pelDev/health-connect/internal/domain/events"
	"github.com/pelDev/health-connect/internal/domain/ports"
)

type inMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[events.DomainEventType][]chan events.DomainEvent
	closed      bool
	closeMu     sync.RWMutex
	closeOnce   sync.Once
	closeCh     chan struct{}
}

func NewInMemoryBus() *inMemoryEventBus {
	return &inMemoryEventBus{
		subscribers: make(map[events.DomainEventType][]chan events.DomainEvent),
		closeCh:     make(chan struct{}),
	}
}

func (b *inMemoryEventBus) Subscribe(
	ctx context.Context,
	eventType events.DomainEventType,
	buffer int,
) (ports.Subscriber, error) {

	// Check if bus is already closed
	b.closeMu.RLock()
	if b.closed {
		b.closeMu.RUnlock()
		return nil, errors.New("event bus is closed")
	}
	b.closeMu.RUnlock()

	ch := make(chan events.DomainEvent, buffer)

	b.mu.Lock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	b.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			b.remove(eventType, ch)
			close(ch)
		case <-b.closeCh:
			// Bus is shutting down, just close the channel
			close(ch)
		}
	}()

	return ch, nil
}

func (b *inMemoryEventBus) Publish(
	ctx context.Context,
	event events.DomainEvent,
) error {

	b.closeMu.RLock()
	if b.closed {
		b.closeMu.RUnlock()
		return errors.New("event bus is closed")
	}
	b.closeMu.RUnlock()

	b.mu.RLock()
	subs := b.subscribers[event.GetEventType()]
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		case <-ctx.Done():
			return ctx.Err()
		case <-b.closeCh:
			return errors.New("event bus closed during publish")
		default:
			// drop or log (design choice)
		}
	}

	return nil
}

// Shutdown gracefully shuts down the event bus
// It closes all subscriber channels and prevents new subscriptions/publishes
func (b *inMemoryEventBus) Shutdown(ctx context.Context) error {
	b.closeOnce.Do(func() {
		b.closeMu.Lock()
		b.closed = true
		b.closeMu.Unlock()
		close(b.closeCh)
	})

	// Wait for all subscribers to be cleaned up or context timeout
	done := make(chan struct{})
	go func() {
		b.waitForSubscribers()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// waitForSubscribers waits for all subscriber goroutines to complete
func (b *inMemoryEventBus) waitForSubscribers() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clear all subscribers
	for _, subs := range b.subscribers {
		for _, ch := range subs {
			// Close each channel (they'll be cleaned up by their goroutines)
			// Don't remove here as the remove method might still be called
			close(ch)
		}
	}

	// Clear the map
	b.subscribers = make(map[events.DomainEventType][]chan events.DomainEvent)
}

func (b *inMemoryEventBus) remove(
	eventType events.DomainEventType,
	target chan events.DomainEvent,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, exists := b.subscribers[eventType]
	if !exists {
		return
	}

	for i, ch := range subs {
		if ch == target {
			// Remove without preserving order
			subs[i] = subs[len(subs)-1]
			subs = subs[:len(subs)-1]
			break
		}
	}

	if len(subs) == 0 {
		delete(b.subscribers, eventType)
	} else {
		b.subscribers[eventType] = subs
	}
}
