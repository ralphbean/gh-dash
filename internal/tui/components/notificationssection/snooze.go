package notificationssection

import (
	"fmt"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
)

// snoozeKey builds the SnoozeStore key for a notification thread ID.
func snoozeKey(id string) string {
	return fmt.Sprintf("notification:%s", id)
}

// applySnooze parses input as a 1-based index into the configured snooze
// presets and, if valid, snoozes notification and removes it from the
// visible list immediately (no need to wait for the next fetch). Invalid
// input (bad index, non-numeric, unrecognized preset) is silently ignored,
// in which case applySnooze returns false.
func (m *Model) applySnooze(input string, notification *notificationrow.Data) bool {
	if notification == nil {
		return false
	}

	id := notification.GetId()
	if !data.ApplySnoozePreset(input, snoozeKey(id), m.Ctx.Config.Defaults.SnoozePresets) {
		return false
	}

	for i, n := range m.Notifications {
		if n.GetId() == id {
			m.Notifications = append(m.Notifications[:i], m.Notifications[i+1:]...)
			break
		}
	}
	m.TotalCount = len(m.Notifications)
	return true
}
