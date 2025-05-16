package sockets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"m/backend/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// BroadcastMessage определяет структуру сообщения для рассылки по чатам.
type BroadcastMessage struct {
	ChatID uint   `json:"chatId"`
	Data   []byte `json:"data"`
}

// Client представляет одно WebSocket‑подключение.
type Client struct {
	Conn   *websocket.Conn // WebSocket‑соединение
	UserID string          // Идентификатор пользователя (например, извлекается из запроса)
	Send   chan []byte     // Канал для отправки сообщений клиенту
	Chats  map[uint]bool   // Список подписанных чатов (chatID -> true)
	// Новое поле: для хранения состояния набора текста для каждого чата
	TypingChats map[uint]bool
	Mutex       sync.Mutex // Для синхронизации доступа к полю Chats
}

// Hub управляет клиентами и рассылкой сообщений по чатам.
type Hub struct {
	Clients           map[*Client]bool          // Все активные клиенты
	ChatSubscriptions map[uint]map[*Client]bool // Для каждого chatID множество клиентов, подписанных на него
	Broadcast         chan BroadcastMessage     // Канал для рассылки сообщений в чаты
	Register          chan *Client              // Новый клиент
	Unregister        chan *Client              // Клиент отключается
	Mutex             sync.RWMutex              // Защита общих структур
}

// Инициализируем глобальный hub.
var hub = Hub{
	Clients:           make(map[*Client]bool),
	ChatSubscriptions: make(map[uint]map[*Client]bool),
	Broadcast:         make(chan BroadcastMessage),
	Register:          make(chan *Client),
	Unregister:        make(chan *Client),
}

// IsUserTypingInChat проверяет, набирает ли пользователь с заданным userID текст в чате chatID.
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
			typing, exists := client.TypingChats[chatID]
			client.Mutex.Unlock()
			if exists && typing {
				return true
			}
		}
	}
	return false
}

// RunHub запускает цикл обработки регистрации, отмены регистрации и рассылки сообщений.
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
				// Удаляем клиента из всех подписок
				for chatID := range client.Chats {
					if subs, exists := hub.ChatSubscriptions[chatID]; exists {
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
			// Отправляем сообщение только клиентам, подписанным на данный чат.
			hub.Mutex.RLock()
			if subs, ok := hub.ChatSubscriptions[msg.ChatID]; ok {
				for client := range subs {
					select {
					case client.Send <- msg.Data:
						logrus.Debugf("Message sent to client %s in chat %d", client.UserID, msg.ChatID)
					default:
						logrus.Warnf("Send channel full for client %s. Closing connection.", client.UserID)
						close(client.Send)
						delete(hub.Clients, client)
					}
				}
			} else {
				logrus.Debugf("No subscribers for chat %d", msg.ChatID)
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
	// В продакшене настройте CheckOrigin по необходимости.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket выполняет апгрейд HTTP-запроса до WebSocket-соединения.
// Здесь в качестве простоты userID извлекается из query-параметра "userID".
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		userID = uuid.New().String()
		logrus.Debug("HandleWebSocket: userID не передан, сгенерирован новый UUID")
	}

	client := &Client{
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		Chats:       make(map[uint]bool),
		TypingChats: make(map[uint]bool),
	}

	hub.Register <- client
	logrus.Infof("HandleWebSocket: client %s подключен", client.UserID)

	go client.writePump()
	client.readPump()
}

// readPump читает входящие сообщения от клиента.
// Ожидается, что клиент будет отправлять JSON-сообщения для подписки/отписки.
func (c *Client) readPump() {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
		logrus.Infof("readPump: закрыто соединение клиента %s", c.UserID)
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump: неожиданная ошибка WebSocket для клиента %s: %v", c.UserID, err)
			} else {
				logrus.Debugf("readPump: клиент %s отключился: %v", c.UserID, err)
			}
			break
		}

		// var req struct {
		// 	Action   string `json:"action"`
		// 	ChatID   string `json:"chatId"`   // для подписки/отписки, если необходимо
		// 	IsOnline *bool  `json:"isOnline"` // для heartbeat-сообщения
		// }
		// if err := json.Unmarshal(msg, &req); err != nil {
		// 	logrus.Errorf("readPump: ошибка разбора сообщения от клиента %s: %v", c.UserID, err)
		// 	continue
		// }

		// chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
		// if err != nil {
		// 	logrus.Warnf("readPump: неверный chat_id '%s' от клиента %s", req.ChatID, c.UserID)
		// 	continue
		// }

		// c.Mutex.Lock()
		// switch req.Action {
		// case "subscribe":
		// 	c.Chats[uint(chatID)] = true
		// 	hub.Mutex.Lock()
		// 	if hub.ChatSubscriptions[uint(chatID)] == nil {
		// 		hub.ChatSubscriptions[uint(chatID)] = make(map[*Client]bool)
		// 	}
		// 	hub.ChatSubscriptions[uint(chatID)][c] = true
		// 	hub.Mutex.Unlock()
		// 	logrus.Infof("readPump: клиент %s подписался на чат %d", c.UserID, chatID)
		// case "unsubscribe":
		// 	delete(c.Chats, uint(chatID))
		// 	hub.Mutex.Lock()
		// 	if subs, exists := hub.ChatSubscriptions[uint(chatID)]; exists {
		// 		delete(subs, c)
		// 		if len(subs) == 0 {
		// 			delete(hub.ChatSubscriptions, uint(chatID))
		// 		}
		// 	}
		// 	hub.Mutex.Unlock()
		// 	logrus.Infof("readPump: клиент %s отписался от чата %d", c.UserID, chatID)
		// case "heartbeat":
		// 	// Если клиент отправил heartbeat, то ожидаем, что поле is_online передано
		// 	if req.IsOnline != nil {
		// 		logrus.Debugf("readPump: получен heartbeat от клиента %s, is_online: %v", c.UserID, *req.IsOnline)
		// 		// Обновляем онлайн-статус пользователя в базе данных.
		// 		go updateUserOnlineStatus(c.UserID, *req.IsOnline)
		// 	} else {
		// 		logrus.Warnf("readPump: heartbeat от клиента %s без поля is_online", c.UserID)
		// 	}
		// default:
		// 	logrus.Warnf("readPump: неизвестное действие '%s' от клиента %s", req.Action, c.UserID)
		// }

		var req struct {
			Action   string `json:"action"`
			ChatID   string `json:"chat_id"`   // вместо "chatId"
			IsOnline *bool  `json:"is_online"` // вместо "isOnline"
			IsTyping *bool  `json:"is_typing"` // вместо "isTyping"
		}

		// var req struct {
		// 	Action   string `json:"action"`
		// 	ChatID   string `json:"chatId"`   // требуется только для subscribe/unsubscribe
		// 	IsOnline *bool  `json:"isOnline"` // используется для heartbeat
		// 	IsTyping *bool  `json:"isTyping"` // новое поле для набора текста
		// }
		if err := json.Unmarshal(msg, &req); err != nil {
			logrus.Errorf("readPump: ошибка разбора сообщения от клиента %s: %v", c.UserID, err)
			continue
		}

		switch req.Action {
		case "subscribe", "unsubscribe":
			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
			if err != nil {
				logrus.Warnf("readPump: неверный chat_id '%s' от клиента %s", req.ChatID, c.UserID)
				continue
			}
			c.Mutex.Lock()
			if req.Action == "subscribe" {
				// логика подписки
				c.Chats[uint(chatID)] = true
				hub.Mutex.Lock()
				if hub.ChatSubscriptions[uint(chatID)] == nil {
					hub.ChatSubscriptions[uint(chatID)] = make(map[*Client]bool)
				}
				hub.ChatSubscriptions[uint(chatID)][c] = true
				hub.Mutex.Unlock()
				logrus.Infof("readPump: клиент %s подписался на чат %d", c.UserID, chatID)
			} else {
				// логика отписки
				delete(c.Chats, uint(chatID))
				hub.Mutex.Lock()
				if subs, exists := hub.ChatSubscriptions[uint(chatID)]; exists {
					delete(subs, c)
					if len(subs) == 0 {
						delete(hub.ChatSubscriptions, uint(chatID))
					}
				}
				hub.Mutex.Unlock()
				logrus.Infof("readPump: клиент %s отписался от чата %d", c.UserID, chatID)
			}
			c.Mutex.Unlock()
		case "heartbeat":
			if req.IsOnline != nil {
				logrus.Debugf("readPump: получен heartbeat от клиента %s, is_online: %v", c.UserID, *req.IsOnline)
				go updateUserOnlineStatus(c.UserID, *req.IsOnline)
			} else {
				logrus.Warnf("readPump: heartbeat от клиента %s без поля is_online", c.UserID)
			}

		case "typing":
			// Обработка события набора текста
			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
			if err != nil {
				logrus.Warnf("readPump: неверный chat_id '%s' в сообщении typing от клиента %s", req.ChatID, c.UserID)
				continue
			}
			if req.IsTyping == nil {
				logrus.Warnf("readPump: поле is_typing отсутствует в сообщении typing от клиента %s", c.UserID)
				continue
			}
			c.Mutex.Lock()
			// Обновляем состояние набора текста для конкретного чата
			c.TypingChats[uint(chatID)] = *req.IsTyping
			c.Mutex.Unlock()
			// Отправляем уведомление другим участникам чата
			uid, err := uuid.Parse(c.UserID)
			if err != nil {
				logrus.Errorf("readPump: не удалось распарсить userID клиента %s: %v", c.UserID, err)
				continue
			}
			go BroadcastTypingNotification(uid, uint(chatID), *req.IsTyping)

		default:
			logrus.Warnf("readPump: неизвестное действие '%s' от клиента %s", req.Action, c.UserID)
		}

		//c.Mutex.Unlock()
	}
}

// writePump отправляет сообщения клиенту и обрабатывает пинги.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		logrus.Infof("writePump: закрыто соединение клиента %s", c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logrus.Errorf("writePump: ошибка получения писателя для клиента %s: %v", c.UserID, err)
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				logrus.Errorf("writePump: ошибка закрытия писателя для клиента %s: %v", c.UserID, err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logrus.Errorf("writePump: ошибка отправки ping для клиента %s: %v", c.UserID, err)
				return
			}
		}
	}
}

// BroadcastNewMessage сериализует объект модели (например, models.Message) и отправляет его в hub.
func BroadcastNewMessage(msg models.Message) error {
	//data, err := json.Marshal(msg)
	// ▶ оборачиваем сообщение в единый протокол с полем type
	payload := map[string]interface{}{
		"type":      "message",
		"chat_id":   msg.ChatID,
		"id":        msg.ID,
		"sender_id": msg.SenderID.String(),
		"content":   msg.Content,
		"timestamp": msg.Timestamp.Unix(),
		"read":      msg.Read,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("BroadcastNewMessage: ошибка маршалинга сообщения: %v", err)
		return err
	}
	hub.Broadcast <- BroadcastMessage{
		ChatID: msg.ChatID,
		Data:   data,
	}
	logrus.Infof("BroadcastNewMessage: сообщение отправлено в чат %d", msg.ChatID)
	return nil
}

// BroadcastNotification отправляет уведомление конкретному пользователю (если он онлайн) по его userID.
func BroadcastNotification(userID uuid.UUID, message string) {
	data, err := json.Marshal(map[string]string{
		"type":    "notification",
		"message": message,
	})
	if err != nil {
		logrus.Errorf("BroadcastNotification: ошибка маршалинга уведомления: %v", err)
		return
	}

	hub.Mutex.RLock()
	defer hub.Mutex.RUnlock()
	for client := range hub.Clients {
		if client.UserID == userID.String() {
			select {
			case client.Send <- data:
				logrus.Infof("BroadcastNotification: уведомление отправлено пользователю %s", userID)
			default:
				logrus.Warnf("BroadcastNotification: канал уведомлений для пользователя %s переполнен", userID)
			}
		}
	}
}

// BroadcastTypingNotification отправляет уведомление о наборе текста в чат.
func BroadcastTypingNotification(userID uuid.UUID, chatID uint, isTyping bool) {
	data, err := json.Marshal(map[string]interface{}{
		"type":      "typing",
		"user_id":   userID.String(),
		"chat_id":   chatID,
		"is_typing": isTyping,
		"timestamp": time.Now().Unix(),
	})
	if err != nil {
		logrus.Errorf("BroadcastTypingNotification: ошибка маршалинга уведомления: %v", err)
		return
	}

	hub.Mutex.RLock()
	subscribers, ok := hub.ChatSubscriptions[chatID]
	hub.Mutex.RUnlock()
	if !ok {
		logrus.Debugf("BroadcastTypingNotification: подписчиков для чата %d не найдено", chatID)
		return
	}

	for client := range subscribers {
		select {
		case client.Send <- data:
			logrus.Infof("BroadcastTypingNotification: уведомление о наборе текста отправлено пользователю %s в чате %d", client.UserID, chatID)
		default:
			logrus.Warnf("BroadcastTypingNotification: канал уведомлений для пользователя %s переполнен", client.UserID)
		}
	}
}

// RunWebSocketServer запускает HTTP-сервер для WebSocket-соединений на заданном адресе.
func RunWebSocketServer(addr string) error {
	go RunHub()
	http.HandleFunc("/ws", HandleWebSocket)
	logrus.Infof("WebSocket server started on %s", addr)
	return http.ListenAndServe(addr, nil)
}
