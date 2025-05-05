package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
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
	ChatID      uint      `json:"chatId"`
	OtherUserID uuid.UUID `json:"otherUserId"`
	OtherUser   *struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"firstName"`
		LastName  string    `json:"lastName"`
		PhotoURL  string    `json:"photoUrl"`
	} `json:"otherUser"`
	LastMessage     MessageSummary `json:"lastMessage"`
	UnreadCount     int            `json:"unreadCount"`
	OtherUserOnline bool           `json:"otherUserOnline"`
	IsTyping        bool           `json:"isTyping"`
	ChatCreatedAt   time.Time      `json:"-"` // Используется для сортировки, но не возвращается клиенту
}

// MessageSummary представляет сводную информацию о последнем сообщении.
type MessageSummary struct {
	ID        uint      `json:"id"`
	SenderID  uuid.UUID `json:"senderId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// GET /chats
// Возвращает список всех чатов, где для каждого:
// - определяется идентификатор другого участника,
// - извлекается последнее сообщение,
// - считается количество непрочитанных сообщений.
// func GetChats(w http.ResponseWriter, r *http.Request) {
// 	userIDStr, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		logrus.Error("GetChats: userID не найден в контексте")
// 		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUserID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		logrus.Errorf("GetChats: неверный userID: %v", err)
// 		http.Error(w, "Invalid userID", http.StatusBadRequest)
// 		return
// 	}

// 	var chats []models.Chat
// 	if err := chatsDB.
// 		Where("user1_id = ? OR user2_id = ?", currentUserID, currentUserID).
// 		Find(&chats).Error; err != nil {
// 		logrus.Errorf("GetChats: ошибка получения чатов: %v", err)
// 		http.Error(w, "Error fetching chats", http.StatusInternalServerError)
// 		return
// 	}

// 	summaries := make([]ChatSummary, 0, len(chats))
// 	for _, chat := range chats {
// 		var otherUserID uuid.UUID
// 		if chat.User1ID == currentUserID {
// 			otherUserID = chat.User2ID
// 		} else {
// 			otherUserID = chat.User1ID
// 		}

// 		var lastMsg models.Message
// 		// if err := chatsDB.
// 		// 	Model(&models.Message{}).
// 		// 	Where("chat_id = ?", chat.ID).
// 		// 	Order("timestamp desc").
// 		// 	Limit(1).
// 		// 	First(&lastMsg).Error; err != nil && err != gorm.ErrRecordNotFound {
// 		// 	logrus.Errorf("GetChats: ошибка получения последнего сообщения для чата %d: %v", chat.ID, err)
// 		// 	http.Error(w, "Error fetching last message", http.StatusInternalServerError)
// 		// 	return
// 		// }
// 		result := chatsDB.
// 			Model(&models.Message{}).
// 			Where("chat_id = ?", chat.ID).
// 			Order("timestamp desc").
// 			Limit(1).
// 			First(&lastMsg)

// 		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			logrus.Errorf("GetChats: ошибка получения последнего сообщения для чата %d: %v", chat.ID, result.Error)
// 			http.Error(w, "Error fetching last message", http.StatusInternalServerError)
// 			return
// 		}

// 		// if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 		// 	logrus.Errorf("GetChats: ошибка получения последнего сообщения для чата %d: %v", chat.ID, err)
// 		// 	http.Error(w, "Error fetching last message", http.StatusInternalServerError)
// 		// 	return
// 		// }

// 		lastMessageSummary := MessageSummary{}
// 		if lastMsg.ID != 0 {
// 			lastMessageSummary = MessageSummary{
// 				ID:        lastMsg.ID,
// 				SenderID:  lastMsg.SenderID,
// 				Content:   lastMsg.Content,
// 				Timestamp: lastMsg.Timestamp,
// 				Read:      lastMsg.Read,
// 			}
// 		}

// 		var unreadCount int64
// 		if err := chatsDB.
// 			Model(&models.Message{}).
// 			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
// 			Count(&unreadCount).Error; err != nil {
// 			logrus.Errorf("GetChats: ошибка подсчёта непрочитанных сообщений для чата %d: %v", chat.ID, err)
// 			http.Error(w, "Error counting unread messages", http.StatusInternalServerError)
// 			return
// 		}

// 		// Извлекаем профиль другого пользователя для индикатора онлайн/офлайн.
// 		var otherProfile models.Profile
// 		otherOnline := false
// 		if err := chatsDB.First(&otherProfile, "user_id = ?", otherUserID).Error; err == nil {
// 			otherOnline = otherProfile.Online
// 		}

// 		summary := ChatSummary{
// 			ChatID:          chat.ID,
// 			OtherUserID:     otherUserID,
// 			LastMessage:     lastMessageSummary,
// 			UnreadCount:     int(unreadCount),
// 			OtherUserOnline: otherOnline,
// 			IsTyping:        false,          // Изначально false; реальное обновление происходит в режиме реального времени через WebSocket.
// 			ChatCreatedAt:   chat.CreatedAt, // Сохраняем время создания чата для сортировки
// 			OtherUser: &struct {
// 				ID        uuid.UUID `json:"id"`
// 				FirstName string    `json:"firstName"`
// 				LastName  string    `json:"lastName"`
// 				PhotoURL  string    `json:"photoUrl"`
// 			}{
// 				ID:        otherUserID,
// 				FirstName: otherProfile.FirstName,
// 				LastName:  otherProfile.LastName,
// 				PhotoURL:  otherProfile.PhotoURL,
// 			},
// 		}

// 		// Проверяем, набирает ли текст другой пользователь в этом чате
// 		if sockets.IsUserTypingInChat(summary.ChatID, summary.OtherUserID.String()) {
// 			summary.IsTyping = true
// 		}
// 		summaries = append(summaries, summary)
// 	}

// 	// Сортируем чаты по времени последней активности.
// 	sort.Slice(summaries, func(i, j int) bool {
// 		var timeI, timeJ time.Time
// 		// Если чат имеет последнее сообщение, используем его Timestamp, иначе – время создания чата.
// 		if !summaries[i].LastMessage.Timestamp.IsZero() {
// 			timeI = summaries[i].LastMessage.Timestamp
// 		} else {
// 			timeI = summaries[i].ChatCreatedAt
// 		}
// 		if !summaries[j].LastMessage.Timestamp.IsZero() {
// 			timeJ = summaries[j].LastMessage.Timestamp
// 		} else {
// 			timeJ = summaries[j].ChatCreatedAt
// 		}
// 		// Сортировка по убыванию времени (более активные чаты наверху)
// 		return timeI.After(timeJ)
// 	})

// 	logrus.Infof("GetChats: получено %d чатов для пользователя %s", len(summaries), currentUserID)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(summaries)
// }

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
	// Query to get chats for the current user (either user1 or user2)
	if err := chatsDB.
		Where("user1_id = ? OR user2_id = ?", currentUserID, currentUserID).
		Find(&chats).Error; err != nil {
		logrus.Errorf("GetChats: ошибка получения чатов: %v", err)
		http.Error(w, "Error fetching chats", http.StatusInternalServerError)
		return
	}

	summaries := make([]ChatSummary, 0, len(chats))
	for _, chat := range chats {
		// Determine the other user in the chat
		var otherUserID uuid.UUID
		if chat.User1ID == currentUserID {
			otherUserID = chat.User2ID
		} else {
			otherUserID = chat.User1ID
		}

		// Fetch the last message for the chat
		var lastMsg models.Message
		result := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ?", chat.ID).
			Order("timestamp desc").
			Limit(1).
			First(&lastMsg)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logrus.Errorf("GetChats: ошибка получения последнего сообщения для чата %d: %v", chat.ID, result.Error)
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

		// Count the number of unread messages
		var unreadCount int64
		if err := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
			Count(&unreadCount).Error; err != nil {
			logrus.Errorf("GetChats: ошибка подсчёта непрочитанных сообщений для чата %d: %v", chat.ID, err)
			http.Error(w, "Error counting unread messages", http.StatusInternalServerError)
			return
		}

		// Fetch the profile of the other user (to check their online status)
		var otherProfile models.Profile
		otherOnline := false
		if err := chatsDB.First(&otherProfile, "user_id = ?", otherUserID).Error; err == nil {
			otherOnline = otherProfile.Online
		}

		// Create a chat summary
		summary := ChatSummary{
			ChatID:          chat.ID,
			OtherUserID:     otherUserID,
			LastMessage:     lastMessageSummary,
			UnreadCount:     int(unreadCount),
			OtherUserOnline: otherOnline,
			IsTyping:        false,          // Placeholder; WebSocket should update this in real-time
			ChatCreatedAt:   chat.CreatedAt, // For sorting
			OtherUser: &struct {
				ID        uuid.UUID `json:"id"`
				FirstName string    `json:"firstName"`
				LastName  string    `json:"lastName"`
				PhotoURL  string    `json:"photoUrl"`
			}{
				ID:        otherUserID,
				FirstName: otherProfile.FirstName,
				LastName:  otherProfile.LastName,
				PhotoURL:  otherProfile.PhotoURL,
			},
		}

		// Check if the other user is typing (via WebSocket)
		if sockets.IsUserTypingInChat(summary.ChatID, summary.OtherUserID.String()) {
			summary.IsTyping = true
		}

		// Append the summary to the list
		summaries = append(summaries, summary)
	}

	// Sort chats by the last activity (latest message timestamp or chat creation time)
	sort.Slice(summaries, func(i, j int) bool {
		var timeI, timeJ time.Time
		if !summaries[i].LastMessage.Timestamp.IsZero() {
			timeI = summaries[i].LastMessage.Timestamp
		} else {
			timeI = summaries[i].ChatCreatedAt
		}
		if !summaries[j].LastMessage.Timestamp.IsZero() {
			timeJ = summaries[j].LastMessage.Timestamp
		} else {
			timeJ = summaries[j].ChatCreatedAt
		}
		return timeI.After(timeJ) // Sort by most recent activity
	})

	logrus.Infof("GetChats: получено %d чатов для пользователя %s", len(summaries), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		logrus.Errorf("GetChats: ошибка кодирования ответа: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// // GET /chats/{chat_id}?page=1&limit=20
// // Возвращает историю сообщений чата с пагинацией и обновляет статус сообщений.
// func GetChatHistory(w http.ResponseWriter, r *http.Request) {
// 	userIDStr, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		logrus.Error("GetChatHistory: userID не найден в контексте")
// 		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUserID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный userID: %v", err)
// 		http.Error(w, "Invalid userID", http.StatusBadRequest)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	chatIDStr := vars["chat_id"]
// 	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный chat_id: %v", err)
// 		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
// 		return
// 	}

// 	var chat models.Chat
// 	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: чат %d не найден: %v", chatID, err)
// 		http.Error(w, "Chat not found", http.StatusNotFound)
// 		return
// 	}
// 	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
// 		logrus.Warnf("GetChatHistory: пользователь %s не является участником чата %d", currentUserID, chatID)
// 		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
// 		return
// 	}

// 	pageStr := r.URL.Query().Get("page")
// 	limitStr := r.URL.Query().Get("limit")
// 	page := 1
// 	limit := 20
// 	if pageStr != "" {
// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
// 			page = p
// 		}
// 	}
// 	if limitStr != "" {
// 		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
// 			limit = l
// 		}
// 	}
// 	offset := (page - 1) * limit

// 	var messages []models.Message
// 	if err := chatsDB.
// 		Where("chat_id = ?", chat.ID).
// 		Order("timestamp asc").
// 		Offset(offset).
// 		Limit(limit).
// 		Find(&messages).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: ошибка получения сообщений для чата %d: %v", chatID, err)
// 		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
// 		return
// 	}

// 	// Фоновое обновление: помечаем как прочитанные все сообщения, отправленные не текущим пользователем.
// 	go func() {
// 		if err := chatsDB.
// 			Model(&models.Message{}).
// 			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
// 			Update("read", true).Error; err != nil {
// 			logrus.Errorf("GetChatHistory: ошибка обновления статуса сообщений для чата %d: %v", chat.ID, err)
// 		}
// 	}()

// 	logrus.Infof("GetChatHistory: получена история сообщений для чата %d (страница %d, лимит %d)", chatID, page, limit)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(messages)
// }

// func GetChatHistory(w http.ResponseWriter, r *http.Request) {
// 	userIDStr, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		logrus.Error("GetChatHistory: userID не найден в контексте")
// 		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUserID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный userID: %v", err)
// 		http.Error(w, "Invalid userID", http.StatusBadRequest)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	chatIDStr := vars["chat_id"]

// 	// Optional: detect whether chat_id is an actual ID or "new"
// 	if chatIDStr == "new" {
// 		// New chat request: must provide other_user_id
// 		otherUserIDStr := r.URL.Query().Get("other_user_id")
// 		if otherUserIDStr == "" {
// 			http.Error(w, "Missing other_user_id for new chat", http.StatusBadRequest)
// 			return
// 		}
// 		otherUserID, err := uuid.Parse(otherUserIDStr)
// 		if err != nil {
// 			http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
// 			return
// 		}

// 		// Look for existing chat between users
// 		var chat models.Chat
// 		err = chatsDB.
// 			Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
// 				currentUserID, otherUserID, otherUserID, currentUserID).
// 			First(&chat).Error

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			// Create new chat
// 			chat = models.Chat{
// 				User1ID: currentUserID,
// 				User2ID: otherUserID,
// 			}
// 			if err := chatsDB.Create(&chat).Error; err != nil {
// 				logrus.Errorf("GetChatHistory: ошибка создания нового чата: %v", err)
// 				http.Error(w, "Error creating chat", http.StatusInternalServerError)
// 				return
// 			}
// 			logrus.Infof("GetChatHistory: создан новый чат между %s и %s", currentUserID, otherUserID)
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode([]models.Message{})
// 			return
// 		} else if err != nil {
// 			http.Error(w, "Error retrieving chat", http.StatusInternalServerError)
// 			return
// 		}

// 		// Redirect to same handler with chat ID
// 		r = mux.SetURLVars(r, map[string]string{"chat_id": fmt.Sprintf("%d", chat.ID)})
// 		GetChatHistory(w, r)
// 		return
// 	}

// 	// Existing chat path
// 	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный chat_id: %v", err)
// 		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
// 		return
// 	}

// 	var chat models.Chat
// 	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: чат %d не найден: %v", chatID, err)
// 		http.Error(w, "Chat not found", http.StatusNotFound)
// 		return
// 	}
// 	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
// 		logrus.Warnf("GetChatHistory: пользователь %s не является участником чата %d", currentUserID, chatID)
// 		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
// 		return
// 	}

// 	// Pagination
// 	pageStr := r.URL.Query().Get("page")
// 	limitStr := r.URL.Query().Get("limit")
// 	page := 1
// 	limit := 20
// 	if pageStr != "" {
// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
// 			page = p
// 		}
// 	}
// 	if limitStr != "" {
// 		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
// 			limit = l
// 		}
// 	}
// 	offset := (page - 1) * limit

// 	var messages []models.Message
// 	if err := chatsDB.
// 		Where("chat_id = ?", chat.ID).
// 		Order("timestamp asc").
// 		Offset(offset).
// 		Limit(limit).
// 		Find(&messages).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: ошибка получения сообщений для чата %d: %v", chatID, err)
// 		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
// 		return
// 	}

// 	// Mark messages as read (background)
// 	go func() {
// 		if err := chatsDB.
// 			Model(&models.Message{}).
// 			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
// 			Update("read", true).Error; err != nil {
// 			logrus.Errorf("GetChatHistory: ошибка обновления статуса сообщений для чата %d: %v", chat.ID, err)
// 		}
// 	}()

// 	logrus.Infof("GetChatHistory: получена история сообщений для чата %d (страница %d, лимит %d)", chatID, page, limit)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(messages)
// }

// func GetChatHistory(w http.ResponseWriter, r *http.Request) {
// 	// Log user context check
// 	userIDStr, ok := r.Context().Value("userID").(string)
// 	if !ok {
// 		logrus.Error("GetChatHistory: userID не найден в контексте")
// 		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
// 		return
// 	}
// 	currentUserID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный userID: %v", err)
// 		http.Error(w, "Invalid userID", http.StatusBadRequest)
// 		return
// 	}
// 	logrus.Infof("GetChatHistory: текущий пользователь %s", currentUserID)

// 	vars := mux.Vars(r)
// 	chatIDStr := vars["chat_id"]
// 	logrus.Infof("GetChatHistory: получен chat_id: %s", chatIDStr)

// 	// Optional: detect whether chat_id is an actual ID or "new"
// 	if chatIDStr == "new" {
// 		// New chat request: must provide other_user_id
// 		otherUserIDStr := r.URL.Query().Get("other_user_id")
// 		if otherUserIDStr == "" {
// 			http.Error(w, "Missing other_user_id for new chat", http.StatusBadRequest)
// 			logrus.Error("GetChatHistory: не указан other_user_id для нового чата")
// 			return
// 		}
// 		otherUserID, err := uuid.Parse(otherUserIDStr)
// 		if err != nil {
// 			http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
// 			logrus.Errorf("GetChatHistory: неверный other_user_id: %v", err)
// 			return
// 		}

// 		// Look for existing chat between users
// 		var chat models.Chat
// 		err = chatsDB.
// 			Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)").
// 			First(&chat, currentUserID, otherUserID, otherUserID, currentUserID).Error

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			// Create new chat
// 			chat = models.Chat{
// 				User1ID: currentUserID,
// 				User2ID: otherUserID,
// 			}
// 			if err := chatsDB.Create(&chat).Error; err != nil {
// 				logrus.Errorf("GetChatHistory: ошибка создания нового чата: %v", err)
// 				http.Error(w, "Error creating chat", http.StatusInternalServerError)
// 				return
// 			}
// 			logrus.Infof("GetChatHistory: создан новый чат между %s и %s", currentUserID, otherUserID)
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode([]models.Message{})
// 			return
// 		} else if err != nil {
// 			http.Error(w, "Error retrieving chat", http.StatusInternalServerError)
// 			logrus.Errorf("GetChatHistory: ошибка получения чата: %v", err)
// 			return
// 		}

// 		// Redirect to same handler with chat ID
// 		logrus.Infof("GetChatHistory: перенаправление на существующий чат с ID %d", chat.ID)
// 		r = mux.SetURLVars(r, map[string]string{"chat_id": fmt.Sprintf("%d", chat.ID)})
// 		GetChatHistory(w, r)
// 		return
// 	}

// 	// Existing chat path
// 	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
// 	if err != nil {
// 		logrus.Errorf("GetChatHistory: неверный chat_id: %v", err)
// 		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
// 		return
// 	}
// 	logrus.Infof("GetChatHistory: обработка существующего чата с ID %d", chatID)

// 	var chat models.Chat
// 	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: чат %d не найден: %v", chatID, err)
// 		http.Error(w, "Chat not found", http.StatusNotFound)
// 		return
// 	}
// 	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
// 		logrus.Warnf("GetChatHistory: пользователь %s не является участником чата %d", currentUserID, chatID)
// 		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
// 		return
// 	}

// 	// Pagination
// 	pageStr := r.URL.Query().Get("page")
// 	limitStr := r.URL.Query().Get("limit")
// 	page := 1
// 	limit := 20
// 	if pageStr != "" {
// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
// 			page = p
// 		}
// 	}
// 	if limitStr != "" {
// 		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
// 			limit = l
// 		}
// 	}
// 	offset := (page - 1) * limit
// 	logrus.Infof("GetChatHistory: пагинация: страница %d, лимит %d, смещение %d", page, limit, offset)

// 	var messages []models.Message
// 	if err := chatsDB.
// 		Where("chat_id = ?", chat.ID).
// 		Order("timestamp asc").
// 		Offset(offset).
// 		Limit(limit).
// 		Find(&messages).Error; err != nil {
// 		logrus.Errorf("GetChatHistory: ошибка получения сообщений для чата %d: %v", chatID, err)
// 		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
// 		return
// 	}

// 	// Log message count
// 	logrus.Infof("GetChatHistory: получено %d сообщений для чата %d", len(messages), chatID)

// 	// Mark messages as read (background)
// 	go func() {
// 		if err := chatsDB.
// 			Model(&models.Message{}).
// 			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
// 			Update("read", true).Error; err != nil {
// 			logrus.Errorf("GetChatHistory: ошибка обновления статуса сообщений для чата %d: %v", chat.ID, err)
// 		} else {
// 			logrus.Infof("GetChatHistory: сообщения чата %d помечены как прочитанные", chat.ID)
// 		}
// 	}()

// 	logrus.Infof("GetChatHistory: отправка истории сообщений для чата %d (страница %d, лимит %d)", chatID, page, limit)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(messages)
// }

func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	// Log user context check
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
	logrus.Infof("GetChatHistory: текущий пользователь %s", currentUserID)

	// Fetch the chat ID from the URL
	vars := mux.Vars(r)
	chatIDStr := vars["chat_id"]
	logrus.Infof("GetChatHistory: получен chat_id: %s", chatIDStr)

	// Check if it's a new chat request
	if chatIDStr == "new" {
		// New chat request: must provide other_user_id
		otherUserIDStr := r.URL.Query().Get("other_user_id")
		if otherUserIDStr == "" {
			http.Error(w, "Missing other_user_id for new chat", http.StatusBadRequest)
			logrus.Error("GetChatHistory: не указан other_user_id для нового чата")
			return
		}
		otherUserID, err := uuid.Parse(otherUserIDStr)
		if err != nil {
			http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
			logrus.Errorf("GetChatHistory: неверный other_user_id: %v", err)
			return
		}

		// Look for existing chat between users
		var chat models.Chat
		err = chatsDB.
			Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)").
			First(&chat, currentUserID, otherUserID, otherUserID, currentUserID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new chat if no chat is found
			chat = models.Chat{
				User1ID: currentUserID,
				User2ID: otherUserID,
			}
			if err := chatsDB.Create(&chat).Error; err != nil {
				logrus.Errorf("GetChatHistory: ошибка создания нового чата: %v", err)
				http.Error(w, "Error creating chat", http.StatusInternalServerError)
				return
			}
			logrus.Infof("GetChatHistory: создан новый чат между %s и %s", currentUserID, otherUserID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Message{})
			return
		} else if err != nil {
			http.Error(w, "Error retrieving chat", http.StatusInternalServerError)
			logrus.Errorf("GetChatHistory: ошибка получения чата: %v", err)
			return
		}

		// Redirect to existing chat handler
		logrus.Infof("GetChatHistory: перенаправление на существующий чат с ID %d", chat.ID)
		r = mux.SetURLVars(r, map[string]string{"chat_id": fmt.Sprintf("%d", chat.ID)})
		GetChatHistory(w, r)
		return
	}

	// Handle existing chat path
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		logrus.Errorf("GetChatHistory: неверный chat_id: %v", err)
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	logrus.Infof("GetChatHistory: обработка существующего чата с ID %d", chatID)

	// Fetch chat details
	var chat models.Chat
	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
		logrus.Errorf("GetChatHistory: чат %d не найден: %v", chatID, err)
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}

	// Check if the current user is part of the chat
	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
		logrus.Warnf("GetChatHistory: пользователь %s не является участником чата %d", currentUserID, chatID)
		http.Error(w, "Forbidden: you are not a participant in this chat", http.StatusForbidden)
		return
	}

	// Pagination setup
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
	logrus.Infof("GetChatHistory: пагинация: страница %d, лимит %d, смещение %d", page, limit, offset)

	// Fetch messages from the chat history
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

	// Log message count
	logrus.Infof("GetChatHistory: получено %d сообщений для чата %d", len(messages), chatID)

	// Mark messages as read in the background
	go func() {
		if err := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
			Update("read", true).Error; err != nil {
			logrus.Errorf("GetChatHistory: ошибка обновления статуса сообщений для чата %d: %v", chat.ID, err)
		} else {
			logrus.Infof("GetChatHistory: сообщения чата %d помечены как прочитанные", chat.ID)
		}
	}()

	// Send back the chat history
	logrus.Infof("GetChatHistory: отправка истории сообщений для чата %d (страница %d, лимит %d)", chatID, page, limit)
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

// // NOT USED YET. CHECK!!!
// func GetChatSummaries(userID uuid.UUID) ([]ChatSummary, error) {
// 	var chatSummaries []ChatSummary

// 	// Get chats for the current user
// 	err := chatsDB.Table("chats").
// 		Select("chats.id AS chat_id, chats.user1_id, chats.user2_id, messages.id AS message_id, messages.sender_id, messages.content, messages.timestamp, messages.read").
// 		Joins("LEFT JOIN messages ON messages.chat_id = chats.id").
// 		Where("chats.user1_id = ? OR chats.user2_id = ?", userID, userID).
// 		Order("messages.timestamp DESC").
// 		Scan(&chatSummaries).Error

// 	if err != nil {
// 		logrus.Errorf("GetChatSummaries: error retrieving chat summaries: %v", err)
// 		return nil, err
// 	}

// 	// Process each chat to populate otherUser and unreadCount
// 	for i, chat := range chatSummaries {
// 		// Determine the other user
// 		if chat.User1ID != userID {
// 			chatSummaries[i].OtherUserID = chat.User1ID
// 			// Fetch user details for other user
// 			otherUser := getUserDetails(chat.User1ID)
// 			chatSummaries[i].OtherUser = &otherUser
// 		} else {
// 			chatSummaries[i].OtherUserID = chat.User2ID
// 			// Fetch user details for other user
// 			otherUser := getUserDetails(chat.User2ID)
// 			chatSummaries[i].OtherUser = &otherUser
// 		}

// 		// Get the unread message count
// 		unreadCount := getUnreadMessagesCount(chat.ChatID, userID)
// 		chatSummaries[i].UnreadCount = unreadCount
// 	}

// 	return chatSummaries, nil
// }

// // Helper function to fetch user details based on userID
// func getUserDetails(userID uuid.UUID) struct {
// 	ID        uuid.UUID `json:"id"`
// 	FirstName string    `json:"firstName"`
// 	LastName  string    `json:"lastName"`
// 	PhotoURL  string    `json:"photoUrl"`
// } {
// 	var user struct {
// 		ID        uuid.UUID `json:"id"`
// 		FirstName string    `json:"firstName"`
// 		LastName  string    `json:"lastName"`
// 		PhotoURL  string    `json:"photoUrl"`
// 	}

// 	chatsDB.First(&user, "id = ?", userID)
// 	return user
// }

// // Helper function to get unread messages count
// func getUnreadMessagesCount(chatID uint, userID uuid.UUID) int {
// 	var count int
// 	chatsDB.Model(&Message{}).
// 		Where("chat_id = ? AND sender_id <> ? AND read = ?", chatID, userID, false).
// 		Count(&count)
// 	return count
// }
