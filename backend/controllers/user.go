package controllers

import (
	"encoding/json"
	"net/http"

	"m/backend/models"
	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Глобальная переменная для подключения к базе данных.
var db *gorm.DB

// InitUserController инициализирует контроллер с подключением к БД.
func InitUserController(database *gorm.DB) {
	db = database
	logrus.Info("User controller initialized")
}

// userHasAccess проверяет, имеет ли текущий пользователь (currentUserID)
// право видеть данные пользователя с идентификатором requestedUserID.
// Доступ разрешается, если:
// - запрошен собственный профиль,
// - существует установленное соединение (status = "accepted"),
// - существует ожидающий запрос (status = "pending"),
// - запрошенный пользователь входит в список рекомендаций текущего пользователя.
func userHasAccess(currentUserID, requestedUserID uuid.UUID) (bool, error) {
	// Если запрошен собственный профиль.
	if currentUserID == requestedUserID {
		logrus.Debugf("User %s accessing own profile", currentUserID)
		return true, nil
	}

	// Проверяем наличие установленного соединения.
	var conn models.Connection
	err := db.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, requestedUserID, requestedUserID, currentUserID, "accepted").
		First(&conn).Error
	if err == nil {
		logrus.Debugf("Connection exists between %s and %s (accepted)", currentUserID, requestedUserID)
		return true, nil
	}

	// Проверяем наличие ожидающего запроса.
	err = db.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, requestedUserID, requestedUserID, currentUserID, "pending").
		First(&conn).Error
	if err == nil {
		logrus.Debugf("Connection request exists between %s and %s (pending)", currentUserID, requestedUserID)
		return true, nil
	}

	// Проверяем, входит ли запрошенный пользователь в рекомендации текущего пользователя.
	recService := services.NewRecommendationService(db)
	recIDs, err := recService.GetRecommendationsForUser(currentUserID)
	if err == nil {
		for _, id := range recIDs {
			if id == requestedUserID {
				logrus.Debugf("User %s is in recommendations for %s", requestedUserID, currentUserID)
				return true, nil
			}
		}
	} else {
		logrus.Errorf("Error fetching recommendations for user %s: %v", currentUserID, err)
	}

	// Если ни одно условие не выполнено — доступ запрещён.
	logrus.Warnf("Access denied: user %s cannot access data for user %s", currentUserID, requestedUserID)
	return false, nil
}

// GET /users/{id}
// Возвращает id, имя (составленное из firstName и lastName) и ссылку на фотографию профиля.
// Если текущий пользователь не имеет права на просмотр данных запрошенного профиля,
// возвращается HTTP404.
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedID := vars["id"]

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetUser: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetUser: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		logrus.Errorf("GetUser: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		logrus.Errorf("GetUser: error checking access for user %s: %v", requestedUserID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user models.User
	if err := db.Preload("Profile").First(&user, "id = ?", requestedUserID).Error; err != nil {
		logrus.Errorf("GetUser: user %s not found: %v", requestedUserID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logrus.Infof("User %s data retrieved by user %s", requestedUserID, currentUserID)
	response := map[string]interface{}{
		"id":        user.ID,
		"name":      user.Profile.FirstName + " " + user.Profile.LastName,
		"photo_url": user.Profile.PhotoURL,
	}
	json.NewEncoder(w).Encode(response)
}

// GET /users/{id}/profile
// Возвращает информацию "Обо мне" из профиля пользователя.
// Если текущий пользователь не имеет права на просмотр — возвращается HTTP404.
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedID := vars["id"]

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetUserProfile: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetUserProfile: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		logrus.Errorf("GetUserProfile: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		logrus.Errorf("GetUserProfile: error checking access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	var profile models.Profile
	if err := db.First(&profile, "user_id = ?", requestedUserID).Error; err != nil {
		logrus.Errorf("GetUserProfile: profile for user %s not found: %v", requestedUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	logrus.Infof("Profile of user %s retrieved by user %s", requestedUserID, currentUserID)
	response := map[string]interface{}{
		"about": profile.About,
	}
	json.NewEncoder(w).Encode(response)
}

// GET /users/{id}/bio
// Возвращает биографические данные пользователя (данные для рекомендаций).
// Если текущий пользователь не имеет права на просмотр — возвращается HTTP404.
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
		logrus.Errorf("GetUserBio: invalid current user ID: %v", err)
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}
	requestedUserID, err := uuid.Parse(requestedID)
	if err != nil {
		logrus.Errorf("GetUserBio: invalid requested user ID: %v", err)
		http.Error(w, "Invalid requested user ID", http.StatusBadRequest)
		return
	}

	allowed, err := userHasAccess(currentUserID, requestedUserID)
	if err != nil {
		logrus.Errorf("GetUserBio: error checking access: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Bio not found", http.StatusNotFound)
		return
	}

	var bio models.Bio
	if err := db.First(&bio, "user_id = ?", requestedUserID).Error; err != nil {
		logrus.Errorf("GetUserBio: bio for user %s not found: %v", requestedUserID, err)
		http.Error(w, "Bio not found", http.StatusNotFound)
		return
	}

	logrus.Infof("Bio of user %s retrieved by user %s", requestedUserID, currentUserID)
	json.NewEncoder(w).Encode(bio)
}

// GET /me
// Возвращает данные аутентифицированного пользователя: id, имя и фото.
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetCurrentUser: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := db.Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
		logrus.Errorf("GetCurrentUser: user %s not found: %v", userID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logrus.Infof("Current user %s data retrieved", userID)
	response := map[string]interface{}{
		"id":        user.ID,
		"name":      user.Profile.FirstName + " " + user.Profile.LastName,
		"photo_url": user.Profile.PhotoURL,
		"email":     user.Email, // email включается в ответ для аутентифицированного пользователя
	}
	json.NewEncoder(w).Encode(response)
}

// GET /me/profile
// Возвращает информацию "Обо мне" для аутентифицированного пользователя.
func GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetCurrentUserProfile: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var profile models.Profile
	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		logrus.Errorf("GetCurrentUserProfile: profile for user %s not found: %v", userID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	logrus.Infof("Profile for current user %s retrieved", userID)
	response := map[string]interface{}{
		"about": profile.About,
	}
	json.NewEncoder(w).Encode(response)
}

// GET /me/bio
// Возвращает биографические данные аутентифицированного пользователя.
func GetCurrentUserBio(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetCurrentUserBio: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var bio models.Bio
	if err := db.First(&bio, "user_id = ?", userID).Error; err != nil {
		logrus.Errorf("GetCurrentUserBio: bio for user %s not found: %v", userID, err)
		http.Error(w, "Bio not found", http.StatusNotFound)
		return
	}

	logrus.Infof("Bio for current user %s retrieved", userID)
	json.NewEncoder(w).Encode(bio)
}
