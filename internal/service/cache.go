package service

import (
	"sync"
	"time"
	"observatory/internal/domain"
)

type StatusCache struct {
	mu        sync.RWMutex
	providers []*domain.Provider
	expires   time.Time
	duration  time.Duration
}

func NewStatusCache(duration time.Duration) *StatusCache {
	return &StatusCache{
		duration: duration,
	}
}

func (c *StatusCache) Get() ([]*domain.Provider, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if time.Now().After(c.expires) {
		return nil, false
	}
	return c.providers, true
}

func (c *StatusCache) Set(providers []*domain.Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.providers = providers
	c.expires = time.Now().Add(c.duration)
}
