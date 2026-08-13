package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"charm.land/log/v2"
)

// StarStore persists item IDs (PRs) that the user has starred. Unlike
// SnoozeStore or DoneStore, entries never expire and are never invalidated
// by new activity - a star is a plain flag the user sets and clears
// explicitly, like starring an email.
type StarStore struct {
	mu       sync.RWMutex
	entries  map[string]struct{}
	filePath string
}

func newStarStore(filename string) *StarStore {
	store := &StarStore{
		entries: make(map[string]struct{}),
	}
	filePath, err := getStateFilePath(filename)
	if err != nil {
		log.Error("Failed to get state file path for starred items", "err", err)
	}
	store.filePath = filePath
	if err := store.load(); err != nil {
		log.Error("Failed to load starred items", "err", err)
	}
	return store
}

// load reads the star store from disk: ["id1", "id2", ...] (array of
// starred IDs).
func (s *StarStore) load() error {
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

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return err
	}

	for _, id := range ids {
		s.entries[id] = struct{}{}
	}
	log.Debug("Loaded starred items", "count", len(s.entries))
	return nil
}

func (s *StarStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	ids := make([]string, 0, len(s.entries))
	for id := range s.entries {
		ids = append(ids, id)
	}

	data, err := json.Marshal(ids)
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

	log.Debug("Saved starred items", "count", len(ids))
	return nil
}

// Star marks the given ID as starred.
func (s *StarStore) Star(id string) {
	s.mu.Lock()
	s.entries[id] = struct{}{}
	s.mu.Unlock()
	go s.save()
}

// Unstar removes the given ID's starred flag.
func (s *StarStore) Unstar(id string) {
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
	go s.save()
}

// IsStarred returns true if the item is currently starred.
func (s *StarStore) IsStarred(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.entries[id]
	return ok
}

// Toggle flips the starred state of the given ID and returns the resulting
// state (true if it is now starred, false if it is now unstarred).
func (s *StarStore) Toggle(id string) bool {
	s.mu.Lock()
	_, wasStarred := s.entries[id]
	if wasStarred {
		delete(s.entries, id)
	} else {
		s.entries[id] = struct{}{}
	}
	s.mu.Unlock()
	go s.save()
	return !wasStarred
}

// Flush forces an immediate synchronous save.
func (s *StarStore) Flush() error {
	return s.save()
}

// Singleton

var (
	starStore     *StarStore
	starStoreOnce sync.Once
)

// GetStarStore returns the singleton star store.
func GetStarStore() *StarStore {
	starStoreOnce.Do(func() {
		starStore = newStarStore("starred.json")
	})
	return starStore
}
