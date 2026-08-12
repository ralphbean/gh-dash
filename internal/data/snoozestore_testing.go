package data

import (
	"time"
)

// NewSnoozeStoreForTesting creates a SnoozeStore backed by the given file path.
func NewSnoozeStoreForTesting(filePath string) *SnoozeStore {
	return &SnoozeStore{
		entries:  make(map[string]time.Time),
		filePath: filePath,
	}
}

// OverrideSnoozeStoreForTesting replaces the singleton SnoozeStore with the
// given store. It returns a function that restores the original store.
func OverrideSnoozeStoreForTesting(store *SnoozeStore) func() {
	// Ensure the singleton is initialized so sync.Once has fired.
	GetSnoozeStore()
	old := snoozeStore
	snoozeStore = store
	return func() { snoozeStore = old }
}
