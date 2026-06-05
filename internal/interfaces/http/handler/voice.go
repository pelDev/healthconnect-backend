package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/application/usecases"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
	"github.com/pelDev/health-connect/internal/utils"
)

type voiceHandler struct {
	voiceChatAdapter application_ports.VoiceChatAdapter
	sessionStore     repositories.SessionStorage
}

func NewVoiceHandler(voiceChatAdapter application_ports.VoiceChatAdapter, sessionStore repositories.SessionStorage) *voiceHandler {
	return &voiceHandler{voiceChatAdapter: voiceChatAdapter, sessionStore: sessionStore}
}

func (h *voiceHandler) InitializeCall(w http.ResponseWriter, r *http.Request) {
	vid := http_middleware.GetVIDFromContext(r.Context())
	if vid == nil {
		utils.RespondWithJson(w, http.StatusUnauthorized, map[string]string{
			"message": "unauthorized",
		})
		return
	}

	usecase := usecases.NewInitializeVoiceChatUseCase(h.voiceChatAdapter, h.sessionStore)

	res, err := usecase.Execute(r.Context(), *vid)
	if err != nil {
		fmt.Println("error initializing call:")
		fmt.Println(err)
		utils.RespondWithJson(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJson(w, http.StatusCreated, res)
}

func (h *voiceHandler) ExchangeOffer(w http.ResponseWriter, r *http.Request) {
	var input application_dto.SendVoiceChatOfferReq
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	sessionIDParam := chi.URLParam(r, "sessionID")
	if sessionIDParam == "" {
		http.Error(w, "orgID is required", http.StatusBadRequest)
		return
	}

	requestedSessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		http.Error(w, "invalid sessionID", http.StatusBadRequest)
		return
	}

	session, err := h.sessionStore.GetSession(requestedSessionID)
	if err != nil || session == nil || session.Reference == nil {
		http.Error(w, "invalid sessionID", http.StatusBadRequest)
		return
	}

	res, err := h.voiceChatAdapter.SendOffer(r.Context(), *session.Reference, input)
	if err != nil {
		fmt.Println("error sending offer:")
		fmt.Println(err)
		utils.RespondWithJson(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJson(w, http.StatusOK, res)
}
