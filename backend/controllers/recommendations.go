package controllers

import (
	"encoding/json"
	"net/http"

	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var recommendationService *services.RecommendationService

// InitRecommendationControllerService инициализирует сервис для рекомендаций.
func InitRecommendationControllerService(db *gorm.DB) {
	recommendationService = services.NewRecommendationService(db, nil)
	logrus.Info("Recommendations controller initialized")
}

// GetRecommendations – HTTP‑обработчик для эндпоинта GET /recommendations.
// Извлекает идентификатор текущего пользователя из контекста и вызывает бизнес-логику сервиса.
func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetRecommendations: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}

	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetRecommendations: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	ids, err := recommendationService.GetRecommendationsForUser(currentUserID)
	if err != nil {
		logrus.Errorf("GetRecommendations: ошибка получения рекомендаций для пользователя %s: %v", currentUserID, err)
		http.Error(w, "Error fetching recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Infof("GetRecommendations: рекомендации успешно получены для пользователя %s", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

// После GetRecommendations добавьте:
func DeclineRecommendation(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	recIDStr := vars["id"]
	recUserID, err := uuid.Parse(recIDStr)
	if err != nil {
		http.Error(w, "Invalid recommendation ID", http.StatusBadRequest)
		return
	}
	// Проверим, что такой кандидат был в рекомендациях (по желанию можно повторно вызвать GetRecommendationsForUser и проверить наличие recUserID)
	// Здесь упрощённо сразу делаем отказ:
	if err := recommendationService.DeclineRecommendation(currentUserID, recUserID); err != nil {
		logrus.Errorf("DeclineRecommendation: error saving decline for user %s -> %s: %v", currentUserID, recUserID, err)
		http.Error(w, "Error declining recommendation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
