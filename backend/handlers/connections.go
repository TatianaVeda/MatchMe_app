package handlers

import (
	"backend/database"
	"backend/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendConnectionRequest handles sending a connection request.
func SendConnectionRequest(c *gin.Context) {
	var req struct {
		ReceiverID uint `json:"receiver_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	sender := currentUser.(models.User)

	// Check if the receiver has already sent a request to the sender
	var incomingRequest models.ConnectionRequest
	if err := database.DB.Where("sender_id = ? AND receiver_id = ?", req.ReceiverID, sender.ID).
		First(&incomingRequest).Error; err == nil {
		incomingRequest.Status = "connected"
		if err := database.DB.Save(&incomingRequest).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish connection"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Connection established successfully"})
		return
	}

	// Check if the sender already sent a request
	var existingRequest models.ConnectionRequest
	if err := database.DB.Where("sender_id = ? AND receiver_id = ?", sender.ID, req.ReceiverID).
		First(&existingRequest).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request already sent"})
		return
	}

	// Create a new pending request
	connectionRequest := models.ConnectionRequest{
		SenderID:   sender.ID,
		ReceiverID: req.ReceiverID,
		Status:     "pending",
	}

	if err := database.DB.Create(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request sent successfully"})
}

// GetConnectionRequests retrieves sent and received requests.
func GetConnectionRequests(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(models.User)

	var sentRequests, receivedRequests []models.ConnectionRequest
	database.DB.Where("sender_id = ?", user.ID).Find(&sentRequests)
	database.DB.Where("receiver_id = ?", user.ID).Find(&receivedRequests)

	c.JSON(http.StatusOK, gin.H{"sent": sentRequests, "received": receivedRequests})
}

// AcceptConnectionRequest marks a request as "connected".
func AcceptConnectionRequest(c *gin.Context) {
	var req struct {
		RequestID uint `json:"request_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(models.User)

	var connectionRequest models.ConnectionRequest
	if err := database.DB.Where("id = ? AND receiver_id = ?", req.RequestID, user.ID).
		First(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection request not found"})
		return
	}

	connectionRequest.Status = "connected"
	if err := database.DB.Save(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request accepted successfully"})
}

// GetConnectedUsers retrieves all users the current user is connected to.
func GetConnectedUsers(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(models.User)

	var connectionRequests []models.ConnectionRequest
	database.DB.Where("(sender_id = ? OR receiver_id = ?) AND status = ?", user.ID, user.ID, "connected").Find(&connectionRequests)

	userIDs := make([]uint, 0)
	for _, conn := range connectionRequests {
		if conn.SenderID == user.ID {
			userIDs = append(userIDs, conn.ReceiverID)
		} else {
			userIDs = append(userIDs, conn.SenderID)
		}
	}

	// Log JSON response
	responseJSON, _ := json.MarshalIndent(gin.H{"connections": userIDs}, "", "  ")
	fmt.Println("GetConnectedUsers Response JSON:", string(responseJSON))

	c.JSON(http.StatusOK, gin.H{"connections": userIDs})
}

// RemoveConnection deletes a connection and adds the user to dismissed users.
func RemoveConnection(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(models.User)

	database.DB.Create(&models.DismissedUser{UserID: user.ID, DismissedUserID: req.UserID})
	database.DB.Create(&models.DismissedUser{UserID: req.UserID, DismissedUserID: user.ID})

	if err := database.DB.Where(
		"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))",
		user.ID, req.UserID, req.UserID, user.ID,
	).Delete(&models.ConnectionRequest{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection removed successfully"})
}

// RejectConnectionRequest deletes a request and marks the sender as dismissed.
func RejectConnectionRequest(c *gin.Context) {
	var req struct {
		RequestID uint `json:"request_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUser.(models.User)

	var connectionRequest models.ConnectionRequest
	err := database.DB.Where("id = ? AND receiver_id = ?", req.RequestID, user.ID).
		First(&connectionRequest).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection request not found"})
		return
	}

	database.DB.Create(&models.DismissedUser{UserID: user.ID, DismissedUserID: connectionRequest.SenderID})

	if err := database.DB.Delete(&connectionRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request rejected successfully"})
}
