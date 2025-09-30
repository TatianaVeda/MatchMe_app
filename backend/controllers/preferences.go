package controllers

import (
	"encoding/json"
	"net/http"

	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var preferencesDB *gorm.DB

// preferences.go - Handles HTTP endpoints for user search and recommendation preferences.
// Provides endpoints to get and update user preferences (e.g., max search radius).
// Automatically creates default preferences if not found.

// InitPreferencesController initializes the preferences controller with a database connection.
// Should be called once at startup.
func InitPreferencesController(db *gorm.DB) {
	preferencesDB = db
	logrus.Info("Preferences controller initialized")
}

// GetPreferences handles GET /me/preferences endpoint.
// Returns the current user's preferences. If not found, creates default preferences.
// Responds with JSON. Handles DB errors and returns appropriate HTTP status.
func GetPreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var pref models.Preference
	err = preferencesDB.
		Where("user_id = ?", uid).
		First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = models.Preference{
			UserID:    uid,
			MaxRadius: 0,
		}
		if err := preferencesDB.Create(&pref).Error; err != nil {
			logrus.Errorf("GetPreferences: error creating default preferences: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		logrus.Errorf("GetPreferences: error fetching preferences: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pref)
}

// UpdatePreferences handles PUT /me/preferences endpoint.
// Updates the current user's preferences (e.g., max search radius).
// If preferences do not exist, creates them. Handles DB errors and returns updated preferences as JSON.
func UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var req struct {
		MaxRadius float64 `json:"maxRadius"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var pref models.Preference
	err = preferencesDB.
		Where("user_id = ?", uid).
		First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = models.Preference{
			UserID:    uid,
			MaxRadius: req.MaxRadius,
		}
		if err := preferencesDB.Create(&pref).Error; err != nil {
			logrus.Errorf("UpdatePreferences: error creating preferences: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		logrus.Errorf("UpdatePreferences: error fetching preferences: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		pref.MaxRadius = req.MaxRadius
		if err := preferencesDB.Save(&pref).Error; err != nil {
			logrus.Errorf("UpdatePreferences: error saving preferences: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pref)
}
