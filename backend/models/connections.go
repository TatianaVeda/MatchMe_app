package models

import "gorm.io/gorm"

// ConnectionRequest represents the request sent by a user to connect with another user.
type ConnectionRequest struct {
	ID         uint   `json:"id"`         // Unique identifier for the connection request
	SenderID   uint   `json:"sender_id"`   // ID of the user sending the request
	ReceiverID uint   `json:"receiver_id"` // ID of the user receiving the request
	Status     string `json:"status"`      // Current status of the request (e.g., "pending", "accepted", "rejected")
	gorm.Model             // Embeds basic model fields (ID, CreatedAt, UpdatedAt, DeletedAt)
}
