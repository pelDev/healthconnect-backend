package inmem

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	eventhandlers "github.com/pelDev/health-connect/internal/application/event_handlers"
	"github.com/pelDev/health-connect/internal/domain/events"
	"github.com/pelDev/health-connect/internal/domain/ports"
)

type EventConsumer struct {
	eventBus   ports.EventSubscriber
	dispatcher *eventhandlers.EventDispatcher
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	workers    map[events.DomainEventType]*worker
}

type worker struct {
	subscriber ports.Subscriber
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewEventConsumer(eventBus ports.EventSubscriber, dispatcher *eventhandlers.EventDispatcher) *EventConsumer {
	return &EventConsumer{
		eventBus:   eventBus,
		dispatcher: dispatcher,
		workers:    make(map[events.DomainEventType]*worker),
	}
}

func (c *EventConsumer) Start(ctx context.Context) error {
	// Register handlers for each event type
	for eventName := range c.dispatcher.Handlers() {
		if err := c.subscribeToEvent(ctx, eventName); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventName, err)
		}
	}

	return nil
}

func (c *EventConsumer) subscribeToEvent(parentCtx context.Context, eventName events.DomainEventType) error {
	// Subscribe to the event
	subscriber, err := c.eventBus.Subscribe(parentCtx, eventName, 100) // buffer size 100
	if err != nil {
		return err
	}

	// Create worker context
	workerCtx, cancel := context.WithCancel(parentCtx)

	worker := &worker{
		subscriber: subscriber,
		cancel:     cancel,
	}

	// Start worker goroutine
	worker.wg.Add(1)
	go c.consumeEvents(workerCtx, worker, eventName)

	c.workers[eventName] = worker

	return nil
}

func (c *EventConsumer) consumeEvents(ctx context.Context, worker *worker, eventName events.DomainEventType) {
	defer worker.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer stopping for event %s: %v", eventName, ctx.Err())
			return

		case event, ok := <-worker.subscriber:
			if !ok {
				log.Printf("Subscriber channel closed for event %s", eventName)
				return
			}

			// Process event with timeout
			c.processEventWithTimeout(ctx, event, eventName)
		}
	}
}

func (c *EventConsumer) processEventWithTimeout(ctx context.Context, event events.DomainEvent, eventName events.DomainEventType) {
	// Set timeout for individual event processing
	eventCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use a channel to handle potential blocking
	done := make(chan error, 1)

	go func() {
		done <- c.dispatcher.Dispatch(eventCtx, event)
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("Error processing event %s: %v", event.GetEventType(), err)
			// Optionally: send to dead letter queue, retry logic, etc.
		} else {
			log.Printf("Successfully processed event: %s", event.GetEventType())
		}

	case <-eventCtx.Done():
		log.Printf("Timeout processing event %s: %v", event.GetEventType(), eventCtx.Err())
		// Optionally: handle timeout (DLQ, retry, etc.)
	}
}

func (c *EventConsumer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down event consumer...")

	// Cancel all workers
	for eventName, worker := range c.workers {
		log.Printf("Stopping worker for event: %s", eventName)
		worker.cancel()
	}

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		for _, worker := range c.workers {
			worker.wg.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stopped gracefully")
		return nil
	case <-ctx.Done():
		log.Println("Shutdown timeout reached, forcing exit")
		return ctx.Err()
	}
}

// Alternative: Using worker pool for better performance
type PooledEventConsumer struct {
	eventBus   ports.EventSubscriber
	dispatcher *eventhandlers.EventDispatcher
	wg         sync.WaitGroup
	workerPool chan struct{}
}

func NewPooledEventConsumer(eventBus ports.EventSubscriber, dispatcher *eventhandlers.EventDispatcher, maxWorkers int) *PooledEventConsumer {
	return &PooledEventConsumer{
		eventBus:   eventBus,
		dispatcher: dispatcher,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

func (c *PooledEventConsumer) consumeEvents(ctx context.Context, subscriber ports.Subscriber, eventName events.DomainEventType) {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-subscriber:
			if !ok {
				return
			}

			// Acquire worker from pool
			select {
			case c.workerPool <- struct{}{}:
				// Process in goroutine
				go func(e events.DomainEvent) {
					defer func() { <-c.workerPool }()

					processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()

					if err := c.dispatcher.Dispatch(processCtx, e); err != nil {
						log.Printf("Error processing event: %v", err)
					}
				}(event)
			case <-ctx.Done():
				return
			}
		}
	}
}
