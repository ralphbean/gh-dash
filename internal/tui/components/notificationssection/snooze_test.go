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
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

// collectMsgs executes cmd, flattening any tea.BatchMsg it produces into the
// individual tea.Msg values that would ultimately be dispatched.
func collectMsgs(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, c := range batch {
			msgs = append(msgs, collectMsgs(t, c)...)
		}
		return msgs
	}
	return []tea.Msg{msg}
}

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

func TestApplySnooze_FiresSnoozeFeedback(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("1")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	var found bool
	for _, msg := range collectMsgs(t, cmd) {
		if finished, ok := msg.(constants.TaskFinishedMsg); ok {
			found = true
			require.Equal(t, m.Id, finished.SectionId)
			require.Equal(t, SectionType, finished.SectionType)
		}
	}
	require.True(t, found, "confirming a snooze should surface footer feedback")
}

func TestApplySnooze_InvalidIndexDoesNotFireSnoozeFeedback(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel(t, "snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("99")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	for _, msg := range collectMsgs(t, cmd) {
		if _, ok := msg.(constants.TaskFinishedMsg); ok {
			t.Fatal("an invalid snooze should not surface footer feedback")
		}
	}
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
