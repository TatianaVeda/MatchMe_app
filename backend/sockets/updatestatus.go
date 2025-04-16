package sockets

import (
	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DB instance для обновления статуса онлайн. Установите его из main или при инициализации WebSocket.
var DB *gorm.DB

// SetDB позволяет установить соединение с базой для пакета sockets.
func SetDB(db *gorm.DB) {
	DB = db
}

// updateUserOnlineStatus вызывается из readPump для обновления статуса пользователя.
func updateUserOnlineStatus(userIDString string, isOnline bool) {
	uid, err := uuid.Parse(userIDString)
	if err != nil {
		logrus.Errorf("updateUserOnlineStatus: неверный userID: %v", err)
		return
	}
	// Проверяем, что DB установлена.
	if DB == nil {
		logrus.Error("updateUserOnlineStatus: DB не установлена")
		return
	}
	if err := models.UpdateUserOnlineStatus(DB, uid, isOnline); err != nil {
		logrus.Errorf("updateUserOnlineStatus: ошибка обновления статуса для пользователя %s: %v", userIDString, err)
	}
}
