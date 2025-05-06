package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	//"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"m/backend/models"
	"m/backend/sockets"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var chatsDB *gorm.DB

func InitChatsController(db *gorm.DB) {
	chatsDB = db
	logrus.Info("Chats controller initialized")
}

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
	ChatCreatedAt   time.Time      `json:"-"`
}

type MessageSummary struct {
	ID        uint      `json:"id"`
	SenderID  uuid.UUID `json:"senderId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

type ChatMessageResponse struct {
	ID         uint      `json:"id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	Read       bool      `json:"read"`
	SenderID   uuid.UUID `json:"sender_id"`
	SenderName string    `json:"sender_name"`
}

// --- НОВЫЙ ХЕНДЛЕР: POST /chats -----------------------

// CreateOrGetChat создает чат между текущим пользователем и other_user_id, либо возвращает существующий.
// func CreateOrGetChat(w http.ResponseWriter, r *http.Request) {
// 	userIDStr, _ := r.Context().Value("userID").(string)
// 	currentUserID, _ := uuid.Parse(userIDStr)

// 	var req struct {
// 		OtherUserID string `json:"otherUserId"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid body", http.StatusBadRequest)
// 		return
// 	}
// 	otherID, err := uuid.Parse(req.OtherUserID)
// 	if err != nil {
// 		http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
// 		return
// 	}

// 	var chat models.Chat
// 	err = chatsDB.
// 		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
// 			currentUserID, otherID, otherID, currentUserID).
// 		First(&chat).Error

// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		chat = models.Chat{User1ID: currentUserID, User2ID: otherID}
// 		if err := chatsDB.Create(&chat).Error; err != nil {
// 			http.Error(w, "Error creating chat", http.StatusInternalServerError)
// 			return
// 		}
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]uint{"chatId": chat.ID})
// }

// --- Существующие хендлеры ----------------------------

func GetChats(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("userID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("invalid userID format: %q, err: %v", userIDStr, err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var chats []models.Chat
	if err := chatsDB.
		Where("user1_id = ? OR user2_id = ?", currentUserID, currentUserID).
		Find(&chats).Error; err != nil {
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
		res := chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ?", chat.ID).
			Order("timestamp desc").
			Limit(1).
			First(&lastMsg) //that is not found if new
		if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Error fetching last message", http.StatusInternalServerError)
			return
		}

		lastSummary := MessageSummary{}
		if lastMsg.ID != 0 {
			lastSummary = MessageSummary{
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
			http.Error(w, "Error counting unread messages", http.StatusInternalServerError)
			return
		}

		var otherProfile models.Profile
		otherOnline := false
		if err := chatsDB.First(&otherProfile, "user_id = ?", otherUserID).Error; err == nil {
			otherOnline = otherProfile.Online
		}

		summary := ChatSummary{
			ChatID:          chat.ID,
			OtherUserID:     otherUserID,
			LastMessage:     lastSummary,
			UnreadCount:     int(unreadCount),
			OtherUserOnline: otherOnline,
			IsTyping:        sockets.IsUserTypingInChat(chat.ID, otherUserID.String()),
			ChatCreatedAt:   chat.CreatedAt,
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
		summaries = append(summaries, summary)
	}

	// Сортируем по времени активности
	sort.Slice(summaries, func(i, j int) bool {
		var ti, tj time.Time
		if !summaries[i].LastMessage.Timestamp.IsZero() {
			ti = summaries[i].LastMessage.Timestamp
		} else {
			ti = summaries[i].ChatCreatedAt
		}
		if !summaries[j].LastMessage.Timestamp.IsZero() {
			tj = summaries[j].LastMessage.Timestamp
		} else {
			tj = summaries[j].ChatCreatedAt
		}
		return ti.After(tj)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Println("userID not found in context GetChatHistory")
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	chatIDStr := vars["chatId"]

	// Обработка "new" — теперь без рекурсии!
	if chatIDStr == "new" {
		otherUserIDStr := r.URL.Query().Get("other_user_id")
		if otherUserIDStr == "" {
			log.Println("otherUserIDStr IS NULL 1056")
			http.Error(w, "Missing other_user_id", http.StatusBadRequest)
			return
		}
		otherUserID, err := uuid.Parse(otherUserIDStr)
		if err != nil {
			log.Println("otherUserIDStr PARSED IS NULL 1056")
			http.Error(w, "Invalid other_user_id", http.StatusBadRequest)
			return
		}

		var chat models.Chat
		err = chatsDB.
			Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
				currentUserID, otherUserID, otherUserID, currentUserID).
			First(&chat).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			chat = models.Chat{User1ID: currentUserID, User2ID: otherUserID}
			if err := chatsDB.Create(&chat).Error; err != nil {
				http.Error(w, "Error creating chat", http.StatusInternalServerError)
				return
			}
		}

		// Возвращаем ID и пустой массив
		resp := struct {
			ChatID   uint             `json:"chatId"`
			Messages []models.Message `json:"messages"`
		}{
			ChatID:   chat.ID,
			Messages: []models.Message{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Обычный путь — получение истории
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		log.Printf("!!!!!!Invalid chat_id: %v (raw value: %q)", err, chatIDStr)
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}

	var chat models.Chat
	if err := chatsDB.First(&chat, "id = ?", chatID).Error; err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}
	if chat.User1ID != currentUserID && chat.User2ID != currentUserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Пагинация
	page, limit := 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		if pi, _ := strconv.Atoi(p); pi > 0 {
			page = pi
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if li, _ := strconv.Atoi(l); li > 0 {
			limit = li
		}
	}
	offset := (page - 1) * limit

	// var messages []models.Message
	// if err := chatsDB.
	// 	Where("chat_id = ?", chat.ID).
	// 	Order("timestamp asc").
	// 	Offset(offset).
	// 	Limit(limit).
	// 	Find(&messages).Error; err != nil {
	// 	http.Error(w, "Error fetching messages", http.StatusInternalServerError)
	// 	return
	// }

	// var messages []models.Message
	// if err := chatsDB.
	// 	Preload("Sender").
	// 	Where("chat_id = ?", chat.ID).
	// 	Order("timestamp asc").
	// 	Offset(offset).
	// 	Limit(limit).
	// 	Preload("Sender.Profile").Find(&messages).Error; err != nil {
	// 	http.Error(w, "Error fetching messages", http.StatusInternalServerError)
	// 	return
	// }

	var messages []models.Message
	if err := chatsDB.
		Preload("Sender.Profile").
		Where("chat_id = ?", chat.ID).
		Order("timestamp asc").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}

	// Background: пометить прочитанными
	go func() {
		_ = chatsDB.
			Model(&models.Message{}).
			Where("chat_id = ? AND sender_id <> ? AND read = ?", chat.ID, currentUserID, false).
			Update("read", true).Error
	}()

	// resp := make([]ChatMessageResponse, len(messages))
	// for i, m := range messages {
	// 	resp[i] = ChatMessageResponse{
	// 		ID:         m.ID,
	// 		Content:    m.Content,
	// 		Timestamp:  m.Timestamp,
	// 		Read:       m.Read,
	// 		SenderID:   m.SenderID,
	// 		SenderName: m.Sender.Username, // ← вот тут имя
	// 	}
	// }

	resp := make([]ChatMessageResponse, len(messages))
	for i, m := range messages {
		fullName := "Unknown"
		if m.Sender.Profile.FirstName != "" || m.Sender.Profile.LastName != "" {
			fullName = strings.TrimSpace(m.Sender.Profile.FirstName + " " + m.Sender.Profile.LastName)
		}
		resp[i] = ChatMessageResponse{
			ID:         m.ID,
			Content:    m.Content,
			Timestamp:  m.Timestamp,
			Read:       m.Read,
			SenderID:   m.SenderID,
			SenderName: fullName,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(messages)
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
	chatIDStr := vars["chatId"]
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

	var fullMsg models.Message
	if err := chatsDB.Preload("Sender.Profile").First(&fullMsg, newMsg.ID).Error; err != nil {
		logrus.Errorf("PostMessage: ошибка при загрузке полного сообщения: %v", err)
		http.Error(w, "Error loading full message", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(fullMsg)

	//json.NewEncoder(w).Encode(newMsg)
}
