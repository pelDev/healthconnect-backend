package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/events"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/interfaces/ws"
)

type doctorNotificationHandler struct {
	hub            *ws.Hub
	sessionStorage repositories.SessionStorage
	requestStorage repositories.AgentRequestStorage
}

func (d *doctorNotificationHandler) HandleMessage(ctx context.Context, event events.DomainEvent) error {
	switch event.GetEventType() {
	case events.DomainEventTypeReferDoctorTriggered:
		return d.handleAgentReferDocRequestCreated(ctx, event)
	case events.DomainEventTypeEmergencyAlertTriggered:
		// return d.handleAgentReferDocRequestCreated(ctx, event)
		panic("unimplemented")
	default:
		return fmt.Errorf("unsupported event type for doctor notification handler: expected %v, got %T", d.SupportedEventNames(), event)
	}
}

func (d *doctorNotificationHandler) handleAgentReferDocRequestCreated(ctx context.Context, event events.DomainEvent) error {
	// Type assert to get the concrete event
	agentRequestEvent, ok := event.(events.AgentReferDocRequestCreatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type for doctor notification handler: expected AgentRequestCreatedEvent, got %T", event)
	}

	log.Printf("Processing doctor notification for event: %s, RequestID: %s, SessionID: %s\n\n",
		event.GetEventType(),
		agentRequestEvent.RequestID,
		agentRequestEvent.SessionID,
	)

	request, err := d.requestStorage.FindByID(ctx, agentRequestEvent.RequestID)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	if request == nil {
		return fmt.Errorf("agent request with ID %s not found", agentRequestEvent.RequestID.String())
	}

	session, err := d.sessionStorage.FindByID(ctx, request.SessionID)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	if session == nil {
		return fmt.Errorf("session with ID %s not found", agentRequestEvent.SessionID.String())
	}

	msgBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v\n", err)
	}

	d.hub.Broadcast <- ws.BroadcastMessage{
		Topic:   string(event.GetEventType()),
		Message: msgBytes,
	}

	userPayload := map[string]string{
		"session_id": session.ID.String(),
	}

	userPayloadBytes, err := json.Marshal(userPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v\n", err)
	}

	d.hub.SendToUser(session.VID, ws.BroadcastMessage{
		Topic:   string(events.DomainEventTypeReferDoctorTriggeredUser),
		Message: userPayloadBytes,
	})

	return nil
}

func (d *doctorNotificationHandler) SupportedEventNames() []events.DomainEventType {
	return []events.DomainEventType{
		events.DomainEventTypeReferDoctorTriggered,
		events.DomainEventTypeEmergencyAlertTriggered,
	}
}

func NewDoctorNotificationHandler(
	hub *ws.Hub,
	requestStorage repositories.AgentRequestStorage,
	sessionStorage repositories.SessionStorage,
) EventHandler {
	return &doctorNotificationHandler{
		requestStorage: requestStorage,
		hub:            hub,
		sessionStorage: sessionStorage,
	}
}
