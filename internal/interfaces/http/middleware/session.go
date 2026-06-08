package http_middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type contextKey string

const vidKey contextKey = "vid"
const sessionIdKey contextKey = "session_id"

type sessionMiddleware struct {
	authSessionStorage repositories.AuthSessionStorage
}

func NewSessionMiddleware(authSessionStorage repositories.AuthSessionStorage) *sessionMiddleware {
	return &sessionMiddleware{
		authSessionStorage: authSessionStorage,
	}
}

func (authn *sessionMiddleware) RequireVID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vid uuid.UUID

		cookie, err := r.Cookie("vid")
		if err == nil && cookie.Value != "" {
			vid = uuid.MustParse(cookie.Value)
		} else {
			vid = uuid.New()
			http.SetCookie(w, &http.Cookie{
				Name:     "vid",
				Value:    vid.String(),
				Expires:  time.Now().UTC().Add(24 * time.Hour),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
			})
		}

		ctx := context.WithValue(r.Context(), vidKey, vid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (authn *sessionMiddleware) RequireAuthSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionID uuid.UUID

		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionID = uuid.MustParse(cookie.Value)
		session, err := authn.authSessionStorage.FindByID(r.Context(), sessionID)
		if err != nil || session.ExpiresAt.Before(time.Now()) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), sessionIdKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetSessionIDFromContext(ctx context.Context) *uuid.UUID {
	if usr, ok := ctx.Value(sessionIdKey).(uuid.UUID); ok {
		return &usr
	}
	return nil
}

func GetVIDFromContext(ctx context.Context) *uuid.UUID {
	if usr, ok := ctx.Value(vidKey).(uuid.UUID); ok {
		return &usr
	}
	return nil
}
