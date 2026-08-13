package prssection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestStarKey_MatchesSnoozeKeyFormat(t *testing.T) {
	pr := &data.PullRequestData{
		Number:     42,
		Repository: data.Repository{NameWithOwner: "owner/repo"},
	}

	require.Equal(t, "pr:owner/repo#42", starKey(pr))
	require.Equal(t, snoozeKey(pr), starKey(pr),
		"starKey should use the same format as snoozeKey")
}
