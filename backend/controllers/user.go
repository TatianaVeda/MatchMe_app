package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"m/backend/models"
	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var db *gorm.DB

// initializes the user controller with the database connection.
func InitUserController(database *gorm.DB) {
	db = database
	logrus.Info("User controller initialized")
}

// userHasAccess checks if the current user has access to the requested user's data.
// Access is granted if:
// - The user requests their own data
// - There is an accepted or pending connection
// - The requested user is in recommendations for the current user
func userHasAccess(currentUserID, requestedUserID uuid.UUID) (bool, error) {
	logrus.Infof("userHasAccess: checking access from %s to %s", currentUserID, requestedUserID)

	// Allow access to own data (user can always see their own info)
	if currentUserID == requestedUserID {
		logrus.Debugf("userHasAccess: user %s accessing own profile", currentUserID)
		return true, nil
	}

	// Check for accepted connection between users (friendship/connection must be accepted)
	var conn models.Connection
	err := db.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, requestedUserID, requestedUserID, currentUserID, "accepted").
		First(&conn).Error

	if err == nil {
		logrus.Infof("userHasAccess: access granted — accepted connection exists between %s and %s", currentUserID, requestedUserID)
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// If DB error (not just not found), return error
		logrus.Errorf("userHasAccess: DB error while checking accepted connection: %v", err)
		return false, err
	} else {
		logrus.Debugf("userHasAccess: no accepted connection between %s and %s", currentUserID, requestedUserID)
	}

	// Check for pending connection request (user can see if there is a pending request)
	err = db.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, requestedUserID, requestedUserID, currentUserID, "pending").
		First(&conn).Error

	if err == nil {
		logrus.Infof("userHasAccess: access granted — pending connection request between %s and %s", currentUserID, requestedUserID)
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// If DB error (not just not found), return error
		logrus.Errorf("userHasAccess: DB error while checking pending connection: %v", err)
		return false, err
	} else {
		logrus.Debugf("userHasAccess: no pending connection between %s and %s", currentUserID, requestedUserID)
	}

	// As a fallback, check if the requested user is in recommendations for the current user
	// This allows users to see public info of those who are recommended to them
	logrus.Debugf("userHasAccess: checking if %s is in recommendations for %s", requestedUserID, currentUserID)
	recService := services.NewRecommendationService(db, nil)
	recIDs, err := recService.GetRecommendationsForUser(currentUserID, "affinity")
	if err != nil {
		logrus.Errorf("userHasAccess: error fetching recommendations for user %s: %v", currentUserID, err)
	} else {
		found := false
		for _, id := range recIDs {
			if id == requestedUserID {
				found = true
				break
			}
		}
		if found {
			logrus.Infof("userHasAccess: access granted — user %s is in recommendations for %s", requestedUserID, currentUserID)
			return true, nil
		} else {
			logrus.Debugf("userHasAccess: user %s not found in recommendations for %s", requestedUserID, currentUserID)
		}
	}

	// If none of the above, deny access
	logrus.Warnf("userHasAccess: access denied — user %s cannot access data for user %s (no connection, no recommendation)", currentUserID, requestedUserID)
	return false, nil
}

// GetUser handles requests to retrieve a user's public information by ID. Requires access check.
func GetUser(w http.ResponseWriter, r *http.Request) {
	logrus.Error("TRIGGERED GetUser")
	vars := mux.Vars(r)
	requestedID := vars["id"]
	logrus.Infof("!!!!!1Requested user ID: %s", requestedID)

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		// If userID is not found in context, the user is not authenticated
		logrus.Error("GetUser: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Malformed UUID in context
		logrus.Errorf("GetUser: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		// Malformed requested user ID
		logrus.Errorf("GetUser: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	// Check access rights using userHasAccess
	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		// Internal error during access check
		logrus.Errorf("GetUser: error checking access for user %s: %v", requestedUserID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		// Access denied: do not reveal if user exists
		//logrus.Errorf("NOT ALLOWED TO SEE USER %s: %v", requestedUserID, err)
		http.Error(w, "User not found", http.StatusNotFound) //404, The users endpoints return HTTP404 when the id is not found
		logrus.Warnf("GetUser: access denied for user %s to user %s (see userHasAccess logs)", currentUserID, requestedUserID)
		return
	}

	// Fetch user profile from DB (includes Profile via Preload)
	var user models.User
	if err := db.Preload("Profile").First(&user, "id = ?", requestedUserID).Error; err != nil {
		logrus.Errorf("!!!!GetUser: user %s not found: %v", requestedUserID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Respond with basic public info (id, firstName, lastName, photoUrl)
	logrus.Infof("User %s data retrieved by user %s", requestedUserID, currentUserID)
	response := map[string]interface{}{
		"id":        user.ID,
		"firstName": user.Profile.FirstName,
		"lastName":  user.Profile.LastName,
		"photoUrl":  user.Profile.PhotoURL,
	}
	json.NewEncoder(w).Encode(response)
}

// GetUserProfile handles requests to retrieve a user's profile by ID. Requires access check.
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedID := vars["id"]

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		// Not authenticated
		logrus.Error("GetUserProfile: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Malformed UUID
		logrus.Errorf("GetUserProfile: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		// Malformed requested user ID
		logrus.Errorf("GetUserProfile: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	// Check access rights (see userHasAccess)
	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		logrus.Errorf("GetUserProfile: error checking access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Fetch profile from DB
	var profile models.Profile
	if err := db.First(&profile, "user_id = ?", requestedUserID).Error; err != nil {
		// Profile not found
		logrus.Errorf("GetUserProfile: profile for user %s not found: %v", requestedUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Respond with 'about' field only (privacy)
	logrus.Infof("Profile of user %s retrieved by user %s", requestedUserID, currentUserID)
	response := map[string]interface{}{
		"about": profile.About,
	}
	json.NewEncoder(w).Encode(response)
}

// GetUserBio handles requests to retrieve a user's bio by ID. Requires access check.
func GetUserBio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedID := vars["id"]

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetUserBio: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Malformed UUID
		logrus.Errorf("GetUserBio: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		// Malformed requested user ID
		logrus.Errorf("GetUserBio: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	// Check access rights
	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		logrus.Errorf("GetUserBio: error checking access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Fetch bio from DB
	var bio models.Bio
	if err := db.First(&bio, "user_id = ?", requestedUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no bio, return empty struct
			logrus.Warnf("GetUserBio: bio for user %s not found, returning empty", requestedUserID)
			empty := models.Bio{UserID: requestedUserID}
			json.NewEncoder(w).Encode(empty)
			return
		}
		logrus.Errorf("GetUserBio: failed to get bio for user %s: %v", requestedUserID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond with full bio
	logrus.Infof("Bio of user %s retrieved by user %s", requestedUserID, currentUserID)
	json.NewEncoder(w).Encode(bio)
}

// GetCurrentUser returns the current authenticated user's basic information.
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		// Not authenticated
		logrus.Error("GetCurrentUser: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch user and profile from DB
	var user models.User
	if err := db.Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
		// User not found
		logrus.Errorf("GetCurrentUser: user %s not found: %v", userID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Respond with basic info (id, name, photoUrl, email)
	logrus.Infof("Current user %s data retrieved", userID)
	response := map[string]interface{}{
		"id":       user.ID,
		"name":     user.Profile.FirstName + " " + user.Profile.LastName,
		"photoUrl": user.Profile.PhotoURL,
		"email":    user.Email,
	}
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUserProfile returns the current authenticated user's profile.
func GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetCurrentUserProfile: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logrus.Infof("🔍 Extracted userID from context: %s", userID)

	var profile models.Profile

	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Profile not found for current user
			logrus.Warnf("Login: profile not found for user %s", userID)

			http.Error(w, "Login error. Please check the entered data.", http.StatusNotFound)
			return
		}
		// DB error
		logrus.Errorf("Login: DB error fetching profile for user %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("✅ Profile found: %+v", profile)

	logrus.Infof("Profile for current user %s retrieved", userID)
	response := map[string]interface{}{
		"about":     profile.About,
		"firstName": profile.FirstName,
		"lastName":  profile.LastName,
		"photoUrl":  profile.PhotoURL,
		"latitude":  profile.Latitude,
		"longitude": profile.Longitude,
		"city":      profile.City,
	}

	logrus.Infof("📤 Sending profile response: %+v", response)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Error encoding response
		logrus.Errorf("❌ Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetCurrentUserBio returns the current authenticated user's bio information.
func GetCurrentUserBio(w http.ResponseWriter, r *http.Request) {
	userIDstr, ok := r.Context().Value("userID").(string)
	userID, _ := uuid.Parse(userIDstr)

	if !ok {
		// Not authenticated
		logrus.Error("GetCurrentUserBio: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var bio models.Bio

	if err := db.First(&bio, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no bio, return empty struct
			logrus.Warnf("GetUserBio: bio for user %s not found, returning empty", userID)
			empty := models.Bio{UserID: userID}
			json.NewEncoder(w).Encode(empty)
			return
		}
		logrus.Errorf("GetUserBio: failed to get bio for user %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If any required bio fields are empty, do not return the bio
	if bio.Interests == "" ||
		bio.Hobbies == "" ||
		bio.Music == "" ||
		bio.Food == "" ||
		bio.Travel == "" {

		return
	}

	// Respond with full bio
	logrus.Infof("Bio for current user %s retrieved", userID)
	json.NewEncoder(w).Encode(bio)
}
