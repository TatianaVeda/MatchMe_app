package middleware

import (
	"m/backend/config"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := config.AppConfig.AllowedOrigins
		origin := r.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, a := range allowedOrigins {
				if strings.TrimSpace(a) == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				logrus.Debugf("CorsMiddleware: allowed origin %s", origin)
			} else {
				logrus.Warnf("CorsMiddleware: origin %s not allowed", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			logrus.Debug("CorsMiddleware: preflight request processed")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
