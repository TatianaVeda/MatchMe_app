package services

import (
	"sync"
	"time"
)

// TokenBlacklist is an in-memory structure for storing blacklisted JWT tokens and their expiration times.
// Uses a mutex for safe concurrent access from multiple goroutines.
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var blacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

// AddToken adds a token to the blacklist with its expiration time.
func AddToken(token string, exp time.Time) {
	// Lock for writing to prevent race conditions
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = exp
}

// IsBlacklisted checks if a token is in the blacklist and not expired.
func IsBlacklisted(token string) bool {
	// Lock for reading to allow concurrent safe access
	blacklist.mu.RLock()
	exp, exists := blacklist.tokens[token]
	blacklist.mu.RUnlock()
	if !exists {
		return false
	}
	// If token is expired, remove it from blacklist
	if time.Now().After(exp) {
		blacklist.mu.Lock()
		delete(blacklist.tokens, token)
		blacklist.mu.Unlock()
		return false
	}
	return true
}
