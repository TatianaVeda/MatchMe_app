package main

import (
	"backend/database"
	"backend/models"
	"backend/routes"
	"backend/websocket"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Initialize the database connection
	db := database.ConnectDB()

	// Apply migrations explicitly if needed
	if err := db.AutoMigrate(&models.Recommendation{}); err != nil {
		log.Printf("Warning: Error migrating Recommendation model: %v", err)
	}

	// Force Gin logs to be colored
	gin.ForceConsoleColor()
	router := gin.Default()

	// Configure CORS to allow frontend communication
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Register routes
	routes.RegisterRoutes(router)

	// Setup WebSocket handler
	router.GET("/ws", websocket.WebSocketHandler)

	// Serve uploaded avatar images
	router.Static("/uploads/avatars", "./uploads/avatars")

	// Add a friendly message at the root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "🚀 Match-Me backend is running! Visit /api endpoints for data access.",
		})
	})

	// Start a goroutine to handle WebSocket message broadcasting
	go websocket.BroadcastMessages()

	// Retrieve the port from environment variables, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddress := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on port %s...", port)

	// Start the server
	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}