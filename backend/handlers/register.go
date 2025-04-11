package handlers

import (
	"fmt"
	"net/http"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// RegisterInput represents the expected request body for user registration
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Avatar   string `json:"avatar"`
}

func RegisterUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already in use"})
		return
	}

	// Create user model
	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
		Avatar:   input.Avatar,
	}

	// Set default avatar if not provided
	if user.Avatar == "" {
		user.Avatar = "/uploads/avatars/default.png"
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Save user in database
	if err := database.DB.Create(&user).Error; err != nil {
		fmt.Printf("Could not register user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register user"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
		},
	})
}