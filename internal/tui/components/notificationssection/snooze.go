package notificationssection

import (
	"fmt"
	"time"

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
// input (bad index, non-numeric, unrecognized preset) is silently ignored.
func (m *Model) applySnooze(input string, notification *notificationrow.Data) {
	if notification == nil {
		return
	}

	wakeAt, ok := data.ResolveSnoozePreset(input, m.Ctx.Config.Defaults.SnoozePresets, time.Now())
	if !ok {
		return
	}

	id := notification.GetId()
	data.GetSnoozeStore().Snooze(snoozeKey(id), wakeAt)

	for i, n := range m.Notifications {
		if n.GetId() == id {
			m.Notifications = append(m.Notifications[:i], m.Notifications[i+1:]...)
			break
		}
	}
	m.TotalCount = len(m.Notifications)
}
