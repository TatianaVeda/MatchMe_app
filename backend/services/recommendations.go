// package services

// import (
// 	"errors"
// 	"math"
// 	"sort"
// 	"strings"

// 	"m/backend/models"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/gorm"
// )

// // RecommendationService инкапсулирует бизнес-логику рекомендаций.
// type RecommendationService struct {
// 	DB *gorm.DB
// }

// // NewRecommendationService создаёт новый экземпляр RecommendationService.
// func NewRecommendationService(db *gorm.DB) *RecommendationService {
// 	logrus.Info("RecommendationService initialized")
// 	return &RecommendationService{DB: db}
// }

// // Candidate – вспомогательная структура для расчёта оценки совпадения.
// type Candidate struct {
// 	User  models.User
// 	Score float64
// }

// // GetRecommendationsForUser возвращает список до 10 id пользователей, рекомендованных для текущего пользователя.
// func (rs *RecommendationService) GetRecommendationsForUser(currentUserID uuid.UUID) ([]uuid.UUID, error) {
// 	var currentUser models.User
// 	// Загружаем текущего пользователя с предзагрузкой связанных моделей.
// 	if err := rs.DB.
// 		Preload("Profile").
// 		Preload("Bio").
// 		Preload("Preference").
// 		First(&currentUser, "id = ?", currentUserID).Error; err != nil {
// 		logrus.Errorf("GetRecommendationsForUser: ошибка загрузки пользователя %s: %v", currentUserID, err)
// 		return nil, err
// 	}

// 	// Проверка: профиль и биография должны быть заполнены.
// 	if currentUser.Profile.ID == 0 || currentUser.Bio.ID == 0 {
// 		logrus.Warnf("GetRecommendationsForUser: профиль или биография пользователя %s не заполнены", currentUserID)
// 		return nil, errors.New("заполните профиль и биографию для получения рекомендаций")
// 	}

// 	if currentUser.Profile.FirstName == "" || currentUser.Profile.LastName == "" {
// 		logrus.Warnf("GetRecommendationsForUser: профиль пользователя %s не заполнен полностью", currentUserID)
// 		return nil, errors.New("пожалуйста, заполните ваш профиль (имя и фамилия)")
// 	}

// 	if currentUser.Bio.Interests == "" ||
// 		currentUser.Bio.Hobbies == "" ||
// 		currentUser.Bio.Music == "" ||
// 		currentUser.Bio.Food == "" ||
// 		currentUser.Bio.Travel == "" {
// 		logrus.Warnf("GetRecommendationsForUser: биография пользователя %s не заполнена полностью", currentUserID)
// 		return nil, errors.New("пожалуйста, заполните вашу биографию: " +
// 			"интересы, хобби, музыка, еда и путешествия")
// 	}

// 	// Извлекаем кандидатов – всех пользователей, кроме текущего.
// 	var candidates []models.User
// 	if err := rs.DB.
// 		Preload("Profile").
// 		Preload("Bio").
// 		Where("id <> ?", currentUserID).
// 		Find(&candidates).Error; err != nil {
// 		logrus.Errorf("GetRecommendationsForUser: ошибка загрузки кандидатов: %v", err)
// 		return nil, err
// 	}

// 	var validCandidates []Candidate
// 	for _, candidate := range candidates {
// 		// Кандидат должен иметь заполненные профиль и биографию.
// 		if candidate.Profile.ID == 0 || candidate.Bio.ID == 0 {
// 			continue
// 		}

// 		// Исключаем кандидата, если ранее был отклонён.
// 		var existingRec models.Recommendation
// 		if err := rs.DB.
// 			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, candidate.ID, "declined").
// 			First(&existingRec).Error; err == nil {
// 			continue
// 		}

// 		// Фильтрация по местоположению, если у текущего пользователя указан максимальный радиус.
// 		if currentUser.Preference.MaxRadius > 0 &&
// 			currentUser.Profile.Latitude != 0 && currentUser.Profile.Longitude != 0 &&
// 			candidate.Profile.Latitude != 0 && candidate.Profile.Longitude != 0 {
// 			distance := haversineDistance(
// 				currentUser.Profile.Latitude,
// 				currentUser.Profile.Longitude,
// 				candidate.Profile.Latitude,
// 				candidate.Profile.Longitude,
// 			)
// 			if distance > currentUser.Preference.MaxRadius {
// 				continue
// 			}
// 		}

// 		// Вычисляем оценку совпадения на основе биографических данных.
// 		score := computeSimilarityScore(currentUser.Bio, candidate.Bio)
// 		if score <= 0 {
// 			continue
// 		}

// 		validCandidates = append(validCandidates, Candidate{
// 			User:  candidate,
// 			Score: score,
// 		})
// 	}
// 	logrus.Debugf("GetRecommendationsForUser: найдено %d кандидатов для пользователя %s", len(validCandidates), currentUserID)

// 	// Сортируем кандидатов по убыванию оценки совпадения.
// 	sort.Slice(validCandidates, func(i, j int) bool {
// 		return validCandidates[i].Score > validCandidates[j].Score
// 	})

// 	// Ограничиваем выборку до 10 кандидатов.
// 	limit := 10
// 	if len(validCandidates) < limit {
// 		limit = len(validCandidates)
// 	}

// 	recommendedIDs := make([]uuid.UUID, 0, limit)
// 	for i := 0; i < limit; i++ {
// 		recommendedIDs = append(recommendedIDs, validCandidates[i].User.ID)
// 	}
// 	logrus.Infof("GetRecommendationsForUser: рекомендации успешно сформированы для пользователя %s", currentUserID)
// 	return recommendedIDs, nil
// }

// // computeSimilarityScore вычисляет оценку совпадения между двумя биографиями.
// // В данном примере объединяются поля Interests и Hobbies, затем подсчитывается количество общих слов.
// func computeSimilarityScore(myBio, candidateBio models.Bio) float64 {
// 	myText := myBio.Interests + " " + myBio.Hobbies
// 	candidateText := candidateBio.Interests + " " + candidateBio.Hobbies

// 	myWords := strings.Fields(strings.ToLower(myText))
// 	candidateWords := strings.Fields(strings.ToLower(candidateText))

// 	// Создаем множество слов текущего пользователя.
// 	mySet := make(map[string]struct{})
// 	for _, word := range myWords {
// 		mySet[word] = struct{}{}
// 	}

// 	// Считаем общее количество общих слов.
// 	commonCount := 0
// 	for _, word := range candidateWords {
// 		if _, exists := mySet[word]; exists {
// 			commonCount++
// 		}
// 	}
// 	return float64(commonCount)
// }

// // haversineDistance вычисляет расстояние между двумя точками (широта, долгота) в километрах.
// func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
// 	const earthRadius = 6371 // Радиус Земли в километрах

// 	dLat := degreesToRadians(lat2 - lat1)
// 	dLon := degreesToRadians(lon2 - lon1)

// 	lat1Rad := degreesToRadians(lat1)
// 	lat2Rad := degreesToRadians(lat2)

// 	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
// 		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
// 	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
// 	return earthRadius * c
// }

// func degreesToRadians(deg float64) float64 {
// 	return deg * math.Pi / 180
// }

// package services

// import (
// 	"errors"
// 	"math"

// 	//"math"

// 	//"reflect"
// 	"sort"
// 	"strings"

// 	"m/backend/models"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/gorm"
// )

// // anyTokenMatch возвращает true, если хотя бы один токен из a содержится среди токенов b.
// func anyTokenMatch(a, b string) bool {
// 	// Разбиваем строки на слова
// 	tokensA := strings.Fields(strings.ToLower(a))
// 	tokensB := strings.Fields(strings.ToLower(b))

// 	// Собираем второй набор в map для быстрой проверки
// 	setB := make(map[string]struct{}, len(tokensB))
// 	for _, t := range tokensB {
// 		setB[t] = struct{}{}
// 	}

// 	// Проверяем, есть ли хоть одно слово из A в B
// 	for _, t := range tokensA {
// 		if _, ok := setB[t]; ok {
// 			return true
// 		}
// 	}
// 	return false
// }

// // FieldConfig описывает одно поле биографии и его вес.
// type FieldConfig struct {
// 	Name      string                    // Человеко-читаемое имя поля
// 	Weight    float64                   // Вес поля
// 	Extractor func(b models.Bio) string // Функция для получения текста из Bio
// }

// // RecommendationService инкапсулирует бизнес-логику рекомендаций.
// type RecommendationService struct {
// 	DB           *gorm.DB
// 	FieldConfigs []FieldConfig
// 	Mode         string
// }

// // NewRecommendationService создаёт новый экземпляр RecommendationService.
// // fieldConfigs можно загрузить из YAML/DB или передать nil для использования значений по умолчанию.
// func NewRecommendationService(db *gorm.DB, fieldConfigs []FieldConfig) *RecommendationService {
// 	rs := &RecommendationService{
// 		DB:   db,
// 		Mode: "affinity",
// 	}
// 	if fieldConfigs == nil || len(fieldConfigs) == 0 {
// 		fieldConfigs = []FieldConfig{
// 			{Name: "Interests", Weight: 1.0, Extractor: func(b models.Bio) string { return b.Interests }},
// 			{Name: "Hobbies", Weight: 1.0, Extractor: func(b models.Bio) string { return b.Hobbies }},
// 			{Name: "Music", Weight: 0.8, Extractor: func(b models.Bio) string { return b.Music }},
// 			{Name: "Food", Weight: 0.5, Extractor: func(b models.Bio) string { return b.Food }},
// 			{Name: "Travel", Weight: 0.5, Extractor: func(b models.Bio) string { return b.Travel }},
// 		}
// 	}
// 	rs.FieldConfigs = fieldConfigs
// 	logrus.Infof("RecommendationService initialized with mode=%s", rs.Mode)
// 	return rs
// }

// type candidate struct {
// 	User     models.User
// 	Score    float64
// 	Distance float64
// }

// // GetRecommendationsForUser возвращает список до 10 id пользователей.
// func (rs *RecommendationService) GetRecommendationsForUser(currentUserID uuid.UUID, mode string) ([]uuid.UUID, error) {
// 	// Устанавливаем режим
// 	if mode != "desire" {
// 		rs.Mode = "affinity"
// 	} else {
// 		rs.Mode = "desire"
// 	}

// 	// Загрузка текущего пользователя с профилем и настройками
// 	var currentUser models.User
// 	if err := rs.DB.Preload("Profile").Preload("Bio").Preload("Preference").
// 		First(&currentUser, "id = ?", currentUserID).Error; err != nil {
// 		logrus.Errorf("GetRecommendationsForUser: не удалось найти пользователя %s: %v", currentUserID, err)
// 		return nil, err
// 	}
// 	if err := validateUserData(currentUser); err != nil {
// 		return nil, err
// 	}

// 	// Загрузка всех других пользователей
// 	var users []models.User
// 	if err := rs.DB.Preload("Profile").Preload("Bio").
// 		Where("id <> ?", currentUserID).
// 		Find(&users).Error; err != nil {
// 		logrus.Errorf("GetRecommendationsForUser: загрузка кандидатов: %v", err)
// 		return nil, err
// 	}

// 	// Сбор кандидатов
// 	var candidates []candidate
// 	for _, u := range users {
// 		// Исключаем отклонённые рекомендации
// 		var rec models.Recommendation
// 		if err := rs.DB.Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
// 			First(&rec).Error; err == nil {
// 			continue
// 		}

// 		// Фильтрация по дистанции
// 		d := computeDistanceIfNeeded(currentUser.Profile, u.Profile, currentUser.Preference.MaxRadius)
// 		if d < 0 {
// 			continue
// 		}

// 		// Режим affinity
// 		var score float64
// 		switch rs.Mode {
// 		case "affinity":
// 			totalWeight := 0.0
// 			for _, fc := range rs.FieldConfigs {
// 				// базовый вес 0.2
// 				w := 0.2
// 				// умножаем на 2, если приоритет задан в Preference
// 				switch fc.Name {
// 				case "Interests":
// 					if currentUser.Preference.PriorityInterests {
// 						w *= 2
// 					}
// 				case "Hobbies":
// 					if currentUser.Preference.PriorityHobbies {
// 						w *= 2
// 					}
// 				case "Music":
// 					if currentUser.Preference.PriorityMusic {
// 						w *= 2
// 					}
// 				case "Food":
// 					if currentUser.Preference.PriorityFood {
// 						w *= 2
// 					}
// 				case "Travel":
// 					if currentUser.Preference.PriorityTravel {
// 						w *= 2
// 					}
// 				}
// 				totalWeight += w

// 				// считаем число общих токенов
// 				setA := make(map[string]struct{})
// 				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(currentUser.Bio))) {
// 					setA[t] = struct{}{}
// 				}
// 				common := 0
// 				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(u.Bio))) {
// 					if _, ok := setA[t]; ok {
// 						common++
// 					}
// 				}
// 				score += float64(common) * w
// 			}
// 			// нормализация (чтобы score был в 0…1)
// 			if totalWeight > 0 {
// 				score /= totalWeight
// 			}

// 		// Режим desire
// 		case "desire":
// 			score = 0
// 			desired := make(map[string]struct{})
// 			for _, t := range strings.Fields(strings.ToLower(currentUser.Bio.LookingFor)) {
// 				desired[t] = struct{}{}
// 			}
// 			for _, t := range strings.Fields(strings.ToLower(u.Bio.LookingFor)) {
// 				if _, ok := desired[t]; ok {
// 					score += 0.1
// 				}
// 			}
// 		}

// 		if score <= 0 {
// 			continue
// 		}
// 		candidates = append(candidates, candidate{User: u, Score: score, Distance: d})
// 	}

// 	// Сортировка: сначала по расстоянию (ближе первыми), потом по score (выше первыми)
// 	sort.Slice(candidates, func(i, j int) bool {
// 		if math.Abs(candidates[i].Distance-candidates[j].Distance) > 1e-9 {
// 			return candidates[i].Distance < candidates[j].Distance
// 		}
// 		return candidates[i].Score > candidates[j].Score
// 	})

// 	// Берём не более 10
// 	limit := 10
// 	if len(candidates) < limit {
// 		limit = len(candidates)
// 	}
// 	result := make([]uuid.UUID, limit)
// 	for i := 0; i < limit; i++ {
// 		result[i] = candidates[i].User.ID
// 	}
// 	logrus.Infof("GetRecommendationsForUser[%s]: сформировано %d рекомендаций", rs.Mode, len(result))
// 	return result, nil
// }

// // validateUserData проверяет, что профиль и биография пользователя заполнены.
// func validateUserData(u models.User) error {
// 	if u.Profile.ID == 0 || u.Bio.ID == 0 {
// 		return errors.New("заполните профиль и биографию для получения рекомендаций")
// 	}
// 	if u.Profile.FirstName == "" || u.Profile.LastName == "" {
// 		return errors.New("пожалуйста, укажите имя и фамилию")
// 	}
// 	s := []struct{ val, msg string }{
// 		{u.Bio.Interests, "интересы"},
// 		{u.Bio.Hobbies, "хобби"},
// 		{u.Bio.Music, "музыку"},
// 		{u.Bio.Food, "еду"},
// 		{u.Bio.Travel, "путешествия"},
// 		{u.Bio.LookingFor, "кого вы ищете"},
// 	}
// 	miss := []string{}
// 	for _, f := range s {
// 		if strings.TrimSpace(f.val) == "" {
// 			miss = append(miss, f.msg)
// 		}
// 	}
// 	if len(miss) > 0 {
// 		return errors.New("пожалуйста, заполните вашу биографию: " + strings.Join(miss, ", "))
// 	}
// 	return nil
// }

// // computeSimilarityScore считает взвешенную оценку совпадения по конфигу.
// func (rs *RecommendationService) computeSimilarityScore(a, b models.Bio) float64 {
// 	var totalWeight float64
// 	for _, fc := range rs.FieldConfigs {
// 		totalWeight += fc.Weight
// 	}
// 	// если нет весов — fallback на простую логику двух полей.
// 	if totalWeight <= 0 {
// 		return simpleSimilarity(a, b)
// 	}

// 	// подсчет для каждого поля
// 	var score float64
// 	scoreField := func(textA, textB string, w float64) float64 {
// 		setA := make(map[string]struct{})
// 		for _, token := range strings.Fields(strings.ToLower(textA)) {
// 			setA[token] = struct{}{}
// 		}
// 		common := 0
// 		for _, token := range strings.Fields(strings.ToLower(textB)) {
// 			if _, ok := setA[token]; ok {
// 				common++
// 			}
// 		}
// 		return float64(common) * w
// 	}
// 	for _, fc := range rs.FieldConfigs {
// 		score += scoreField(fc.Extractor(a), fc.Extractor(b), fc.Weight)
// 	}
// 	return score
// }

// // simpleSimilarity считает количество общих слов только в Interests+Hobbies.
// func simpleSimilarity(a, b models.Bio) float64 {
// 	textA := a.Interests + " " + a.Hobbies
// 	textB := b.Interests + " " + b.Hobbies
// 	setA := make(map[string]struct{})
// 	for _, token := range strings.Fields(strings.ToLower(textA)) {
// 		setA[token] = struct{}{}
// 	}
// 	common := 0
// 	for _, token := range strings.Fields(strings.ToLower(textB)) {
// 		if _, ok := setA[token]; ok {
// 			common++
// 		}
// 	}
// 	return float64(common)
// }

// // computeDistanceIfNeeded рассчитывает расстояние и проверяет радиус, возвращает -1 если вне радиуса.
// func computeDistanceIfNeeded(p1, p2 models.Profile, maxRadius float64) float64 {
// 	if maxRadius > 0 && p1.Latitude != 0 && p1.Longitude != 0 && p2.Latitude != 0 && p2.Longitude != 0 {
// 		d := haversineDistance(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
// 		if d > maxRadius {
// 			return -1
// 		}
// 		return d
// 	}
// 	return 0 // не проверялось или в радиусе
// }

// // haversineDistance ...
// func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
// 	const earthRadius = 6371
// 	dLat := degreesToRadians(lat2 - lat1)
// 	dLon := degreesToRadians(lon2 - lon1)
// 	lat1Rad := degreesToRadians(lat1)
// 	lat2Rad := degreesToRadians(lat2)
// 	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
// 		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
// 	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
// 	return earthRadius * c
// }

// func degreesToRadians(deg float64) float64 {
// 	return deg * math.Pi / 180
// }

// // Добавьте в RecommendationService:
// func (rs *RecommendationService) DeclineRecommendation(currentUserID, recUserID uuid.UUID) error {
// 	// Проверим, что мы не дублируем отказ
// 	var existing models.Recommendation
// 	if err := rs.DB.
// 		Where("user_id = ? AND rec_user_id = ?", currentUserID, recUserID).
// 		First(&existing).Error; err == nil {
// 		// Обновляем статус, если уже была запись
// 		existing.Status = "declined"
// 		return rs.DB.Save(&existing).Error
// 	}

// 	// Иначе создаём новую запись
// 	rec := models.Recommendation{
// 		UserID:    currentUserID,
// 		RecUserID: recUserID,
// 		Status:    "declined",
// 	}
// 	return rs.DB.Create(&rec).Error
// }

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

// FieldConfig описывает одно поле биографии и его вес.
type FieldConfig struct {
	Name      string                    // Человеко-читаемое имя поля
	Weight    float64                   // Вес поля
	Extractor func(b models.Bio) string // Функция для получения текста из Bio
}

// RecommendationService инкапсулирует бизнес-логику рекомендаций.
type RecommendationService struct {
	DB           *gorm.DB
	FieldConfigs []FieldConfig
	Mode         string
}

// NewRecommendationService создаёт новый экземпляр RecommendationService.
func NewRecommendationService(db *gorm.DB, fieldConfigs []FieldConfig) *RecommendationService {
	rs := &RecommendationService{DB: db, Mode: "affinity"}
	if len(fieldConfigs) == 0 {
		fieldConfigs = []FieldConfig{
			{Name: "Interests", Weight: 1.0, Extractor: func(b models.Bio) string { return b.Interests }},
			{Name: "Hobbies", Weight: 1.0, Extractor: func(b models.Bio) string { return b.Hobbies }},
			{Name: "Music", Weight: 0.8, Extractor: func(b models.Bio) string { return b.Music }},
			{Name: "Food", Weight: 0.5, Extractor: func(b models.Bio) string { return b.Food }},
			{Name: "Travel", Weight: 0.5, Extractor: func(b models.Bio) string { return b.Travel }},
		}
	}
	rs.FieldConfigs = fieldConfigs
	logrus.Infof("RecommendationService initialized with mode=%s", rs.Mode)
	return rs
}

// anyTokenMatch возвращает true, если хотя бы один токен из a содержится среди токенов b.
func anyTokenMatch(a, b string) bool {
	tokensA := strings.Fields(strings.ToLower(a))
	tokensB := strings.Fields(strings.ToLower(b))
	setB := make(map[string]struct{}, len(tokensB))
	for _, t := range tokensB {
		setB[t] = struct{}{}
	}
	for _, t := range tokensA {
		if _, ok := setB[t]; ok {
			return true
		}
	}
	return false
}

// RecommendationWithDistance — структура "ID + Distance" для контроллера.
type RecommendationWithDistance struct {
	UserID   uuid.UUID
	Distance float64
}

type candidate struct {
	User     models.User
	Score    float64
	Distance float64
}

// GetRecommendationsForUser возвращает список до 10 id пользователей.
func (rs *RecommendationService) GetRecommendationsForUser(currentUserID uuid.UUID, mode string) ([]uuid.UUID, error) {
	// Устанавливаем режим
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	// Загружаем пользователя и проверяем данные
	var currentUser models.User
	if err := rs.DB.Preload("Profile").Preload("Bio").Preload("Preference").
		First(&currentUser, "id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("GetRecommendationsForUser: не удалось загрузить пользователя %s: %v", currentUserID, err)
		return nil, err
	}
	if err := validateUserData(currentUser); err != nil {
		return nil, err
	}

	// Загружаем всех остальных
	var users []models.User
	if err := rs.DB.Preload("Profile").Preload("Bio").
		Where("id <> ?", currentUserID).Find(&users).Error; err != nil {
		logrus.Errorf("GetRecommendationsForUser: загрузка кандидатов: %v", err)
		return nil, err
	}

	// Сбор кандидатов
	var cands []candidate
	for _, u := range users {
		// Пропускаем отклонённых
		var rec models.Recommendation
		if err := rs.DB.Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		// Фильтрация по радиусу
		d := computeDistanceIfNeeded(currentUser.Profile, u.Profile, currentUser.Preference.MaxRadius)
		if d < 0 {
			continue
		}

		// Вычисление score
		var score float64
		switch rs.Mode {
		case "affinity":
			totalW := 0.0
			for _, fc := range rs.FieldConfigs {
				w := 0.2
				switch fc.Name {
				case "Interests":
					if currentUser.Preference.PriorityInterests {
						w *= 2
					}
				case "Hobbies":
					if currentUser.Preference.PriorityHobbies {
						w *= 2
					}
				case "Music":
					if currentUser.Preference.PriorityMusic {
						w *= 2
					}
				case "Food":
					if currentUser.Preference.PriorityFood {
						w *= 2
					}
				case "Travel":
					if currentUser.Preference.PriorityTravel {
						w *= 2
					}
				}
				totalW += w
				// Считаем общие токены
				setA := make(map[string]struct{})
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(currentUser.Bio))) {
					setA[t] = struct{}{}
				}
				common := 0
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(u.Bio))) {
					if _, ok := setA[t]; ok {
						common++
					}
				}
				score += float64(common) * w
			}
			if totalW > 0 {
				score /= totalW
			}

		case "desire":
			// Считаем совпадения LookingFor
			tokens := strings.Fields(strings.ToLower(currentUser.Bio.LookingFor))
			for _, tok := range tokens {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.1
				}
			}
		}

		if score <= 0 {
			continue
		}
		cands = append(cands, candidate{User: u, Score: score, Distance: d})
	}

	// Сортировка по дистанции↑, затем score↓
	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

	// Ограничение до 10
	limit := 10
	if len(cands) < limit {
		limit = len(cands)
	}
	result := make([]uuid.UUID, limit)
	for i := 0; i < limit; i++ {
		result[i] = cands[i].User.ID
	}
	logrus.Infof("GetRecommendationsForUser[%s]: сформировано %d рекомендаций", rs.Mode, len(result))
	return result, nil
}

// GetRecommendationsWithDistance возвращает до 10 рекомендаций вместе с рассчитанным расстоянием.
func (rs *RecommendationService) GetRecommendationsWithDistance(currentUserID uuid.UUID, mode string) ([]RecommendationWithDistance, error) {
	// 1) Устанавливаем режим работы
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	// 2) Загружаем текущего пользователя
	var currentUser models.User
	if err := rs.DB.Preload("Profile").Preload("Bio").Preload("Preference").
		First(&currentUser, "id = ?", currentUserID).Error; err != nil {
		logrus.Errorf("GetRecommendationsWithDistance: не удалось загрузить пользователя %s: %v", currentUserID, err)
		return nil, err
	}
	if err := validateUserData(currentUser); err != nil {
		return nil, err
	}

	// 3) Загружаем всех остальных кандидатов
	var users []models.User
	if err := rs.DB.Preload("Profile").Preload("Bio").
		Where("id <> ?", currentUserID).
		Find(&users).Error; err != nil {
		logrus.Errorf("GetRecommendationsWithDistance: ошибка загрузки кандидатов: %v", err)
		return nil, err
	}

	// 4) Формируем список кандидатов с оценкой и дистанцией
	type cand struct {
		ID              uuid.UUID
		Score, Distance float64
	}
	var cands []cand
	for _, u := range users {
		// Пропускаем отклонённые
		var rec models.Recommendation
		if err := rs.DB.Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		// Расстояние (или -1 если вне радиуса)
		d := computeDistanceIfNeeded(currentUser.Profile, u.Profile, currentUser.Preference.MaxRadius)
		if d < 0 {
			continue
		}

		// Оцениваем совпадение
		var score float64
		switch rs.Mode {
		case "affinity":
			totalW := 0.0
			for _, fc := range rs.FieldConfigs {
				w := 0.2
				switch fc.Name {
				case "Interests":
					if currentUser.Preference.PriorityInterests {
						w *= 2
					}
				case "Hobbies":
					if currentUser.Preference.PriorityHobbies {
						w *= 2
					}
				case "Music":
					if currentUser.Preference.PriorityMusic {
						w *= 2
					}
				case "Food":
					if currentUser.Preference.PriorityFood {
						w *= 2
					}
				case "Travel":
					if currentUser.Preference.PriorityTravel {
						w *= 2
					}
				}
				totalW += w
				// Считаем общие токены
				setA := make(map[string]struct{})
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(currentUser.Bio))) {
					setA[t] = struct{}{}
				}
				common := 0
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(u.Bio))) {
					if _, ok := setA[t]; ok {
						common++
					}
				}
				score += float64(common) * w
			}
			if totalW > 0 {
				score /= totalW
			}

		case "desire":
			score = 0
			want := make(map[string]struct{})
			for _, t := range strings.Fields(strings.ToLower(currentUser.Bio.LookingFor)) {
				want[t] = struct{}{}
			}
			for _, t := range strings.Fields(strings.ToLower(u.Bio.LookingFor)) {
				if _, ok := want[t]; ok {
					score += 0.1
				}
			}
		}
		if score <= 0 {
			continue
		}
		cands = append(cands, cand{ID: u.ID, Score: score, Distance: d})
	}

	// 5) Сортировка по distance↑, потом score↓
	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

	// 6) Формируем результат до 10 элементов
	limit := 10
	if len(cands) < limit {
		limit = len(cands)
	}
	out := make([]RecommendationWithDistance, limit)
	for i := 0; i < limit; i++ {
		out[i] = RecommendationWithDistance{UserID: cands[i].ID, Distance: cands[i].Distance}
	}

	logrus.Infof("GetRecommendationsWithDistance[%s]: вернуло %d рекомендаций", rs.Mode, len(out))
	return out, nil
}

// validateUserData проверяет, что профиль и биография пользователя заполнены.
func validateUserData(u models.User) error {
	if u.Profile.ID == 0 || u.Bio.ID == 0 {
		return errors.New("заполните профиль и биографию для получения рекомендаций")
	}
	if u.Profile.FirstName == "" || u.Profile.LastName == "" {
		return errors.New("пожалуйста, укажите имя и фамилию")
	}
	s := []struct{ val, msg string }{
		{u.Bio.Interests, "интересы"},
		{u.Bio.Hobbies, "хобби"},
		{u.Bio.Music, "музыку"},
		{u.Bio.Food, "еду"},
		{u.Bio.Travel, "путешествия"},
		{u.Bio.LookingFor, "кого вы ищете"},
	}
	miss := []string{}
	for _, f := range s {
		if strings.TrimSpace(f.val) == "" {
			miss = append(miss, f.msg)
		}
	}
	if len(miss) > 0 {
		return errors.New("пожалуйста, заполните вашу биографию: " + strings.Join(miss, ", "))
	}
	return nil
}

// computeDistanceIfNeeded рассчитывает расстояние и проверяет радиус.
func computeDistanceIfNeeded(p1, p2 models.Profile, maxRadius float64) float64 {
	if maxRadius > 0 && p1.Latitude != 0 && p1.Longitude != 0 && p2.Latitude != 0 && p2.Longitude != 0 {
		d := haversineDistance(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
		if d > maxRadius {
			return -1
		}
		return d
	}
	return 0 // не проверялось или в радиусе
}

// haversineDistance ...
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371
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

// DeclineRecommendation сохраняет отказ в БД.
func (rs *RecommendationService) DeclineRecommendation(currentUserID, recUserID uuid.UUID) error {
	var existing models.Recommendation
	if err := rs.DB.Where("user_id = ? AND rec_user_id = ?", currentUserID, recUserID).
		First(&existing).Error; err == nil {
		existing.Status = "declined"
		return rs.DB.Save(&existing).Error
	}
	rec := models.Recommendation{UserID: currentUserID, RecUserID: recUserID, Status: "declined"}
	return rs.DB.Create(&rec).Error
}
