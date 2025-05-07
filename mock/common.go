package mock

import (
	"sync"
)

const (
	DummyUserID = 1 // Static user ID for testing
)

// Common storage structure
type Storage struct {
	mu     sync.RWMutex
	lastID int
}
