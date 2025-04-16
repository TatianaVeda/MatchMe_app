package services

import (
	"sync"
	"time"
)

// TokenBlacklist хранит отозванные токены вместе с временем их истечения.
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var blacklist = &TokenBlacklist{
	tokens: make(map[string]time.Time),
}

// AddToken добавляет токен в чёрный список с указанным временем истечения.
func AddToken(token string, exp time.Time) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = exp
}

// IsBlacklisted проверяет, находится ли токен в чёрном списке.
// Если срок его действия уже истёк, его можно удалить из списка.
func IsBlacklisted(token string) bool {
	blacklist.mu.RLock()
	exp, exists := blacklist.tokens[token]
	blacklist.mu.RUnlock()
	if !exists {
		return false
	}
	// Если токен уже истёк, удаляем его из списка и возвращаем false.
	if time.Now().After(exp) {
		blacklist.mu.Lock()
		delete(blacklist.tokens, token)
		blacklist.mu.Unlock()
		return false
	}
	return true
}
