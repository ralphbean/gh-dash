package notificationssection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
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

// newTestModel creates a real, fully-initialized Model (mirroring
// TestUpdateNotificationKeepsCursorOnNewLastItem in commands_test.go) with
// the prompt confirmation box focused for the given action and a single
// notification so GetCurrNotification returns non-nil.
func newTestModel(t *testing.T, action string) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{Config: &cfg}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)
	ctx.StartTask = func(task context.Task) tea.Cmd {
		return func() tea.Msg { return nil }
	}

	m := NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	m.Notifications = []notificationrow.Data{
		{Notification: data.NotificationData{Id: "notif-A"}},
	}
	m.TotalCount = len(m.Notifications)
	m.Table.SetRows(m.BuildRows())

	m.SetPromptConfirmationAction(action)
	m.SetIsPromptConfirmationShown(true)
	m.PromptConfirmationBox.Focus()

	return m
}

func TestApplySnooze_ValidPresetSnoozesNotification(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("1")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.True(t, data.GetSnoozeStore().IsSnoozed("notification:notif-A"),
		"notification should be snoozed after confirming preset 1")
}

func TestApplySnooze_InvalidIndexIsIgnored(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("99")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, data.GetSnoozeStore().IsSnoozed("notification:notif-A"),
		"invalid preset index should be silently ignored")
}

func TestApplySnooze_NonNumericInputIsIgnored(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("abc")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, data.GetSnoozeStore().IsSnoozed("notification:notif-A"),
		"non-numeric input should be silently ignored")
}

func TestApplySnooze_RemovesNotificationImmediately(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("1")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, _ := m.Update(msg)
	m = *updated.(*Model)

	require.Empty(t, m.Notifications,
		"snoozed notification should be removed from the list immediately")
	require.Equal(t, 0, m.TotalCount)
}
