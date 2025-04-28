package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"m/backend/models"
	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RecommendationOutput — DTO для ответа с расстоянием.
type RecommendationOutput struct {
	ID       uuid.UUID `json:"id"`
	Distance float64   `json:"distance"`
}

var recommendationService *services.RecommendationService

// InitRecommendationControllerService вызывается из main.go или routes.go
func InitRecommendationControllerService(db *gorm.DB) {
	recommendationService = services.NewRecommendationService(db, nil)
	logrus.Info("Recommendations controller initialized")
}

// parseMode проверяет параметр mode и возвращает либо "affinity", либо "desire".
func parseMode(q string) (string, error) {
	switch q {
	case "", "affinity":
		return "affinity", nil
	case "desire":
		return "desire", nil
	default:
		return "", fmt.Errorf("invalid mode %q", q)
	}
}

// GetRecommendations отдаёт до 10 рекомендаций.
// GET /recommendations?mode={affinity|desire}&withDistance=true
func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// 1) userID
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, _ := uuid.Parse(userIDStr)

	// 2) Проверяем, что у пользователя есть координаты
	var profile models.Profile
	if err := recommendationService.DB.
		Select("latitude, longitude").
		Where("user_id = ?", currentUserID).
		First(&profile).Error; err != nil {
		logrus.WithField("userID", currentUserID).Errorf("failed to load profile geo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if profile.Latitude == 0 || profile.Longitude == 0 {
		http.Error(w, "Пожалуйста, укажите ваш город", http.StatusBadRequest)
		return
	}

	// 3) Парсим mode
	mode, err := parseMode(r.URL.Query().Get("mode"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4) Разрешаем выдачу с расстоянием или без
	withDist := r.URL.Query().Get("withDistance") == "true"
	w.Header().Set("Content-Type", "application/json")

	if withDist {
		raw, err := recommendationService.GetRecommendationsWithDistance(currentUserID, mode)
		if err != nil {
			logrus.WithFields(logrus.Fields{"userID": currentUserID, "mode": mode}).
				Errorf("GetRecommendationsWithDistance failed: %v", err)
			http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
			return
		}
		// Маппим в DTO
		out := make([]RecommendationOutput, len(raw))
		for i, r := range raw {
			out[i] = RecommendationOutput{ID: r.UserID, Distance: r.Distance}
		}
		if err := json.NewEncoder(w).Encode(out); err != nil {
			logrus.WithError(err).Error("Failed to serialize recommendations with distance")
		}
	} else {
		ids, err := recommendationService.GetRecommendationsForUser(currentUserID, mode)
		if err != nil {
			logrus.WithFields(logrus.Fields{"userID": currentUserID, "mode": mode}).
				Errorf("GetRecommendationsForUser failed: %v", err)
			http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(ids); err != nil {
			logrus.WithError(err).Error("Failed to serialize recommendation IDs")
		}
	}
}

// DeclineRecommendation обрабатывает отказ от рекомендации.
// POST /recommendations/{id}/decline
func DeclineRecommendation(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, _ := uuid.Parse(userIDStr)

	vars := mux.Vars(r)
	recID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recommendation ID", http.StatusBadRequest)
		return
	}

	if err := recommendationService.DeclineRecommendation(currentUserID, recID); err != nil {
		logrus.WithFields(logrus.Fields{"userID": currentUserID, "recID": recID}).
			Errorf("DeclineRecommendation failed: %v", err)
		http.Error(w, "Error declining recommendation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
