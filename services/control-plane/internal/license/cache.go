// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See LICENSE-COMMERCIAL for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

package license

import (
	"sync"
	"time"
)

type cacheEntry struct {
	features  []string
	expiresAt time.Time
}

type cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func newCache() *cache {
	return &cache{entries: make(map[string]cacheEntry)}
}

func (c *cache) get(key string) ([]string, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.features, true
}

func (c *cache) set(key string, features []string, expiresAt time.Time) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{features: features, expiresAt: expiresAt}
	c.mu.Unlock()
}
