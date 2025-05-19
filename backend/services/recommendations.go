// package services

// import (
// 	"errors"
// 	"fmt"
// 	"m/backend/models"
// 	"math"
// 	"sort"
// 	"strings"
// 	"unicode"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/gorm"
// )

// type Nearby struct {
// 	ID       uuid.UUID
// 	Distance float64
// }
// type FieldConfig struct {
// 	Name      string
// 	Weight    float64
// 	Extractor func(b models.Bio) string
// }
// type RecommendationWithDistance struct {
// 	UserID   uuid.UUID `json:"id"`
// 	Distance float64   `json:"distance"`
// 	Score    float64   `json:"score"`
// }
// type candidate struct {
// 	User     models.User
// 	Score    float64
// 	Distance float64
// }
// type RecommendationService struct {
// 	DB           *gorm.DB
// 	FieldConfigs []FieldConfig
// 	Mode         string
// }

// func NewRecommendationService(db *gorm.DB, fieldConfigs []FieldConfig) *RecommendationService {
// 	rs := &RecommendationService{DB: db, Mode: "affinity"}
// 	if len(fieldConfigs) == 0 {
// 		fieldConfigs = []FieldConfig{
// 			{Name: "Interests", Weight: 0.005, Extractor: func(b models.Bio) string { return b.Interests }},
// 			{Name: "Hobbies", Weight: 0.005, Extractor: func(b models.Bio) string { return b.Hobbies }},
// 			{Name: "Music", Weight: 0.005, Extractor: func(b models.Bio) string { return b.Music }},
// 			{Name: "Food", Weight: 0.005, Extractor: func(b models.Bio) string { return b.Food }},
// 			{Name: "Travel", Weight: 0.005, Extractor: func(b models.Bio) string { return b.Travel }},
// 		}
// 	}
// 	rs.FieldConfigs = fieldConfigs
// 	logrus.Infof("RecommendationService initialized with mode=%s", rs.Mode)
// 	return rs
// }
// func splitTokens(s string) []string {
// 	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
// 		return r == ',' || unicode.IsSpace(r)
// 	})
// }
// func countCommon(a, b []string) int {
// 	set := make(map[string]struct{}, len(a))
// 	for _, x := range a {
// 		set[x] = struct{}{}
// 	}
// 	cnt := 0
// 	for _, y := range b {
// 		if _, ok := set[y]; ok {
// 			cnt++
// 		}
// 	}
// 	return cnt
// }
// func (rs *RecommendationService) GetNearbyUsers(
// 	lat, lon, maxRadius float64,
// 	limit int,
// 	excludeID uuid.UUID,
// ) ([]Nearby, error) {
// 	maxMeters := maxRadius * 1000.0
// 	fmt.Printf("[DEBUG] Searching nearby users from lat=%.6f, lon=%.6f, radius=%.2f km\n", lat, lon, maxRadius)
// 	fmt.Printf("[DEBUG] Excluding user ID: %s\n", excludeID)
// 	var list []Nearby
// 	err := rs.DB.
// 		Raw(`
//             SELECT
//               user_id   AS id,
//               earth_distance(earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
//             FROM profiles
//             WHERE
//               earth_box(ll_to_earth(?, ?), ?) @> earth_loc
//               AND earth_distance(earth_loc, ll_to_earth(?, ?)) <= ?
//               AND user_id != ?
//             ORDER BY distance ASC
//             LIMIT ?
//         `,
// 			lat, lon,
// 			lat, lon, maxMeters,
// 			lat, lon, maxMeters,
// 			excludeID, limit,
// 		).
// 		Scan(&list).Error
// 	if err != nil {
// 		fmt.Printf("[ERROR] Nearby query failed: %v\n", err)
// 		return nil, err
// 	}
// 	return list, nil
// }
// func (rs *RecommendationService) GetRecommendationsForUser(
// 	currentUserID uuid.UUID,
// 	mode string,
// ) ([]uuid.UUID, error) {
// 	if mode != "desire" {
// 		rs.Mode = "affinity"
// 	} else {
// 		rs.Mode = "desire"
// 	}
// 	var me models.User
// 	if err := rs.DB.
// 		Preload("Profile").Preload("Bio").Preload("Preference").
// 		First(&me, "id = ?", currentUserID).Error; err != nil {
// 		return nil, err
// 	}
// 	if err := validateUserData(me); err != nil {
// 		return nil, err
// 	}
// 	nearby, err := rs.GetNearbyUsers(
// 		me.Profile.Latitude,
// 		me.Profile.Longitude,
// 		me.Preference.MaxRadius,
// 		50,
// 		currentUserID,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(nearby) == 0 {
// 		return nil, nil
// 	}
// 	ids := make([]uuid.UUID, len(nearby))
// 	distMap := make(map[uuid.UUID]float64, len(nearby))
// 	for i, n := range nearby {
// 		ids[i] = n.ID
// 		distMap[n.ID] = n.Distance
// 	}
// 	var users []models.User
// 	if err := rs.DB.Preload("Bio").
// 		Where("id IN ?", ids).
// 		Find(&users).Error; err != nil {
// 		return nil, err
// 	}
// 	var cands []candidate
// 	for _, u := range users {
// 		var rec models.Recommendation
// 		if err := rs.DB.
// 			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
// 			First(&rec).Error; err == nil {
// 			continue
// 		}
// 		d := distMap[u.ID]
// 		var score float64
// 		if rs.Mode == "affinity" {
// 			totalW := 0.0
// 			for _, fc := range rs.FieldConfigs {
// 				w := fc.Weight
// 				switch fc.Name {
// 				case "Interests":
// 					if me.Preference.PriorityInterests {
// 						w *= 2
// 					}
// 				case "Hobbies":
// 					if me.Preference.PriorityHobbies {
// 						w *= 2
// 					}
// 				case "Music":
// 					if me.Preference.PriorityMusic {
// 						w *= 2
// 					}
// 				case "Food":
// 					if me.Preference.PriorityFood {
// 						w *= 2
// 					}
// 				case "Travel":
// 					if me.Preference.PriorityTravel {
// 						w *= 2
// 					}
// 				}
// 				totalW += w
// 				setA := make(map[string]struct{})
// 				for _, t := range splitTokens(fc.Extractor(me.Bio)) {
// 					setA[t] = struct{}{}
// 				}
// 				common := 0
// 				for _, t := range splitTokens(fc.Extractor(u.Bio)) {
// 					if _, ok := setA[t]; ok {
// 						common++
// 					}
// 				}
// 				score += float64(common) * w
// 			}
// 			if totalW > 0 {
// 				score /= totalW
// 			}
// 		} else {
// 			for _, tok := range splitTokens(me.Bio.LookingFor) {
// 				if anyTokenMatch(tok, u.Bio.LookingFor) {
// 					score += 0.005
// 				}
// 			}
// 		}
// 		if score > 1 {
// 			score = 1
// 		}
// 		if score > 0 {
// 			cands = append(cands, candidate{User: u, Score: score, Distance: d})
// 		}
// 	}
// 	sort.Slice(cands, func(i, j int) bool {
// 		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
// 			return cands[i].Distance < cands[j].Distance
// 		}
// 		return cands[i].Score > cands[j].Score
// 	})
// 	limit := 10
// 	if len(cands) < limit {
// 		limit = len(cands)
// 	}
// 	out := make([]uuid.UUID, limit)
// 	for i := 0; i < limit; i++ {
// 		out[i] = cands[i].User.ID
// 	}
// 	return out, nil
// }
// func (rs *RecommendationService) GetRecommendationsWithDistance(
// 	currentUserID uuid.UUID, mode string,
// ) ([]RecommendationWithDistance, error) {
// 	if mode != "desire" {
// 		rs.Mode = "affinity"
// 	} else {
// 		rs.Mode = "desire"
// 	}
// 	var me models.User
// 	if err := rs.DB.
// 		Preload("Profile").Preload("Bio").Preload("Preference").
// 		First(&me, "id = ?", currentUserID).Error; err != nil {
// 		return nil, err
// 	}
// 	if err := validateUserData(me); err != nil {
// 		return nil, err
// 	}
// 	nearby, err := rs.GetNearbyUsers(
// 		me.Profile.Latitude,
// 		me.Profile.Longitude,
// 		me.Preference.MaxRadius,
// 		100,
// 		currentUserID,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(nearby) == 0 {
// 		return []RecommendationWithDistance{}, nil
// 	}
// 	ids := make([]uuid.UUID, len(nearby))
// 	distMap := make(map[uuid.UUID]float64, len(nearby))
// 	for i, n := range nearby {
// 		ids[i] = n.ID
// 		distMap[n.ID] = n.Distance
// 	}
// 	var users []models.User
// 	if err := rs.DB.Preload("Bio").
// 		Where("id IN ?", ids).
// 		Find(&users).Error; err != nil {
// 		return nil, err
// 	}
// 	type outCand struct {
// 		ID       uuid.UUID
// 		Score    float64
// 		Distance float64
// 	}
// 	var cands []outCand
// 	for _, u := range users {
// 		var rec models.Recommendation
// 		if err := rs.DB.
// 			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
// 			First(&rec).Error; err == nil {
// 			continue
// 		}
// 		d := distMap[u.ID]
// 		var score float64
// 		if rs.Mode == "affinity" {
// 			totalW := 0.0
// 			for _, fc := range rs.FieldConfigs {
// 				w := fc.Weight
// 				switch fc.Name {
// 				case "Interests":
// 					if me.Preference.PriorityInterests {
// 						w *= 2
// 					}
// 				case "Hobbies":
// 					if me.Preference.PriorityHobbies {
// 						w *= 2
// 					}
// 				case "Music":
// 					if me.Preference.PriorityMusic {
// 						w *= 2
// 					}
// 				case "Food":
// 					if me.Preference.PriorityFood {
// 						w *= 2
// 					}
// 				case "Travel":
// 					if me.Preference.PriorityTravel {
// 						w *= 2
// 					}
// 				}
// 				totalW += w
// 				setA := make(map[string]struct{})
// 				for _, t := range splitTokens(fc.Extractor(me.Bio)) {
// 					setA[t] = struct{}{}
// 				}
// 				common := 0
// 				for _, t := range splitTokens(fc.Extractor(u.Bio)) {
// 					if _, ok := setA[t]; ok {
// 						common++
// 					}
// 				}
// 				score += float64(common) * w
// 			}
// 			if totalW > 0 {
// 				score /= totalW
// 			}
// 		} else {
// 			for _, tok := range splitTokens(me.Bio.LookingFor) {
// 				if anyTokenMatch(tok, u.Bio.LookingFor) {
// 					score += 0.05
// 				}
// 			}
// 		}
// 		if score > 1 {
// 			score = 1
// 		}
// 		if score > 0 {
// 			cands = append(cands, outCand{ID: u.ID, Score: score, Distance: d})
// 		}
// 	}
// 	sort.Slice(cands, func(i, j int) bool {
// 		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
// 			return cands[i].Distance < cands[j].Distance
// 		}
// 		return cands[i].Score > cands[j].Score
// 	})
// 	limit := 10
// 	if len(cands) < limit {
// 		limit = len(cands)
// 	}
// 	out := make([]RecommendationWithDistance, limit)
// 	for i := 0; i < limit; i++ {
// 		out[i] = RecommendationWithDistance{
// 			UserID:   cands[i].ID,
// 			Distance: cands[i].Distance,
// 			Score:    cands[i].Score,
// 		}
// 	}
// 	return out, nil
// }
// func (rs *RecommendationService) GetRecommendationsWithFiltersWithDistance(
// 	currentUserID uuid.UUID,
// 	mode string,
// 	lat, lon float64,
// 	interests []string, prioInterests bool,
// 	hobbies []string, prioHobbies bool,
// 	music []string, prioMusic bool,
// 	food []string, prioFood bool,
// 	travel []string, prioTravel bool,
// 	lookingFor string,
// ) ([]RecommendationWithDistance, error) {
// 	if mode != "desire" {
// 		rs.Mode = "affinity"
// 	} else {
// 		rs.Mode = "desire"
// 	}
// 	var me models.User
// 	if err := rs.DB.
// 		Preload("Profile").
// 		Preload("Preference").
// 		First(&me, "id = ?", currentUserID).
// 		Error; err != nil {
// 		return nil, err
// 	}
// 	me.Profile.Latitude = lat
// 	me.Profile.Longitude = lon
// 	nearby, err := rs.GetNearbyUsers(
// 		lat, lon,
// 		me.Preference.MaxRadius,
// 		100,
// 		currentUserID,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(nearby) == 0 {
// 		return []RecommendationWithDistance{}, nil
// 	}
// 	ids := make([]uuid.UUID, len(nearby))
// 	distMap := make(map[uuid.UUID]float64, len(nearby))
// 	for i, n := range nearby {
// 		ids[i] = n.ID
// 		distMap[n.ID] = n.Distance
// 	}
// 	var users []models.User
// 	if err := rs.DB.Preload("Bio").
// 		Where("id IN ?", ids).
// 		Find(&users).Error; err != nil {
// 		return nil, err
// 	}
// 	meInts := splitTokens(strings.Join(interests, " "))
// 	meHobs := splitTokens(strings.Join(hobbies, " "))
// 	meMusic := splitTokens(strings.Join(music, " "))
// 	meFood := splitTokens(strings.Join(food, " "))
// 	meTrav := splitTokens(strings.Join(travel, " "))
// 	type outCand struct {
// 		ID       uuid.UUID
// 		Score    float64
// 		Distance float64
// 	}
// 	var cands []outCand
// 	for _, u := range users {
// 		var rec models.Recommendation
// 		if err := rs.DB.
// 			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
// 			First(&rec).Error; err == nil {
// 			continue
// 		}
// 		d := distMap[u.ID]
// 		var score float64
// 		if rs.Mode == "affinity" {
// 			totalW := 0.0
// 			w := 0.001
// 			if prioInterests {
// 				w *= 2
// 			}
// 			totalW += w
// 			score += float64(countCommon(meInts, splitTokens(u.Bio.Interests))) * w
// 			w = 0.001
// 			if prioHobbies {
// 				w *= 2
// 			}
// 			totalW += w
// 			score += float64(countCommon(meHobs, splitTokens(u.Bio.Hobbies))) * w
// 			w = 0.001
// 			if prioMusic {
// 				w *= 2
// 			}
// 			totalW += w
// 			score += float64(countCommon(meMusic, splitTokens(u.Bio.Music))) * w
// 			w = 0.001
// 			if prioFood {
// 				w *= 2
// 			}
// 			totalW += w
// 			score += float64(countCommon(meFood, splitTokens(u.Bio.Food))) * w
// 			w = 0.001
// 			if prioTravel {
// 				w *= 2
// 			}
// 			totalW += w
// 			score += float64(countCommon(meTrav, splitTokens(u.Bio.Travel))) * w

// 			if totalW > 0 {
// 				score /= totalW
// 			}
// 		} else {
// 			for _, tok := range splitTokens(lookingFor) {
// 				if anyTokenMatch(tok, u.Bio.LookingFor) {
// 					score += 0.05
// 				}
// 			}
// 		}
// 		if score > 1 {
// 			score = 1
// 		}
// 		if score > 0 {
// 			cands = append(cands, outCand{ID: u.ID, Score: score, Distance: d})
// 		}
// 	}
// 	sort.Slice(cands, func(i, j int) bool {
// 		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
// 			return cands[i].Distance < cands[j].Distance
// 		}
// 		return cands[i].Score > cands[j].Score
// 	})
// 	limit := 10
// 	if len(cands) < limit {
// 		limit = len(cands)
// 	}
// 	out := make([]RecommendationWithDistance, limit)
// 	for i := 0; i < limit; i++ {
// 		out[i] = RecommendationWithDistance{
// 			UserID:   cands[i].ID,
// 			Distance: cands[i].Distance,
// 			Score:    cands[i].Score,
// 		}
// 	}
// 	return out, nil
// }
// func (rs *RecommendationService) DeclineRecommendation(
// 	currentUserID, recUserID uuid.UUID,
// ) error {
// 	var existing models.Recommendation
// 	if err := rs.DB.
// 		Where("user_id = ? AND rec_user_id = ?", currentUserID, recUserID).
// 		First(&existing).Error; err == nil {
// 		existing.Status = "declined"
// 		return rs.DB.Save(&existing).Error
// 	}
// 	rec := models.Recommendation{
// 		UserID:    currentUserID,
// 		RecUserID: recUserID,
// 		Status:    "declined",
// 	}
// 	return rs.DB.Create(&rec).Error
// }
// func anyTokenMatch(a, b string) bool {
// 	for _, t := range splitTokens(a) {
// 		for _, u := range splitTokens(b) {
// 			if t == u {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
// func validateUserData(u models.User) error {
// 	if u.Profile.ID == 0 || u.Bio.ID == 0 {
// 		return errors.New("заполните профиль и биографию для получения рекомендаций")
// 	}
// 	if u.Profile.FirstName == "" || u.Profile.LastName == "" {
// 		return errors.New("пожалуйста, укажите имя и фамилию")
// 	}
// 	must := []struct{ val, msg string }{
// 		{u.Bio.Interests, "интересы"},
// 		{u.Bio.Hobbies, "хобби"},
// 		{u.Bio.Music, "музыку"},
// 		{u.Bio.Food, "еду"},
// 		{u.Bio.Travel, "путешествия"},
// 		{u.Bio.LookingFor, "кого вы ищете"},
// 	}
// 	var miss []string
// 	for _, f := range must {
// 		if strings.TrimSpace(f.val) == "" {
// 			miss = append(miss, f.msg)
// 		}
// 	}
// 	if len(miss) > 0 {
// 		return errors.New("пожалуйста, заполните вашу биографию: " + strings.Join(miss, ", "))
// 	}
// 	return nil
// }

package services

import (
	"errors"
	"fmt"
	"m/backend/models"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Nearby struct {
	ID       uuid.UUID
	Distance float64
}

type FieldConfig struct {
	Name      string
	Weight    float64
	Extractor func(b models.Bio) string
}

type RecommendationWithDistance struct {
	UserID   uuid.UUID `json:"id"`
	Distance float64   `json:"distance"`
	Score    float64   `json:"score"`
}

type candidate struct {
	User     models.User
	Score    float64
	Distance float64
}

type RecommendationService struct {
	DB           *gorm.DB
	FieldConfigs []FieldConfig
	Mode         string
}

func NewRecommendationService(db *gorm.DB, fieldConfigs []FieldConfig) *RecommendationService {
	rs := &RecommendationService{DB: db, Mode: "affinity"}
	if len(fieldConfigs) == 0 {
		fieldConfigs = []FieldConfig{
			{Name: "Interests", Weight: 0.02, Extractor: func(b models.Bio) string { return b.Interests }},
			{Name: "Hobbies", Weight: 0.02, Extractor: func(b models.Bio) string { return b.Hobbies }},
			{Name: "Music", Weight: 0.02, Extractor: func(b models.Bio) string { return b.Music }},
			{Name: "Food", Weight: 0.02, Extractor: func(b models.Bio) string { return b.Food }},
			{Name: "Travel", Weight: 0.02, Extractor: func(b models.Bio) string { return b.Travel }},
		}
	}
	rs.FieldConfigs = fieldConfigs
	logrus.Infof("RecommendationService initialized with mode=%s", rs.Mode)
	return rs
}

func splitTokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
}

func countCommon(a, b []string) int {
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[x] = struct{}{}
	}
	cnt := 0
	for _, y := range b {
		if _, ok := set[y]; ok {
			cnt++
		}
	}
	return cnt
}

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
		// SELECT
		//   user_id   AS id,
		//   earth_distance(earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
		// FROM profiles
		// WHERE
		//   earth_box(ll_to_earth(?, ?), ?) @> earth_loc
		//   AND earth_distance(earth_loc, ll_to_earth(?, ?)) <= ?
		//   AND user_id != ?
		// ORDER BY distance ASC
		// LIMIT ?
		Raw(`
			SELECT
			p.user_id   AS id,
			earth_distance(p.earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
		  FROM profiles p
		  WHERE 
			earth_box(ll_to_earth(?, ?), ?) @> p.earth_loc
			AND earth_distance(p.earth_loc, ll_to_earth(?, ?)) <= ?
			AND p.user_id != ?
			AND NOT EXISTS (
			  SELECT 1
			  FROM recommendations r
			  WHERE
				r.user_id       = ?
				AND r.rec_user_id = p.user_id
				AND r.status     = 'declined'
			)
		  ORDER BY distance ASC
		  LIMIT ?
        `,
			lat, lon,
			lat, lon, maxMeters,
			lat, lon, maxMeters,
			excludeID, excludeID, limit,
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

// func (rs *RecommendationService) GetNearbyUsers(
// 	lat, lon, maxRadius float64,
// 	limit int,
// 	excludeID uuid.UUID,
// ) ([]Nearby, error) {
// 	maxMeters := maxRadius * 1000.0
// 	fmt.Printf("[DEBUG] Searching nearby users from lat=%.6f, lon=%.6f, radius=%.2f km\n", lat, lon, maxRadius)
// 	fmt.Printf("[DEBUG] Excluding user ID: %s\n", excludeID)
// 	var list []Nearby
// 	err := rs.DB.
// 		Raw(`
//             SELECT
//               user_id   AS id,
//               earth_distance(earth_loc, ll_to_earth(?, ?)) / 1000.0 AS distance
//             FROM profiles
//             WHERE
//               earth_box(ll_to_earth(?, ?), ?) @> earth_loc
//               AND earth_distance(earth_loc, ll_to_earth(?, ?)) <= ?
// 			  AND user_id != ?
// 			  AND NOT EXISTS (
// 				SELECT 1 FROM recommendations
// 				WHERE
// 				  (
// 					(user_id = ? AND rec_user_id = profiles.user_id)
// 					OR
// 					(user_id = profiles.user_id AND rec_user_id = ?)
// 				  )
// 				  AND status = 'declined'
// 			)
//             ORDER BY distance ASC
//             LIMIT ?
//         `,
// 			lat, lon, // for SELECT and earth_box
// 			lat, lon, maxMeters, // for earth_box
// 			lat, lon, maxMeters, // for earth_distance
// 			excludeID,            // for user_id != ?
// 			excludeID, excludeID, // for NOT EXISTS subquery (2 times)
// 			limit,
// 		).
// 		Scan(&list).Error
// 	if err != nil {
// 		fmt.Printf("[ERROR] Nearby query failed: %v\n", err)
// 		return nil, err
// 	}
// 	fmt.Printf("[DEBUG] Nearby users found: %d\n", len(list))
// 	for _, n := range list {
// 		fmt.Printf("[DEBUG] User %s at %.2f km\n", n.ID, n.Distance)
// 	}
// 	return list, nil
// }

func (rs *RecommendationService) GetRecommendationsForUser(
	currentUserID uuid.UUID,
	mode string,
) ([]uuid.UUID, error) {
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	var me models.User
	if err := rs.DB.
		Preload("Profile").Preload("Bio").Preload("Preference").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		return nil, err
	}
	if err := validateUserData(me); err != nil {
		return nil, err
	}

	nearby, err := rs.GetNearbyUsers(
		me.Profile.Latitude,
		me.Profile.Longitude,
		me.Preference.MaxRadius,
		50,
		currentUserID,
	)
	if err != nil {
		return nil, err
	}
	if len(nearby) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}

	var users []models.User
	// if err := rs.DB.Preload("Bio").
	// 	Where("id IN ?", ids).
	// 	Find(&users).Error; err != nil {
	// 	return nil, err
	// }
	if err := rs.DB.Preload("Profile").Preload("Bio").
		Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	var cands []candidate
	for _, u := range users {
		// пропускаем отклонённые
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		if rs.Mode == "affinity" {
			for _, fc := range rs.FieldConfigs {
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

				// считаем совпадения подкатегорий
				setA := make(map[string]struct{})
				for _, t := range splitTokens(fc.Extractor(me.Bio)) {
					setA[t] = struct{}{}
				}
				common := 0
				for _, t := range splitTokens(fc.Extractor(u.Bio)) {
					if _, ok := setA[t]; ok {
						common++
					}
				}
				score += float64(common) * w
			}
		} else {
			for _, tok := range splitTokens(me.Bio.LookingFor) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.005
				}
			}
		}

		if score > 1 {
			score = 1
		}
		if score > 0 {
			cands = append(cands, candidate{User: u, Score: score, Distance: d})
		}
	}

	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

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

func (rs *RecommendationService) GetRecommendationsWithDistance(
	currentUserID uuid.UUID, mode string,
) ([]RecommendationWithDistance, error) {
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	var me models.User
	if err := rs.DB.
		Preload("Profile").Preload("Bio").Preload("Preference").
		First(&me, "id = ?", currentUserID).Error; err != nil {
		fmt.Printf("[ERROR] Failed to load user: %v\n", err)
		return nil, err
	}
	if err := validateUserData(me); err != nil {
		return nil, err
	}

	nearby, err := rs.GetNearbyUsers(
		me.Profile.Latitude,
		me.Profile.Longitude,
		me.Preference.MaxRadius,
		100,
		currentUserID,
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

	type outCand struct {
		ID       uuid.UUID
		Score    float64
		Distance float64
	}
	var cands []outCand

	for _, u := range users {
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		if rs.Mode == "affinity" {
			for _, fc := range rs.FieldConfigs {
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

				setA := make(map[string]struct{})
				for _, t := range splitTokens(fc.Extractor(me.Bio)) {
					setA[t] = struct{}{}
				}
				common := 0
				for _, t := range splitTokens(fc.Extractor(u.Bio)) {
					if _, ok := setA[t]; ok {
						common++
					}
				}
				score += float64(common) * w
			}
		} else {
			for _, tok := range splitTokens(me.Bio.LookingFor) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}

		if score > 1 {
			score = 1
		}
		if score > 0 {
			cands = append(cands, outCand{ID: u.ID, Score: score, Distance: d})
		}
	}

	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

	limit := 50
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

func (rs *RecommendationService) GetRecommendationsWithFiltersWithDistance(
	currentUserID uuid.UUID,
	mode string,
	lat, lon float64,
	interests []string, prioInterests bool,
	hobbies []string, prioHobbies bool,
	music []string, prioMusic bool,
	food []string, prioFood bool,
	travel []string, prioTravel bool,
	lookingFor string,
) ([]RecommendationWithDistance, error) {
	if mode != "desire" {
		rs.Mode = "affinity"
	} else {
		rs.Mode = "desire"
	}

	var me models.User
	if err := rs.DB.
		Preload("Profile").
		Preload("Preference").
		First(&me, "id = ?", currentUserID).
		Error; err != nil {
		return nil, err
	}
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

	meInts := splitTokens(strings.Join(interests, " "))
	meHobs := splitTokens(strings.Join(hobbies, " "))
	meMusic := splitTokens(strings.Join(music, " "))
	meFood := splitTokens(strings.Join(food, " "))
	meTrav := splitTokens(strings.Join(travel, " "))

	type outCand struct {
		ID       uuid.UUID
		Score    float64
		Distance float64
	}
	var cands []outCand

	for _, u := range users {
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		if rs.Mode == "affinity" {
			// Interests
			w := 0.02
			if prioInterests {
				w *= 2
			}
			score += float64(countCommon(meInts, splitTokens(u.Bio.Interests))) * w

			// Hobbies
			w = 0.02
			if prioHobbies {
				w *= 2
			}
			score += float64(countCommon(meHobs, splitTokens(u.Bio.Hobbies))) * w

			// Music
			w = 0.02
			if prioMusic {
				w *= 2
			}
			score += float64(countCommon(meMusic, splitTokens(u.Bio.Music))) * w

			// Food
			w = 0.02
			if prioFood {
				w *= 2
			}
			score += float64(countCommon(meFood, splitTokens(u.Bio.Food))) * w

			// Travel
			w = 0.02
			if prioTravel {
				w *= 2
			}
			score += float64(countCommon(meTrav, splitTokens(u.Bio.Travel))) * w
		} else {
			for _, tok := range splitTokens(lookingFor) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.05
				}
			}
		}

		if score > 1 {
			score = 1
		}
		if score > 0 {
			cands = append(cands, outCand{ID: u.ID, Score: score, Distance: d})
		}
	}

	sort.Slice(cands, func(i, j int) bool {
		if math.Abs(cands[i].Distance-cands[j].Distance) > 1e-9 {
			return cands[i].Distance < cands[j].Distance
		}
		return cands[i].Score > cands[j].Score
	})

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
	return out, nil
}

// DeclineRecommendation позволяет пометить рекомендацию как "declined".
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

// anyTokenMatch проверяет, есть ли хотя бы один совпадающий токен
func anyTokenMatch(a, b string) bool {
	for _, t := range splitTokens(a) {
		for _, u := range splitTokens(b) {
			if t == u {
				return true
			}
		}
	}
	return false
}

// validateUserData гарантирует, что у пользователя заполнены все необходимые поля для рекомендаций
func validateUserData(u models.User) error {
	// профиль и биография должны существовать
	if u.Profile.ID == 0 || u.Bio.ID == 0 {
		return errors.New("заполните профиль и биографию для получения рекомендаций")
	}
	// обязательны имя и фамилия
	if strings.TrimSpace(u.Profile.FirstName) == "" || strings.TrimSpace(u.Profile.LastName) == "" {
		return errors.New("пожалуйста, укажите имя и фамилию")
	}
	// проверяем, что в био заполнены все разделы
	required := []struct {
		val string
		msg string
	}{
		{u.Bio.Interests, "интересы"},
		{u.Bio.Hobbies, "хобби"},
		{u.Bio.Music, "музыку"},
		{u.Bio.Food, "еду"},
		{u.Bio.Travel, "путешествия"},
		{u.Bio.LookingFor, "кого вы ищете"},
	}
	var missing []string
	for _, field := range required {
		if strings.TrimSpace(field.val) == "" {
			missing = append(missing, field.msg)
		}
	}
	if len(missing) > 0 {
		return errors.New("пожалуйста, заполните вашу биографию: " + strings.Join(missing, ", "))
	}
	return nil
}
