package models

import (
	"time"

	"gorm.io/gorm"
)

// Recommendation represents a user recommendation with a score.
type Recommendation struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	UserID            uint           `json:"user_id" gorm:"not null;uniqueIndex:user_recommendation_idx"`             // Part of composite unique index
	RecommendedUserID uint           `json:"recommended_user_id" gorm:"not null;uniqueIndex:user_recommendation_idx"` // Part of composite unique index
	Score             int            `json:"score" gorm:"not null;default:0"`                                         // Default score to prevent null values

	// Foreign Key Relations
	User            User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`            // Cascade delete on User
	RecommendedUser User `gorm:"foreignKey:RecommendedUserID;constraint:OnDelete:CASCADE;"` // Cascade delete on RecommendedUser
}
