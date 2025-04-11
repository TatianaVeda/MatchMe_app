package websocket

import (
	"log"
	"net/http"
	"sync"
	"strconv"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[uint]*websocket.Conn) // Map user IDs to WebSocket connections
	Broadcast = make(chan Message)             // Broadcast channel
	mu        sync.Mutex                       // Mutex for clients map
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allows connections from any origin
		},
	}
)

// Message represents a message sent via WebSocket
type Message struct {
	ID         uint   `json:"id"`          // Unique message ID
	SenderID   uint   `json:"sender_id"`   // ID of the sender
	ReceiverID uint   `json:"receiver_id"` // ID of the receiver
	Message    string `json:"message"`     // Message content
	Timestamp  string `json:"timestamp"`   // Timestamp of the message
	IsRead     bool   `json:"is_read"`     // Whether the message has been read
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return err
	}
	defer conn.Close()

	// Extract user ID from query parameters or headers
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.Println("User ID is required")
		return nil // You might want to return an error if user ID is missing
	}

	// Convert userID to uint
	uid := parseUserID(userID)
	if uid == 0 {
		log.Println("Invalid user ID:", userID)
		return nil // You might want to return an error for invalid UID
	}

	// Add the connection to the clients map
	mu.Lock()
	clients[uid] = conn
	mu.Unlock()

	// Handle incoming messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			cleanupConnection(uid)
			break
		}

		// Broadcast the message to the intended recipient
		Broadcast <- msg
	}

	return nil // No error in successful execution
}

// BroadcastMessages broadcasts messages to the intended recipient
func BroadcastMessages() {
	for msg := range Broadcast {
		mu.Lock()
		if conn, ok := clients[msg.ReceiverID]; ok {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing JSON to user %d: %v", msg.ReceiverID, err)
				cleanupConnection(msg.ReceiverID)
			}
		}
		mu.Unlock()
	}
}

// parseUserID converts a string user ID to uint
func parseUserID(userID string) uint {
	uid, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Invalid user ID format: %v", userID)
		return 0
	}
	return uint(uid)
}

// cleanupConnection removes a WebSocket connection and its user from the clients map
func cleanupConnection(uid uint) {
	mu.Lock()
	defer mu.Unlock()
	if conn, ok := clients[uid]; ok {
		conn.Close()
		delete(clients, uid)
		log.Printf("Connection for user %d closed and removed", uid)
	}
}
