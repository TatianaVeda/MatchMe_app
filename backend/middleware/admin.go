package middleware

import (
	"encoding/json"
	"net/http"

	"m/backend/models"
	//"m/backend/middleware"
	"m/backend/config"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AdminOnly возвращает middleware, которое проверяет, имеет ли текущий пользователь административные права.
// Здесь для демонстрации считается, что пользователь с email "admin@example.com" является администратором.
// Если у вас есть булево поле IsAdmin в модели User, замените условие проверки соответственно.
func AdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем userID, установленный предыдущим middleware (например, JWT-аутентификация).
			userIDStr, ok := r.Context().Value("userID").(string)
			if !ok {
				logrus.Warn("AdminOnly: userID не найден в контексте")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			uid, err := uuid.Parse(userIDStr)
			if err != nil {
				logrus.Errorf("AdminOnly: неверный userID: %v", err)
				http.Error(w, "Invalid user id", http.StatusBadRequest)
				return
			}

			var user models.User
			if err := db.First(&user, "id = ?", uid).Error; err != nil {
				logrus.Errorf("AdminOnly: пользователь %s не найден: %v", uid, err)
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Проверка административных прав.
			//if user.Email != fixtures.AdminEmail {
			if user.Email != config.AdminEmail {
				logrus.Warnf("AdminOnly: пользователь %s не является администратором", uid)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Admin access required"})
				return
			}

			logrus.Infof("AdminOnly: административный доступ подтвержден для пользователя %s", uid)
			next.ServeHTTP(w, r)
		})
	}
}
