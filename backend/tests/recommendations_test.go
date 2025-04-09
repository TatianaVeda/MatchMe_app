package tests

import (
	"testing"

	"m/backend/services"
)

func TestGetRecommendationsForUser(t *testing.T) {
	db := SetupTestDB(t)
	// Создаём текущего пользователя с определёнными интересами.
	currentUser := CreateTestUser(db, "current@example.com", "Current", "music sports", "reading")
	// Создаём кандидата.
	candidate1 := CreateTestUser(db, "cand1@example.com", "Candidate1", "music movies", "hiking")
	// Если переменная candidate2 не используется, её можно убрать.
	// candidate2 := CreateTestUser(db, "cand2@example.com", "Candidate2", "cooking travel", "swimming")

	recService := services.NewRecommendationService(db)
	recIDs, err := recService.GetRecommendationsForUser(currentUser.ID)
	if err != nil {
		t.Fatalf("Ошибка получения рекомендаций: %v", err)
	}
	if len(recIDs) == 0 {
		t.Errorf("Ожидались рекомендации, но получено 0")
	}
	// Проверяем, что кандидат1, у которого есть общее слово "music", присутствует в рекомендациях.
	found := false
	for _, id := range recIDs {
		if id == candidate1.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Ожидалось, что кандидат1 будет рекомендован")
	}
}
