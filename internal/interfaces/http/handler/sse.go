package handler

import (
	"fmt"
	"net/http"
	"time"

	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/events"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
	"github.com/pelDev/health-connect/internal/interfaces/ws"
)

type SseHandler struct {
	hub          *ws.Hub
	sessionStore repositories.AuthSessionStorage
	userStore    repositories.UserStorage
}

func NewSseHandler(
	hub *ws.Hub,
	sessionStore repositories.AuthSessionStorage,
	userStore repositories.UserStorage,
) *SseHandler {
	return &SseHandler{
		hub:          hub,
		sessionStore: sessionStore,
		userStore:    userStore,
	}
}

func (h *SseHandler) ConnectForDocEvents(w http.ResponseWriter, r *http.Request) {
	sessionId := http_middleware.GetSessionIDFromContext(r.Context())
	if sessionId == nil {
		handleError(w, domain_errors.ErrNotFound("session", "context"))
		return
	}

	session, err := h.sessionStore.FindByID(r.Context(), *sessionId)
	if err != nil {
		handleError(w, domain_errors.ErrDatabase(err))
		return
	}

	if session == nil {
		handleError(w, domain_errors.ErrNotFound("session", sessionId))
		return
	}

	user, err := h.userStore.FindByID(r.Context(), session.UserID)
	if err != nil {
		handleError(w, domain_errors.ErrDatabase(err))
		return
	}

	if user == nil {
		handleError(w, domain_errors.ErrNotFound("user", session.UserID))
		return
	}

	// Set http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Create a context with shutdown awareness
	ctx := r.Context()

	clientGone := ctx.Done()

	// Creating channel
	ch := make(chan ws.BroadcastMessage)
	heartbeatCh := make(chan []byte)
	client := ws.Client{
		Channel:          ch,
		HeartbeatChannel: heartbeatCh,
		UserID:           user.ID,
		DeviceID:         session.ID,
		Topics:           []string{string(events.DomainEventTypeEmergencyAlertTriggered), string(events.DomainEventTypeReferDoctorTriggered)},
	}

	// Register connect device to hub
	h.hub.Register <- client
	defer func() {
		h.hub.Unregister <- client
		fmt.Println("Client unregistered:", client.DeviceID)
	}()

	rc := http.NewResponseController(w)

	fmt.Fprintf(w, ": connection established\n\n")
	rc.Flush()

	for {
		select {
		case <-h.hub.Done:
			// Hub is shutting down - send shutdown message
			fmt.Println("Hub shutdown detected for device:", client.DeviceID)
			shutdownMsg := fmt.Sprintf("event: %s\ndata: %s\n\n", "shutdown", `{"reason":"server shutdown"}`)
			fmt.Fprintf(w, shutdownMsg, shutdownMsg)
			rc.Flush()

			// Give client a moment to process and close
			time.Sleep(200 * time.Millisecond)
			return

		case <-clientGone:
			// Client disconnected or server forcing shutdown
			fmt.Println("Context cancelled for device:", client.DeviceID)
			return

		// we got message to send!
		case msg := <-ch:
			_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.Topic, msg.Message)
			if err != nil {
				fmt.Println("Error writing data to sse", client.DeviceID, err)
				return
			}
			if err := rc.Flush(); err != nil {
				fmt.Println("Error flushing data to sse", client.DeviceID, err)
				return
			}

		// Send heart beat
		case msg := <-heartbeatCh:
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				fmt.Println("Error writing heartbeat to sse", client.DeviceID, err)
				return
			}
			fmt.Printf("Sent heartbeat to device: %s\n", client.DeviceID)
			if err := rc.Flush(); err != nil {
				fmt.Println("Error flushing heartbeat to sse", client.DeviceID, err)
				return
			}
		}
	}
}
