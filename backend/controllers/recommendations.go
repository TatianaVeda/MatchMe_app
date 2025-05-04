package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	Score    float64   `json:"score"`
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

	fmt.Println("Extracted userID from context:", userIDStr)

	// 2) Проверяем mode
	mode, err := parseMode(r.URL.Query().Get("mode"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3) Определяем, возвращать расстояния или нет
	withDist := r.URL.Query().Get("withDistance") == "true"
	//it's always true!!!
	w.Header().Set("Content-Type", "application/json")

	// 4) Читаем флаг useProfile (по умолчанию true)
	useProfile := r.URL.Query().Get("useProfile") != "false"

	fmt.Println("Mode:", mode)
	fmt.Println("UseProfile:", useProfile)

	var (
		idsWithDist []services.RecommendationWithDistance
		ids         []uuid.UUID
	)

	if useProfile {
		fmt.Println("Using saved profile filters")
		idsWithDist, err = recommendationService.GetRecommendationsWithDistance(currentUserID, mode)
		//always with distance
		// if withDist {
		// 	fmt.Println("Calling GetRecommendationsWithDistance for user:", currentUserID)
		// 	idsWithDist, err = recommendationService.GetRecommendationsWithDistance(currentUserID, mode)
		// } else {
		// 	fmt.Println("Calling GetRecommendationsForUser for user:", currentUserID)
		// 	ids, err = recommendationService.GetRecommendationsForUser(currentUserID, mode)
		// }
	} else {
		// Новый путь: читаем фильтры из query-параметров
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("cityLat"), 64)
		lon, _ := strconv.ParseFloat(r.URL.Query().Get("cityLon"), 64)

		// affinity-фильтры
		interests := strings.FieldsFunc(r.URL.Query().Get("interests"), func(r rune) bool { return r == ',' })
		priorityInterests, _ := strconv.ParseBool(r.URL.Query().Get("priorityInterests"))

		hobbies := strings.FieldsFunc(r.URL.Query().Get("hobbies"), func(r rune) bool { return r == ',' })
		priorityHobbies, _ := strconv.ParseBool(r.URL.Query().Get("priorityHobbies"))

		music := strings.FieldsFunc(r.URL.Query().Get("music"), func(r rune) bool { return r == ',' })
		priorityMusic, _ := strconv.ParseBool(r.URL.Query().Get("priorityMusic"))

		food := strings.FieldsFunc(r.URL.Query().Get("food"), func(r rune) bool { return r == ',' })
		priorityFood, _ := strconv.ParseBool(r.URL.Query().Get("priorityFood"))

		travel := strings.FieldsFunc(r.URL.Query().Get("travel"), func(r rune) bool { return r == ',' })
		priorityTravel, _ := strconv.ParseBool(r.URL.Query().Get("priorityTravel"))

		// desire-фильтр
		lookingFor := r.URL.Query().Get("lookingFor")

		fmt.Println("Lat:", lat, "Lon:", lon)
		fmt.Println("Interests:", interests, "Priority:", priorityInterests)
		fmt.Println("Hobbies:", hobbies, "Priority:", priorityHobbies)
		fmt.Println("Music:", music, "Priority:", priorityMusic)
		fmt.Println("Food:", food, "Priority:", priorityFood)
		fmt.Println("Travel:", travel, "Priority:", priorityTravel)
		fmt.Println("LookingFor:", lookingFor)

		idsWithDist, err = recommendationService.GetRecommendationsWithFiltersWithDistance(
			currentUserID, mode,
			lat, lon,
			interests, priorityInterests,
			hobbies, priorityHobbies,
			music, priorityMusic,
			food, priorityFood,
			travel, priorityTravel,
			lookingFor,
		)

		// NOT USED!!! ALWAYS WITH DISTANCE
		// if withDist {
		// 	idsWithDist, err = recommendationService.GetRecommendationsWithFiltersWithDistance(
		// 		currentUserID, mode,
		// 		lat, lon,
		// 		interests, priorityInterests,
		// 		hobbies, priorityHobbies,
		// 		music, priorityMusic,
		// 		food, priorityFood,
		// 		travel, priorityTravel,
		// 		lookingFor,
		// 	)
		// } else
		// {
		// 	ids, err = recommendationService.GetRecommendationsWithFilters(
		// 		currentUserID, mode,
		// 		lat, lon,
		// 		interests, priorityInterests,
		// 		hobbies, priorityHobbies,
		// 		music, priorityMusic,
		// 		food, priorityFood,
		// 		travel, priorityTravel,
		// 		lookingFor,
		// 	)
		// }
	}

	// if err != nil {

	// 	logrus.WithFields(logrus.Fields{"userID": currentUserID, "mode": mode}).Errorf("GetRecommendations failed: %v", err)
	// 	http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
	// 	return
	// }

	if err == nil {
		fmt.Println("Received", len(idsWithDist), "recommendations with distance")
		// if withDist {
		// 	fmt.Println("Received", len(idsWithDist), "recommendations with distance")
		// } else {
		// 	fmt.Println("Received", len(ids), "recommendations")
		// }
	}

	if err != nil {
		// если профиль или био неполные — возвращаем просто пустой массив вместо 500
		msg := err.Error()
		if strings.Contains(msg, "пожалуйста, заполните вашу биографию") ||
			strings.Contains(msg, "пожалуйста, укажите имя и фамилию") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
			return
		}
		// все остальные ошибки — настоящая 500, логируем точный текст
		logrus.WithFields(logrus.Fields{
			"userID": currentUserID,
			"mode":   mode,
		}).Errorf("GetRecommendations failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5) Сериализуем ответ
	if withDist {
		out := make([]RecommendationOutput, len(idsWithDist))
		for i, rec := range idsWithDist {
			out[i] = RecommendationOutput{ID: rec.UserID, Distance: rec.Distance, Score: rec.Score}
		}
		json.NewEncoder(w).Encode(out)
	} else {
		json.NewEncoder(w).Encode(ids)
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
