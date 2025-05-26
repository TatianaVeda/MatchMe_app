package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"m/backend/config"
	"m/backend/models"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var fixturesDB *gorm.DB

func InitFixturesController(db *gorm.DB) {
	fixturesDB = db
	logrus.Info("Fixtures controller initialized")
}
func ResetFixtures(w http.ResponseWriter, r *http.Request) {
	modelsToDrop := []interface{}{
		&models.User{}, &models.Profile{}, &models.Bio{}, &models.Preference{},
		&models.Recommendation{}, &models.Connection{}, &models.Chat{},
		&models.Message{}, &models.FakeUser{},
	}
	if err := fixturesDB.Migrator().DropTable(modelsToDrop...); err != nil {
		logrus.Errorf("ResetFixtures: error dropping tables: %v", err)
		http.Error(w, fmt.Sprintf("Error dropping tables: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.Info("ResetFixtures: tables dropped successfully")
	if err := models.Migrate(fixturesDB); err != nil {
		logrus.Errorf("ResetFixtures: database migration error: %v", err)
		http.Error(w, fmt.Sprintf("Database migration error: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.Info("ResetFixtures: database migration completed successfully")
	adminUUID, _ := uuid.Parse(config.AdminID)
	var existing models.User
	err := fixturesDB.First(&existing, "id = ?", adminUUID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, _ := models.HashPassword(config.AdminPassword)
		admin := models.User{
			ID:           adminUUID,
			Email:        config.AdminEmail,
			PasswordHash: hash,
		}
		if err := fixturesDB.Create(&admin).Error; err != nil {
			logrus.Errorf("ResetFixtures: failed to create admin: %v", err)
		} else {
			logrus.Infof("ResetFixtures: admin %s created (ID=%s)", config.AdminEmail, config.AdminID)
		}
	} else {
		logrus.Info("ResetFixtures: admin already exists, creation skipped")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Database reset, administrator saved",
	})
}

func GenerateFixtures(w http.ResponseWriter, r *http.Request) {
	numUsers := 100
	if param := r.URL.Query().Get("num"); param != "" {
		if n, err := strconv.Atoi(param); err == nil && n > 0 {
			numUsers = n
		} else {
			logrus.Warnf("GenerateFixtures: invalid value for num parameter (%s), using %d", param, numUsers)
		}
	}
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= numUsers; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		password := "password123"
		user, err := models.CreateUser(fixturesDB, email, password)
		if err != nil {
			logrus.Warnf("GenerateFixtures: error creating user %s: %v", email, err)
			continue
		}

		latitude, longitude, city := randomLocationWithCity()
		profile := models.Profile{
			UserID:    user.ID,
			FirstName: randomFirstName(),
			LastName:  randomLastName(),
			About:     "Test user for demonstration purposes.",
			PhotoURL:  "/static/images/default.png",
			Online:    false,
			Latitude:  latitude,
			Longitude: longitude,
			City:      city,
		}
		if err := fixturesDB.Save(&profile).Error; err != nil {
			logrus.Warnf("GenerateFixtures: error saving profile %s: %v", email, err)
		}

		bioUpdates := map[string]interface{}{
			"interests": randomInterests(),
			"hobbies":   randomHobbies(),
			"music":     randomMusic(),
			"food":      randomFood(),
			"travel":    randomTravel(),
		}
		if err := fixturesDB.Model(&models.Bio{}).
			Where("user_id = ?", user.ID).
			Updates(bioUpdates).Error; err != nil {
			logrus.Warnf("GenerateFixtures: error updating bio %s: %v", email, err)
		}

		if err := fixturesDB.Exec(`
			UPDATE profiles
			SET earth_loc = ll_to_earth(?, ?)
			WHERE user_id = ?`,
			profile.Latitude,
			profile.Longitude,
			profile.UserID,
		).Error; err != nil {
			logrus.Warnf("Error updating earth_loc for %s: %v", email, err)
		}

		logrus.Debugf("GenerateFixtures: user %s created", email)
	}
	logrus.Infof("GenerateFixtures: %d fake users created", numUsers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("%d fake users generated", numUsers),
	})
}

func randomFirstName() string {
	arr := []string{"Anna", "Jan", "Maria", "Alex", "Olga", "Danny", "Eelena", "Sergio", "Natalie", "Michael"}
	return arr[rand.Intn(len(arr))]
}
func randomLastName() string {
	arr := []string{"Agricola", "Petrov", "Sibelius", "Kuznets", "Saminen", "Gogol", "Novi", "Feducci", "Savolainen", "Gagarin"}
	return arr[rand.Intn(len(arr))]
}
func randomInterests() string {
	arr := []string{"movies", "sports", "music", "technology", "art", "travel", "literature", "photography"}
	return arr[rand.Intn(len(arr))]
}
func randomHobbies() string {
	arr := []string{"reading", "running", "drawing", "games", "cooking", "gardening", "swimming", "travel"}
	return arr[rand.Intn(len(arr))]
}
func randomMusic() string {
	arr := []string{"rock", "jazz", "classical", "pop", "hip-hop", "electronic", "blues"}
	return arr[rand.Intn(len(arr))]
}
func randomFood() string {
	arr := []string{"italian", "asian", "russian", "french", "mexican", "japanese"}
	return arr[rand.Intn(len(arr))]
}
func randomTravel() string {
	arr := []string{"beach vacation", "mountains", "cultural tours", "expeditions", "city trip"}
	return arr[rand.Intn(len(arr))]
}
func randomLatitude() float64 {
	return 41 + rand.Float64()*(82-41)
}
func randomLongitude() float64 {
	return 19 + rand.Float64()*(169-19)
}
func randomLocationWithCity() (float64, float64, string) {
	if rand.Float64() < 0.8 {
		c := finnishCities[rand.Intn(len(finnishCities))]
		jitterLat := c.Latitude + (rand.Float64()-0.5)*0.02
		jitterLon := c.Longitude + (rand.Float64()-0.5)*0.02
		return jitterLat, jitterLon, c.Name
	}
	lat := randomLatitude()
	lon := randomLongitude()
	return lat, lon, "Unknown"
}
