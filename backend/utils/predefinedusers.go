package utils

import (
	"fmt"
	"m/backend/database"
	"m/backend/models"
)

func GeneratePredefinedUsers() {
	predefinedUsers := []models.User{
		{
			Username:         "Predefined1",
			FirstName:        "John",
			LastName:         "Doe",
			Email:            "predefined1@mail.com",
			Password:         "1234",
			Location:         "Helsinki",
			Latitude:         60.1695,
			Longitude:        24.9354,
			AboutMe:          "I love movies!",
			FavoriteGenre:    "Action",
			FavoriteMovie:    "Ben-Hur",
			FavoriteDirector: "Christopher Nolan",
			FavoriteActor:    "Leonardo DiCaprio",
			FavoriteActress:  "Meryl Streep",
		},
		{
			Username:         "Predefined2",
			FirstName:        "Jane",
			LastName:         "Smith",
			Email:            "predefined2@mail.com",
			Password:         "1234",
			Location:         "Kuopio",
			Latitude:         62.8924,
			Longitude:        27.6770,
			AboutMe:          "Cinema is my passion!",
			FavoriteGenre:    "Action",
			FavoriteMovie:    "Ben-Hur",
			FavoriteDirector: "Christopher Nolan",
			FavoriteActor:    "Leonardo DiCaprio",
			FavoriteActress:  "Meryl Streep",
		},
		{
			Username:         "Predefined3",
			FirstName:        "Mike",
			LastName:         "Johnson",
			Email:            "predefined3@mail.com",
			Password:         "1234",
			Location:         "Tampere",
			Latitude:         61.4978,
			Longitude:        23.7610,
			AboutMe:          "I enjoy classic films!",
			FavoriteGenre:    "Action",
			FavoriteMovie:    "Ben-Hur",
			FavoriteDirector: "Christopher Nolan",
			FavoriteActor:    "Leonardo DiCaprio",
			FavoriteActress:  "Meryl Streep",
		},
	}

	for _, user := range predefinedUsers {
		// Check if the user already exists in the database
		var count int
		err := database.DB.Raw("SELECT COUNT(*) FROM users WHERE username = ?", user.Username).Scan(&count).Error
		if err != nil {
			fmt.Printf("Error checking existence of user %s: %v\n", user.Username, err)
			continue
		}

		// If the user does not exist, create them
		if count == 0 {
			if err := user.HashPassword(); err != nil {
				fmt.Printf("Failed to hash password for user %s: %v\n", user.Username, err)
				continue
			}
			if err := database.DB.Create(&user).Error; err != nil {
				fmt.Printf("Failed to create user %s: %v\n", user.Username, err)
			} else {
				fmt.Printf("User %s created successfully\n", user.Username)
			}
		} else {
			fmt.Printf("User %s already exists, skipping...\n", user.Username)
		}
	}
}
