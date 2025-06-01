package sockets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"m/backend/models"
	"m/backend/services"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var presenceSvc *services.PresenceService
var chatsDB *gorm.DB

// This module implements real-time chat and notifications using WebSockets, that provide a persistent, full-duplex connection between client and server,
// enabling instant message delivery, typing indicators, and presence updates.
// The Hub manages all connected clients, chat subscriptions, and message broadcasting.
// Each client maintains its own connection, send queue, and chat subscriptions.
// The design ensures thread safety and efficient delivery of events to all relevant users.
// This approach is essential for responsive, interactive chat and social features.

type BroadcastMessage struct {
	ChatID uint   `json:"chatId"`
	Data   []byte `json:"data"`
}

type Client struct {
	Conn        *websocket.Conn
	UserID      string
	Send        chan []byte
	Chats       map[uint]bool
	TypingChats map[uint]bool
	Mutex       sync.Mutex
}

// Hub manages all WebSocket clients and chat subscriptions.
// It handles registration, unregistration, and broadcasting messages to the correct chat subscribers.
type Hub struct {
	Clients           map[*Client]bool
	ChatSubscriptions map[uint]map[*Client]bool
	Broadcast         chan BroadcastMessage
	Register          chan *Client
	Unregister        chan *Client
	Mutex             sync.RWMutex
}

var hub = Hub{
	Clients:           make(map[*Client]bool),
	ChatSubscriptions: make(map[uint]map[*Client]bool),
	Broadcast:         make(chan BroadcastMessage),
	Register:          make(chan *Client),
	Unregister:        make(chan *Client),
}

func IsUserTypingInChat(chatID uint, userID string) bool {
	hub.Mutex.RLock()
	defer hub.Mutex.RUnlock()
	subs, ok := hub.ChatSubscriptions[chatID]
	if !ok {
		return false
	}
	for client := range subs {
		if client.UserID == userID {
			client.Mutex.Lock()
			typing := client.TypingChats[chatID]
			client.Mutex.Unlock()
			if typing {
				return true
			}
		}
	}
	return false
}

func RunHub() {
	// Main event loop for the WebSocket hub.
	// Handles registration/unregistration of clients and broadcasting messages to chat subscribers.
	for {
		select {
		case client := <-hub.Register:
			// Register a new client connection
			hub.Mutex.Lock()
			hub.Clients[client] = true
			hub.Mutex.Unlock()
			logrus.Infof("Client registered: %s", client.UserID)

		case client := <-hub.Unregister:
			// Unregister a client and clean up all its subscriptions
			hub.Mutex.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				// Remove client from all chat subscriptions
				for chatID := range client.Chats {
					if subs := hub.ChatSubscriptions[chatID]; subs != nil {
						delete(subs, client)
						if len(subs) == 0 {
							delete(hub.ChatSubscriptions, chatID)
						}
					}
				}
				close(client.Send)
				logrus.Infof("Client unregistered: %s", client.UserID)
			}
			hub.Mutex.Unlock()

		case msg := <-hub.Broadcast:
			// Broadcast a message to all clients subscribed to the chat
			hub.Mutex.RLock()
			if subs := hub.ChatSubscriptions[msg.ChatID]; subs != nil {
				for client := range subs {
					select {
					case client.Send <- msg.Data:
						logrus.Debugf("Message sent to client %s in chat %d", client.UserID, msg.ChatID)
					default:
						// If the send buffer is full, close the connection to avoid blocking the hub
						logrus.Warnf("Send channel full for client %s. Closing.", client.UserID)
						close(client.Send)
						delete(hub.Clients, client)
					}
				}
			}
			hub.Mutex.RUnlock()
		}
	}
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract userID from query or generate a new one
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		userID = uuid.New().String()
		logrus.Debug("HandleWebSocket: generated userID")
	}

	// Create a new client instance for this connection
	client := &Client{
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		Chats:       make(map[uint]bool),
		TypingChats: make(map[uint]bool),
	}

	// Mark user as online in presence service
	if presenceSvc != nil {
		if err := presenceSvc.Touch(userID); err != nil {
			logrus.Warnf("presence.Touch failed for %s: %v", userID, err)
		}
	}

	if presenceSvc != nil {
		if err := presenceSvc.Touch(userID); err != nil {
			logrus.Warnf("presence.Touch failed for %s: %v", userID, err)
		}
	}
	// Register client in the hub and start pumps
	hub.Register <- client
	logrus.Infof("HandleWebSocket: client %s connected", client.UserID)

	// Start writePump in a separate goroutine to send messages
	go client.writePump()
	// Start readPump in the current goroutine to receive messages
	client.readPump()
}

//Helper function for retrieving message IDs

func getMessageIDs(msgs []models.Message) []uint {
	ids := make([]uint, len(msgs))
	for i, m := range msgs {
		ids[i] = m.ID
	}
	return ids
}

func (c *Client) readPump() {
	// Main loop for reading messages from the WebSocket connection.
	// Handles subscription, typing, heartbeat, and other client events.
	defer func() {
		// On disconnect, mark user as offline and unregister client
		if presenceSvc != nil {
			if err := presenceSvc.SetOffline(c.UserID); err != nil {
				logrus.Warnf("presence.SetOffline failed for %s: %v", c.UserID, err)
			}
		}
		hub.Unregister <- c
		c.Conn.Close()
		logrus.Infof("readPump: connection closed for %s", c.UserID)
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump error for %s: %v", c.UserID, err)
			}
			break
		}

		// Parse incoming JSON message
		var req struct {
			Action   string `json:"action"`
			ChatID   string `json:"chat_id"`
			IsOnline *bool  `json:"is_online"`
			IsTyping *bool  `json:"is_typing"`
		}
		if err := json.Unmarshal(msg, &req); err != nil {
			logrus.Errorf("readPump: unmarshal error for %s: %v", c.UserID, err)
			continue
		}

		// Handle different types of client actions
		switch req.Action {
		case "subscribe", "unsubscribe":
			// Subscribe/unsubscribe client to/from a chat room
			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
			if err != nil {
				logrus.Warnf("readPump: bad chat_id '%s' from %s", req.ChatID, c.UserID)
				continue
			}
			c.Mutex.Lock()
			if req.Action == "subscribe" {
				c.Chats[uint(chatID)] = true
				hub.Mutex.Lock()
				if hub.ChatSubscriptions[uint(chatID)] == nil {
					hub.ChatSubscriptions[uint(chatID)] = make(map[*Client]bool)
				}
				hub.ChatSubscriptions[uint(chatID)][c] = true
				hub.Mutex.Unlock()
				logrus.Infof("%s subscribed to chat %d", c.UserID, chatID)
			} else {
				delete(c.Chats, uint(chatID))
				hub.Mutex.Lock()
				if subs := hub.ChatSubscriptions[uint(chatID)]; subs != nil {
					delete(subs, c)
					if len(subs) == 0 {
						delete(hub.ChatSubscriptions, uint(chatID))
					}
				}
				hub.Mutex.Unlock()
				logrus.Infof("%s unsubscribed from chat %d", c.UserID, chatID)
			}
			c.Mutex.Unlock()

		case "heartbeat":
			// Heartbeat to keep connection alive and update presence
			logrus.Debugf("readPump heartbeat from %s", c.UserID)
			if presenceSvc != nil {
				if err := presenceSvc.Touch(c.UserID); err != nil {
					logrus.Warnf("presence.Touch failed for %s: %v", c.UserID, err)
				}
			}

		case "typing":
			// Update typing status for this client in a chat
			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
			if err != nil {
				logrus.Warnf("readPump: bad chat_id in typing from %s", c.UserID)
				continue
			}
			if req.IsTyping == nil {
				logrus.Warnf("readPump: missing is_typing from %s", c.UserID)
				continue
			}
			c.Mutex.Lock()
			c.TypingChats[uint(chatID)] = *req.IsTyping
			c.Mutex.Unlock()

			uid, err := uuid.Parse(c.UserID)
			if err == nil {
				// Broadcast typing status to all chat participants
				go BroadcastTypingNotification(uid, uint(chatID), *req.IsTyping)
			}

		case "read":
			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
			if err != nil {
				logrus.Warnf("readPump: bad chat_id in read from %s", c.UserID)
				continue
			}
			uid, err := uuid.Parse(c.UserID)
			if err != nil {
				logrus.Warnf("readPump: invalid userID in read from %s", c.UserID)
				continue
			}
			var unreadMsgs []models.Message
			if err := chatsDB.
				Model(&models.Message{}).
				Where("chat_id = ? AND sender_id <> ? AND read = ?", chatID, uid, false).
				Find(&unreadMsgs).Error; err != nil {
				logrus.Errorf("readPump: failed to find unread messages for chat %d: %v", chatID, err)
				continue
			}
			if len(unreadMsgs) == 0 {
				continue // Защита от лишних событий!
			}
			if err := chatsDB.
				Model(&models.Message{}).
				Where("id IN ?", getMessageIDs(unreadMsgs)).
				Update("read", true).Error; err != nil {
				logrus.Errorf("readPump: failed to mark messages as read for chat %d: %v", chatID, err)
				continue
			}
			for _, msg := range unreadMsgs {
				payload := map[string]interface{}{
					"type":       "read",
					"chat_id":    chatID,
					"message_id": msg.ID,
				}
				data, _ := json.Marshal(payload)
				hub.Broadcast <- BroadcastMessage{
					ChatID: uint(chatID),
					Data:   data,
				}
			}

		default:
			logrus.Warnf("readPump: unknown action '%s' from %s", req.Action, c.UserID)
		}
	}
}

func (c *Client) writePump() {
	// Main loop for sending messages from the server to the client over WebSocket.
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		logrus.Infof("writePump: connection closed for %s", c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed, send close message to client
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Send all queued messages in the channel
			for i := 0; i < len(c.Send); i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}
			w.Close()

		case <-ticker.C:
			// Send periodic ping to keep connection alive
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func BroadcastNewMessage(msg models.Message) error {
	// Broadcast a new chat message to all clients subscribed to the chat
	senderName := "Unknown"
	if msg.Sender.Profile.FirstName != "" || msg.Sender.Profile.LastName != "" {
		senderName = strings.TrimSpace(
			msg.Sender.Profile.FirstName + " " +
				msg.Sender.Profile.LastName,
		)
	}

	payload := map[string]interface{}{
		"type":        "message",
		"chat_id":     msg.ChatID,
		"id":          msg.ID,
		"sender_id":   msg.SenderID.String(),
		"sender_name": senderName,
		"content":     msg.Content,
		"timestamp":   msg.Timestamp.UnixNano() / int64(time.Millisecond),
		"read":        msg.Read,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("BroadcastNewMessage: marshal error: %v", err)
		return err
	}

	hub.Broadcast <- BroadcastMessage{
		ChatID: msg.ChatID,
		Data:   data,
	}
	logrus.Infof("BroadcastNewMessage: sent message %d to chat %d", msg.ID, msg.ChatID)
	return nil
}

func BroadcastNotification(userID uuid.UUID, message string) {
	data, err := json.Marshal(map[string]string{
		"type":    "notification",
		"message": message,
	})
	if err != nil {
		logrus.Errorf("BroadcastNotification: marshal error: %v", err)
		return
	}

	hub.Mutex.RLock()
	defer hub.Mutex.RUnlock()
	for client := range hub.Clients {
		if client.UserID == userID.String() {
			select {
			case client.Send <- data:
				logrus.Infof("Notification sent to %s", userID)
			default:
				logrus.Warnf("Notification channel full for %s", userID)
			}
		}
	}
}

func BroadcastTypingNotification(userID uuid.UUID, chatID uint, isTyping bool) {
	// Broadcast typing status to all clients in the chat
	data, err := json.Marshal(map[string]interface{}{
		"type":      "typing",
		"user_id":   userID.String(),
		"chat_id":   chatID,
		"is_typing": isTyping,
		"timestamp": time.Now().Unix(),
	})
	if err != nil {
		logrus.Errorf("BroadcastTypingNotification: marshal error: %v", err)
		return
	}

	hub.Mutex.RLock()
	subs, ok := hub.ChatSubscriptions[chatID]
	hub.Mutex.RUnlock()
	if !ok {
		return
	}
	for client := range subs {
		select {
		case client.Send <- data:
		default:
			logrus.Warnf("Typing channel full for %s", client.UserID)
		}
	}
}

func BroadcastConnectionRequest(userID uuid.UUID) {
	payload := map[string]interface{}{
		"type": "connection_request",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("BroadcastConnectionRequest: marshal error: %v", err)
		return
	}

	hub.Mutex.RLock()
	defer hub.Mutex.RUnlock()
	for client := range hub.Clients {
		if client.UserID == userID.String() {
			select {
			case client.Send <- data:
				logrus.Infof("ConnectionRequest event sent to %s", userID)
			default:
				logrus.Warnf("ConnectionRequest channel full for %s", userID)
			}
		}
	}
}

func InitWebSocketServer(ps *services.PresenceService, addr string) error {
	presenceSvc = ps
	go RunHub()
	http.HandleFunc("/ws", HandleWebSocket)
	logrus.Infof("WebSocket server started on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func SetChatsDB(db *gorm.DB) {
	chatsDB = db
}
