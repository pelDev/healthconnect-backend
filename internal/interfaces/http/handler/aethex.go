package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/utils"
)

type aethexHandler struct {
	eventBus     ports.EventBus
	sessionStore repositories.SessionStorage
}

// Request Models
type ReferDoctorData struct {
	Summary  string `json:"summary,required"`
	Symptoms string `json:"symptoms,required"`
}

type RaiseEmergencyData struct {
	Summary string `json:"summary,required"`
}

type AethexRequest[T any] struct {
	Arguments      T      `json:"arguments,required"`
	AgentID        string `json:"agent_id,required"`
	ConversationID string `json:"conversation_id,required"`
	CallID         string `json:"call_id"`
}

func NewAethexHandler(eventBus ports.EventBus, sessionStore repositories.SessionStorage) *aethexHandler {
	return &aethexHandler{
		eventBus:     eventBus,
		sessionStore: sessionStore,
	}
}

func (handler *aethexHandler) ReferToDoctor(w http.ResponseWriter, r *http.Request) {
	var request AethexRequest[ReferDoctorData]
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusUnprocessableEntity)
		return
	}

	sessionReference := request.ConversationID

	session, err := handler.sessionStore.GetSessionByReference(r.Context(), sessionReference)
	if err != nil {
		utils.RespondWithJson(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not find session",
		})
		return
	}

	if session == nil {
		utils.RespondWithJson(w, http.StatusNotFound, map[string]string{
			"message": "Session not found",
		})
		return
	}
}

func (handler *aethexHandler) RaiseEmergency(w http.ResponseWriter, r *http.Request) {
	var request AethexRequest[RaiseEmergencyData]
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusUnprocessableEntity)
		return
	}

	sessionReference := request.ConversationID

	log.Println(fmt.Sprintf("Raise emergency for conversation: %s Summary: %s", request.ConversationID, request.Arguments.Summary))

	session, err := handler.sessionStore.GetSessionByReference(r.Context(), sessionReference)
	if err != nil {
		utils.RespondWithJson(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not find session",
		})
		return
	}

	if session == nil {
		utils.RespondWithJson(w, http.StatusNotFound, map[string]string{
			"message": "Session not found",
		})
		return
	}

	utils.RespondWithJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Call has been set out for direct out-reach to emergency services",
	})
}
