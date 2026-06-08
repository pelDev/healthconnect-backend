package http_interface

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	chi_middleware "github.com/go-chi/chi/middleware"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/interfaces/http/handler"
	http_middleware "github.com/pelDev/health-connect/internal/interfaces/http/middleware"
	"github.com/pelDev/health-connect/internal/interfaces/ws"
)

func NewRouter(
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error),
	sessionStore repositories.SessionStorage,
	messageStore repositories.MessageStorage,
	authSessionStore repositories.AuthSessionStorage,
	userStore repositories.UserStorage,
	agentRequestsStorage repositories.AgentRequestStorage,
	voiceChatAdapter application_ports.VoiceChatAdapter,
	eventBus ports.EventBus,
	hub *ws.Hub,
) http.Handler {
	r := chi.NewRouter()

	// workDir, _ := os.Getwd()
	// filesDir := http.Dir(filepath.Join(workDir, "dist/assets"))

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
		"http://localhost:8001": {},
		"https://healthconnect-backend-production-25ff.up.railway.app": {},
	}))

	// -------------------
	// Handlers
	// -------------------
	voiceHandler := handler.NewVoiceHandler(voiceChatAdapter, sessionStore, uowFactory)
	aethexHandler := handler.NewAethexHandler(eventBus, sessionStore, uowFactory)
	authHandler := handler.NewAuthHandler(uowFactory, authSessionStore, userStore)
	sseHandler := handler.NewSseHandler(hub, authSessionStore, userStore)
	agentRequestHandler := handler.NewAgentRequestsHandler(authSessionStore, agentRequestsStorage, userStore)

	// -------------------
	// Middleware
	// -------------------
	sessionMiddleware := http_middleware.NewSessionMiddleware(authSessionStore)

	// -------------------
	// Routes
	// -------------------

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(sessionMiddleware.RequireVID)

			r.Route("/voice/sessions", func(r chi.Router) {
				r.Post("/", voiceHandler.InitializeCall)

				r.Route("/{sessionID}", func(r chi.Router) {
					r.Post("/offer", voiceHandler.ExchangeOffer)
				})
			})
		})

		r.Route("/aethex/function", func(r chi.Router) {
			r.Post("/trigger_emergency_alert", aethexHandler.RaiseEmergency)
			r.Post("/refer_to_doctor", aethexHandler.ReferToDoctor)
		})

		r.Group(func(r chi.Router) {
			r.Use(sessionMiddleware.RequireAuthSession)
			// r.Use(middleware.CSRFMiddleware)

			r.Get("/me", authHandler.Me)

			r.Route("/doc", func(r chi.Router) {
				r.Get("/sse", sseHandler.ConnectForDocEvents)
				r.Get("/requests", agentRequestHandler.ListDoctorRequests)
			})
		})

	})

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./dist/assets"))))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists in public folder
		filePath := filepath.Join("./dist", r.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./dist/index.html")
	})

	return r
}
