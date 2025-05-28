package middleware

import (
	"encoding/json"
	"net/http"

	"m/backend/config"
	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AdminOnly is a middleware that allows access only to administrators (email from config).
// Checks userID in context, looks up the user in the database, and compares the email.
func AdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userIDStr, ok := r.Context().Value("userID").(string)
			if !ok {
				logrus.Warn("AdminOnly: userID not found in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			uid, err := uuid.Parse(userIDStr)
			if err != nil {
				logrus.Errorf("AdminOnly: invalid userID: %v", err)
				http.Error(w, "Invalid user id", http.StatusBadRequest)
				return
			}

			var user models.User
			if err := db.First(&user, "id = ?", uid).Error; err != nil {
				logrus.Errorf("AdminOnly: user %s not found: %v", uid, err)
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			if user.Email != config.AdminEmail {
				logrus.Warnf("AdminOnly: user %s is not an administrator", uid)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Admin access required"})
				return
			}

			logrus.Infof("AdminOnly: admin access granted for user %s", uid)
			next.ServeHTTP(w, r)
		})
	}
}
