// backend/services/presence.go
package services

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	PresencePrefix = "presence:" // ключ = "presence:<userID>"
	TTL            = 120 * time.Second
)

type PresenceService struct {
	Rdb *redis.Client
	Ctx context.Context
}

func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{Rdb: rdb, Ctx: context.Background()}
}

// Touch обновляет TTL ключа при получении heartbeat
func (ps *PresenceService) Touch(userID string) error {
	key := PresencePrefix + userID
	return ps.Rdb.Set(ps.Ctx, key, "1", TTL).Err()
}

// IsOnline возвращает true, если ключ ещё не истёк
// func (ps *PresenceService) IsOnline(userID string) (bool, error) {
// 	key := PresencePrefix + userID
// 	exists, err := ps.Rdb.Exists(ps.Ctx, key).Result()
// 	return exists == 1, err
// }

func (ps *PresenceService) IsOnline(userID string) (bool, error) {
	key := PresencePrefix + userID
	exists, err := ps.Rdb.Exists(ps.Ctx, key).Result()
	if err != nil {
		log.Printf("Redis EXISTS error for key %s: %v", key, err)
	}
	return exists == 1, err
}

// func GetUserOnlineStatusHandler(ps *PresenceService) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		userID := r.URL.Query().Get("user_id")
// 		if userID == "" {
// 			http.Error(w, "missing user_id", http.StatusBadRequest)
// 			return
// 		}
// 		online, err := ps.IsOnline(userID)
// 		if err != nil {
// 			http.Error(w, "error checking presence", http.StatusInternalServerError)
// 			return
// 		}
// 		json.NewEncoder(w).Encode(map[string]bool{"online": online})
// 	}
// }

func GetUserOnlineStatusHandler(ps *PresenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}

		if ps == nil {
			http.Error(w, "PresenceService is nil", http.StatusInternalServerError)
			return
		}

		log.Printf("Checking online status for user: %s", userID)
		online, err := ps.IsOnline(userID)
		if err != nil {
			http.Error(w, "error checking presence", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"online": online})
	}
}
