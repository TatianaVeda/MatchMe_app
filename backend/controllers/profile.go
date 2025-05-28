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

var profileDB *gorm.DB

// InitProfileController initializes the profile controller with the database connection.
func InitProfileController(db *gorm.DB) {
	profileDB = db
	logrus.Info("Profile controller initialized")
}

// UpdateCurrentUserProfile updates the current user's profile information.
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
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("UpdateCurrentUserProfile: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.FirstName, 255); err != nil {
		http.Error(w, "First name is too long", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.LastName, 255); err != nil {
		http.Error(w, "Last name is too long, max 255 chars", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateStringLength(reqBody.About, 1000); err != nil {
		http.Error(w, "Description is too long, max 1000 chars", http.StatusBadRequest)
		return
	}
	if reqBody.City == "" {
		http.Error(w, "City cannot be empty", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	err = profileDB.First(&profile, "user_id = ?", currentUserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

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
		profile.Latitude = reqBody.Latitude
		profile.Longitude = reqBody.Longitude
		if err := profileDB.Save(&profile).Error; err != nil {
			http.Error(w, "Error updating location", http.StatusInternalServerError)
			return
		}
		logrus.Infof("Coordinates for user %s updated successfully", currentUserID)
	}

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

// UpdateCurrentUserBio updates the current user's bio information.
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

	var pref models.Preference
	if err := profileDB.
		Where("user_id = ?", currentUserID).
		First(&pref).Error; err != nil {
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bio":         bio,
		"preferences": pref,
	})
}

// UploadUserPhoto handles uploading a new profile photo for the current user.
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

// DeleteUserPhoto deletes the current user's profile photo and resets it to default.
func DeleteUserPhoto(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("DeleteUserPhoto: userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("DeleteUserPhoto: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	if err := profileDB.First(&profile, "user_id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("DeleteUserPhoto: profile not found for user %s: %v", currentUserID, err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	defaultPhotoURL := "/static/images/default.png"

	if profile.PhotoURL != "" && profile.PhotoURL != defaultPhotoURL {
		uploadDir := config.AppConfig.MediaUploadDir
		fileName := strings.TrimPrefix(profile.PhotoURL, "/static/images/")
		filePath := filepath.Join(uploadDir, fileName)
		if err := os.Remove(filePath); err != nil {
			logrus.Warnf("DeleteUserPhoto: error deleting file %s: %v", filePath, err)
		} else {
			logrus.Infof("DeleteUserPhoto: file %s deleted successfully", filePath)
		}
	}

	profile.PhotoURL = defaultPhotoURL
	if err := profileDB.Save(&profile).Error; err != nil {
		logrus.Errorf("DeleteUserPhoto: error updating profile for user %s: %v", currentUserID, err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	logrus.Infof("DeleteUserPhoto: profile photo for user %s reset to default", currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
