package controllers

import (
	"context"
	"net/http"
	"strings"

	"m/backend/config"
	"m/backend/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware проверяет наличие и корректность JWT-токена в заголовке Authorization,
// извлекает из него userID и помещает его в контекст запроса.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Warn("AuthMiddleware: отсутствует заголовок Authorization")
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Ожидаем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logrus.Warn("AuthMiddleware: неверный формат заголовка Authorization")
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Парсим JWT-токен с использованием структуры JWTClaims, определенной в models/auth.go
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Секрет для подписи берется из конфигурации
			return []byte(config.AppConfig.JWTSecret), nil
		})
		if err != nil {
			logrus.Errorf("AuthMiddleware: ошибка при парсинге токена: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			logrus.Warn("AuthMiddleware: токен не валиден")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			logrus.Error("AuthMiddleware: неверные claims токена")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		logrus.Infof("AuthMiddleware: успешно аутентифицирован пользователь %s", claims.UserID.String())
		// Помещаем идентификатор пользователя (userID) в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", claims.UserID.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
