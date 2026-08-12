package prssection

import (
	"fmt"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// snoozeKey builds the SnoozeStore key for a PR.
func snoozeKey(pr data.RowData) string {
	return fmt.Sprintf("pr:%s#%d", pr.GetRepoNameWithOwner(), pr.GetNumber())
}

// applySnooze parses input as a 1-based index into the configured snooze
// presets and, if valid, snoozes pr until the computed wake time. Invalid
// input (bad index, non-numeric, unrecognized preset) is silently ignored.
func (m *Model) applySnooze(input string, pr data.RowData) {
	if pr == nil {
		return
	}

	wakeAt, ok := data.ResolveSnoozePreset(input, m.Ctx.Config.Defaults.SnoozePresets, time.Now())
	if !ok {
		return
	}

	data.GetSnoozeStore().Snooze(snoozeKey(pr), wakeAt)
}
