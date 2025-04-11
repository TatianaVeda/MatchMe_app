package handlers

import (
	"backend/database"
	"backend/models"
	"backend/websocket"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handle message sending
func SendChatMessage(c *gin.Context) {
	var message models.Chat
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, authenticated := c.Get("currentUser")
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	message.SenderID = user.(models.User).ID
	message.Timestamp = time.Now()

	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save message"})
		return
	}

	// Broadcast via WebSocket
	go func() {
		wsMessage := websocket.Message{
			ID:         message.ID,
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			Message:    message.Message,
			Timestamp:  message.Timestamp.Format(time.RFC3339),
		}
		websocket.Broadcast <- wsMessage
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Message sent", "chat": message})
}

// Retrieve chat messages
func FetchChatMessages(c *gin.Context) {
	receiverID, err := strconv.Atoi(c.Query("receiver_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	user, authenticated := c.Get("currentUser")
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	userID := user.(models.User).ID

	var messages []models.Chat
	dbQuery := database.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, receiverID, receiverID, userID,
	).Order("timestamp DESC").Limit(limit).Offset(offset).Find(&messages)

	if dbQuery.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": dbQuery.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Get timestamp of the last exchanged message
func RetrieveLastMessageTimestamp(c *gin.Context) {
	receiverID, err := strconv.Atoi(c.Query("receiver_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	user, authenticated := c.Get("currentUser")
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	userID := user.(models.User).ID

	var lastMessage models.Chat
	dbQuery := database.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, receiverID, receiverID, userID,
	).Order("timestamp DESC").First(&lastMessage)

	if dbQuery.Error != nil {
		if errors.Is(dbQuery.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"lastMessageTimestamp": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": dbQuery.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lastMessageTimestamp": lastMessage.Timestamp})
}

// Mark messages as read
func MarkMessagesRead(c *gin.Context) {
	senderID, err := strconv.Atoi(c.Query("sender_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender ID"})
		return
	}

	user, authenticated := c.Get("currentUser")
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	receiverID := user.(models.User).ID

	updateQuery := database.DB.Model(&models.Chat{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", senderID, receiverID, false).
		Update("is_read", true)

	if updateQuery.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read", "count": updateQuery.RowsAffected})
}

// Get unread message counts per sender
func FetchUnreadMessageCounts(c *gin.Context) {
	user, authenticated := c.Get("currentUser")
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User authentication required"})
		return
	}
	receiverID := user.(models.User).ID

	type UnreadMessages struct {
		SenderID uint `json:"sender_id"`
		Count    int  `json:"count"`
	}

	var unreadList []UnreadMessages
	dbQuery := database.DB.Model(&models.Chat{}).
		Select("sender_id, count(*) as count").
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Group("sender_id").Find(&unreadList)

	if dbQuery.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve unread messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_counts": unreadList})
}
