package request

import (
	"time"
)

type RequestManager struct {
	capacity   int
	tokens     int
	lastAccess time.Time
}

func NewRequestManager(rate int) *RequestManager {
	return &RequestManager{
		capacity:   rate,
		tokens:     rate,
		lastAccess: time.Now(),
	}
}

func (manager *RequestManager) Next() bool {
	now := time.Now()
	elapsed := now.Sub(manager.lastAccess)

	if elapsed.Seconds() >= 1 {
		manager.tokens = manager.capacity
	}

	if manager.tokens > 0 {
		manager.tokens--
		return true
	}

	return false
}

func (manager *RequestManager) Try() {
	for !manager.Next() {
		time.Sleep(time.Millisecond)
	}
}

func (manager *RequestManager) UpdateAccess() {
	manager.lastAccess = time.Now()
}
