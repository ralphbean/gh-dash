package prssection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
)

func withTestSnoozeStore(t *testing.T) *data.SnoozeStore {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "gh-dash-snooze-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	store := data.NewSnoozeStoreForTesting(filepath.Join(tempDir, "snoozed.json"))
	restore := data.OverrideSnoozeStoreForTesting(store)
	t.Cleanup(restore)
	return store
}

func TestApplySnooze_ValidPresetSnoozesPR(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("1")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.True(t, data.GetSnoozeStore().IsSnoozed("pr:#42"),
		"PR should be snoozed after confirming preset 1")
}

func TestApplySnooze_InvalidIndexIsIgnored(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("99")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, data.GetSnoozeStore().IsSnoozed("pr:#42"),
		"invalid preset index should be silently ignored")
}

func TestApplySnooze_NonNumericInputIsIgnored(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("abc")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, data.GetSnoozeStore().IsSnoozed("pr:#42"),
		"non-numeric input should be silently ignored")
}

func TestBuildRows_FiltersSnoozedPRs(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("pr:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Prs = []prrow.Data{
		{Primary: &data.PullRequestData{
			Number:     1,
			Repository: data.Repository{NameWithOwner: "owner/repo"},
		}},
		{Primary: &data.PullRequestData{
			Number:     2,
			Repository: data.Repository{NameWithOwner: "owner/repo"},
		}},
	}

	rows := m.BuildRows()
	require.Len(t, rows, 1, "snoozed PR should be filtered out of BuildRows")
}

func TestGetCurrRow_IndexesIntoVisiblePRsOnly(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("pr:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Prs = []prrow.Data{
		{Primary: &data.PullRequestData{
			Number:     1,
			Repository: data.Repository{NameWithOwner: "owner/repo"},
		}},
		{Primary: &data.PullRequestData{
			Number:     2,
			Repository: data.Repository{NameWithOwner: "owner/repo"},
		}},
	}
	// The table cursor is at index 0, which should now refer to the second
	// (visible) PR, since the first is snoozed.
	row := m.GetCurrRow()
	require.NotNil(t, row)
	require.Equal(t, 2, row.GetNumber())
}
