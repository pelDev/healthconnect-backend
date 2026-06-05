package http_middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const vidKey contextKey = "vid"

type sessionMiddleware struct{}

func NewSessionMiddleware() *sessionMiddleware {
	return &sessionMiddleware{}
}

func (authn *sessionMiddleware) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vid uuid.UUID

		cookie, err := r.Cookie("vid")
		if err == nil {
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

func GetVIDFromContext(ctx context.Context) *uuid.UUID {
	if usr, ok := ctx.Value(vidKey).(uuid.UUID); ok {
		return &usr
	}
	return nil
}
