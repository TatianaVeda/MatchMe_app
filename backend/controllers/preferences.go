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

// InitPreferencesController нужно вызвать при старте (см. следующий пункт)
func InitPreferencesController(db *gorm.DB) {
	preferencesDB = db
	logrus.Info("Preferences controller initialized")
}

// GetPreferences возвращает запись Preference для текущего пользователя.
// Если записи нет — создаёт с дефолтными значениями.
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
		// создаём дефолтную запись
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

// UpdatePreferences обновляет MaxRadius у записи Preference.
// Ожидает JSON { "maxRadius": <число> }.
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
		// если нет — создаём
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
		// обновляем существующую
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
