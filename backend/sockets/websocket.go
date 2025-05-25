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
)

var presenceSvc *services.PresenceService

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
	for {
		select {
		case client := <-hub.Register:
			hub.Mutex.Lock()
			hub.Clients[client] = true
			hub.Mutex.Unlock()
			logrus.Infof("Client registered: %s", client.UserID)

		case client := <-hub.Unregister:
			hub.Mutex.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
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
			hub.Mutex.RLock()
			if subs := hub.ChatSubscriptions[msg.ChatID]; subs != nil {
				for client := range subs {
					select {
					case client.Send <- msg.Data:
						logrus.Debugf("Message sent to client %s in chat %d", client.UserID, msg.ChatID)
					default:
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		userID = uuid.New().String()
		logrus.Debug("HandleWebSocket: generated userID")
	}

	client := &Client{
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		Chats:       make(map[uint]bool),
		TypingChats: make(map[uint]bool),
	}

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

	hub.Register <- client
	logrus.Infof("HandleWebSocket: client %s connected", client.UserID)

	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
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

		switch req.Action {
		case "subscribe", "unsubscribe":
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
			logrus.Debugf("readPump heartbeat from %s", c.UserID)
			if presenceSvc != nil {
				if err := presenceSvc.Touch(c.UserID); err != nil {
					logrus.Warnf("presence.Touch failed for %s: %v", c.UserID, err)
				}
			}

		case "typing":
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
				go BroadcastTypingNotification(uid, uint(chatID), *req.IsTyping)
			}

		default:
			logrus.Warnf("readPump: unknown action '%s' from %s", req.Action, c.UserID)
		}
	}
}

func (c *Client) writePump() {
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
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			for i := 0; i < len(c.Send); i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}
			w.Close()

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func BroadcastNewMessage(msg models.Message) error {
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
