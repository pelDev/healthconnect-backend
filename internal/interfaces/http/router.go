package http_interface

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	chi_middleware "github.com/go-chi/chi/middleware"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/interfaces/http/handler"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
)

func NewRouter(
	sessionStore repositories.SessionStorage,
	messageStore repositories.MessageStorage,
	voiceChatAdapter application_ports.VoiceChatAdapter,
) http.Handler {
	r := chi.NewRouter()

	// -------------------
	// Global middleware
	// -------------------
	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(chi_middleware.Timeout(30 * time.Second))
	r.Use(http_middleware.CORSMiddleware(map[string]struct{}{
		"http://localhost:5173": {},
	}))

	// -------------------
	// Handlers
	// -------------------
	voiceHandler := handler.NewVoiceHandler(voiceChatAdapter, sessionStore)

	// -------------------
	// Middleware
	// -------------------
	sessionMiddleware := http_middleware.NewSessionMiddleware()

	// -------------------
	// Routes
	// -------------------

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(sessionMiddleware.RequireSession)

		r.Route("/voice/sessions", func(r chi.Router) {
			r.Post("/", voiceHandler.InitializeCall)

			r.Route("/{sessionID}", func(r chi.Router) {
				r.Post("/offer", voiceHandler.ExchangeOffer)
			})
		})
	})

	return r
}
