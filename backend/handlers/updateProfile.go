// backend/handlers/updateProfile.go
package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateUserProfile updates the profile information of the authenticated user.
func UpdateUserProfile(c *gin.Context) {
	var userInput models.User
	// Bind the incoming JSON request to the User struct
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the currently authenticated user
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user data"})
		return
	}

	// Type assertion to ensure we are working with a User object
	user := currentUser.(models.User)

	// Define predefined locations and their latitudes/longitudes
	availableLocations := []map[string]interface{}{
		{"city": "Helsinki", "latitude": 60.1695, "longitude": 24.9354},
		{"city": "Espoo", "latitude": 60.2055, "longitude": 24.6559},
		{"city": "Tampere", "latitude": 61.4978, "longitude": 23.7610},
		// Additional locations...
	}

	// If the location has changed, update the latitude and longitude values
	if user.Location != userInput.Location {
		for _, loc := range availableLocations {
			if loc["city"] == userInput.Location {
				user.Latitude = loc["latitude"].(float64)
				user.Longitude = loc["longitude"].(float64)
				break
			}
		}
	}

	// Update the user's profile with the new data
	user.FirstName = userInput.FirstName
	user.LastName = userInput.LastName
	user.Location = userInput.Location
	user.AboutMe = userInput.AboutMe
	user.Avatar = userInput.Avatar
	user.FavoriteGenre = userInput.FavoriteGenre
	user.FavoriteMovie = userInput.FavoriteMovie
	user.FavoriteDirector = userInput.FavoriteDirector
	user.FavoriteActor = userInput.FavoriteActor
	user.FavoriteActress = userInput.FavoriteActress

	// Save the updated user data in the database
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile successfully updated"})
}
