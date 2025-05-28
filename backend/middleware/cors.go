package middleware

import (
	"m/backend/config"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// CorsMiddleware adds CORS headers to HTTP responses and handles preflight requests.
// Uses settings from the config.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from config and the Origin header from the request
		allowedOrigins := config.AppConfig.AllowedOrigins
		origin := r.Header.Get("Origin")
		if origin != "" {
			allowed := false
			// Check if the request's origin is in the allowed list
			for _, a := range allowedOrigins {
				if strings.TrimSpace(a) == origin {
					allowed = true
					break
				}
			}
			if allowed {
				// Allow CORS for this origin
				w.Header().Set("Access-Control-Allow-Origin", origin)
				logrus.Debugf("CorsMiddleware: allowed origin %s", origin)
			} else {
				// Log and do not set CORS headers for disallowed origins
				logrus.Warnf("CorsMiddleware: origin %s not allowed", origin)
			}
		}

		// Set standard CORS headers for all responses
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight (OPTIONS) requests: respond with no content
		if r.Method == "OPTIONS" {
			logrus.Debug("CorsMiddleware: preflight request processed")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler for non-OPTIONS requests
		next.ServeHTTP(w, r)
	})
}
