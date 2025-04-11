package models

import (
	"time"

	"gorm.io/gorm"
)

// Chat represents a message exchanged between users.
type Chat struct {
	gorm.Model
	SenderID   uint      `json:"sender_id" gorm:"not null;index"` // Indexed for faster queries
	ReceiverID uint      `json:"receiver_id" gorm:"not null;index"` // Indexed for efficient lookups
	Message    string    `json:"message" gorm:"type:text;not null"` // Storing long messages safely
	Timestamp  time.Time `json:"timestamp" gorm:"not null;default:CURRENT_TIMESTAMP"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`
}
