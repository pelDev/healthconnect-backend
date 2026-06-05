package http_middleware

import (
	"net/http"
)

func CORSMiddleware(allowedOrigins map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowedOrigins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Vary", "Origin")
				} else {
					http.Error(w, "CORS origin denied", http.StatusForbidden)
					return
				}
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set(
					"Access-Control-Allow-Headers",
					"Content-Type, X-CSRF-Token",
				)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
