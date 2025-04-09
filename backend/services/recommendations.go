package services

import (
	"errors"
	"math"
	"sort"
	"strings"

	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RecommendationService инкапсулирует бизнес-логику рекомендаций.
type RecommendationService struct {
	DB *gorm.DB
}

// NewRecommendationService создаёт новый экземпляр RecommendationService.
func NewRecommendationService(db *gorm.DB) *RecommendationService {
	logrus.Info("RecommendationService initialized")
	return &RecommendationService{DB: db}
}

// Candidate – вспомогательная структура для расчёта оценки совпадения.
type Candidate struct {
	User  models.User
	Score float64
}

// GetRecommendationsForUser возвращает список до 10 id пользователей, рекомендованных для текущего пользователя.
func (rs *RecommendationService) GetRecommendationsForUser(currentUserID uuid.UUID) ([]uuid.UUID, error) {
	var currentUser models.User
	// Загружаем текущего пользователя с предзагрузкой связанных моделей.
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Preload("Preference").
		First(&currentUser, "id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("GetRecommendationsForUser: ошибка загрузки пользователя %s: %v", currentUserID, err)
		return nil, err
	}

	// Проверка: профиль и биография должны быть заполнены.
	if currentUser.Profile.ID == 0 || currentUser.Bio.ID == 0 {
		logrus.Warnf("GetRecommendationsForUser: профиль или биография пользователя %s не заполнены", currentUserID)
		return nil, errors.New("заполните профиль и биографию для получения рекомендаций")
	}

	// Извлекаем кандидатов – всех пользователей, кроме текущего.
	var candidates []models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Where("id <> ?", currentUserID).
		Find(&candidates).Error; err != nil {
		logrus.Errorf("GetRecommendationsForUser: ошибка загрузки кандидатов: %v", err)
		return nil, err
	}

	var validCandidates []Candidate
	for _, candidate := range candidates {
		// Кандидат должен иметь заполненные профиль и биографию.
		if candidate.Profile.ID == 0 || candidate.Bio.ID == 0 {
			continue
		}

		// Исключаем кандидата, если ранее был отклонён.
		var existingRec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, candidate.ID, "declined").
			First(&existingRec).Error; err == nil {
			continue
		}

		// Фильтрация по местоположению, если у текущего пользователя указан максимальный радиус.
		if currentUser.Preference.MaxRadius > 0 &&
			currentUser.Profile.Latitude != 0 && currentUser.Profile.Longitude != 0 &&
			candidate.Profile.Latitude != 0 && candidate.Profile.Longitude != 0 {
			distance := haversineDistance(
				currentUser.Profile.Latitude,
				currentUser.Profile.Longitude,
				candidate.Profile.Latitude,
				candidate.Profile.Longitude,
			)
			if distance > currentUser.Preference.MaxRadius {
				continue
			}
		}

		// Вычисляем оценку совпадения на основе биографических данных.
		score := computeSimilarityScore(currentUser.Bio, candidate.Bio)
		if score <= 0 {
			continue
		}

		validCandidates = append(validCandidates, Candidate{
			User:  candidate,
			Score: score,
		})
	}
	logrus.Debugf("GetRecommendationsForUser: найдено %d кандидатов для пользователя %s", len(validCandidates), currentUserID)

	// Сортируем кандидатов по убыванию оценки совпадения.
	sort.Slice(validCandidates, func(i, j int) bool {
		return validCandidates[i].Score > validCandidates[j].Score
	})

	// Ограничиваем выборку до 10 кандидатов.
	limit := 10
	if len(validCandidates) < limit {
		limit = len(validCandidates)
	}

	recommendedIDs := make([]uuid.UUID, 0, limit)
	for i := 0; i < limit; i++ {
		recommendedIDs = append(recommendedIDs, validCandidates[i].User.ID)
	}
	logrus.Infof("GetRecommendationsForUser: рекомендации успешно сформированы для пользователя %s", currentUserID)
	return recommendedIDs, nil
}

// computeSimilarityScore вычисляет оценку совпадения между двумя биографиями.
// В данном примере объединяются поля Interests и Hobbies, затем подсчитывается количество общих слов.
func computeSimilarityScore(myBio, candidateBio models.Bio) float64 {
	myText := myBio.Interests + " " + myBio.Hobbies
	candidateText := candidateBio.Interests + " " + candidateBio.Hobbies

	myWords := strings.Fields(strings.ToLower(myText))
	candidateWords := strings.Fields(strings.ToLower(candidateText))

	// Создаем множество слов текущего пользователя.
	mySet := make(map[string]struct{})
	for _, word := range myWords {
		mySet[word] = struct{}{}
	}

	// Считаем общее количество общих слов.
	commonCount := 0
	for _, word := range candidateWords {
		if _, exists := mySet[word]; exists {
			commonCount++
		}
	}
	return float64(commonCount)
}

// haversineDistance вычисляет расстояние между двумя точками (широта, долгота) в километрах.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Радиус Земли в километрах

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
