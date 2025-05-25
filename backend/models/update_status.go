package models

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func UpdateUserOnlineStatus(db *gorm.DB, userID uuid.UUID, isOnline bool) error {

	result := db.Model(&Profile{}).Where("user_id = ?", userID).Update("online", isOnline)
	if result.Error != nil {
		logrus.Errorf("UpdateUserOnlineStatus: error updating status for user %s: %v", userID, result.Error)
		return result.Error
	}
	logrus.Infof("UpdateUserOnlineStatus: status of user %s updated to %v", userID, isOnline)
	return nil
}
