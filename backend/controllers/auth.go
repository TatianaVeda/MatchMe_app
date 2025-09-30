package controllers

import (
	"context"
	"m/backend/config"
	"m/backend/models"
	"m/backend/services"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware is an HTTP middleware that authenticates requests using JWT tokens.
// - Validates Authorization header format and parses JWT.
// - Checks token expiration and blacklist (revoked tokens).
// - On success, injects userID into request context for downstream handlers.
// - On failure, responds with appropriate HTTP status and error message.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Warn("AuthMiddleware: missing Authorization header")
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Warn("AuthMiddleware: invalid Authorization header format")
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})
		if err != nil {
			logrus.Errorf("AuthMiddleware: error parsing token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			logrus.Warn("AuthMiddleware: token is invalid")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			logrus.Error("AuthMiddleware: invalid token claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		if claims.ExpiresAt < time.Now().Unix() {
			logrus.Warn("AuthMiddleware: token expired")
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}
		if services.IsBlacklisted(tokenString) {
			logrus.Warn("AuthMiddleware: token is in the blacklist (revoked)")
			http.Error(w, "Token revoked", http.StatusUnauthorized)
			return
		}
		logrus.Infof("AuthMiddleware: successfully authenticated user %s", claims.UserID.String())
		ctx := context.WithValue(r.Context(), "userID", claims.UserID.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
