package controllers

import (
	"encoding/json"
	"m/backend/models"
	"m/backend/services"
	"m/backend/sockets"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var connectionsDB *gorm.DB

// InitConnectionsController инициализирует контроллер подключений.
func InitConnectionsController(db *gorm.DB) {
	connectionsDB = db
	logrus.Info("Connections controller initialized")
}

// GetConnections возвращает список взаимных подключений (только статус "accepted").
func GetConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetConnections: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetConnections: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var connections []models.Connection
	if err := connectionsDB.
		Where("(user_id = ? OR connection_id = ?) AND status = ?", currentUserID, currentUserID, "accepted").
		Find(&connections).Error; err != nil {
		logrus.Errorf("GetConnections: ошибка получения подключений для пользователя %s: %v", currentUserID, err)
		http.Error(w, "Error fetching connections", http.StatusInternalServerError)
		return
	}

	connectedIDs := make([]uuid.UUID, 0, len(connections))
	for _, conn := range connections {
		if conn.UserID == currentUserID {
			connectedIDs = append(connectedIDs, conn.ConnectionID)
		} else {
			connectedIDs = append(connectedIDs, conn.UserID)
		}
	}
	logrus.Infof("GetConnections: найдено %d подключений для пользователя %s", len(connectedIDs), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedIDs)
}

// PostConnection обрабатывает отправку запроса на подключение.
func PostConnection(w http.ResponseWriter, r *http.Request) {
	// Извлечение идентификатора текущего пользователя из контекста.
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("PostConnection: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("PostConnection: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	// Получаем идентификатор целевого пользователя из URL.
	vars := mux.Vars(r)
	targetIDStr := vars["id"]
	targetUserID, err := uuid.Parse(targetIDStr)
	if err != nil {
		logrus.Errorf("PostConnection: неверный target user ID: %v", err)
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}
	if currentUserID == targetUserID {
		logrus.Warn("PostConnection: попытка отправки запроса самому себе")
		http.Error(w, "Cannot send connection request to yourself", http.StatusBadRequest)
		return
	}

	// Проверка: существует ли уже ожидающий запрос от целевого пользователя к текущему.
	var existing models.Connection
	if err := connectionsDB.
		Where("user_id = ? AND connection_id = ? AND status = ?", targetUserID, currentUserID, "pending").
		First(&existing).Error; err == nil {
		// Обратный запрос найден – обновляем его до accepted (взаимное подключение).
		existing.Status = "accepted"
		if err := connectionsDB.Save(&existing).Error; err != nil {
			logrus.Errorf("PostConnection: ошибка обновления запроса от %s к %s: %v", targetUserID, currentUserID, err)
			http.Error(w, "Error updating connection request", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PostConnection: взаимное подключение между %s и %s", currentUserID, targetUserID)
		sockets.BroadcastNotification(targetUserID, "Your connection request has been mutually accepted!")
		chatService := services.NewChatService(connectionsDB)
		chat, err := chatService.CreateChat(targetUserID, currentUserID)
		if err != nil {
			logrus.Errorf("PostConnection: ошибка создания чата между %s и %s: %v", targetUserID, currentUserID, err)
		} else {
			logrus.Infof("PostConnection: чат с ID %d создан между %s и %s", chat.ID, targetUserID, currentUserID)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection mutually accepted"})
		return
	}

	// Проверка на дублирование запроса (в любом направлении).
	var duplicate models.Connection
	if err := connectionsDB.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?))",
			currentUserID, targetUserID, targetUserID, currentUserID).
		First(&duplicate).Error; err == nil {
		logrus.Warnf("PostConnection: дублирующий запрос между %s и %s", currentUserID, targetUserID)
		http.Error(w, "Connection request already exists or connection already established", http.StatusBadRequest)
		return
	}

	// Создаем новый запрос на подключение со статусом "pending".
	newConn := models.Connection{
		UserID:       currentUserID,
		ConnectionID: targetUserID,
		Status:       "pending",
	}
	if err := connectionsDB.Create(&newConn).Error; err != nil {
		logrus.Errorf("PostConnection: ошибка создания запроса от %s к %s: %v", currentUserID, targetUserID, err)
		http.Error(w, "Error creating connection request", http.StatusInternalServerError)
		return
	}
	logrus.Infof("PostConnection: запрос на подключение отправлен от %s к %s", currentUserID, targetUserID)
	sockets.BroadcastNotification(targetUserID, "You have a new connection request!")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection request sent"})
}

// PutConnection обрабатывает принятие или отклонение запроса на подключение.
func PutConnection(w http.ResponseWriter, r *http.Request) {
	// Извлекаем идентификатор текущего пользователя.
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("PutConnection: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("PutConnection: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	// Получаем идентификатор отправителя запроса из URL.
	vars := mux.Vars(r)
	senderIDStr := vars["id"]
	senderUserID, err := uuid.Parse(senderIDStr)
	if err != nil {
		logrus.Errorf("PutConnection: неверный sender user ID: %v", err)
		http.Error(w, "Invalid sender user ID", http.StatusBadRequest)
		return
	}

	var connection models.Connection
	if err := connectionsDB.
		Where("user_id = ? AND connection_id = ? AND status = ?", senderUserID, currentUserID, "pending").
		First(&connection).Error; err != nil {
		logrus.Warnf("PutConnection: запрос от %s к %s не найден", senderUserID, currentUserID)
		http.Error(w, "Connection request not found", http.StatusNotFound)
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logrus.Errorf("PutConnection: ошибка декодирования тела запроса: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switch body.Action {
	case "accept":
		connection.Status = "accepted"
		if err := connectionsDB.Save(&connection).Error; err != nil {
			logrus.Errorf("PutConnection: ошибка обновления запроса от %s к %s: %v", senderUserID, currentUserID, err)
			http.Error(w, "Error updating connection", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PutConnection: пользователь %s принял запрос от %s", currentUserID, senderUserID)
		sockets.BroadcastNotification(senderUserID, "Your connection request has been accepted!")
		chatService := services.NewChatService(connectionsDB)
		chat, err := chatService.CreateChat(senderUserID, currentUserID)
		if err != nil {
			logrus.Errorf("PutConnection: ошибка создания чата между %s и %s: %v", senderUserID, currentUserID, err)
		} else {
			logrus.Infof("PutConnection: чат с ID %d создан между %s и %s", chat.ID, senderUserID, currentUserID)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection accepted"})
	case "decline":
		if err := connectionsDB.Delete(&connection).Error; err != nil {
			logrus.Errorf("PutConnection: ошибка удаления запроса от %s к %s: %v", senderUserID, currentUserID, err)
			http.Error(w, "Error deleting connection request", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PutConnection: пользователь %s отклонил запрос от %s", currentUserID, senderUserID)
		sockets.BroadcastNotification(senderUserID, "Your connection request has been declined.")
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection declined"})
	default:
		logrus.Warn("PutConnection: получено неверное действие")
		http.Error(w, "Invalid action. Must be 'accept' or 'decline'", http.StatusBadRequest)
		return
	}
}

// DeleteConnection удаляет уже установленное взаимное подключение.
func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("DeleteConnection: userID не найден в контексте")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("DeleteConnection: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	targetIDStr := vars["id"]
	targetUserID, err := uuid.Parse(targetIDStr)
	if err != nil {
		logrus.Errorf("DeleteConnection: неверный target user ID: %v", err)
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}
	var connection models.Connection
	if err := connectionsDB.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, targetUserID, targetUserID, currentUserID, "accepted").
		First(&connection).Error; err != nil {
		logrus.Warnf("DeleteConnection: взаимное подключение между %s и %s не найдено", currentUserID, targetUserID)
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	if err := connectionsDB.Delete(&connection).Error; err != nil {
		logrus.Errorf("DeleteConnection: ошибка удаления подключения между %s и %s: %v", currentUserID, targetUserID, err)
		http.Error(w, "Error deleting connection", http.StatusInternalServerError)
		return
	}
	logrus.Infof("DeleteConnection: подключение между %s и %s успешно удалено", currentUserID, targetUserID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Disconnected successfully"})
}

// Новый метод для входящих запросов
func GetPendingConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetPendingConnections: userID не найден в контексте")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetPendingConnections: неверный userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var conns []models.Connection
	if err := connectionsDB.
		Where("connection_id = ? AND status = ?", currentUserID, "pending").
		Find(&conns).Error; err != nil {
		logrus.Errorf("GetPendingConnections: ошибка получения pending-запросов: %v", err)
		http.Error(w, "Error fetching pending connections", http.StatusInternalServerError)
		return
	}

	pendingIDs := make([]uuid.UUID, 0, len(conns))
	for _, c := range conns {
		pendingIDs = append(pendingIDs, c.UserID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingIDs)
}
