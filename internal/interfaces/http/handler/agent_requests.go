package handler

import (
	"net/http"

	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
	"github.com/pelDev/health-connect/internal/utils"
)

type agentRequestsHandler struct {
	sessionStore         repositories.AuthSessionStorage
	agentRequestsStorage repositories.AgentRequestStorage
	userStorage          repositories.UserStorage
}

func NewAgentRequestsHandler(
	sessionStore repositories.AuthSessionStorage,
	agentRequestsStorage repositories.AgentRequestStorage,
	userStorage repositories.UserStorage,
) *agentRequestsHandler {
	return &agentRequestsHandler{
		sessionStore:         sessionStore,
		agentRequestsStorage: agentRequestsStorage,
		userStorage:          userStorage,
	}
}

func (h *agentRequestsHandler) ListDoctorRequests(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userStorage.FindByID(r.Context(), session.UserID)
	if err != nil {
		handleError(w, domain_errors.ErrDatabase(err))
		return
	}

	if user == nil {
		handleError(w, domain_errors.ErrNotFound("user", session.UserID))
		return
	}

	requests, err := h.agentRequestsStorage.ListDoctorRequests(r.Context(), user.ID)
	if err != nil {
		handleError(w, domain_errors.ErrDatabase(err))
		return
	}

	utils.RespondWithJson(w, http.StatusOK, requests)
}
