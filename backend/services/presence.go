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

func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{
		Rdb: rdb,
		Ctx: context.Background(),
	}
}

func (ps *PresenceService) Touch(userID string) error {
	key := PresencePrefix + userID
	return ps.Rdb.Set(ps.Ctx, key, "1", PresenceTTL).Err()
}

func (ps *PresenceService) SetOffline(userID string) error {
	key := PresencePrefix + userID
	return ps.Rdb.Del(ps.Ctx, key).Err()
}

func (ps *PresenceService) IsOnline(userID string) (bool, error) {
	key := PresencePrefix + userID
	cnt, err := ps.Rdb.Exists(ps.Ctx, key).Result()
	return cnt == 1, err
}
