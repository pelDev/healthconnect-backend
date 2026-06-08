package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/application/usecases"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/utils"
)

type aethexHandler struct {
	eventBus     ports.EventBus
	sessionStore repositories.SessionStorage
	uowFactory   func(ctx context.Context) (application_ports.UnitOfWork, error)
}

type AethexRequest[T any] struct {
	Arguments      T      `json:"arguments,required"`
	AgentID        string `json:"agent_id,required"`
	ConversationID string `json:"conversation_id,required"`
	CallID         string `json:"call_id"`
}

func NewAethexHandler(eventBus ports.EventBus, sessionStore repositories.SessionStorage, uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)) *aethexHandler {
	return &aethexHandler{
		eventBus:     eventBus,
		sessionStore: sessionStore,
		uowFactory:   uowFactory,
	}
}

func (handler *aethexHandler) ReferToDoctor(w http.ResponseWriter, r *http.Request) {
	var request AethexRequest[application_dto.ReferDoctorData]
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusUnprocessableEntity)
		return
	}

	useCase := usecases.NewReferToDoctorUseCase(
		handler.eventBus,
		handler.sessionStore,
		handler.uowFactory,
	)

	err := useCase.Execute(r.Context(), request.ConversationID, request.Arguments)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Call has been set out for direct out-reach to healthcare practitioners",
	})
}

func (handler *aethexHandler) RaiseEmergency(w http.ResponseWriter, r *http.Request) {
	var request AethexRequest[application_dto.RaiseEmergencyData]
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
