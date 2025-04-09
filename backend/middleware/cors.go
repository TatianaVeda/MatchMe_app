package middleware

import (
	"m/backend/config"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// CorsMiddleware разрешает CORS-запросы согласно настройкам из конфигурации.
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
				logrus.Debugf("CorsMiddleware: разрешен origin %s", origin)
			} else {
				logrus.Warnf("CorsMiddleware: origin %s не разрешен", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Обработка preflight-запросов.
		if r.Method == "OPTIONS" {
			logrus.Debug("CorsMiddleware: обработан preflight запрос")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
