package models

import "gorm.io/gorm"

// DismissedUser represents a user who has been dismissed by another user.
type DismissedUser struct {
	gorm.Model
	UserID          uint `json:"user_id" gorm:"not null;index"`           // Indexed for faster lookups
	DismissedUserID uint `json:"dismissed_user_id" gorm:"not null;index"` // Indexed for efficient queries
}
