package handlers

import (
	"backend/database"
	"backend/models"
	//"encoding/json"
	//"fmt"
	"math"
	"net/http"
	"sort"
	//"time"

	"github.com/gin-gonic/gin"
)

// Haversine formula to calculate distance between two coordinates
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c // Distance in km
}

// Calculate match score based on user preferences
func calculateMatchScore(currentUser, user models.User) int {
	score := 0
	if user.FavoriteGenre == currentUser.FavoriteGenre {
		score += 40
	}
	if user.FavoriteMovie == currentUser.FavoriteMovie {
		score += 30
	}
	if user.FavoriteDirector == currentUser.FavoriteDirector {
		score += 10
	}
	if user.FavoriteActor == currentUser.FavoriteActor {
		score += 10
	}
	if user.FavoriteActress == currentUser.FavoriteActress {
		score += 10
	}
	return score
}

// Helper to get the current user
func getCurrentUser(c *gin.Context) (models.User, bool) {
	currentUserInterface, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return models.User{}, false
	}
	return currentUserInterface.(models.User), true
}

// Find similar users and store recommendations in the database
func MatchUsers(c *gin.Context) {
	var users []models.User
	var excludedUserIDs []uint

	// Get the current user
	currentUser, exists := getCurrentUser(c)
	if !exists {
		return
	}

	// Remove old recommendations
	database.DB.Where("user_id = ?", currentUser.ID).Delete(&models.Recommendation{})

	// Exclude connected and dismissed users
	excludedUserIDs = getExcludedUserIDs(currentUser.ID)

	// Fetch users excluding dismissed & connected ones
	query := database.DB
	if len(excludedUserIDs) > 0 {
		query = query.Where("id NOT IN ?", excludedUserIDs)
	}
	if err := query.Limit(500).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Process matches
	matchedUsers := generateMatches(currentUser, users)

	// Store and return recommendations
	database.DB.Create(&matchedUsers)
	c.JSON(http.StatusOK, gin.H{"user_ids": extractUserIDs(matchedUsers)})
}

// Get stored recommendations for the user
func GetRecommendations(c *gin.Context) {
	currentUser, exists := getCurrentUser(c)
	if !exists {
		return
	}

	var recommendations []models.Recommendation
	database.DB.Where("user_id = ?", currentUser.ID).Order("score DESC").Limit(10).Find(&recommendations)
	c.JSON(http.StatusOK, gin.H{"user_ids": extractUserIDs(recommendations)})
}

// Set search radius for the current user
func SetRadius(c *gin.Context) {
	var req struct {
		Radius float64 `json:"radius"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Radius < 50 || req.Radius > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Radius must be between 50 and 500 km"})
		return
	}

	currentUser, exists := getCurrentUser(c)
	if !exists {
		return
	}

	database.DB.Model(&currentUser).Update("search_radius", req.Radius)
	c.JSON(http.StatusOK, gin.H{"message": "Search radius updated successfully", "radius": req.Radius})
}

// Utility functions
func getExcludedUserIDs(userID uint) []uint {
	var excludedIDs []uint
	var connectedUsers []uint
	database.DB.Raw(`
    SELECT DISTINCT sender_id FROM connection_requests 
    WHERE receiver_id = ? AND status IN ('pending', 'connected')
    UNION 
    SELECT DISTINCT receiver_id FROM connection_requests 
    WHERE sender_id = ? AND status IN ('pending', 'connected')
	`, userID, userID).Scan(&connectedUsers)

	excludedIDs = append(excludedIDs, connectedUsers...)
	database.DB.Table("dismissed_users").Where("user_id = ?", userID).Pluck("dismissed_user_id", &excludedIDs)
	database.DB.Table("dismissed_users").Where("dismissed_user_id = ?", userID).Pluck("user_id", &excludedIDs)

	return excludedIDs
}

func generateMatches(currentUser models.User, users []models.User) []models.Recommendation {
	var matches []models.Recommendation

	for _, user := range users {
		if user.ID == currentUser.ID {
			continue
		}

		distance := haversine(currentUser.Latitude, currentUser.Longitude, user.Latitude, user.Longitude)
		if distance > currentUser.SearchRadius {
			continue
		}

		score := calculateMatchScore(currentUser, user)
		if score > 0 {
			matches = append(matches, models.Recommendation{
				UserID:            currentUser.ID,
				RecommendedUserID: user.ID,
				Score:             score,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > 10 {
		matches = matches[:10]
	}

	return matches
}

func extractUserIDs(recommendations []models.Recommendation) []uint {
	var ids []uint
	for _, rec := range recommendations {
		ids = append(ids, rec.RecommendedUserID)
	}
	return ids
}
