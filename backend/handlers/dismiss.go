package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DismissUser(c *gin.Context) {
	var req struct {
		DismissedUserID uint `json:"dismissed_user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	currentUser, exists := getCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	dismissedUser := models.DismissedUser{
		UserID:          currentUser.ID,
		DismissedUserID: req.DismissedUserID,
	}

	if err := database.DB.Create(&dismissedUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not dismiss user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User dismissed successfully"})
}
