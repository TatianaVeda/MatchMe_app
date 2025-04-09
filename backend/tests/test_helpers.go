package tests

import (
	"m/backend/models"
	"testing"
	"time"

	"github.com/google/uuid"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB инициализирует in‑memory базу данных для тестов.
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Ошибка открытия тестовой БД: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Bio{},
		&models.Preference{},
		&models.Recommendation{},
		&models.Connection{},
		&models.Chat{},
		&models.Message{},
	); err != nil {
		t.Fatalf("Ошибка миграции тестовой БД: %v", err)
	}
	return db
}

// createTestUser создаёт тестового пользователя с заданными параметрами.
func CreateTestUser(db *gorm.DB, email, firstName, interests, hobbies string) models.User {
	user := models.User{
		// Генерируем новый UUID.
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "dummyHash", // Для теста достаточно заглушки
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user.Profile = models.Profile{
		UserID:    user.ID,
		FirstName: firstName,
		LastName:  "Test",
	}
	user.Bio = models.Bio{
		UserID:    user.ID,
		Interests: interests,
		Hobbies:   hobbies,
	}
	user.Preference = models.Preference{
		UserID:    user.ID,
		MaxRadius: 100,
	}
	db.Create(&user)
	return user
}
