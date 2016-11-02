package main

import (
	"net/http"
	"sync"
)

var (
	healthzStatus = http.StatusOK
	mu            sync.RWMutex
)

func HealthzStatus() int {
	mu.RLock()
	defer mu.RUnlock()
	return healthzStatus
}

func SetHealthzStatus(status int) {
	mu.Lock()
	healthzStatus = status
	mu.Unlock()
}
