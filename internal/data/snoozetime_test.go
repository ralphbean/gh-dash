package data

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

// Jan 1, 2024 is a Monday, so:
//
//	Jan 8  = Monday
//	Jan 10 = Wednesday
//	Jan 12 = Friday
//	Jan 15 = Monday (next week)
//	Jan 19 = Friday (next week)
func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		t.Fatalf("failed to parse test date %q: %v", s, err)
	}
	return parsed
}

func TestComputeWakeTime(t *testing.T) {
	t.Run("duration presets", func(t *testing.T) {
		now := mustDate(t, "2024-01-10 12:00")
		cases := map[string]time.Duration{
			"10m": 10 * time.Minute,
			"1h":  time.Hour,
			"4h":  4 * time.Hour,
			"-2w": -2 * 7 * 24 * time.Hour,
		}
		for preset, want := range cases {
			got, err := ComputeWakeTime(preset, now)
			if err != nil {
				t.Fatalf("preset %q: unexpected error: %v", preset, err)
			}
			if !got.Equal(now.Add(want)) {
				t.Errorf("preset %q: got %v, want %v", preset, got, now.Add(want))
			}
		}
	})

	t.Run("tomorrow always advances exactly one day", func(t *testing.T) {
		// Wednesday
		now := mustDate(t, "2024-01-10 23:30")
		got, err := ComputeWakeTime("tomorrow", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-11 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("this friday from a wednesday goes to this week's friday", func(t *testing.T) {
		now := mustDate(t, "2024-01-10 09:00") // Wednesday
		got, err := ComputeWakeTime("this friday", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-12 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("this friday when today is friday before cutoff resolves to today", func(t *testing.T) {
		now := mustDate(t, "2024-01-12 07:00") // Friday, before 8am
		got, err := ComputeWakeTime("this friday", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-12 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("this friday when today is friday rolls to next week", func(t *testing.T) {
		now := mustDate(t, "2024-01-12 09:00") // Friday, already past 8am
		got, err := ComputeWakeTime("this friday", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-19 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("next week from a wednesday goes to the upcoming monday", func(t *testing.T) {
		now := mustDate(t, "2024-01-10 09:00") // Wednesday
		got, err := ComputeWakeTime("next week", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-15 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("next week when today is monday rolls to next monday", func(t *testing.T) {
		now := mustDate(t, "2024-01-08 07:00") // Monday, before 8am
		got, err := ComputeWakeTime("next week", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-15 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("keyword presets are case-insensitive", func(t *testing.T) {
		now := mustDate(t, "2024-01-10 09:00") // Wednesday
		got, err := ComputeWakeTime("This Friday", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := mustDate(t, "2024-01-12 08:00")
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unrecognized preset returns an error", func(t *testing.T) {
		now := mustDate(t, "2024-01-10 09:00")
		if _, err := ComputeWakeTime("bogus", now); err == nil {
			t.Error("expected an error for an unrecognized preset")
		}
	})
}

func withTestSnoozeStore(t *testing.T) *SnoozeStore {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "gh-dash-snooze-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	store := NewSnoozeStoreForTesting(filepath.Join(tempDir, "snoozed.json"))
	restore := OverrideSnoozeStoreForTesting(store)
	t.Cleanup(restore)
	return store
}

func TestApplySnoozePreset(t *testing.T) {
	presets := []config.SnoozePreset{
		{Label: "10m", After: "10m"},
	}

	t.Run("valid index snoozes the key and reports applied", func(t *testing.T) {
		withTestSnoozeStore(t)
		applied := ApplySnoozePreset("1", "some-key", presets)
		require.True(t, applied)
		require.True(t, GetSnoozeStore().IsSnoozed("some-key"))
	})

	t.Run("invalid index does not snooze and reports not applied", func(t *testing.T) {
		withTestSnoozeStore(t)
		applied := ApplySnoozePreset("99", "some-key", presets)
		require.False(t, applied)
		require.False(t, GetSnoozeStore().IsSnoozed("some-key"))
	})
}

func TestResolveSnoozePreset(t *testing.T) {
	presets := []config.SnoozePreset{
		{Label: "10m", After: "10m"},
		{Label: "1h", After: "1h"},
	}
	now := mustDate(t, "2024-01-10 09:00")

	t.Run("valid 1-based index resolves the wake time", func(t *testing.T) {
		wakeAt, ok := ResolveSnoozePreset("1", presets, now)
		if !ok {
			t.Fatal("expected ok=true for a valid index")
		}
		if want := now.Add(10 * time.Minute); !wakeAt.Equal(want) {
			t.Errorf("got %v, want %v", wakeAt, want)
		}
	})

	t.Run("non-numeric input is rejected", func(t *testing.T) {
		if _, ok := ResolveSnoozePreset("abc", presets, now); ok {
			t.Error("expected ok=false for non-numeric input")
		}
	})

	t.Run("out-of-range index is rejected", func(t *testing.T) {
		if _, ok := ResolveSnoozePreset("99", presets, now); ok {
			t.Error("expected ok=false for an out-of-range index")
		}
		if _, ok := ResolveSnoozePreset("0", presets, now); ok {
			t.Error("expected ok=false for index 0")
		}
	})
}
