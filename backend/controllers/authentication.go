package controllers

import (
	"encoding/json"
	"errors"
	"m/backend/config"
	"m/backend/models"
	"m/backend/services"
	"m/backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var authDB *gorm.DB

func InitAuthenticationController(db *gorm.DB) {
	authDB = db
	logrus.Info("Authentication controller initialized")
}

// Signup handles user registration requests. Validates input and creates a new user.
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
		"userId": user.ID,
		"email":  user.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login handles user login requests. Validates credentials and returns JWT tokens.
func Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("Login: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := models.AuthenticateUser(authDB, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrInvalidCredentials) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			logrus.Errorf("Login: error authenticating user %s: %v", req.Email, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	accessToken, err := models.GenerateAccessToken(user.ID, config.AppConfig.JWTSecret)
	if err != nil {
		logrus.Errorf("Login: error generating access token for user %s: %v", user.Email, err)
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := models.GenerateRefreshToken(user.ID, config.AppConfig.JWTSecret, config.AppConfig.JWTRefreshExpiresIn)
	if err != nil {
		logrus.Errorf("Login: error generating refresh token for user %s: %v", user.Email, err)
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout requests. Adds the token to the blacklist.
func Logout(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
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
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// RefreshToken handles requests to refresh JWT tokens.
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("RefreshToken: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(reqBody.RefreshToken) == "" {
		logrus.Warn("RefreshToken: missing refresh token in request")
		http.Error(w, "Missing refreshToken", http.StatusBadRequest)
		return
	}
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
	newAccessToken, err := models.GenerateAccessToken(claims.UserID, config.AppConfig.JWTSecret)
	if err != nil {
		logrus.Errorf("RefreshToken: failed to generate new access token: %v", err)
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := models.GenerateRefreshToken(claims.UserID, config.AppConfig.JWTSecret, config.AppConfig.JWTRefreshExpiresIn)
	if err != nil {
		logrus.Errorf("RefreshToken: failed to generate new refresh token: %v", err)
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateEmail handles requests to update the user's email address.
func UpdateEmail(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateEmail(body.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var count int64
	authDB.Model(&models.User{}).
		Where("email = ?", body.Email).
		Count(&count)
	if count > 0 {
		http.Error(w, "Email already in use", http.StatusBadRequest)
		return
	}
	if err := authDB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("email", body.Email).Error; err != nil {
		logrus.Errorf("UpdateEmail: error updating email for user %s: %v", userID, err)
		http.Error(w, "Error updating email", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"email": body.Email,
	})
}

// UpdatePassword handles requests to update the user's password.
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	var body struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var user models.User
	if err := authDB.First(&user, "id = ?", userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if !models.CheckPasswordHash(body.Current, user.PasswordHash) {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}
	if err := utils.ValidatePassword(body.New); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashed, err := models.HashPassword(body.New)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	if err := authDB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", hashed).Error; err != nil {
		logrus.Errorf("UpdatePassword: error updating password for user %s: %v", userID, err)
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
}
