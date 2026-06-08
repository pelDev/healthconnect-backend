package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/application/usecases"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
	"github.com/pelDev/health-connect/internal/utils"
)

type authHandler struct {
	uowFactory     func(ctx context.Context) (application_ports.UnitOfWork, error)
	sessionStorage repositories.AuthSessionStorage
	userStorage    repositories.UserStorage
}

func NewAuthHandler(
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error),
	sessionStorage repositories.AuthSessionStorage,
	userStorage repositories.UserStorage,
) *authHandler {
	return &authHandler{
		uowFactory:     uowFactory,
		sessionStorage: sessionStorage,
		userStorage:    userStorage,
	}
}

func (handler *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req application_dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJson(w, http.StatusBadRequest, map[string]string{
			"message": "invalid request body",
		})
		return
	}

	useCase := usecases.NewLoginUserUseCase(handler.uowFactory)

	res, err := useCase.Execute(r.Context(), req)
	if err != nil {
		fmt.Println("error logging in", err)
		handleError(w, err)
		return
	}

	setAuthCookies(w, res.SessionID)

	utils.RespondWithJson(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (handler *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := http_middleware.GetSessionIDFromContext(r.Context())

	if sessionID == nil {
		utils.RespondWithJson(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
		return
	}

	useCase := usecases.NewRevokeAuthSessionUseCase(handler.uowFactory)

	err := useCase.Execute(r.Context(), *sessionID)
	if err != nil {
		handleError(w, err)
		return
	}

	clearAuthCookies(w)

	utils.RespondWithJson(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (handler *authHandler) Me(w http.ResponseWriter, r *http.Request) {
	sessionID := http_middleware.GetSessionIDFromContext(r.Context())
	if sessionID == nil {
		utils.RespondWithJson(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	session, err := handler.sessionStorage.FindByID(r.Context(), *sessionID)
	if err != nil {
		handleError(w, err)
		return
	}

	if session == nil {
		handleError(w, domain_errors.ErrNotFound("session", sessionID))
		return
	}

	user, err := handler.userStorage.FindByID(r.Context(), session.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	if user == nil {
		handleError(w, domain_errors.ErrNotFound("user", session.UserID))
	}

	utils.RespondWithJson(w, http.StatusOK, user)
}

func setAuthCookies(w http.ResponseWriter, sessionID uuid.UUID) {
	// set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	csrfToken := generateCSRFToken()

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  expired,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Expires:  expired,
		MaxAge:   -1,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
