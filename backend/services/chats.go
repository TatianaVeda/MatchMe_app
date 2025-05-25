package services

import (
	"encoding/json"
	"errors"
	"m/backend/models"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ChatService struct {
	DB *gorm.DB
}

func NewChatService(db *gorm.DB) *ChatService {
	logrus.Info("ChatService initialized")
	return &ChatService{DB: db}
}

func (cs *ChatService) CreateChat(user1ID, user2ID uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	if err := cs.DB.
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			user1ID, user2ID, user2ID, user1ID).
		First(&chat).Error; err == nil {
		logrus.Debugf("CreateChat: chat already exists between %s and %s", user1ID, user2ID)
		return &chat, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Errorf("CreateChat: error searching for chat: %v", err)
		return nil, err
	}

	chat = models.Chat{
		User1ID:   user1ID,
		User2ID:   user2ID,
		CreatedAt: time.Now(),
	}
	if err := cs.DB.Create(&chat).Error; err != nil {
		logrus.Errorf("CreateChat: error creating new chat: %v", err)
		return nil, err
	}
	logrus.Infof("CreateChat: new chat created between %s and %s with ID %d", user1ID, user2ID, chat.ID)
	return &chat, nil
}

func (cs *ChatService) GetChatMessages(chatID uint, page, limit int) ([]models.Message, error) {
	offset := (page - 1) * limit
	var messages []models.Message
	if err := cs.DB.
		Where("chat_id = ?", chatID).
		Order("timestamp asc").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		logrus.Errorf("GetChatMessages: error fetching messages for chat %d: %v", chatID, err)
		return nil, err
	}
	logrus.Debugf("GetChatMessages: %d messages fetched for chat %d", len(messages), chatID)
	return messages, nil
}

func TypingNotification(userID uuid.UUID, chatID uint, isTyping bool) ([]byte, error) {
	notification := map[string]interface{}{
		"type":      "typing",
		"userId":    userID.String(),
		"chatId":    chatID,
		"isTyping":  isTyping,
		"timestamp": time.Now().Unix(),
	}
	data, err := json.Marshal(notification)
	if err != nil {
		logrus.Errorf("TypingNotification: error marshaling notification: %v", err)
		return nil, err
	}
	logrus.Debugf("TypingNotification: typing notification created for user %s in chat %d", userID, chatID)
	return data, nil
}
