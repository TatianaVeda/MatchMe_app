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

/*
RecommendationService — Core Recommendation Algorithm Overview
--------------------------------------------------------------
This service provides user-to-user recommendations based on profile similarity, preferences, and geolocation.

Key Principles:
- Two modes: "affinity" (profile similarity, weighted fields) and "desire" (matching by 'LookingFor').
- Geospatial filtering: Only users within a preferred radius (using PostgreSQL earthdistance/cube).
- Score calculation: Weighted overlap of interests, hobbies, music, food, travel (weights can be doubled by user priorities).
- Filtering: Excludes declined users and those with incomplete profiles.
- Sorting: Recommendations are sorted by distance (ascending), then by score (descending).
- Extensible: Field weights and extractors are configurable for future algorithm tuning.

Typical Flow:
1. Load current user with profile, bio, and preferences.
2. Find nearby users within radius, excluding declined.
3. For each candidate, calculate score based on field overlap and priorities.
4. Filter out candidates with zero score.
5. Sort by distance, then score. Limit output.

See also:
- GetRecommendationsForUser: Main entry for recommendations.
- GetRecommendationsWithFiltersWithDistance: Advanced search with custom filters.
- validateUserData: Ensures profile completeness.
*/

// Used internally for sorting and filtering nearby users.
type Nearby struct {
	ID       uuid.UUID
	Distance float64
}

// FieldConfig defines a field and its weight for recommendation scoring.
// Each field has a name, weight (importance in scoring), and an extractor function
// that retrieves the relevant data from a user's Bio.
type FieldConfig struct {
	Name      string
	Weight    float64
	Extractor func(b models.Bio) string
}

// RecommendationWithDistance contains a recommended user with their distance
// and match score. Used for API responses and client-side display.
type RecommendationWithDistance struct {
	UserID   uuid.UUID `json:"id"`
	Distance float64   `json:"distance"`
	Score    float64   `json:"score"`
}

// candidate is an internal struct for scoring and sorting candidates.
// Contains the full user model along with calculated score and distance.
type candidate struct {
	User     models.User
	Score    float64
	Distance float64
}

// RecommendationService provides methods for generating user recommendations
// based on profile data, preferences, and geolocation.
// Supports different recommendation modes (affinity/desire) and customizable
// field weights for scoring.
type RecommendationService struct {
	DB           *gorm.DB
	FieldConfigs []FieldConfig
	Mode         string
}

// NewRecommendationService creates a new RecommendationService with optional
// custom field configurations. If no configs are provided, uses default weights
// for interests, hobbies, music, food, and travel preferences.
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

// GetNearbyUsers returns a list of users within maxRadius km from the given coordinates, excluding the specified user.
func (rs *RecommendationService) GetNearbyUsers(
	lat, lon, maxRadius float64,
	limit int,
	excludeID uuid.UUID,
) ([]Nearby, error) {
	// Convert radius to meters for geospatial query
	maxMeters := maxRadius * 1000.0
	fmt.Printf("[DEBUG] Searching nearby users from lat=%.6f, lon=%.6f, radius=%.2f km\n", lat, lon, maxRadius)
	fmt.Printf("[DEBUG] Excluding user ID: %s\n", excludeID)
	var list []Nearby
	// Use PostgreSQL earthdistance/cube extensions for efficient geospatial search
	// Exclude users who have been declined in recommendations
	err := rs.DB.
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

// GetRecommendationsForUser returns a list of recommended user IDs for the given user, using either affinity or desire mode.
// Recommendations are sorted by distance and match score.
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

	// Find nearby users within the user's preferred radius
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

	// Map user IDs to distances for later scoring
	ids := make([]uuid.UUID, len(nearby))
	distMap := make(map[uuid.UUID]float64, len(nearby))
	for i, n := range nearby {
		ids[i] = n.ID
		distMap[n.ID] = n.Distance
	}

	var users []models.User

	// Preload profiles and bios for all nearby users
	if err := rs.DB.Preload("Profile").Preload("Bio").
		Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	var cands []candidate
	for _, u := range users {
		// Skip users who have been declined
		var rec models.Recommendation
		if err := rs.DB.
			Where("user_id = ? AND rec_user_id = ? AND status = ?", currentUserID, u.ID, "declined").
			First(&rec).Error; err == nil {
			continue
		}

		d := distMap[u.ID]
		var score float64

		if rs.Mode == "affinity" {
			// Affinity mode: score by field similarity and user preferences
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

				// Tokenize and compare fields for overlap
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
			// Desire mode: match by 'LookingFor' field
			for _, tok := range splitTokens(me.Bio.LookingFor) {
				if anyTokenMatch(tok, u.Bio.LookingFor) {
					score += 0.005
				}
			}
		}

		// Cap score at 1
		if score > 1 {
			score = 1
		}
		if score > 0 {
			cands = append(cands, candidate{User: u, Score: score, Distance: d})
		}
	}

	// Sort by distance (asc), then by score (desc)
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

// GetRecommendationsWithDistance returns recommendations with distance and score for the given user.
// Used for displaying recommendations with additional info.
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

// GetRecommendationsWithFiltersWithDistance returns recommendations using custom filters (interests, hobbies, etc.) and location.
// Used for advanced search and filtering in recommendations.
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

// DeclineRecommendation marks a recommendation as declined for the current user.
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

// anyTokenMatch checks if any token from a matches b (case-insensitive substring).
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

// validateUserData checks if the user profile and bio are sufficiently filled for recommendations.
func validateUserData(u models.User) error {
	if u.Profile.ID == 0 || u.Bio.ID == 0 {
		return errors.New("please fill in your profile and biography to get recommendations")
	}
	if strings.TrimSpace(u.Profile.FirstName) == "" || strings.TrimSpace(u.Profile.LastName) == "" {
		return errors.New("please provide your first and last name")
	}
	required := []struct {
		val string
		msg string
	}{
		{u.Bio.Interests, "interests"},
		{u.Bio.Hobbies, "hobbies"},
		{u.Bio.Music, "music"},
		{u.Bio.Food, "food"},
		{u.Bio.Travel, "travel"},
		{u.Bio.LookingFor, "who you are looking for"},
	}
	var missing []string
	for _, field := range required {
		if strings.TrimSpace(field.val) == "" {
			missing = append(missing, field.msg)
		}
	}
	if len(missing) > 0 {
		return errors.New("please complete your biography: " + strings.Join(missing, ", "))
	}
	return nil
}
