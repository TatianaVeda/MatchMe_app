package services

import (
	"sync"
	"time"
)

type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var blacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

func AddToken(token string, exp time.Time) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = exp
}

func IsBlacklisted(token string) bool {
	blacklist.mu.RLock()
	exp, exists := blacklist.tokens[token]
	blacklist.mu.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(exp) {
		blacklist.mu.Lock()
		delete(blacklist.tokens, token)
		blacklist.mu.Unlock()
		return false
	}
	return true
}
