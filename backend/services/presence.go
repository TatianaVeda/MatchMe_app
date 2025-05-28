package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	PresencePrefix = "presence:"
	PresenceTTL    = 60 * time.Second
)

type PresenceService struct {
	Rdb *redis.Client
	Ctx context.Context
}

// NewPresenceService creates a service for tracking user online status using Redis.
func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{
		Rdb: rdb,
		Ctx: context.Background(),
	}
}

// Touch updates the TTL of a user's online status in Redis.
func (ps *PresenceService) Touch(userID string) error {
	key := PresencePrefix + userID
	return ps.Rdb.Set(ps.Ctx, key, "1", PresenceTTL).Err()
}

// SetOffline marks a user as offline in Redis.
func (ps *PresenceService) SetOffline(userID string) error {
	key := PresencePrefix + userID
	return ps.Rdb.Del(ps.Ctx, key).Err()
}

// IsOnline checks if a user is online (by presence of a key in Redis).
func (ps *PresenceService) IsOnline(userID string) (bool, error) {
	key := PresencePrefix + userID
	cnt, err := ps.Rdb.Exists(ps.Ctx, key).Result()
	return cnt == 1, err
}
