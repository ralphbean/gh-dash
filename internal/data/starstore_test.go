package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStarStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gh-dash-starstore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Star and IsStarred", func(t *testing.T) {
		store := NewStarStoreForTesting(filepath.Join(tempDir, "test1.json"))

		store.Star("id1")
		if !store.IsStarred("id1") {
			t.Error("Should be starred after Star")
		}
	})

	t.Run("Unstar and IsStarred", func(t *testing.T) {
		store := NewStarStoreForTesting(filepath.Join(tempDir, "test2.json"))

		store.Star("id1")
		store.Unstar("id1")
		if store.IsStarred("id1") {
			t.Error("Should NOT be starred after Unstar")
		}
	})

	t.Run("Unknown id is not starred", func(t *testing.T) {
		store := NewStarStoreForTesting(filepath.Join(tempDir, "test3.json"))

		if store.IsStarred("unknown") {
			t.Error("Unknown id should not be starred")
		}
	})

	t.Run("Toggle stars an unstarred id and returns true", func(t *testing.T) {
		store := NewStarStoreForTesting(filepath.Join(tempDir, "test4.json"))

		result := store.Toggle("id1")
		if !result {
			t.Error("Toggle should return true when starring")
		}
		if !store.IsStarred("id1") {
			t.Error("Should be starred after Toggle")
		}
	})

	t.Run("Toggle unstars a starred id and returns false", func(t *testing.T) {
		store := NewStarStoreForTesting(filepath.Join(tempDir, "test5.json"))

		store.Star("id1")
		result := store.Toggle("id1")
		if result {
			t.Error("Toggle should return false when unstarring")
		}
		if store.IsStarred("id1") {
			t.Error("Should NOT be starred after Toggle")
		}
	})

	t.Run("Star persists across reload from the same file path", func(t *testing.T) {
		path := filepath.Join(tempDir, "test6.json")
		store := NewStarStoreForTesting(path)
		store.Star("id1")
		if err := store.Flush(); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

		reloaded := NewStarStoreForTesting(path)
		if err := reloaded.load(); err != nil {
			t.Fatalf("Failed to load: %v", err)
		}
		if !reloaded.IsStarred("id1") {
			t.Error("Starred id should persist across reload")
		}
	})
}
