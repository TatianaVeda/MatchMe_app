package models

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UpdateUserOnlineStatus обновляет статус онлайн для пользователя в таблице Profile.
func UpdateUserOnlineStatus(db *gorm.DB, userID uuid.UUID, isOnline bool) error {
	// Обновляем поле Online в модели Profile
	result := db.Model(&Profile{}).Where("user_id = ?", userID).Update("online", isOnline)
	if result.Error != nil {
		logrus.Errorf("UpdateUserOnlineStatus: ошибка обновления статуса для пользователя %s: %v", userID, result.Error)
		return result.Error
	}
	logrus.Infof("UpdateUserOnlineStatus: статус пользователя %s обновлён на %v", userID, isOnline)
	return nil
}
