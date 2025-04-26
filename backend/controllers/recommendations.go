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

// ПРИ НЕОБХОДИМОСТИ УДАЛИТЬ RecommendationOutput — структура для отдачи id + distance в JSON
type RecommendationOutput struct {
	ID       uuid.UUID `json:"id"`
	Distance float64   `json:"distance"`
}

var recommendationService *services.RecommendationService

// InitRecommendationControllerService инициализирует сервис для рекомендаций.
func InitRecommendationControllerService(db *gorm.DB) {
	recommendationService = services.NewRecommendationService(db, nil)
	logrus.Info("Recommendations controller initialized")
}

// GetRecommendations – HTTP-обработчик для GET /recommendations
func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// 1) Извлечь userID из контекста
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

	// 2) Прочитать режим из query-параметра ?mode=
	mode := r.URL.Query().Get("mode")
	if mode != "desire" {
		mode = "affinity"
	}
	// по умолчанию отдавать и расстояния, если UI их захочет
	withDist := r.URL.Query().Get("withDistance") == "true"

	var resp interface{}
	if withDist {
		// возвращаем и ID, и Distance
		raw, err := recommendationService.GetRecommendationsWithDistance(currentUserID, mode)
		if err != nil {
			logrus.Errorf("GetRecommendations: ошибка получения рекомендаций для %s: %v", currentUserID, err)
			http.Error(w, "Error fetching recommendations: "+err.Error(), http.StatusInternalServerError)
			return
		}

		type Out struct {
			ID       uuid.UUID `json:"id"`
			Distance float64   `json:"distance"`
		}
		// 4) Преобразовать в выходную структуру для JSON
		out := make([]Out, len(raw))
		for i, c := range raw {
			out[i] = Out{ID: c.UserID, Distance: c.Distance}
		}
		resp = out
	} else {
		// возвращаем только список ID
		ids, err := recommendationService.GetRecommendationsForUser(currentUserID, mode)
		if err != nil {
			logrus.Errorf("GetRecommendations: ошибка получения рекомендаций для %s: %v", currentUserID, err)
			http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
			return
		}
		resp = ids
	}

	// 4) отдаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
	// Здесь упрощённо сразу делаем отказ:
	if err := recommendationService.DeclineRecommendation(currentUserID, recUserID); err != nil {
		logrus.Errorf("DeclineRecommendation: error saving decline for user %s -> %s: %v", currentUserID, recUserID, err)
		http.Error(w, "Error declining recommendation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
