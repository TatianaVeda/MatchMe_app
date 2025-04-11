package main

import (
	"backend/database"
	"backend/utils"
	"fmt"
	"log"
)

func main() {
	// Establish a connection to the database
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Begin generating predefined users
	fmt.Println("Starting the generation of predefined users...")

	// Generate and insert predefined users
	if err := utils.GeneratePredefinedUsers(); err != nil {
		log.Fatalf("Error generating predefined users: %v", err)
	}

	// Notify user of successful completion
	fmt.Println("Predefined user generation completed successfully!")
}
