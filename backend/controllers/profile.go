package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"m/backend/config"
	"m/backend/models"
	"m/backend/utils"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// profileDB используется для операций с профилем и биографией.
var profileDB *gorm.DB

// InitProfileController инициализирует контроллеры профиля, устанавливая подключение к базе данных.
func InitProfileController(db *gorm.DB) {
	profileDB = db
	logrus.Info("Profile controller initialized")
}

// UpdateCurrentUserProfile обновляет информацию "Обо мне" (например, first name, last name, about).
// PUT /me/profile
// Доступ разрешён только для аутентифицированного пользователя.
func UpdateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("UpdateCurrentUserProfile: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("UpdateCurrentUserProfile: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		About     string `json:"about"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("UpdateCurrentUserProfile: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.FirstName, 255); err != nil {
		http.Error(w, "Слишком длинное имя", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.LastName, 255); err != nil {
		http.Error(w, "Слишком длинная фамилия", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.About, 1000); err != nil {
		http.Error(w, "Описание слишком длинное", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	if err := profileDB.First(&profile, "user_id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("UpdateCurrentUserProfile: profile not found for user %s: %v", currentUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	profile.FirstName = reqBody.FirstName
	profile.LastName = reqBody.LastName
	profile.About = reqBody.About

	if err := profileDB.Save(&profile).Error; err != nil {
		logrus.Errorf("UpdateCurrentUserProfile: error updating profile for user %s: %v", currentUserID, err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Profile for user %s updated successfully", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// UpdateCurrentUserBio обновляет биографические данные пользователя.
// PUT /me/bio
// Доступ разрешён только для аутентифицированного пользователя.
func UpdateCurrentUserBio(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("UpdateCurrentUserBio: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("UpdateCurrentUserBio: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Interests string `json:"interests"`
		Hobbies   string `json:"hobbies"`
		Music     string `json:"music"`
		Food      string `json:"food"`
		Travel    string `json:"travel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("UpdateCurrentUserBio: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var bio models.Bio
	if err := profileDB.First(&bio, "user_id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("UpdateCurrentUserBio: bio not found for user %s: %v", currentUserID, err)
		http.Error(w, "Bio not found", http.StatusNotFound)
		return
	}

	bio.Interests = reqBody.Interests
	bio.Hobbies = reqBody.Hobbies
	bio.Music = reqBody.Music
	bio.Food = reqBody.Food
	bio.Travel = reqBody.Travel

	if err := profileDB.Save(&bio).Error; err != nil {
		logrus.Errorf("UpdateCurrentUserBio: error updating bio for user %s: %v", currentUserID, err)
		http.Error(w, "Error updating bio", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Bio for user %s updated successfully", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bio)
}

// UploadUserPhoto обрабатывает загрузку/изменение фотографии профиля.
// POST /me/photo
// Ожидается multipart/form-data с файлом под именем "photo".
// Проверяется формат файла (JPEG/PNG), файл сохраняется в MediaUploadDir,
// и в поле PhotoURL профиля сохраняется относительный путь.
// Доступ разрешён только для аутентифицированного пользователя.
func UploadUserPhoto(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		logrus.Errorf("UploadUserPhoto: error parsing multipart form: %v", err)
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("UploadUserPhoto: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("UploadUserPhoto: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		logrus.Errorf("UploadUserPhoto: error retrieving file: %v", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Считываем первые 512 байт для определения типа контента.
	fileBytes := make([]byte, 512)
	if _, err := file.Read(fileBytes); err != nil {
		logrus.Errorf("UploadUserPhoto: error reading file: %v", err)
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" && fileType != "image/png" {
		logrus.Warnf("UploadUserPhoto: unsupported file type: %s", fileType)
		http.Error(w, "Only JPEG and PNG images are allowed", http.StatusBadRequest)
		return
	}
	// Возвращаем указатель в начало файла.
	file.Seek(0, 0)

	ext := filepath.Ext(fileHeader.Filename)
	newFileName := currentUserID.String() + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ext
	uploadDir := config.AppConfig.MediaUploadDir

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		logrus.Errorf("UploadUserPhoto: error creating upload directory: %v", err)
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}
	fullPath := filepath.Join(uploadDir, newFileName)

	dst, err := os.Create(fullPath)
	if err != nil {
		logrus.Errorf("UploadUserPhoto: error creating file: %v", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logrus.Errorf("UploadUserPhoto: error saving file: %v", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	var profile models.Profile
	if err := profileDB.First(&profile, "user_id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("UploadUserPhoto: profile not found for user %s: %v", currentUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}
	profile.PhotoURL = "/static/images/" + newFileName

	if err := profileDB.Save(&profile).Error; err != nil {
		logrus.Errorf("UploadUserPhoto: error updating profile photo for user %s: %v", currentUserID, err)
		http.Error(w, "Error updating profile photo", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Profile photo for user %s updated successfully", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
