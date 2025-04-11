// backend/handlers/updateAvatar.go
package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 1 * 1024 * 1024 // Maximum file size set to 1 MB

// UpdateUserAvatar handles updating the avatar for the current user.
func UpdateUserAvatar(c *gin.Context) {
	// Retrieve the uploaded avatar file from the form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to process file"})
		return
	}

	// Ensure the file does not exceed the maximum allowed size
	if file.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File exceeds the 1 MB limit"})
		return
	}

	// Fetch the current user from the context
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve current user"})
		return
	}

	user := currentUser.(models.User)

	// Ensure the uploads directory exists, create it if not
	uploadDirectory := "uploads/avatars"
	if err := os.MkdirAll(uploadDirectory, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Define the file path for saving the avatar
	filePath := filepath.Join(uploadDirectory, fmt.Sprintf("%d.png", user.ID))

	// Save the file to the server
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
		return
	}

	// Update the user's avatar URL in the database
	user.Avatar = fmt.Sprintf("/uploads/avatars/%d.png", user.ID)

	// Save the updated user profile in the database
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar URL"})
		return
	}

	// Fetch the updated user profile from the database
	var updatedUser models.User
	if err := database.DB.First(&updatedUser, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated user"})
		return
	}

	// Respond with the updated avatar URL
	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar updated successfully",
		"avatar":  updatedUser.Avatar,
	})
}

// ResetUserAvatar handles resetting the avatar to the default avatar.
func ResetUserAvatar(c *gin.Context) {
	// Fetch the current user from the context
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve current user"})
		return
	}

	user := currentUser.(models.User)

	// Define the path to the avatar file
	avatarFilePath := filepath.Join("uploads/avatars", fmt.Sprintf("%d.png", user.ID))

	// Check if the avatar file exists and delete it if found
	if _, err := os.Stat(avatarFilePath); err == nil {
		os.Remove(avatarFilePath)
	}

	// Reset to the default avatar URL
	user.Avatar = "/uploads/avatars/default.png"

	// Save the reset avatar information to the database
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset avatar"})
		return
	}

	// Return a successful response with the default avatar URL
	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar successfully reset to default",
		"avatar":  user.Avatar,
	})
}
