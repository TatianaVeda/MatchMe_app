package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// PresenceService uses Redis as a fast in-memory store to track user online status.
// Redis is a high-performance key-value database, often used for caching and real-time data.
// In this service, each user gets a presence key with a short TTL (time-to-live).
// If the key exists, the user is considered online; if not, offline.
// This approach is efficient for real-time presence tracking in chat and social apps.

const (
	PresencePrefix = "presence:"
	// PresenceTTL defines how long a user is considered online after their last activity.
	// If the key expires, the user is marked offline automatically.
	PresenceTTL = 60 * time.Second
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
// Called on user activity (e.g., ping, message) to keep them online.
func (ps *PresenceService) Touch(userID string) error {
	// Set a key in Redis with a short TTL to mark the user as online
	key := PresencePrefix + userID
	return ps.Rdb.Set(ps.Ctx, key, "1", PresenceTTL).Err()
}

// SetOffline marks a user as offline in Redis.
// Used when user logs out or disconnects explicitly.
func (ps *PresenceService) SetOffline(userID string) error {
	// Remove the user's presence key from Redis to mark as offline
	key := PresencePrefix + userID
	return ps.Rdb.Del(ps.Ctx, key).Err()
}

// IsOnline checks if a user is online (by presence of a key in Redis).
// Returns true if the presence key exists (user is online), false otherwise.
func (ps *PresenceService) IsOnline(userID string) (bool, error) {
	// Check if the user's presence key exists in Redis
	key := PresencePrefix + userID
	cnt, err := ps.Rdb.Exists(ps.Ctx, key).Result()
	return cnt == 1, err
}
