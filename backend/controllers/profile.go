package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		About     string  `json:"about"`
		City      string  `json:"city"`
		Latitude  float64 `json:"latitude"` // ← новые поля
		Longitude float64 `json:"longitude"`
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
	if reqBody.City == "" {
		http.Error(w, "Город не может быть пустым", http.StatusBadRequest)
		return
	}

	// var profile models.Profile
	// if err := profileDB.First(&profile, "user_id = ?", currentUserID).Error; err != nil {
	// 	logrus.Errorf("UpdateCurrentUserProfile: profile not found for user %s: %v", currentUserID, err)
	// 	http.Error(w, "Profile not found", http.StatusNotFound)
	// 	return
	// }

	var profile models.Profile
	err = profileDB.First(&profile, "user_id = ?", currentUserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new profile if not found
			profile = models.Profile{
				UserID:    currentUserID,
				FirstName: reqBody.FirstName,
				LastName:  reqBody.LastName,
				About:     reqBody.About,
				City:      reqBody.City,
				Latitude:  reqBody.Latitude,
				Longitude: reqBody.Longitude,
			}
			if err := profileDB.Create(&profile).Error; err != nil {
				logrus.Errorf("UpdateCurrentUserProfile: error creating profile for user %s: %v", currentUserID, err)
				http.Error(w, "Error creating profile", http.StatusInternalServerError)
				return
			}
			logrus.Infof("Profile for user %s created successfully", currentUserID)
		} else {
			logrus.Errorf("UpdateCurrentUserProfile: error querying profile for user %s: %v", currentUserID, err)
			http.Error(w, "Error reading profile", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing profile
		profile.FirstName = reqBody.FirstName
		profile.LastName = reqBody.LastName
		profile.About = reqBody.About
		profile.City = reqBody.City
		if reqBody.Latitude != 0 || reqBody.Longitude != 0 {
			profile.Latitude = reqBody.Latitude
			profile.Longitude = reqBody.Longitude
		}

		if err := profileDB.Save(&profile).Error; err != nil {
			logrus.Errorf("UpdateCurrentUserProfile: error updating profile for user %s: %v", currentUserID, err)
			http.Error(w, "Error updating profile", http.StatusInternalServerError)
			return
		}
		logrus.Infof("Profile for user %s updated successfully", currentUserID)
	}

	// profile.FirstName = reqBody.FirstName
	// profile.LastName = reqBody.LastName
	// profile.About = reqBody.About
	// profile.City = reqBody.City

	// // // Если пришли геокоординаты — сохраняем
	// if reqBody.Latitude != 0 || reqBody.Longitude != 0 {
	// 	profile.Latitude = reqBody.Latitude
	// 	profile.Longitude = reqBody.Longitude
	// }

	// if err := profileDB.Save(&profile).Error; err != nil {
	// 	logrus.Errorf("UpdateCurrentUserProfile: error updating profile for user %s: %v", currentUserID, err)
	// 	http.Error(w, "Error updating profile", http.StatusInternalServerError)
	// 	return
	// }

	// Save earth_loc in PostgreSQL
	if profile.Latitude != 0 && profile.Longitude != 0 {
		if err := profileDB.Exec(`
		UPDATE profiles
		SET earth_loc = ll_to_earth(?, ?)
		WHERE user_id = ?
	`, profile.Latitude, profile.Longitude, profile.UserID).Error; err != nil {
			logrus.Errorf("Error setting earth_loc for user %s: %v", profile.UserID, err)
			http.Error(w, "Failed to update earth location", http.StatusInternalServerError)
			return
		}
	}

	logrus.Infof("Profile for user %s updated successfully", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// PUT /me/location
func UpdateCurrentUserLocation(w http.ResponseWriter, r *http.Request) {
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

	var reqBody struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	err = profileDB.First(&profile, "user_id = ?", currentUserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new profile with coordinates
			profile = models.Profile{
				UserID:    currentUserID,
				Latitude:  reqBody.Latitude,
				Longitude: reqBody.Longitude,
			}
			if err := profileDB.Create(&profile).Error; err != nil {
				http.Error(w, "Error creating profile with location", http.StatusInternalServerError)
				return
			}
			logrus.Infof("New profile created with coordinates for user %s", currentUserID)
		} else {
			http.Error(w, "Error retrieving profile", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing coordinates
		profile.Latitude = reqBody.Latitude
		profile.Longitude = reqBody.Longitude
		if err := profileDB.Save(&profile).Error; err != nil {
			http.Error(w, "Error updating location", http.StatusInternalServerError)
			return
		}
		logrus.Infof("Coordinates for user %s updated successfully", currentUserID)
	}

	// 🌍 Update the earth_loc column using ll_to_earth
	if reqBody.Latitude != 0 && reqBody.Longitude != 0 {
		if err := profileDB.Exec(`
			UPDATE profiles
			SET earth_loc = ll_to_earth(?, ?)
			WHERE user_id = ?
		`, reqBody.Latitude, reqBody.Longitude, currentUserID).Error; err != nil {
			logrus.Errorf("Error updating earth_loc for user %s: %v", currentUserID, err)
			http.Error(w, "Error updating earth location", http.StatusInternalServerError)
			return
		}
		logrus.Infof("earth_loc updated for user %s", currentUserID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// func UpdateCurrentUserLocation(w http.ResponseWriter, r *http.Request) {
// 	userIDStr, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUserID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		http.Error(w, "Invalid userID", http.StatusBadRequest)
// 		return
// 	}

// 	var reqBody struct {
// 		Latitude  float64 `json:"latitude"`
// 		Longitude float64 `json:"longitude"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	var profile models.Profile
// 	err = profileDB.First(&profile, "user_id = ?", currentUserID).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			// Create new profile with coordinates
// 			profile = models.Profile{
// 				UserID:    currentUserID,
// 				Latitude:  reqBody.Latitude,
// 				Longitude: reqBody.Longitude,
// 			}
// 			if err := profileDB.Create(&profile).Error; err != nil {
// 				http.Error(w, "Error creating profile with location", http.StatusInternalServerError)
// 				return
// 			}
// 			logrus.Infof("New profile created with coordinates for user %s", currentUserID)
// 		} else {
// 			http.Error(w, "Error retrieving profile", http.StatusInternalServerError)
// 			return
// 		}
// 	} else {
// 		// Update existing coordinates
// 		profile.Latitude = reqBody.Latitude
// 		profile.Longitude = reqBody.Longitude
// 		if err := profileDB.Save(&profile).Error; err != nil {
// 			http.Error(w, "Error updating location", http.StatusInternalServerError)
// 			return
// 		}
// 		logrus.Infof("Coordinates for user %s updated successfully", currentUserID)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(profile)
// }

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
		Interests         string `json:"interests"`
		Hobbies           string `json:"hobbies"`
		Music             string `json:"music"`
		Food              string `json:"food"`
		Travel            string `json:"travel"`
		LookingFor        string `json:"lookingFor"`
		PriorityInterests bool   `json:"priorityInterests"`
		PriorityHobbies   bool   `json:"priorityHobbies"`
		PriorityMusic     bool   `json:"priorityMusic"`
		PriorityFood      bool   `json:"priorityFood"`
		PriorityTravel    bool   `json:"priorityTravel"`
	}

	logrus.Infof("UpdateCurrentUserBio: reqBody: %+v", reqBody)

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
	bio.LookingFor = reqBody.LookingFor

	if err := profileDB.Save(&bio).Error; err != nil {
		logrus.Errorf("UpdateCurrentUserBio: error updating bio for user %s: %v", currentUserID, err)
		http.Error(w, "Error updating bio", http.StatusInternalServerError)
		return
	}

	// Дополнительно: сохраняем приоритетные флаги в Preference
	var pref models.Preference
	if err := profileDB.
		Where("user_id = ?", currentUserID).
		First(&pref).Error; err != nil {
		// Если записи нет — создаём новую
		pref = models.Preference{
			UserID:            currentUserID,
			PriorityInterests: reqBody.PriorityInterests,
			PriorityHobbies:   reqBody.PriorityHobbies,
			PriorityMusic:     reqBody.PriorityMusic,
			PriorityFood:      reqBody.PriorityFood,
			PriorityTravel:    reqBody.PriorityTravel,
		}
		if err := profileDB.Create(&pref).Error; err != nil {
			logrus.Errorf("UpdateCurrentUserBio: error creating preferences for user %s: %v", currentUserID, err)
		}
	} else {
		// Обновляем существующие флаги
		pref.PriorityInterests = reqBody.PriorityInterests
		pref.PriorityHobbies = reqBody.PriorityHobbies
		pref.PriorityMusic = reqBody.PriorityMusic
		pref.PriorityFood = reqBody.PriorityFood
		pref.PriorityTravel = reqBody.PriorityTravel
		if err := profileDB.Save(&pref).Error; err != nil {
			logrus.Errorf("UpdateCurrentUserBio: error updating preferences for user %s: %v", currentUserID, err)
		}
	}

	logrus.Infof("Bio for user %s updated successfully", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(bio)
	// Возвращаем обновлённую Bio + Preferences вместе в одном ответе
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bio":         bio,
		"preferences": pref,
	})
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

// DeleteUserPhoto – endpoint для удаления (сброса) фотографии профиля
func DeleteUserPhoto(w http.ResponseWriter, r *http.Request) {
	// Получаем ID текущего пользователя из контекста (установленный в AuthMiddleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("DeleteUserPhoto: userID не найден в контексте")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("DeleteUserPhoto: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	// Загружаем профиль пользователя
	var profile models.Profile
	if err := profileDB.First(&profile, "user_id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("DeleteUserPhoto: профиль для пользователя %s не найден: %v", currentUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Значение по умолчанию для фото (установленное при инициализации профиля)
	defaultPhotoURL := "/static/images/default.png"

	// Если фото не является значением по умолчанию, попробуем удалить физический файл
	if profile.PhotoURL != "" && profile.PhotoURL != defaultPhotoURL {
		uploadDir := config.AppConfig.MediaUploadDir // например, "./static/images"
		// Предполагаем, что profile.PhotoURL имеет вид "/static/images/имя_файла.ext"
		fileName := strings.TrimPrefix(profile.PhotoURL, "/static/images/")
		filePath := filepath.Join(uploadDir, fileName)
		if err := os.Remove(filePath); err != nil {
			// Если файла нет или возникла другая ошибка, можно залогировать предупреждение, но не прерывать выполнение
			logrus.Warnf("DeleteUserPhoto: ошибка удаления файла %s: %v", filePath, err)
		} else {
			logrus.Infof("DeleteUserPhoto: файл %s успешно удалён", filePath)
		}
	}

	// Сброс значения photo_url до значения по умолчанию
	profile.PhotoURL = defaultPhotoURL
	if err := profileDB.Save(&profile).Error; err != nil {
		logrus.Errorf("DeleteUserPhoto: ошибка обновления профиля для пользователя %s: %v", currentUserID, err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	logrus.Infof("DeleteUserPhoto: фото профиля для пользователя %s сброшено на значение по умолчанию", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
