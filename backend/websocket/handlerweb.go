package websocket

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// WebSocketHandler handles the incoming WebSocket requests and establishes a WebSocket connection.
func WebSocketHandler(c *gin.Context) {
	// Try to upgrade the HTTP connection to WebSocket.
	if err := HandleWebSocket(c.Writer, c.Request); err != nil {
		// Log the error if WebSocket connection fails.
		log.Printf("Error establishing WebSocket connection: %v", err)
		// Send an error response to the client.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}

	// Successfully established the WebSocket connection.
	c.Status(http.StatusOK)
}
