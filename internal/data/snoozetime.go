package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

// snoozeMorningHour is the local hour used for "morning" snooze presets
// (tomorrow, this <weekday>, next week), matching Gmail's own defaults.
const snoozeMorningHour = 8

var weekdayNames = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

// ComputeWakeTime computes the time at which a snoozed item should reappear,
// given a preset string and the current time.
//
// Keyword presets are checked before duration parsing: utils.ParseDuration
// returns a zero duration (and no error) for any string with no numeric
// component, which would otherwise silently swallow keywords like "tomorrow".
func ComputeWakeTime(preset string, now time.Time) (time.Time, error) {
	lower := strings.ToLower(strings.TrimSpace(preset))

	switch {
	case lower == "tomorrow":
		return atMorning(now.AddDate(0, 0, 1)), nil
	case lower == "next week":
		return nextOccurrenceOf(now, time.Monday), nil
	case strings.HasPrefix(lower, "this "):
		if wd, ok := weekdayNames[strings.TrimPrefix(lower, "this ")]; ok {
			return nextOccurrenceOf(now, wd), nil
		}
	}

	// utils.ParseDuration returns a zero duration (and no error) for any
	// string with no digits at all, so only attempt it for strings that
	// could plausibly be a duration - otherwise every unrecognized keyword
	// would silently resolve to "now".
	if strings.ContainsAny(preset, "0123456789") {
		if d, err := utils.ParseDuration(preset); err == nil {
			return now.Add(d), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized snooze preset: %q", preset)
}

// ResolveSnoozePreset parses input as a 1-based index into presets and
// resolves it to a wake-at time. ok is false for invalid input (non-numeric,
// out of range) or an unrecognized preset, in which case callers should
// silently ignore the input.
func ResolveSnoozePreset(
	input string,
	presets []config.SnoozePreset,
	now time.Time,
) (wakeAt time.Time, ok bool) {
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(presets) {
		return time.Time{}, false
	}

	wakeAt, err = ComputeWakeTime(presets[idx-1].After, now)
	if err != nil {
		return time.Time{}, false
	}

	return wakeAt, true
}

// atMorning returns t with its time-of-day set to snoozeMorningHour:00:00.
func atMorning(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), snoozeMorningHour, 0, 0, 0, t.Location())
}

// nextOccurrenceOf returns the next occurrence of the given weekday at
// snoozeMorningHour, at least one day out. If now falls on the target
// weekday, it rolls to next week's occurrence rather than possibly returning
// a time already in the past today.
func nextOccurrenceOf(now time.Time, target time.Weekday) time.Time {
	daysUntil := (int(target) - int(now.Weekday()) + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7
	}
	return atMorning(now.AddDate(0, 0, daysUntil))
}
