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

// Nearby — результат быстрого геопоиска через earth_loc
type Nearby struct {
	ID       uuid.UUID
	Distance float64
}

// FieldConfig описывает одно поле биографии и его "вес".
type FieldConfig struct {
	Name      string
	Weight    float64
	Extractor func(b models.Bio) string
}

// RecommendationWithDistance — структура для контроллера.
type RecommendationWithDistance struct {
	UserID   uuid.UUID
	Distance float64
}

// candidate — промежуточная структура для сортировки.
type candidate struct {
	User     models.User
	Score    float64
	Distance float64
}

// RecommendationService инкапсулирует логику рекомендаций.
type RecommendationService struct {
	DB           *gorm.DB
	FieldConfigs []FieldConfig
	Mode         string // "affinity" или "desire"
}

// NewRecommendationService создаёт сервис рекомендаций.
func NewRecommendationService(db *gorm.DB, fieldConfigs []FieldConfig) *RecommendationService {
	rs := &RecommendationService{DB: db, Mode: "affinity"}
	if len(fieldConfigs) == 0 {
		fieldConfigs = []FieldConfig{
			{Name: "Interests", Weight: 0.1, Extractor: func(b models.Bio) string { return b.Interests }},
			{Name: "Hobbies", Weight: 0.1, Extractor: func(b models.Bio) string { return b.Hobbies }},
			{Name: "Music", Weight: 0.1, Extractor: func(b models.Bio) string { return b.Music }},
			{Name: "Food", Weight: 0.1, Extractor: func(b models.Bio) string { return b.Food }},
			{Name: "Travel", Weight: 0.1, Extractor: func(b models.Bio) string { return b.Travel }},
		}
	}
	rs.FieldConfigs = fieldConfigs
	logrus.Infof("RecommendationService initialized with mode=%s", rs.Mode)
	return rs
}

// GetNearbyUsers возвращает до `limit` пользователей в радиусе maxRadius км от (lat, lon),
// исключая пользователя excludeID. Вся геолокационная логика выполняется в Postgres.
func (rs *RecommendationService) GetNearbyUsers(
	lat, lon, maxRadius float64,
	limit int,
	excludeID uuid.UUID,
) ([]Nearby, error) {
	maxMeters := maxRadius * 1000.0
	var list []Nearby
	err := rs.DB.
		Raw(`
			SELECT
			  user_id   AS id,
			  earth_distance(earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
			FROM profiles
			WHERE 
			  earth_box(ll_to_earth(?, ?), ?) @> earth_loc
			  AND earth_distance(earth_loc, ll_to_earth(?, ?)) <= ?
			  AND user_id != ?
			ORDER BY distance ASC
			LIMIT ?
		`,
			// placeholders:
			lat, lon, // first earth_distance
			lat, lon, maxMeters, // earth_box
			lat, lon, maxMeters, // distance filter
			excludeID, limit,
		).
		Scan(&list).Error
	return list, err
}

// GetRecommendationsForUser возвращает до 10 UUID рекомендаций.
func (rs *RecommendationService) GetRecommendationsForUser(
	currentUserID uuid.UUID,
	mode string,
) ([]uuid.UUID, error) {
	// 1) Устанавливаем режим
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	// 2) Загружаем текущего пользователя с Profile, Bio и Preference
	var me models.User
	if err := rs.DB.
		Preload("Profile").Preload("Bio").Preload("Preference").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		return nil, err
	}
	if err := validateUserData(me); err != nil {
		return nil, err
	}

	// 3) Быстрый отбор по локации
	nearby, err := rs.GetNearbyUsers(
		me.Profile.Latitude,
		me.Profile.Longitude,
		me.Preference.MaxRadius,
		50,            // берём чуть больше, чтобы потом отобрать топ-10 по score
		currentUserID, // исключаем себя
	)
	if err != nil {
		return nil, err
	}
	if len(nearby) == 0 {
		return nil, nil
	}

	// 4) Собираем ID и мапу дистанций
	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}

	// 5) Подгружаем Profile и Bio этих кандидатов
	var users []models.User
	if err := rs.DB.Preload("Profile").Preload("Bio").
		Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	// 6) Считаем score по affinity/desire
	var cands []candidate
	for _, u := range users {
		// пропускаем уже отклонённых
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		if rs.Mode == "affinity" {
			totalW := 0.0
			for _, fc := range rs.FieldConfigs {
				w := fc.Weight
				// удваиваем вес, если приоритетно
				switch fc.Name {
				case "Interests":
					if me.Preference.PriorityInterests {
						w *= 2
					}
				case "Hobbies":
					if me.Preference.PriorityHobbies {
						w *= 2
					}
				case "Music":
					if me.Preference.PriorityMusic {
						w *= 2
					}
				case "Food":
					if me.Preference.PriorityFood {
						w *= 2
					}
				case "Travel":
					if me.Preference.PriorityTravel {
						w *= 2
					}
				}
				totalW += w

				// считаем общее количество совпавших токенов
				setA := make(map[string]struct{})
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(me.Bio))) {
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

		} else {
			// desire
			for _, tok := range strings.Fields(strings.ToLower(me.Bio.LookingFor)) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}

		// гарантируем, что score в [0,1]
		if score > 1 {
			score = 1
		}

		if score > 0 {
			cands = append(cands, candidate{User: u, Score: score, Distance: d})
		}

	}

	// 7) Сортируем по distance ↑, затем score ↓
	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

	// 8) Берём до 10 результатов
	limit := 10
	if len(cands) < limit {
		limit = len(cands)
	}
	out := make([]uuid.UUID, limit)
	for i := 0; i < limit; i++ {
		out[i] = cands[i].User.ID
	}

	return out, nil
}

// GetRecommendationsWithDistance возвращает до 10 рекомендаций вместе с рассчитанным расстоянием.
func (rs *RecommendationService) GetRecommendationsWithDistance(
	currentUserID uuid.UUID,
	mode string,
) ([]RecommendationWithDistance, error) {
	// 1) Устанавливаем режим работы
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	// 2) Загружаем текущего пользователя с Profile, Bio и Preference
	var me models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Preload("Preference").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		return nil, err
	}
	if err := validateUserData(me); err != nil {
		return nil, err
	}

	// 3) Отбираем кандидатов по радиусу через GetNearbyUsers
	nearby, err := rs.GetNearbyUsers(
		me.Profile.Latitude,
		me.Profile.Longitude,
		me.Preference.MaxRadius,
		100,           // берём чуть больше, чтобы потом отсечь по score
		currentUserID, // исключаем самого себя
	)
	if err != nil {
		return nil, err
	}
	if len(nearby) == 0 {
		return []RecommendationWithDistance{}, nil
	}

	// 4) Собираем все id и создаём мапу id→distance
	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}

	// 5) Подгружаем профили и био этих кандидатов
	var users []models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 6) Считаем score и собираем окончательный слайс
	type cand struct {
		ID       uuid.UUID
		Score    float64
		Distance float64
	}
	var cands []cand

	for _, u := range users {
		// пропускаем тех, кого пользователь уже отклонил
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		switch rs.Mode {
		case "affinity":
			totalW := 0.0
			for _, fc := range rs.FieldConfigs {
				// берём вес из конфигурации и применяем приоритеты
				w := fc.Weight
				switch fc.Name {
				case "Interests":
					if me.Preference.PriorityInterests {
						w *= 2
					}
				case "Hobbies":
					if me.Preference.PriorityHobbies {
						w *= 2
					}
				case "Music":
					if me.Preference.PriorityMusic {
						w *= 2
					}
				case "Food":
					if me.Preference.PriorityFood {
						w *= 2
					}
				case "Travel":
					if me.Preference.PriorityTravel {
						w *= 2
					}
				}
				totalW += w

				// считаем пересечение токенов
				setA := make(map[string]struct{}, 0)
				for _, t := range strings.Fields(strings.ToLower(fc.Extractor(me.Bio))) {
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
			for _, tok := range strings.Fields(strings.ToLower(me.Bio.LookingFor)) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}

		// удерживаем score в диапазоне [0,1]
		if score > 1 {
			score = 1
		}

		// фильтрация: только положительный score
		if score > 0 {
			cands = append(cands, cand{
				ID:       u.ID,
				Score:    score,
				Distance: d,
			})
		}
	}

	// 7) Сортируем: сначала по distance ↑, потом по score ↓
	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

	// 8) Формируем выходной массив до 10 элементов
	limit := 10
	if len(cands) < limit {
		limit = len(cands)
	}
	out := make([]RecommendationWithDistance, limit)
	for i := 0; i < limit; i++ {
		out[i] = RecommendationWithDistance{
			UserID:   cands[i].ID,
			Distance: cands[i].Distance,
		}
	}

	return out, nil
}

// DeclineRecommendation сохраняет отказ пользователя.
func (rs *RecommendationService) DeclineRecommendation(
	currentUserID, recUserID uuid.UUID,
) error {
	var existing models.Recommendation
	if err := rs.DB.
		Where("user_id = ? AND rec_user_id = ?", currentUserID, recUserID).
		First(&existing).Error; err == nil {
		existing.Status = "declined"
		return rs.DB.Save(&existing).Error
	}
	rec := models.Recommendation{
		UserID:    currentUserID,
		RecUserID: recUserID,
		Status:    "declined",
	}
	return rs.DB.Create(&rec).Error
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

// validateUserData проверяет, что профиль и биография пользователя заполнены.
func validateUserData(u models.User) error {
	if u.Profile.ID == 0 || u.Bio.ID == 0 {
		return errors.New("заполните профиль и биографию для получения рекомендаций")
	}
	if u.Profile.FirstName == "" || u.Profile.LastName == "" {
		return errors.New("пожалуйста, укажите имя и фамилию")
	}
	must := []struct{ val, msg string }{
		{u.Bio.Interests, "интересы"},
		{u.Bio.Hobbies, "хобби"},
		{u.Bio.Music, "музыку"},
		{u.Bio.Food, "еду"},
		{u.Bio.Travel, "путешествия"},
		{u.Bio.LookingFor, "кого вы ищете"},
	}
	miss := []string{}
	for _, f := range must {
		if strings.TrimSpace(f.val) == "" {
			miss = append(miss, f.msg)
		}
	}
	if len(miss) > 0 {
		return errors.New("пожалуйста, заполните вашу биографию: " + strings.Join(miss, ", "))
	}
	return nil
}
