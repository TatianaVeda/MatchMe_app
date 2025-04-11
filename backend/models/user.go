package models

import (
	"time"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the application
type User struct {
	ID               uint           `json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Username         string         `json:"username" gorm:"unique;not null"`
	FirstName        string         `json:"firstname"`
	LastName         string         `json:"lastname"`
	Email            string         `json:"email" gorm:"unique;not null"`
	Password         string         `json:"-"`
	Avatar           string         `json:"avatar" gorm:"default:'/uploads/avatars/default.png'"`
	Location         string         `json:"location,omitempty"`
	Latitude         float64        `json:"latitude,omitempty"`
	Longitude        float64        `json:"longitude,omitempty"`
	SearchRadius     float64        `json:"search_radius" gorm:"default:50"`
	AboutMe          string         `json:"aboutme,omitempty"`
	FavoriteGenre    string         `json:"favorite_genre,omitempty"`
	FavoriteMovie    string         `json:"favorite_movie,omitempty"`
	FavoriteDirector string         `json:"favorite_director,omitempty"`
	FavoriteActor    string         `json:"favorite_actor,omitempty"`
	FavoriteActress  string         `json:"favorite_actress,omitempty"`
}

// HashPassword hashes the user's password before saving to the database
func (user *User) HashPassword() error {
	if user.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password to the hashed version
	user.Password = string(hashedPassword)
	return nil
}

// ValidatePassword checks if the provided password matches the stored password
func (user *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}