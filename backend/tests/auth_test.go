package tests

import (
	"testing"

	"m/backend/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "password123"
	hash, err := models.HashPassword(password)
	if err != nil {
		t.Fatalf("Ошибка при хэшировании пароля: %v", err)
	}
	if !models.CheckPasswordHash(password, hash) {
		t.Errorf("Пароль должен соответствовать хэшу")
	}
	if models.CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("Неверный пароль не должен соответствовать хэшу")
	}
}
