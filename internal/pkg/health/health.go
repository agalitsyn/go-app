package health

import (
	"net/http"
	"sync"
)

var (
	readinessStatus = http.StatusOK
	mu              sync.RWMutex
)

func ReadinessStatus() int {
	mu.RLock()
	defer mu.RUnlock()
	return readinessStatus
}

func SetReadinessStatus(status int) {
	mu.Lock()
	readinessStatus = status
	mu.Unlock()
}
