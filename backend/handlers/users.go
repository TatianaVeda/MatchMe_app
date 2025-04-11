package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// validateUserAccess checks whether the requester can access the target user's data.
func validateUserAccess(requesterID, targetID uint) bool {
	if requesterID == targetID {
		return true
	}

	var count int64
	database.DB.Raw(`
		SELECT COUNT(*) FROM connection_requests 
		WHERE (sender_id = ? AND receiver_id = ? OR sender_id = ? AND receiver_id = ?) 
		AND status IN ('connected', 'pending') 
		UNION ALL
		SELECT COUNT(*) FROM recommendations 
		WHERE user_id = ? AND recommended_user_id = ?
	`, requesterID, targetID, targetID, requesterID, requesterID, targetID).Scan(&count)

	return count > 0
}

// GetUserNameAndPic - Returns basic user info (username & avatar)
func GetUserNameAndPic(c *gin.Context) {
	requester, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if !validateUserAccess(requester.(models.User).ID, uint(userID)) {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	var user struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}

	if err := database.DB.Table("users").Select("id, username, avatar").Where("id = ?", userID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	logResponse("GetUserNameAndPic", user)
	c.JSON(http.StatusOK, user)
}

// GetUserInfo - Returns detailed user profile information
func GetUserInfo(c *gin.Context) {
	requester, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if !validateUserAccess(requester.(models.User).ID, uint(userID)) {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	var user struct {
		ID        uint   `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Location  string `json:"location"`
		AboutMe   string `json:"aboutme"`
	}

	if err := database.DB.Table("users").Select("id, first_name, last_name, location, about_me").Where("id = ?", userID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	logResponse("GetUserInfo", user)
	c.JSON(http.StatusOK, user)
}

// GetUserFavs - Returns user's favorite movies, genres, etc.
func GetUserFavs(c *gin.Context) {
	requester, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if !validateUserAccess(requester.(models.User).ID, uint(userID)) {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	var user struct {
		ID                uint   `json:"id"`
		FavoriteGenre     string `json:"favorite_genre"`
		FavoriteMovie     string `json:"favorite_movie"`
		FavoriteDirector  string `json:"favorite_director"`
		FavoriteActor     string `json:"favorite_actor"`
		FavoriteActress   string `json:"favorite_actress"`
	}

	if err := database.DB.Table("users").Select("id, favorite_genre, favorite_movie, favorite_director, favorite_actor, favorite_actress").Where("id = ?", userID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error 404": "User not found"})
		return
	}

	logResponse("GetUserFavs", user)
	c.JSON(http.StatusOK, user)
}

// logResponse formats and prints the JSON response
func logResponse(endpoint string, data interface{}) {
	responseJSON, _ := json.MarshalIndent(data, "", "  ")
	fmt.Printf("%s Response JSON: %s\n", endpoint, string(responseJSON))
}
