package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"m/backend/models"
	"m/backend/sockets" // Предполагается, что здесь реализована логика WebSocket оповещений.

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Глобальное подключение к базе данных для работы с чатами.
var chatsDB *gorm.DB

// InitChatsController устанавливает подключение к базе данных для работы с чатами.
func InitChatsController(db *gorm.DB) {
	chatsDB = db
	logrus.Info("Chats controller initialized")
}

// ChatSummary представляет сводную информацию о чате.
type ChatSummary struct {
	ChatID      uint           `json:"chat_id"`
	OtherUserID uuid.UUID      `json:"other_user_id"`
	LastMessage MessageSummary `json:"last_message"`
	UnreadCount int            `json:"unread_count"`
}

// MessageSummary представляет сводную информацию о последнем сообщении.
type MessageSummary struct {
	ID        uint      `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// GET /chats
// Возвращает список всех чатов, где для каждого:
// - определяется идентификатор другого участника,
// - извлекается последнее сообщение,
// - считается количество непрочитанных сообщений.
func GetChats(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetChats: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetChats: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var chats []models.Chat
	if err := chatsDB.
		Where("user1_id = ? OR user2_id = ?", currentUserID, currentUserID).
		Find(&chats).Error; err != nil {
		logrus.Errorf("GetChats: ошибка получения чатов: %v", err)
		http.Error(w, "Error fetching chats", http.StatusInternalServerError)
		return
	}

	summaries := make([]ChatSummary, 0, len(chats))
	for _, chat := range chats {
		var otherUserID uuid.UUID
		if chat.User1ID == currentUserID {
			otherUserID = chat.User2ID
		} else {
			otherUserID = chat.User1ID
		}

		var lastMsg models.Message
		if err := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ?", chat.ID).
			Order("timestamp desc").
			Limit(1).
			First(&lastMsg).Error; err != nil && err != gorm.ErrRecordNotFound {
			logrus.Errorf("GetChats: ошибка получения последнего сообщения для чата %d: %v", chat.ID, err)
			http.Error(w, "Error fetching last message", http.StatusInternalServerError)
			return
		}
		lastMessageSummary := MessageSummary{}
		if lastMsg.ID != 0 {
			lastMessageSummary = MessageSummary{
				ID:        lastMsg.ID,
				SenderID:  lastMsg.SenderID,
				Content:   lastMsg.Content,
				Timestamp: lastMsg.Timestamp,
				Read:      lastMsg.Read,
			}
		}

		var unreadCount int64
		if err := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
			Count(&unreadCount).Error; err != nil {
			logrus.Errorf("GetChats: ошибка подсчёта непрочитанных сообщений для чата %d: %v", chat.ID, err)
			http.Error(w, "Error counting unread messages", http.StatusInternalServerError)
			return
		}

		summaries = append(summaries, ChatSummary{
			ChatID:      chat.ID,
			OtherUserID: otherUserID,
			LastMessage: lastMessageSummary,
			UnreadCount: int(unreadCount),
		})
	}

	logrus.Infof("GetChats: получено %d чатов для пользователя %s", len(summaries), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// GET /chats/{chat_id}?page=1&limit=20
// Возвращает историю сообщений чата с пагинацией и обновляет статус сообщений.
func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetChatHistory: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetChatHistory: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	chatIDStr := vars["chat_id"]
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		logrus.Errorf("GetChatHistory: неверный chat_id: %v", err)
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	var chat models.Chat
	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
		logrus.Errorf("GetChatHistory: чат %d не найден: %v", chatID, err)
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}
	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
		logrus.Warnf("GetChatHistory: пользователь %s не является участником чата %d", currentUserID, chatID)
		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page := 1
	limit := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := (page - 1) * limit

	var messages []models.Message
	if err := chatsDB.
		Where("chat_id = ?", chat.ID).
		Order("timestamp asc").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		logrus.Errorf("GetChatHistory: ошибка получения сообщений для чата %d: %v", chatID, err)
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}

	// Фоновое обновление: помечаем как прочитанные все сообщения, отправленные не текущим пользователем.
	go func() {
		if err := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
			Update("read", true).Error; err != nil {
			logrus.Errorf("GetChatHistory: ошибка обновления статуса сообщений для чата %d: %v", chat.ID, err)
		}
	}()

	logrus.Infof("GetChatHistory: получена история сообщений для чата %d (страница %d, лимит %d)", chatID, page, limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// POST /chats/{chat_id}/messages
// Создает новое сообщение в чате и уведомляет участников через WebSocket.
func PostMessage(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("PostMessage: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("PostMessage: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	chatIDStr := vars["chat_id"]
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		logrus.Errorf("PostMessage: неверный chat_id: %v", err)
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	var chat models.Chat
	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
		logrus.Errorf("PostMessage: чат %d не найден: %v", chatID, err)
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}
	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
		logrus.Warnf("PostMessage: пользователь %s не является участником чата %d", currentUserID, chatID)
		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
		return
	}

	var reqBody struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logrus.Errorf("PostMessage: ошибка декодирования запроса: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if reqBody.Content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	newMsg := models.Message{
		ChatID:    uint(chatID),
		SenderID:  currentUserID,
		Content:   reqBody.Content,
		Timestamp: time.Now(),
		Read:      false,
	}
	if err := chatsDB.Create(&newMsg).Error; err != nil {
		logrus.Errorf("PostMessage: ошибка создания сообщения: %v", err)
		http.Error(w, "Error creating message", http.StatusInternalServerError)
		return
	}

	go func(msg models.Message) {
		if err := sockets.BroadcastNewMessage(msg); err != nil {
			logrus.Errorf("PostMessage: ошибка отправки сообщения через WebSocket: %v", err)
		}
	}(newMsg)

	logrus.Infof("PostMessage: новое сообщение создано в чате %d отправителем %s", chatID, currentUserID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMsg)
}
