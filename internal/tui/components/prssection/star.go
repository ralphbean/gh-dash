package prssection

import (
	"fmt"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

// starKey builds the StarStore key for a PR. Same format as snoozeKey, but
// kept independent since each store is keyed and persisted separately.
func starKey(pr data.RowData) string {
	return fmt.Sprintf("pr:%s#%d", pr.GetRepoNameWithOwner(), pr.GetNumber())
}
