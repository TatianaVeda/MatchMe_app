package controllers

import (
	"encoding/json"
	"fmt"
	"m/backend/models"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func containsUUID(arr []uuid.UUID, x uuid.UUID) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

type RecommendationOutput struct {
	ID uuid.UUID `json:"id"`
	//Distance float64   `json:"distance"`
	//Score    float64   `json:"score"`
	//Online   bool      `json:"online"`
}

var recommendationService *services.RecommendationService
var presenceService *services.PresenceService

// InitRecommendationControllerService initializes the recommendation and presence services for this controller.
// Should be called once at startup.
func InitRecommendationControllerService(db *gorm.DB, ps *services.PresenceService) {
	recommendationService = services.NewRecommendationService(db, nil)
	presenceService = ps // ✅ added
	logrus.Info("Recommendations controller initialized")
}

// parseMode parses the recommendation mode from query string.
// Returns "affinity" (default) or "desire". Returns error for invalid values.
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

// GetRecommendations handles GET /recommendations endpoint.
// Returns a list of recommended users for the current user, with optional distance and score.
// Supports two modes: profile-based (uses saved user preferences) and custom-filtered (uses query params).
// Filters out users with pending/declined connections. Adds online status via presenceService.
// Handles errors and incomplete profiles gracefully (returns empty array for known validation errors).
func GetRecommendations(w http.ResponseWriter, r *http.Request) {

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, _ := uuid.Parse(userIDStr)

	fmt.Println("Extracted userID from context:", userIDStr)

	mode, err := parseMode(r.URL.Query().Get("mode"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	withDist := r.URL.Query().Get("withDistance") == "true"

	w.Header().Set("Content-Type", "application/json")

	useProfile := r.URL.Query().Get("useProfile") != "false"

	fmt.Println("Mode:", mode)
	fmt.Println("UseProfile:", useProfile)

	var pendingIDs, declinedIDs []uuid.UUID

	recommendationService.DB.
		Model(&models.Connection{}).
		Where("connection_id = ? AND status = ?", currentUserID, "pending").
		Pluck("user_id", &pendingIDs)

	recommendationService.DB.
		Model(&models.Recommendation{}).
		Where("user_id = ? AND status = ?", currentUserID, "declined").
		Pluck("rec_user_id", &declinedIDs)

	var (
		idsWithDist []services.RecommendationWithDistance
		ids         []uuid.UUID
	)

	if useProfile {
		fmt.Println("Using saved profile filters")
		idsWithDist, err = recommendationService.GetRecommendationsWithDistance(currentUserID, mode)

	} else {

		lat, _ := strconv.ParseFloat(r.URL.Query().Get("cityLat"), 64)
		lon, _ := strconv.ParseFloat(r.URL.Query().Get("cityLon"), 64)

		interests := strings.FieldsFunc(r.URL.Query().Get("interests"), func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		priorityInterests, _ := strconv.ParseBool(r.URL.Query().Get("priorityInterests"))

		hobbies := strings.FieldsFunc(r.URL.Query().Get("hobbies"), func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		priorityHobbies, _ := strconv.ParseBool(r.URL.Query().Get("priorityHobbies"))

		music := strings.FieldsFunc(r.URL.Query().Get("music"), func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		priorityMusic, _ := strconv.ParseBool(r.URL.Query().Get("priorityMusic"))

		food := strings.FieldsFunc(r.URL.Query().Get("food"), func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		priorityFood, _ := strconv.ParseBool(r.URL.Query().Get("priorityFood"))

		travel := strings.FieldsFunc(r.URL.Query().Get("travel"), func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
		priorityTravel, _ := strconv.ParseBool(r.URL.Query().Get("priorityTravel"))

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

	}

	if err == nil {
		fmt.Println("Received", len(idsWithDist), "recommendations with distance")

	}

	if err != nil {

		msg := err.Error()
		if strings.Contains(msg, "please fill in your profile and biography to get recommendations") ||
			strings.Contains(msg, "please provide your first and last name") ||
			strings.Contains(msg, "please complete your biography") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
			return
		}

		logrus.WithFields(logrus.Fields{
			"userID": currentUserID,
			"mode":   mode,
		}).Errorf("GetRecommendations failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := make([]services.RecommendationWithDistance, 0, len(idsWithDist))
	for _, rec := range idsWithDist {
		if containsUUID(pendingIDs, rec.UserID) || containsUUID(declinedIDs, rec.UserID) {
			continue
		}
		filtered = append(filtered, rec)
	}
	idsWithDist = filtered

	if withDist {
		out := make([]RecommendationOutput, len(idsWithDist))
		for i, rec := range idsWithDist {
			//online, _ := presenceService.IsOnline(rec.UserID.String()) // ✅ added
			out[i] = RecommendationOutput{
				ID: rec.UserID,
				//Distance: rec.Distance,
				//Score:    rec.Score,
				//Online:   online,
			}
		}
		json.NewEncoder(w).Encode(out)
	} else {
		json.NewEncoder(w).Encode(ids)
	}
}

// DeclineRecommendation handles POST /recommendations/{id}/decline endpoint.
// Marks a recommendation as declined for the current user.
// Returns 204 No Content on success, or error status on failure.
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
