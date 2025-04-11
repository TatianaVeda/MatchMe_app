package main

import (
	"backend/database"
	"backend/utils"
	"fmt"
	"log"
)

func main() {
	// Establish database connection
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Start user generation
	fmt.Println("Starting the generation of 100 random users...")
	
	// Generate and insert random users
	if err := utils.GenerateUsers(); err != nil {
		log.Fatalf("Error generating users: %v", err)
	}
	
	// Notify completion
	fmt.Println("User generation completed successfully!")
}
