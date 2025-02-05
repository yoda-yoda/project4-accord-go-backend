package utils

import (
	"fmt"
	"sync"
)

var (
	idCounter int
	mu        sync.Mutex
)

func init() {
	idCounter = 1
}

func GenerateID() string {
	mu.Lock()
	defer mu.Unlock()

	id := idCounter
	idCounter++
	return fmt.Sprintf("%d", id)
}
