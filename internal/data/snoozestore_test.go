package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSnoozeStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gh-dash-snoozestore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Snooze and IsSnoozed while wake time is in the future", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test1.json"),
		}

		store.Snooze("id1", time.Now().Add(1*time.Hour))
		if !store.IsSnoozed("id1") {
			t.Error("Should be snoozed when wake time is in the future")
		}
	})

	t.Run("IsSnoozed after wake time has passed returns false", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test2.json"),
		}

		store.Snooze("id1", time.Now().Add(-1*time.Hour))
		if store.IsSnoozed("id1") {
			t.Error("Should NOT be snoozed once wake time has passed")
		}
	})

	t.Run("IsSnoozed for unknown ID returns false", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test3.json"),
		}

		if store.IsSnoozed("unknown") {
			t.Error("Should NOT be snoozed for an ID not in the store")
		}
	})

	t.Run("new activity does not un-snooze", func(t *testing.T) {
		// Unlike DoneStore, SnoozeStore has no concept of item updatedAt at all -
		// IsSnoozed only depends on wall-clock time versus the stored wake time.
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test4.json"),
		}
		store.Snooze("id1", time.Now().Add(1*time.Hour))
		if !store.IsSnoozed("id1") {
			t.Error("Should remain snoozed regardless of any external activity")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test5.json"),
		}

		store.Snooze("id1", time.Now().Add(1*time.Hour))
		store.Remove("id1")
		if store.IsSnoozed("id1") {
			t.Error("Should NOT be snoozed after Remove")
		}
	})

	t.Run("Snooze overwrites wake time on same ID", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test6.json"),
		}

		store.Snooze("id1", time.Now().Add(-1*time.Hour))
		if store.IsSnoozed("id1") {
			t.Error("Should not be snoozed with a past wake time")
		}
		store.Snooze("id1", time.Now().Add(1*time.Hour))
		if !store.IsSnoozed("id1") {
			t.Error("Should be snoozed after re-snoozing with a future wake time")
		}
	})

	t.Run("persistence round-trip", func(t *testing.T) {
		persistFile := filepath.Join(tempDir, "persist.json")
		wakeAt := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)

		store1 := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: persistFile,
		}
		store1.Snooze("id1", wakeAt)
		if err := store1.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		store2 := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: persistFile,
		}
		if err := store2.load(); err != nil {
			t.Fatalf("load failed: %v", err)
		}

		if !store2.IsSnoozed("id1") {
			t.Error("Loaded store should have id1 as snoozed")
		}
	})

	t.Run("load from non-existent file", func(t *testing.T) {
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "nonexistent.json"),
		}
		if err := store.load(); err != nil {
			t.Errorf("load from non-existent file should not error, got: %v", err)
		}
	})

	t.Run("load from corrupted JSON", func(t *testing.T) {
		corruptedFile := filepath.Join(tempDir, "corrupted.json")
		if err := os.WriteFile(corruptedFile, []byte("{invalid json"), 0o644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: corruptedFile,
		}
		if err := store.load(); err == nil {
			t.Error("load from corrupted JSON should return an error")
		}
	})

	t.Run("load prunes entries whose wake time has already passed", func(t *testing.T) {
		pruneFile := filepath.Join(tempDir, "prune.json")
		now := time.Now().UTC().Truncate(time.Second)

		tsMap := map[string]string{
			"future":  now.Add(1 * time.Hour).Format(time.RFC3339),
			"past":    now.Add(-1 * time.Hour).Format(time.RFC3339),
			"ancient": now.Add(-90 * 24 * time.Hour).Format(time.RFC3339),
		}
		raw, err := json.Marshal(tsMap)
		if err != nil {
			t.Fatalf("Failed to marshal test data: %v", err)
		}
		if err := os.WriteFile(pruneFile, raw, 0o644); err != nil {
			t.Fatalf("Failed to write prune file: %v", err)
		}

		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: pruneFile,
		}
		if err := store.load(); err != nil {
			t.Fatalf("load failed: %v", err)
		}

		if !store.IsSnoozed("future") {
			t.Error("Future wake time entry should survive pruning")
		}
		if _, ok := store.entries["past"]; ok {
			t.Error("Past wake time entry should be pruned on load")
		}
		if _, ok := store.entries["ancient"]; ok {
			t.Error("Ancient wake time entry should be pruned on load")
		}
	})

	t.Run("save then load preserves RFC3339 format", func(t *testing.T) {
		file := filepath.Join(tempDir, "format.json")
		store := &SnoozeStore{
			entries:  make(map[string]time.Time),
			filePath: file,
		}
		store.Snooze("id1", time.Now().Add(1*time.Hour))
		if err := store.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		var tsMap map[string]string
		if err := json.Unmarshal(raw, &tsMap); err != nil {
			t.Fatalf("Should be a JSON object: %v", err)
		}
		if _, ok := tsMap["id1"]; !ok {
			t.Error("Should have id1 key")
		}
		if _, err := time.Parse(time.RFC3339, tsMap["id1"]); err != nil {
			t.Errorf("Timestamp should be RFC3339: %v", err)
		}
	})
}
