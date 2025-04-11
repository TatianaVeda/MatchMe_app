// backend/handlers/profile.go
package handlers

import (
	"net/http"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GetUserProfile retrieves basic user information (username, avatar).
func GetUserProfile(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user data"})
		return
	}

	// Ensure user is of correct type
	currentUser := user.(models.User)

	// Fetch user's avatar from database
	if err := database.DB.Model(&currentUser).Select("avatar").First(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user avatar"})
		return
	}

	// Check if avatar exists, if not, assign default
	if currentUser.Avatar == "" {
		currentUser.Avatar = "/uploads/avatars/default.png"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       currentUser.ID,
		"username": currentUser.Username,
		"avatar":   currentUser.Avatar,
	})
}

// GetUserProfileDetails retrieves additional profile details like name, location, and about me.
func GetUserProfileDetails(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user data"})
		return
	}

	// Ensure user is of correct type
	currentUser := user.(models.User)

	// Fetch detailed user profile from database
	if err := database.DB.Model(&currentUser).Select("first_name, last_name, location, about_me").First(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user profile details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"firstName": currentUser.FirstName,
		"lastName":  currentUser.LastName,
		"location":  currentUser.Location,
		"aboutMe":   currentUser.AboutMe,
	})
}

// GetUserFavorites retrieves the user's favorite genres, movies, and actors.
func GetUserFavorites(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user data"})
		return
	}

	// Ensure user is of correct type
	currentUser := user.(models.User)

	// Fetch user favorites from database
	if err := database.DB.Model(&currentUser).Select("favorite_genre, favorite_movie, favorite_director, favorite_actor, favorite_actress").First(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favoriteGenre":    currentUser.FavoriteGenre,
		"favoriteMovie":    currentUser.FavoriteMovie,
		"favoriteDirector": currentUser.FavoriteDirector,
		"favoriteActor":    currentUser.FavoriteActor,
		"favoriteActress":  currentUser.FavoriteActress,
	})
}

// GetUserEmail retrieves the email address of the authenticated user.
func GetUserEmail(c *gin.Context) {
	user, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user data"})
		return
	}

	// Ensure user is of correct type
	currentUser := user.(models.User)

	// Fetch user email from the database
	if err := database.DB.Model(&currentUser).Select("email").First(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": currentUser.Email,
	})
}
