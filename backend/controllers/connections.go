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

func InitConnectionsController(db *gorm.DB) {
	connectionsDB = db
	logrus.Info("Connections controller initialized")
}

func GetConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetConnections: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetConnections: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var connections []models.Connection
	if err := connectionsDB.
		Where("(user_id = ? OR connection_id = ?) AND status = ?", currentUserID, currentUserID, "accepted").
		Find(&connections).Error; err != nil {
		logrus.Errorf("GetConnections: error fetching connections for user %s: %v", currentUserID, err)
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
	logrus.Infof("GetConnections: found %d connections for user %s", len(connectedIDs), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedIDs)
}

func PostConnection(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("PostConnection: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("PostConnection: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	targetIDStr := vars["id"]
	targetUserID, err := uuid.Parse(targetIDStr)
	if err != nil {
		logrus.Errorf("PostConnection: invalid target user ID: %v", err)
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}
	if currentUserID == targetUserID {
		logrus.Warn("PostConnection: attempt to send request to self")
		http.Error(w, "Cannot send connection request to yourself", http.StatusBadRequest)
		return
	}

	var existing models.Connection
	if err := connectionsDB.
		Where("user_id = ? AND connection_id = ? AND status = ?", targetUserID, currentUserID, "pending").
		First(&existing).Error; err == nil {
		existing.Status = "accepted"
		if err := connectionsDB.Save(&existing).Error; err != nil {
			logrus.Errorf("PostConnection: error updating request from %s to %s: %v", targetUserID, currentUserID, err)
			http.Error(w, "Error updating connection request", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PostConnection: mutual connection between %s and %s", currentUserID, targetUserID)

		chatService := services.NewChatService(connectionsDB)
		chat, err := chatService.CreateChat(targetUserID, currentUserID)
		if err != nil {
			logrus.Errorf("PostConnection: error creating chat between %s and %s: %v", targetUserID, currentUserID, err)
		} else {
			logrus.Infof("PostConnection: chat with ID %d created between %s and %s", chat.ID, targetUserID, currentUserID)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection mutually accepted"})
		return
	}

	var duplicate models.Connection
	if err := connectionsDB.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?))",
			currentUserID, targetUserID, targetUserID, currentUserID).
		First(&duplicate).Error; err == nil {
		logrus.Warnf("PostConnection: duplicate request between %s and %s", currentUserID, targetUserID)
		http.Error(w, "Connection request already exists or connection already established", http.StatusBadRequest)
		return
	}

	newConn := models.Connection{
		UserID:       currentUserID,
		ConnectionID: targetUserID,
		Status:       "pending",
	}
	if err := connectionsDB.Create(&newConn).Error; err != nil {
		logrus.Errorf("PostConnection: error creating request from %s to %s: %v", currentUserID, targetUserID, err)
		http.Error(w, "Error creating connection request", http.StatusInternalServerError)
		return
	}
	logrus.Infof("PostConnection: connection request sent from %s to %s", currentUserID, targetUserID)

	go sockets.BroadcastNotification(targetUserID, `{"type":"connection_request"}`)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection request sent"})
}

func PutConnection(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("PutConnection: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("PutConnection: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	senderIDStr := vars["id"]
	senderUserID, err := uuid.Parse(senderIDStr)
	if err != nil {
		logrus.Errorf("PutConnection: invalid sender user ID: %v", err)
		http.Error(w, "Invalid sender user ID", http.StatusBadRequest)
		return
	}

	var connection models.Connection
	if err := connectionsDB.
		Where("user_id = ? AND connection_id = ? AND status = ?", senderUserID, currentUserID, "pending").
		First(&connection).Error; err != nil {
		logrus.Warnf("PutConnection: request from %s to %s not found", senderUserID, currentUserID)
		http.Error(w, "Connection request not found", http.StatusNotFound)
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logrus.Errorf("PutConnection: error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switch body.Action {
	case "accept":
		connection.Status = "accepted"
		if err := connectionsDB.Save(&connection).Error; err != nil {
			logrus.Errorf("PutConnection: error updating request from %s to %s: %v", senderUserID, currentUserID, err)
			http.Error(w, "Error updating connection", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PutConnection: user %s accepted request from %s", currentUserID, senderUserID)

		go sockets.BroadcastNotification(senderUserID, `{"type":"connection_request_accepted"}`)
		chatService := services.NewChatService(connectionsDB)
		chat, err := chatService.CreateChat(senderUserID, currentUserID)
		if err != nil {
			logrus.Errorf("PutConnection: error creating chat between %s and %s: %v", senderUserID, currentUserID, err)
		} else {
			logrus.Infof("PutConnection: chat with ID %d created between %s and %s", chat.ID, senderUserID, currentUserID)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection accepted"})
	case "decline":
		if err := connectionsDB.Delete(&connection).Error; err != nil {
			logrus.Errorf("PutConnection: error deleting request from %s to %s: %v", senderUserID, currentUserID, err)
			http.Error(w, "Error deleting connection request", http.StatusInternalServerError)
			return
		}
		logrus.Infof("PutConnection: user %s declined request from %s", currentUserID, senderUserID)

		go sockets.BroadcastNotification(senderUserID, `{"type":"connection_request_declined"}`)
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection declined"})
	default:
		logrus.Warn("PutConnection: invalid action received")
		http.Error(w, "Invalid action. Must be 'accept' or 'decline'", http.StatusBadRequest)
		return
	}
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("DeleteConnection: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("DeleteConnection: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	targetIDStr := vars["id"]
	targetUserID, err := uuid.Parse(targetIDStr)
	if err != nil {
		logrus.Errorf("DeleteConnection: invalid target user ID: %v", err)
		http.Error(w, "Invalid target user ID", http.StatusBadRequest)
		return
	}
	var connection models.Connection
	if err := connectionsDB.
		Where("((user_id = ? AND connection_id = ?) OR (user_id = ? AND connection_id = ?)) AND status = ?",
			currentUserID, targetUserID, targetUserID, currentUserID, "accepted").
		First(&connection).Error; err != nil {
		logrus.Warnf("DeleteConnection: mutual connection between %s and %s not found", currentUserID, targetUserID)
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	if err := connectionsDB.Delete(&connection).Error; err != nil {
		logrus.Errorf("DeleteConnection: error deleting connection between %s and %s: %v", currentUserID, targetUserID, err)
		http.Error(w, "Error deleting connection", http.StatusInternalServerError)
		return
	}
	logrus.Infof("DeleteConnection: connection between %s and %s successfully deleted", currentUserID, targetUserID)

	if err := connectionsDB.
		Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
			currentUserID, targetUserID, targetUserID, currentUserID).
		Delete(&models.Chat{}).Error; err != nil {
		logrus.Warnf("DeleteConnection: failed to delete chat between %s and %s: %v", currentUserID, targetUserID, err)
	} else {
		logrus.Infof("DeleteConnection: chat between %s and %s successfully deleted", currentUserID, targetUserID)
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Disconnected successfully"})
}

func GetPendingConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetPendingConnections: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetPendingConnections: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var conns []models.Connection
	if err := connectionsDB.
		Where("connection_id = ? AND status = ?", currentUserID, "pending").
		Find(&conns).Error; err != nil {
		logrus.Errorf("GetPendingConnections: error fetching incoming requests: %v", err)
		http.Error(w, "Error fetching pending connections", http.StatusInternalServerError)
		return
	}

	incomingIDs := make([]uuid.UUID, 0, len(conns))
	for _, c := range conns {
		incomingIDs = append(incomingIDs, c.UserID)
	}

	logrus.Infof("GetPendingConnections: found %d incoming requests for %s", len(incomingIDs), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incomingIDs)
}

func GetSentConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		logrus.Error("GetSentConnections: userID not found in context")
		http.Error(w, "Unauthorized: userID not found in context", http.StatusUnauthorized)
		return
	}
	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.Errorf("GetSentConnections: invalid userID: %v", err)
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var conns []models.Connection
	if err := connectionsDB.
		Where("user_id = ? AND status = ?", currentUserID, "pending").
		Find(&conns).Error; err != nil {
		logrus.Errorf("GetSentConnections: error fetching outgoing requests for %s: %v", currentUserID, err)
		http.Error(w, "Error fetching sent connections", http.StatusInternalServerError)
		return
	}

	sentIDs := make([]uuid.UUID, 0, len(conns))
	for _, c := range conns {
		sentIDs = append(sentIDs, c.ConnectionID)
	}

	logrus.Infof("GetSentConnections: found %d outgoing requests for %s", len(sentIDs), currentUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sentIDs)
}
