package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"charm.land/log/v2"
)

// SnoozeStore persists item IDs (PRs, Issues, or Notifications) along with
// the "wake-at" timestamp at which they should reappear. Unlike DoneStore,
// new activity on a snoozed item does not un-snooze it - IsSnoozed only
// checks the stored wake-at time against the current time.
type SnoozeStore struct {
	mu       sync.RWMutex
	entries  map[string]time.Time // id -> wake-at time
	filePath string
}

func newSnoozeStore(filename string) *SnoozeStore {
	store := &SnoozeStore{
		entries: make(map[string]time.Time),
	}
	filePath, err := getStateFilePath(filename)
	if err != nil {
		log.Error("Failed to get state file path for snoozed items", "err", err)
	}
	store.filePath = filePath
	if err := store.load(); err != nil {
		log.Error("Failed to load snoozed items", "err", err)
	}
	return store
}

// load reads the snooze store from disk: {"id": "2024-01-15T10:30:00Z", ...}
// (map of ID → RFC 3339 wake-at timestamp).
func (s *SnoozeStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var tsMap map[string]string
	if err := json.Unmarshal(data, &tsMap); err != nil {
		return err
	}

	for id, raw := range tsMap {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			log.Warn(
				"Skipping snooze entry with invalid timestamp",
				"id",
				id,
				"raw",
				raw,
				"err",
				err,
			)
			continue
		}
		s.entries[id] = t
	}
	s.prune()
	log.Debug("Loaded snoozed items", "count", len(s.entries))
	return nil
}

// prune removes entries whose wake-at time has already passed: an expired
// snooze entry serves no purpose.
func (s *SnoozeStore) prune() {
	now := time.Now()
	for id, wakeAt := range s.entries {
		if !now.Before(wakeAt) {
			delete(s.entries, id)
		}
	}
}

func (s *SnoozeStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	tsMap := make(map[string]string, len(s.entries))
	for id, t := range s.entries {
		tsMap[id] = t.Format(time.RFC3339)
	}

	data, err := json.Marshal(tsMap)
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	log.Debug("Saved snoozed items", "count", len(tsMap))
	return nil
}

// Snooze records the wake-at time for the given ID.
func (s *SnoozeStore) Snooze(id string, until time.Time) {
	s.mu.Lock()
	s.entries[id] = until
	s.mu.Unlock()
	go s.save()
}

// IsSnoozed returns true if the item is still snoozed, i.e. its wake-at
// time is in the future.
func (s *SnoozeStore) IsSnoozed(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wakeAt, ok := s.entries[id]
	if !ok {
		return false
	}
	return time.Now().Before(wakeAt)
}

// Remove removes an item from the snooze store.
func (s *SnoozeStore) Remove(id string) {
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
	go s.save()
}

// Flush forces an immediate synchronous save.
func (s *SnoozeStore) Flush() error {
	return s.save()
}

// Singleton

var (
	snoozeStore     *SnoozeStore
	snoozeStoreOnce sync.Once
)

// GetSnoozeStore returns the singleton snooze store.
func GetSnoozeStore() *SnoozeStore {
	snoozeStoreOnce.Do(func() {
		snoozeStore = newSnoozeStore("snoozed.json")
	})
	return snoozeStore
}
