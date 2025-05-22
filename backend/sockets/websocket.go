// // package sockets

// // import (
// // 	"encoding/json"
// // 	"net/http"
// // 	"strconv"
// // 	"sync"
// // 	"time"

// // 	"m/backend/models"
// // 	"m/backend/services"

// // 	"github.com/google/uuid"
// // 	"github.com/gorilla/websocket"
// // 	"github.com/sirupsen/logrus"
// // )

// // // presenceSvc хранит ссылку на PresenceService, установленную из main.go
// // var presenceSvc *services.PresenceService

// // // BroadcastMessage определяет структуру сообщения для рассылки по чатам.
// // type BroadcastMessage struct {
// // 	ChatID uint   `json:"chatId"`
// // 	Data   []byte `json:"data"`
// // }

// // // Client представляет одно WebSocket‑подключение.
// // type Client struct {
// // 	Conn   *websocket.Conn // WebSocket‑соединение
// // 	UserID string          // Идентификатор пользователя (например, извлекается из запроса)
// // 	Send   chan []byte     // Канал для отправки сообщений клиенту
// // 	Chats  map[uint]bool   // Список подписанных чатов (chatID -> true)
// // 	// Новое поле: для хранения состояния набора текста для каждого чата
// // 	TypingChats map[uint]bool
// // 	Mutex       sync.Mutex // Для синхронизации доступа к полю Chats
// // }

// // // Hub управляет клиентами и рассылкой сообщений по чатам.
// // type Hub struct {
// // 	Clients           map[*Client]bool          // Все активные клиенты
// // 	ChatSubscriptions map[uint]map[*Client]bool // Для каждого chatID множество клиентов, подписанных на него
// // 	Broadcast         chan BroadcastMessage     // Канал для рассылки сообщений в чаты
// // 	Register          chan *Client              // Новый клиент
// // 	Unregister        chan *Client              // Клиент отключается
// // 	Mutex             sync.RWMutex              // Защита общих структур
// // }

// // // Инициализируем глобальный hub.
// // var hub = Hub{
// // 	Clients:           make(map[*Client]bool),
// // 	ChatSubscriptions: make(map[uint]map[*Client]bool),
// // 	Broadcast:         make(chan BroadcastMessage),
// // 	Register:          make(chan *Client),
// // 	Unregister:        make(chan *Client),
// // }

// // // IsUserTypingInChat проверяет, набирает ли пользователь с заданным userID текст в чате chatID.
// // func IsUserTypingInChat(chatID uint, userID string) bool {
// // 	hub.Mutex.RLock()
// // 	defer hub.Mutex.RUnlock()
// // 	subs, ok := hub.ChatSubscriptions[chatID]
// // 	if !ok {
// // 		return false
// // 	}
// // 	for client := range subs {
// // 		if client.UserID == userID {
// // 			client.Mutex.Lock()
// // 			typing, exists := client.TypingChats[chatID]
// // 			client.Mutex.Unlock()
// // 			if exists && typing {
// // 				return true
// // 			}
// // 		}
// // 	}
// // 	return false
// // }

// // // RunHub запускает цикл обработки регистрации, отмены регистрации и рассылки сообщений.
// // func RunHub() {
// // 	for {
// // 		select {
// // 		case client := <-hub.Register:
// // 			hub.Mutex.Lock()
// // 			hub.Clients[client] = true
// // 			hub.Mutex.Unlock()
// // 			logrus.Infof("Client registered: %s", client.UserID)
// // 		case client := <-hub.Unregister:
// // 			hub.Mutex.Lock()
// // 			if _, ok := hub.Clients[client]; ok {
// // 				delete(hub.Clients, client)
// // 				// Удаляем клиента из всех подписок
// // 				for chatID := range client.Chats {
// // 					if subs, exists := hub.ChatSubscriptions[chatID]; exists {
// // 						delete(subs, client)
// // 						if len(subs) == 0 {
// // 							delete(hub.ChatSubscriptions, chatID)
// // 						}
// // 					}
// // 				}
// // 				close(client.Send)
// // 				logrus.Infof("Client unregistered: %s", client.UserID)
// // 			}
// // 			hub.Mutex.Unlock()
// // 		case msg := <-hub.Broadcast:
// // 			// Отправляем сообщение только клиентам, подписанным на данный чат.
// // 			hub.Mutex.RLock()
// // 			if subs, ok := hub.ChatSubscriptions[msg.ChatID]; ok {
// // 				for client := range subs {
// // 					select {
// // 					case client.Send <- msg.Data:
// // 						logrus.Debugf("Message sent to client %s in chat %d", client.UserID, msg.ChatID)
// // 					default:
// // 						logrus.Warnf("Send channel full for client %s. Closing connection.", client.UserID)
// // 						close(client.Send)
// // 						delete(hub.Clients, client)
// // 					}
// // 				}
// // 			} else {
// // 				logrus.Debugf("No subscribers for chat %d", msg.ChatID)
// // 			}
// // 			hub.Mutex.RUnlock()
// // 		}
// // 	}
// // }

// // const (
// // 	writeWait  = 10 * time.Second
// // 	pongWait   = 60 * time.Second
// // 	pingPeriod = (pongWait * 9) / 10
// // )

// // var upgrader = websocket.Upgrader{
// // 	ReadBufferSize:  1024,
// // 	WriteBufferSize: 1024,
// // 	// В продакшене настройте CheckOrigin по необходимости.
// // 	CheckOrigin: func(r *http.Request) bool {
// // 		return true
// // 	},
// // }

// // // HandleWebSocket выполняет апгрейд HTTP-запроса до WebSocket-соединения.
// // // Здесь в качестве простоты userID извлекается из query-параметра "userID".
// // // func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// // // 	conn, err := upgrader.Upgrade(w, r, nil)
// // // 	if err != nil {
// // // 		logrus.Errorf("WebSocket upgrade error: %v", err)
// // // 		return
// // // 	}

// // // 	userID := r.URL.Query().Get("userID")
// // // 	if userID == "" {
// // // 		userID = uuid.New().String()
// // // 		logrus.Debug("HandleWebSocket: userID не передан, сгенерирован новый UUID")
// // // 	}

// // // 	client := &Client{
// // // 		Conn:        conn,
// // // 		Send:        make(chan []byte, 256),
// // // 		UserID:      userID,
// // // 		Chats:       make(map[uint]bool),
// // // 		TypingChats: make(map[uint]bool),
// // // 	}

// // // 	hub.Register <- client
// // // 	logrus.Infof("HandleWebSocket: client %s подключен", client.UserID)

// // // 	go client.writePump()
// // // 	client.readPump()
// // // }

// // func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// // 	conn, err := upgrader.Upgrade(w, r, nil)
// // 	if err != nil {
// // 		logrus.Errorf("WebSocket upgrade error: %v", err)
// // 		return
// // 	}

// // 	userID := r.URL.Query().Get("userID")
// // 	if userID == "" {
// // 		userID = uuid.New().String()
// // 		logrus.Debug("HandleWebSocket: userID не передан, сгенерирован новый UUID")
// // 	}

// // 	client := &Client{
// // 		Conn:        conn,
// // 		Send:        make(chan []byte, 256),
// // 		UserID:      userID,
// // 		Chats:       make(map[uint]bool),
// // 		TypingChats: make(map[uint]bool),
// // 	}

// // 	// ✅ Mark user as online
// // 	if presenceSvc != nil {
// // 		if err := presenceSvc.Touch(userID); err != nil {
// // 			logrus.Warnf("Failed to update presence for %s: %v", userID, err)
// // 		}
// // 	}

// // 	hub.Register <- client
// // 	logrus.Infof("HandleWebSocket: client %s подключен", client.UserID)

// // 	go client.writePump()

// // 	// ✅ Start listening and updating presence on each read
// // 	client.readPump()

// // 	// ⚠️ When client exits, no need to delete key — TTL will expire it
// // }

// // // readPump читает входящие сообщения от клиента.
// // // Ожидается, что клиент будет отправлять JSON-сообщения для подписки/отписки.
// // func (c *Client) readPump() {
// // 	defer func() {
// // 		hub.Unregister <- c
// // 		c.Conn.Close()
// // 		logrus.Infof("readPump: закрыто соединение клиента %s", c.UserID)

// // 		// 🧹 Удаляем presence-ключ при отключении
// // 		if presenceSvc != nil {
// // 			if err := presenceSvc.Rdb.Del(presenceSvc.Ctx, services.PresencePrefix+c.UserID).Err(); err != nil {
// // 				logrus.Warnf("readPump: не удалось удалить presence ключ клиента %s: %v", c.UserID, err)
// // 			}
// // 		}
// // 	}()

// // 	c.Conn.SetReadLimit(512)
// // 	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
// // 	c.Conn.SetPongHandler(func(appData string) error {
// // 		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
// // 		return nil
// // 	})

// // 	for {
// // 		_, msg, err := c.Conn.ReadMessage()
// // 		if err != nil {
// // 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// // 				logrus.Errorf("readPump: неожиданная ошибка WebSocket для клиента %s: %v", c.UserID, err)
// // 			} else {
// // 				logrus.Debugf("readPump: клиент %s отключился: %v", c.UserID, err)
// // 			}
// // 			break
// // 		}

// // 		// var req struct {
// // 		// 	Action   string `json:"action"`
// // 		// 	ChatID   string `json:"chatId"`   // для подписки/отписки, если необходимо
// // 		// 	IsOnline *bool  `json:"isOnline"` // для heartbeat-сообщения
// // 		// }
// // 		// if err := json.Unmarshal(msg, &req); err != nil {
// // 		// 	logrus.Errorf("readPump: ошибка разбора сообщения от клиента %s: %v", c.UserID, err)
// // 		// 	continue
// // 		// }

// // 		// chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
// // 		// if err != nil {
// // 		// 	logrus.Warnf("readPump: неверный chat_id '%s' от клиента %s", req.ChatID, c.UserID)
// // 		// 	continue
// // 		// }

// // 		// c.Mutex.Lock()
// // 		// switch req.Action {
// // 		// case "subscribe":
// // 		// 	c.Chats[uint(chatID)] = true
// // 		// 	hub.Mutex.Lock()
// // 		// 	if hub.ChatSubscriptions[uint(chatID)] == nil {
// // 		// 		hub.ChatSubscriptions[uint(chatID)] = make(map[*Client]bool)
// // 		// 	}
// // 		// 	hub.ChatSubscriptions[uint(chatID)][c] = true
// // 		// 	hub.Mutex.Unlock()
// // 		// 	logrus.Infof("readPump: клиент %s подписался на чат %d", c.UserID, chatID)
// // 		// case "unsubscribe":
// // 		// 	delete(c.Chats, uint(chatID))
// // 		// 	hub.Mutex.Lock()
// // 		// 	if subs, exists := hub.ChatSubscriptions[uint(chatID)]; exists {
// // 		// 		delete(subs, c)
// // 		// 		if len(subs) == 0 {
// // 		// 			delete(hub.ChatSubscriptions, uint(chatID))
// // 		// 		}
// // 		// 	}
// // 		// 	hub.Mutex.Unlock()
// // 		// 	logrus.Infof("readPump: клиент %s отписался от чата %d", c.UserID, chatID)
// // 		// case "heartbeat":
// // 		// 	// Если клиент отправил heartbeat, то ожидаем, что поле is_online передано
// // 		// 	if req.IsOnline != nil {
// // 		// 		logrus.Debugf("readPump: получен heartbeat от клиента %s, is_online: %v", c.UserID, *req.IsOnline)
// // 		// 		// Обновляем онлайн-статус пользователя в базе данных.
// // 		// 		go updateUserOnlineStatus(c.UserID, *req.IsOnline)
// // 		// 	} else {
// // 		// 		logrus.Warnf("readPump: heartbeat от клиента %s без поля is_online", c.UserID)
// // 		// 	}
// // 		// default:
// // 		// 	logrus.Warnf("readPump: неизвестное действие '%s' от клиента %s", req.Action, c.UserID)
// // 		// }

// // 		var req struct {
// // 			Action   string `json:"action"`
// // 			ChatID   string `json:"chat_id"`   // вместо "chatId"
// // 			IsOnline *bool  `json:"is_online"` // вместо "isOnline"
// // 			IsTyping *bool  `json:"is_typing"` // вместо "isTyping"
// // 		}

// // 		// var req struct {
// // 		// 	Action   string `json:"action"`
// // 		// 	ChatID   string `json:"chatId"`   // требуется только для subscribe/unsubscribe
// // 		// 	IsOnline *bool  `json:"isOnline"` // используется для heartbeat
// // 		// 	IsTyping *bool  `json:"isTyping"` // новое поле для набора текста
// // 		// }
// // 		if err := json.Unmarshal(msg, &req); err != nil {
// // 			logrus.Errorf("readPump: ошибка разбора сообщения от клиента %s: %v", c.UserID, err)
// // 			continue
// // 		}

// // 		switch req.Action {
// // 		case "subscribe", "unsubscribe":
// // 			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
// // 			if err != nil {
// // 				logrus.Warnf("readPump: неверный chat_id '%s' от клиента %s", req.ChatID, c.UserID)
// // 				continue
// // 			}
// // 			c.Mutex.Lock()
// // 			if req.Action == "subscribe" {
// // 				// логика подписки
// // 				c.Chats[uint(chatID)] = true
// // 				hub.Mutex.Lock()
// // 				if hub.ChatSubscriptions[uint(chatID)] == nil {
// // 					hub.ChatSubscriptions[uint(chatID)] = make(map[*Client]bool)
// // 				}
// // 				hub.ChatSubscriptions[uint(chatID)][c] = true
// // 				hub.Mutex.Unlock()
// // 				logrus.Infof("readPump: клиент %s подписался на чат %d", c.UserID, chatID)
// // 			} else {
// // 				// логика отписки
// // 				delete(c.Chats, uint(chatID))
// // 				hub.Mutex.Lock()
// // 				if subs, exists := hub.ChatSubscriptions[uint(chatID)]; exists {
// // 					delete(subs, c)
// // 					if len(subs) == 0 {
// // 						delete(hub.ChatSubscriptions, uint(chatID))
// // 					}
// // 				}
// // 				hub.Mutex.Unlock()
// // 				logrus.Infof("readPump: клиент %s отписался от чата %d", c.UserID, chatID)
// // 			}
// // 			c.Mutex.Unlock()
// // 		case "heartbeat":
// // 			// if req.IsOnline != nil {
// // 			// 	logrus.Debugf("readPump: получен heartbeat от клиента %s, is_online: %v", c.UserID, *req.IsOnline)
// // 			// 	go updateUserOnlineStatus(c.UserID, *req.IsOnline)
// // 			// } else {
// // 			// 	logrus.Warnf("readPump: heartbeat от клиента %s без поля is_online", c.UserID)
// // 			// }

// // 			logrus.Debugf("readPump: получен heartbeat от клиента %s", c.UserID)
// // 			// Обновляем TTL в Redis: считаем пользователя онлайн
// // 			if presenceSvc != nil {
// // 				if err := presenceSvc.Touch(c.UserID); err != nil {
// // 					logrus.Warnf("presence.Touch failed for %s: %v", c.UserID, err)
// // 				}
// // 			} else {
// // 				logrus.Warn("presenceSvc не инициализирован, heartbeat не обработан")
// // 			}

// // 		case "typing":
// // 			// Обработка события набора текста
// // 			chatID, err := strconv.ParseUint(req.ChatID, 10, 64)
// // 			if err != nil {
// // 				logrus.Warnf("readPump: неверный chat_id '%s' в сообщении typing от клиента %s", req.ChatID, c.UserID)
// // 				continue
// // 			}
// // 			if req.IsTyping == nil {
// // 				logrus.Warnf("readPump: поле is_typing отсутствует в сообщении typing от клиента %s", c.UserID)
// // 				continue
// // 			}
// // 			c.Mutex.Lock()
// // 			// Обновляем состояние набора текста для конкретного чата
// // 			c.TypingChats[uint(chatID)] = *req.IsTyping
// // 			c.Mutex.Unlock()
// // 			// Отправляем уведомление другим участникам чата
// // 			uid, err := uuid.Parse(c.UserID)
// // 			if err != nil {
// // 				logrus.Errorf("readPump: не удалось распарсить userID клиента %s: %v", c.UserID, err)
// // 				continue
// // 			}
// // 			go BroadcastTypingNotification(uid, uint(chatID), *req.IsTyping)

// // 		default:
// // 			logrus.Warnf("readPump: неизвестное действие '%s' от клиента %s", req.Action, c.UserID)
// // 		}

// // 		//c.Mutex.Unlock()
// // 	}
// // }

// // // writePump отправляет сообщения клиенту и обрабатывает пинги.
// // func (c *Client) writePump() {
// // 	ticker := time.NewTicker(pingPeriod)
// // 	defer func() {
// // 		ticker.Stop()
// // 		c.Conn.Close()
// // 		logrus.Infof("writePump: закрыто соединение клиента %s", c.UserID)
// // 	}()

// // 	for {
// // 		select {
// // 		case message, ok := <-c.Send:
// // 			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
// // 			if !ok {
// // 				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
// // 				return
// // 			}
// // 			w, err := c.Conn.NextWriter(websocket.TextMessage)
// // 			if err != nil {
// // 				logrus.Errorf("writePump: ошибка получения писателя для клиента %s: %v", c.UserID, err)
// // 				return
// // 			}
// // 			w.Write(message)

// // 			n := len(c.Send)
// // 			for i := 0; i < n; i++ {
// // 				w.Write([]byte("\n"))
// // 				w.Write(<-c.Send)
// // 			}

// // 			if err := w.Close(); err != nil {
// // 				logrus.Errorf("writePump: ошибка закрытия писателя для клиента %s: %v", c.UserID, err)
// // 				return
// // 			}
// // 		case <-ticker.C:
// // 			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
// // 			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// // 				logrus.Errorf("writePump: ошибка отправки ping для клиента %s: %v", c.UserID, err)
// // 				return
// // 			}
// // 		}
// // 	}
// // }

// // // BroadcastNewMessage сериализует объект модели (например, models.Message) и отправляет его в hub.
// // func BroadcastNewMessage(msg models.Message) error {
// // 	//data, err := json.Marshal(msg)
// // 	// ▶ оборачиваем сообщение в единый протокол с полем type
// // 	payload := map[string]interface{}{
// // 		"type":      "message",
// // 		"chat_id":   msg.ChatID,
// // 		"id":        msg.ID,
// // 		"sender_id": msg.SenderID.String(),
// // 		"content":   msg.Content,
// // 		//"timestamp": msg.Timestamp.Unix(),
// // 		"timestamp": msg.Timestamp.UnixNano() / int64(time.Millisecond),
// // 		"read":      msg.Read,
// // 	}
// // 	data, err := json.Marshal(payload)
// // 	if err != nil {
// // 		logrus.Errorf("BroadcastNewMessage: ошибка маршалинга сообщения: %v", err)
// // 		return err
// // 	}
// // 	hub.Broadcast <- BroadcastMessage{
// // 		ChatID: msg.ChatID,
// // 		Data:   data,
// // 	}
// // 	logrus.Infof("BroadcastNewMessage: сообщение отправлено в чат %d", msg.ChatID)
// // 	return nil
// // }

// // // BroadcastNotification отправляет уведомление конкретному пользователю (если он онлайн) по его userID.
// // func BroadcastNotification(userID uuid.UUID, message string) {
// // 	data, err := json.Marshal(map[string]string{
// // 		"type":    "notification",
// // 		"message": message,
// // 	})
// // 	if err != nil {
// // 		logrus.Errorf("BroadcastNotification: ошибка маршалинга уведомления: %v", err)
// // 		return
// // 	}

// // 	hub.Mutex.RLock()
// // 	defer hub.Mutex.RUnlock()
// // 	for client := range hub.Clients {
// // 		if client.UserID == userID.String() {
// // 			select {
// // 			case client.Send <- data:
// // 				logrus.Infof("BroadcastNotification: уведомление отправлено пользователю %s", userID)
// // 			default:
// // 				logrus.Warnf("BroadcastNotification: канал уведомлений для пользователя %s переполнен", userID)
// // 			}
// // 		}
// // 	}
// // }

// // // BroadcastTypingNotification отправляет уведомление о наборе текста в чат.
// // func BroadcastTypingNotification(userID uuid.UUID, chatID uint, isTyping bool) {
// // 	data, err := json.Marshal(map[string]interface{}{
// // 		"type":      "typing",
// // 		"user_id":   userID.String(),
// // 		"chat_id":   chatID,
// // 		"is_typing": isTyping,
// // 		"timestamp": time.Now().Unix(),
// // 	})
// // 	if err != nil {
// // 		logrus.Errorf("BroadcastTypingNotification: ошибка маршалинга уведомления: %v", err)
// // 		return
// // 	}

// // 	hub.Mutex.RLock()
// // 	subscribers, ok := hub.ChatSubscriptions[chatID]
// // 	hub.Mutex.RUnlock()
// // 	if !ok {
// // 		logrus.Debugf("BroadcastTypingNotification: подписчиков для чата %d не найдено", chatID)
// // 		return
// // 	}

// // 	for client := range subscribers {
// // 		select {
// // 		case client.Send <- data:
// // 			logrus.Infof("BroadcastTypingNotification: уведомление о наборе текста отправлено пользователю %s в чате %d", client.UserID, chatID)
// // 		default:
// // 			logrus.Warnf("BroadcastTypingNotification: канал уведомлений для пользователя %s переполнен", client.UserID)
// // 		}
// // 	}
// // }

// // // InitWebSocketServer инициализирует presenceSvc, а затем запускает хаб и HTTP-сервер.
// // func InitWebSocketServer(ps *services.PresenceService, addr string) error {
// // 	// Сохраняем сервис
// // 	presenceSvc = ps
// // 	// Запускаем цикл хаба
// // 	go RunHub()
// // 	// Регистрируем обработчик WS
// // 	http.HandleFunc("/ws", HandleWebSocket)
// // 	logrus.Infof("WebSocket server started on %s", addr)
// // 	return http.ListenAndServe(addr, nil)
// // }

// RunWebSocketServer запускает HTTP-сервер для WebSocket-соединений на заданном адресе.
// func RunWebSocketServer(addr string) error {
// 	go RunHub()
// 	http.HandleFunc("/ws", HandleWebSocket)
// 	logrus.Infof("WebSocket server started on %s", addr)
// 	return http.ListenAndServe(addr, nil)
// }

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

// presenceSvc хранит ссылку на PresenceService, установленную из main.go
var presenceSvc *services.PresenceService

// BroadcastMessage определяет структуру сообщения для рассылки по чатам.
type BroadcastMessage struct {
	ChatID uint   `json:"chatId"`
	Data   []byte `json:"data"`
}

// Client представляет одно WebSocket-подключение.
type Client struct {
	Conn        *websocket.Conn // WebSocket-соединение
	UserID      string          // Идентификатор пользователя
	Send        chan []byte     // Канал для отправки сообщений клиенту
	Chats       map[uint]bool   // Список подписанных чатов (chatID -> true)
	TypingChats map[uint]bool   // Состояние «печатает» для каждого чата
	Mutex       sync.Mutex      // Синхронизация доступа к Chats и TypingChats
}

// Hub управляет клиентами и рассылкой сообщений по чатам.
type Hub struct {
	Clients           map[*Client]bool          // Все активные клиенты
	ChatSubscriptions map[uint]map[*Client]bool // Для каждого chatID — подписанные клиенты
	Broadcast         chan BroadcastMessage     // Канал для рассылки сообщений
	Register          chan *Client              // Канал для регистрации новых клиентов
	Unregister        chan *Client              // Канал для отключения клиентов
	Mutex             sync.RWMutex              // Защита общих структур
}

// Глобальный экземпляр хаба
var hub = Hub{
	Clients:           make(map[*Client]bool),
	ChatSubscriptions: make(map[uint]map[*Client]bool),
	Broadcast:         make(chan BroadcastMessage),
	Register:          make(chan *Client),
	Unregister:        make(chan *Client),
}

// IsUserTypingInChat проверяет, печатает ли пользователь userID в чате chatID.
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

// RunHub запускает главный цикл: регистрация, отписка и рассылка.
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
				// Убираем из всех подписок
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
	CheckOrigin:     func(r *http.Request) bool { return true }, // в проде сузьте
}

// HandleWebSocket апгрейдит HTTP->WebSocket, регистрирует клиента и стартует read/write-петли.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		userID = uuid.New().String()
		logrus.Debug("HandleWebSocket: сгенерирован userID")
	}

	client := &Client{
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		Chats:       make(map[uint]bool),
		TypingChats: make(map[uint]bool),
	}

	// Отметить онлайн в Redis
	if presenceSvc != nil {
		if err := presenceSvc.Touch(userID); err != nil {
			logrus.Warnf("presence.Touch failed for %s: %v", userID, err)
		}
	}

	hub.Register <- client
	logrus.Infof("HandleWebSocket: клиент %s подключён", client.UserID)

	go client.writePump()
	client.readPump()
}

// readPump слушает клиентские сообщения: подписка/отписка, heartbeat, typing.
func (c *Client) readPump() {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
		logrus.Infof("readPump: закрыто соединение %s", c.UserID)
		// Очистка presence
		if presenceSvc != nil {
			_ = presenceSvc.Rdb.Del(presenceSvc.Ctx, services.PresencePrefix+c.UserID).Err()
		}
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
			//go updateUserOnlineStatus(c.UserID, *req.IsOnline)!!!!!!!!!!!!!!!!!!!!!!!
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

// writePump шлёт сообщения по каналу Send и отправляет пинги.
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

			// Отправляем всё, что накопилось
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
	// Составляем полное имя отправителя из загруженного msg.Sender.Profile
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
		"sender_name": senderName, // теперь берётся из Profile
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

// BroadcastNotification шлёт нотификацию конкретному пользователю.
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

// BroadcastTypingNotification рассылает «печатает…» всем в чате.
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

// BroadcastConnectionRequest посылает всем WS-клиентам именно событие о новом запросе в друзья.
func BroadcastConnectionRequest(userID uuid.UUID) {
	payload := map[string]interface{}{
		"type": "connection_request",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("BroadcastConnectionRequest: marshal error: %v", err)
		return
	}

	// шлём только тому клиенту, чей userID совпал
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

// InitWebSocketServer инициализирует presenceSvc, запускает хаб и HTTP-сервер.
func InitWebSocketServer(ps *services.PresenceService, addr string) error {
	presenceSvc = ps
	go RunHub()
	http.HandleFunc("/ws", HandleWebSocket)
	logrus.Infof("WebSocket server started on %s", addr)
	return http.ListenAndServe(addr, nil)
}
