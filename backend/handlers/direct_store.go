package handlers

import (
	"sync"
	"time"
)

var (
	directSessions      = make(map[string]*DirectSession)
	directSessionsMutex sync.RWMutex
)

func CleanupDirectSessions() {
	for {
		directSessionsMutex.Lock()
		now := time.Now()
		for token, session := range directSessions {
			if now.Sub(session.CreatedAt) > 10*time.Minute {
				delete(directSessions, token)
			}
		}
		directSessionsMutex.Unlock()
		time.Sleep(1 * time.Minute)
	}
}
