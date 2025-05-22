// controllers/presence.go
package controllers

import (
	"encoding/json"
	"m/backend/services"
	"net/http"
)

// NewPresenceController takes the service and returns all presence-related handlers.
type PresenceController struct {
	svc *services.PresenceService
}

func NewPresenceController(svc *services.PresenceService) *PresenceController {
	return &PresenceController{svc: svc}
}

func (pc *PresenceController) GetOnlineStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}
	online, err := pc.svc.IsOnline(userID)
	if err != nil {
		http.Error(w, "error checking presence", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"online": online})
}
