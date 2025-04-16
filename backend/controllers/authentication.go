package controllers

import (
	"encoding/json"
	"m/backend/config"
	"m/backend/models"
	"m/backend/services"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// authDB используется для доступа к базе данных в функциях аутентификации.
var authDB *gorm.DB

// InitAuthenticationController инициализирует контроллер аутентификации.
func InitAuthenticationController(db *gorm.DB) {
	authDB = db
	logrus.Info("Authentication controller initialized")
}

// Signup – endpoint для регистрации нового пользователя.
//
// Принимает POST-запрос с JSON телом:
//
//	{
//	   "email": "user@example.com",
//	   "password": "пароль"
//	}
//
// Если регистрация прошла успешно, возвращает JSON с информацией о новом пользователе.
func Signup(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("Signup: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := models.CreateUser(authDB, reqBody.Email, reqBody.Password)
	if err != nil {
		logrus.Errorf("Signup: error creating user %s: %v", reqBody.Email, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.Infof("Signup: user %s successfully registered", user.Email)
	response := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout – endpoint для выхода из системы.
// Так как JWT является stateless-токеном, сервер не может "аннулировать" его без дополнительной логики (например, blacklist).
// Здесь мы просто возвращаем клиенту сообщение о том, что выход выполнен успешно.
func Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста, установленного AuthMiddleware.
	userID, _ := r.Context().Value("userID").(string)

	// Извлекаем токен из заголовка, чтобы добавить его в чёрный список.
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			// Парсим токен, чтобы получить время истечения.
			token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.AppConfig.JWTSecret), nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*models.JWTClaims); ok {
					expirationTime := time.Unix(claims.ExpiresAt, 0)
					services.AddToken(tokenString, expirationTime)
				}
			}
		}
	}

	logrus.Infof("Logout: user %s logged out", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// RefreshToken – endpoint для обновления access и refresh токенов.
//
// Клиент отправляет POST-запрос с JSON телом:
//
//	{
//	   "refresh_token": "старый_refresh_токен"
//	}
//
// Если refresh токен проходит валидацию, сервер генерирует новый access-token и новый refresh-token,
// используя настройку времени жизни refresh-токена из конфигурации (JWTRefreshExpiresIn).
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("RefreshToken: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Парсим refresh токен с использованием наших claims.
	token, err := jwt.ParseWithClaims(reqBody.RefreshToken, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		logrus.Errorf("RefreshToken: invalid refresh token: %v", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		logrus.Error("RefreshToken: invalid token claims")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Генерируем новый access token (например, с коротким временем жизни, 15 минут).
	newAccessToken, err := models.GenerateAccessToken(claims.UserID, config.AppConfig.JWTSecret)
	if err != nil {
		logrus.Errorf("RefreshToken: failed to generate new access token: %v", err)
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	// Генерируем новый refresh token, передавая время жизни из конфигурации.
	newRefreshToken, err := models.GenerateRefreshToken(claims.UserID, config.AppConfig.JWTSecret, config.AppConfig.JWTRefreshExpiresIn)
	if err != nil {
		logrus.Errorf("RefreshToken: failed to generate new refresh token: %v", err)
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
