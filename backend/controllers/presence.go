package controllers

import (
	"encoding/json"
	"m/backend/services"
	"net/http"
	"strings"
)

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

func (pc *PresenceController) GetMultipleOnlineStatus(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query().Get("ids")
	if qs == "" {
		http.Error(w, "missing ids", http.StatusBadRequest)
		return
	}
	ids := strings.Split(qs, ",")
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		online, _ := pc.svc.IsOnline(id)
		result[id] = online
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
