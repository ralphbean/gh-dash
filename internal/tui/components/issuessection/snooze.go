package issuessection

import (
	"fmt"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// snoozeKey builds the SnoozeStore key for an issue.
func snoozeKey(issue data.RowData) string {
	return fmt.Sprintf("issue:%s#%d", issue.GetRepoNameWithOwner(), issue.GetNumber())
}

// applySnooze parses input as a 1-based index into the configured snooze
// presets and, if valid, snoozes issue until the computed wake time. Invalid
// input (bad index, non-numeric, unrecognized preset) is silently ignored,
// in which case applySnooze returns false.
func (m *Model) applySnooze(input string, issue data.RowData) bool {
	if issue == nil {
		return false
	}

	return data.ApplySnoozePreset(input, snoozeKey(issue), m.Ctx.Config.Defaults.SnoozePresets)
}
