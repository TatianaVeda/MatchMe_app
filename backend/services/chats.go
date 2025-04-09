package services

import (
	"encoding/json"
	"errors"
	"time"

	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ChatService инкапсулирует бизнес-логику для работы с чатами.
type ChatService struct {
	DB *gorm.DB
}

// NewChatService возвращает новый экземпляр ChatService.
func NewChatService(db *gorm.DB) *ChatService {
	logrus.Info("ChatService initialized")
	return &ChatService{DB: db}
}

// CreateChat создает уникальный чат между двумя пользователями, если он не существует.
func (cs *ChatService) CreateChat(user1ID, user2ID uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	if err := cs.DB.
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			user1ID, user2ID, user2ID, user1ID).
		First(&chat).Error; err == nil {
		logrus.Debugf("CreateChat: чат уже существует между %s и %s", user1ID, user2ID)
		return &chat, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Errorf("CreateChat: ошибка при поиске чата: %v", err)
		return nil, err
	}

	chat = models.Chat{
		User1ID:   user1ID,
		User2ID:   user2ID,
		CreatedAt: time.Now(),
	}
	if err := cs.DB.Create(&chat).Error; err != nil {
		logrus.Errorf("CreateChat: ошибка создания нового чата: %v", err)
		return nil, err
	}
	logrus.Infof("CreateChat: новый чат создан между %s и %s с ID %d", user1ID, user2ID, chat.ID)
	return &chat, nil
}

// GetChatMessages возвращает сообщения для указанного чата с пагинацией.
func (cs *ChatService) GetChatMessages(chatID uint, page, limit int) ([]models.Message, error) {
	offset := (page - 1) * limit
	var messages []models.Message
	if err := cs.DB.
		Where("chat_id = ?", chatID).
		Order("timestamp asc").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		logrus.Errorf("GetChatMessages: ошибка получения сообщений для чата %d: %v", chatID, err)
		return nil, err
	}
	logrus.Debugf("GetChatMessages: получено %d сообщений для чата %d", len(messages), chatID)
	return messages, nil
}

// TypingNotification формирует уведомление о наборе текста для передачи через WebSocket.
func TypingNotification(userID uuid.UUID, chatID uint, isTyping bool) ([]byte, error) {
	notification := map[string]interface{}{
		"type":      "typing",
		"user_id":   userID.String(),
		"chat_id":   chatID,
		"is_typing": isTyping,
		"timestamp": time.Now().Unix(),
	}
	data, err := json.Marshal(notification)
	if err != nil {
		logrus.Errorf("TypingNotification: ошибка маршалинга уведомления: %v", err)
		return nil, err
	}
	logrus.Debugf("TypingNotification: уведомление о наборе текста сформировано для пользователя %s в чате %d", userID, chatID)
	return data, nil
}
