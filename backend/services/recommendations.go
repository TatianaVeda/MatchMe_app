package services

import (
	"errors"
	"fmt"
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
	UserID   uuid.UUID `json:"id"`
	Distance float64   `json:"distance"`
	Score    float64   `json:"score"`
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
// func (rs *RecommendationService) GetNearbyUsers(
// 	lat, lon, maxRadius float64,
// 	limit int,
// 	excludeID uuid.UUID,
// ) ([]Nearby, error) {
// 	maxMeters := maxRadius * 1000.0
// 	var list []Nearby
// 	err := rs.DB.
// 		Raw(`
// 			SELECT
// 			  user_id   AS id,
// 			  earth_distance(earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
// 			FROM profiles
// 			WHERE
// 			  earth_box(ll_to_earth(?, ?), ?) @> earth_loc
// 			  AND earth_distance(earth_loc, ll_to_earth(?, ?)) <= ?
// 			  AND user_id != ?
// 			ORDER BY distance ASC
// 			LIMIT ?
// 		`,
// 			// placeholders:
// 			lat, lon, // first earth_distance
// 			lat, lon, maxMeters, // earth_box
// 			lat, lon, maxMeters, // distance filter
// 			excludeID, limit,
// 		).
// 		Scan(&list).Error
// 	return list, err
// }

func (rs *RecommendationService) GetNearbyUsers(
	lat, lon, maxRadius float64,
	limit int,
	excludeID uuid.UUID,
) ([]Nearby, error) {
	maxMeters := maxRadius * 1000.0
	fmt.Printf("[DEBUG] Searching nearby users from lat=%.6f, lon=%.6f, radius=%.2f km\n", lat, lon, maxRadius)
	fmt.Printf("[DEBUG] Excluding user ID: %s\n", excludeID)

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
			lat, lon, // distance
			lat, lon, maxMeters, // box
			lat, lon, maxMeters, // filter
			excludeID, limit,
		).
		Scan(&list).Error

	if err != nil {
		fmt.Printf("[ERROR] Nearby query failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Nearby users found: %d\n", len(list))
	for _, n := range list {
		fmt.Printf("[DEBUG] User %s at %.2f km\n", n.ID, n.Distance)
	}
	return list, nil
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

	fmt.Printf("[DEBUG] Mode: %s\n", rs.Mode)

	// 2) Загружаем текущего пользователя с Profile, Bio и Preference
	var me models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Preload("Preference").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		fmt.Printf("[ERROR] Failed to load user: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Current User: %+v\n", me)
	fmt.Printf("[DEBUG] Current user: %s\n", me.ID)
	fmt.Printf("[DEBUG] Interests: %v\n", strings.Fields(me.Bio.Interests))
	fmt.Printf("[DEBUG] Hobbies: %v\n", strings.Fields(me.Bio.Hobbies))
	fmt.Printf("[DEBUG] Music: %v\n", strings.Fields(me.Bio.Music))

	if err := validateUserData(me); err != nil {
		fmt.Printf("[ERROR] Validation failed: %v\n", err)
		return nil, err
	}

	// 3) Отбираем кандидатов по радиусу через GetNearbyUsers
	nearby, err := rs.GetNearbyUsers(
		me.Profile.Latitude,
		me.Profile.Longitude,
		me.Preference.MaxRadius,
		//wtf?!
		10000,         // берём чуть больше, чтобы потом отсечь по score?????
		currentUserID, // исключаем самого себя
	)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get nearby users: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Found %d nearby users\n", len(nearby))
	for _, u := range nearby {
		fmt.Printf("[DEBUG] Nearby user: %s at %.2f km\n", u.ID, u.Distance)
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
	fmt.Printf("[DEBUG] Distance map: %+v\n", distMap)

	// 5) Подгружаем профили и био этих кандидатов
	var users []models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Bio").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		fmt.Printf("[ERROR] Failed to load candidate users: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Loaded %d candidate users\n", len(users))

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
		fmt.Printf("[DEBUG] Checking user: %s\n", u.ID)
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			fmt.Printf("[DEBUG] Skipping declined user: %s\n", u.ID)
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
				fmt.Printf("[DEBUG] Comparing %s: me=%v, other=%v, common=%d, weight=%.2f\n",
					fc.Name,
					strings.Fields(strings.ToLower(fc.Extractor(me.Bio))),
					strings.Fields(strings.ToLower(fc.Extractor(u.Bio))),
					common,
					w)

				score += float64(common) * w
				fmt.Printf("[DEBUG] Final score for user %s: %.4f (distance: %.2f)\n", u.ID, score, d)

			}
			if totalW > 0 {
				score /= totalW
				fmt.Printf("[DEBUG] Final score for user %s: %.4f (distance: %.2f)\n", u.ID, score, d)
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
			fmt.Printf("[DEBUG] Candidate %s: score=%.2f, distance=%.2f\n", u.ID, score, d)
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
			Score:    cands[i].Score,
		}
	}
	fmt.Printf("[DEBUG] Returning %d recommendations\n", len(out))

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

// GetRecommendationsWithFilters возвращает до 10 UUID рекомендаций,
// применяя переданные вручную фильтры вместо того, что хранится в профиле.
func (rs *RecommendationService) GetRecommendationsWithFilters(
	currentUserID uuid.UUID,
	mode string,
	lat, lon float64,
	interests []string, prioInterests bool,
	hobbies []string, prioHobbies bool,
	music []string, prioMusic bool,
	food []string, prioFood bool,
	travel []string, prioTravel bool,
	lookingFor string, // используется только если mode=="desire"
) ([]uuid.UUID, error) {
	// 1) Устанавливаем режим
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	// 2) Загружаем только профиль (не нужен Bio и Preference для "me")
	var me models.User
	if err := rs.DB.Preload("Profile").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		return nil, err
	}
	// Подменяем координаты на переданные
	me.Profile.Latitude = lat
	me.Profile.Longitude = lon

	// 3) Быстрый отбор по локации (использует me.Preference.MaxRadius, можно тоже передать его через аргумент)
	nearby, err := rs.GetNearbyUsers(
		lat, lon,
		me.Preference.MaxRadius, // или переданный maxRadius
		50,
		currentUserID,
	)
	if err != nil || len(nearby) == 0 {
		return nil, err
	}
	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}

	// 4) Подгружаем кандидатов с их Bio
	var users []models.User
	if err := rs.DB.Preload("Bio").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 5) Считаем score
	type cand struct {
		id       uuid.UUID
		score    float64
		distance float64
	}
	var candidates []cand

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
			// для каждого поля считаем пересечение и умножаем на вес
			totalW := 0.0
			// Interests
			w := 0.1
			if prioInterests {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(interests, strings.Fields(u.Bio.Interests))) * w

			// Hobbies
			w = 0.1
			if prioHobbies {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(hobbies, strings.Fields(u.Bio.Hobbies))) * w

			// Music
			w = 0.1
			if prioMusic {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(music, strings.Fields(u.Bio.Music))) * w

			// Food
			w = 0.1
			if prioFood {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(food, strings.Fields(u.Bio.Food))) * w

			// Travel
			w = 0.1
			if prioTravel {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(travel, strings.Fields(u.Bio.Travel))) * w

			if totalW > 0 {
				score = score / totalW
			}
		} else {
			// desire — один список lookingFor
			for _, tok := range strings.Fields(strings.ToLower(lookingFor)) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}

		if score > 0 {
			if score > 1 {
				score = 1
			}
			candidates = append(candidates, cand{u.ID, score, d})
		}
	}

	// 6) Сортировка: сначала distance ↑, потом score ↓
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].score > candidates[j].score
	})

	// 7) Берём топ-10
	limit := 10
	if len(candidates) < limit {
		limit = len(candidates)
	}
	result := make([]uuid.UUID, limit)
	for i := 0; i < limit; i++ {
		result[i] = candidates[i].id
	}
	return result, nil
}

// countCommon считает пересечение двух срезов строк
func countCommon(a, b []string) int {
	set := make(map[string]struct{})
	for _, x := range a {
		set[strings.ToLower(x)] = struct{}{}
	}
	cnt := 0
	for _, y := range b {
		if _, ok := set[strings.ToLower(y)]; ok {
			cnt++
		}
	}
	return cnt
}

// GetRecommendationsWithFiltersWithDistance возвращает до 10 рекомендаций вместе с рассчитанным расстоянием,
// применяя пользовательские фильтры вместо данных из хранимого профиля.
func (rs *RecommendationService) GetRecommendationsWithFiltersWithDistance(
	currentUserID uuid.UUID,
	mode string,
	lat, lon float64,
	interests []string, prioInterests bool,
	hobbies []string, prioHobbies bool,
	music []string, prioMusic bool,
	food []string, prioFood bool,
	travel []string, prioTravel bool,
	lookingFor string, // для режима desire
) ([]RecommendationWithDistance, error) {
	fmt.Println("Mode:", mode)
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}
	fmt.Println("Recommendation mode set to:", rs.Mode)
	var me models.User
	if err := rs.DB.Preload("Profile").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		return nil, err
	}
	fmt.Printf("Loaded current user: %s, setting coordinates: (%f, %f)\n", me.ID, lat, lon)
	me.Profile.Latitude = lat
	me.Profile.Longitude = lon
	nearby, err := rs.GetNearbyUsers(
		lat, lon,
		me.Preference.MaxRadius,
		100,
		currentUserID,
	)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Nearby users found: %d\n", len(nearby))
	if len(nearby) == 0 {
		return []RecommendationWithDistance{}, nil
	}
	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}
	var users []models.User
	if err := rs.DB.Preload("Bio").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return nil, err
	}
	fmt.Printf("Loaded bios for %d users\n", len(users))
	type candWithDist struct {
		userID   uuid.UUID
		Score    float64
		distance float64
	}
	var cands []candWithDist
	for _, u := range users {
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			fmt.Printf("User %s was previously declined. Skipping.\n", u.ID)
			continue
		}
		d := distMap[u.ID]
		var score float64
		if rs.Mode == "affinity" {
			totalW := 0.0
			w := 0.1
			if prioInterests {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(interests, strings.Fields(u.Bio.Interests))) * w
			fmt.Printf("User %s - Interest Score: %.2f\n", u.ID, score)
			w = 0.1
			if prioHobbies {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(hobbies, strings.Fields(u.Bio.Hobbies))) * w
			w = 0.1
			if prioMusic {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(music, strings.Fields(u.Bio.Music))) * w
			w = 0.1
			if prioFood {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(food, strings.Fields(u.Bio.Food))) * w
			w = 0.1
			if prioTravel {
				w *= 2
			}
			totalW += w
			score += float64(countCommon(travel, strings.Fields(u.Bio.Travel))) * w

			if totalW > 0 {
				score /= totalW
			}
		} else {
			for _, tok := range strings.Fields(strings.ToLower(lookingFor)) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}
		if score > 0 {
			if score > 1 {
				score = 1
			}
			cands = append(cands, candWithDist{u.ID, score, d})
			fmt.Printf("User %s added to recommendations with score %.2f and distance %.2f\n", u.ID, score, d)
		} else {
			fmt.Printf("User %s skipped due to zero score.\n", u.ID)
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].distance != cands[j].distance {
			return cands[i].distance < cands[j].distance
		}
		return cands[i].Score > cands[j].Score
	})
	limit := 10
	if len(cands) < limit {
		limit = len(cands)
	}
	fmt.Printf("Returning top %d candidates\n", limit)
	out := make([]RecommendationWithDistance, limit)
	for i := 0; i < limit; i++ {
		out[i] = RecommendationWithDistance{
			UserID:   cands[i].userID,
			Distance: cands[i].distance,
			Score:    cands[i].Score,
		}
	}
	return out, nil
}
