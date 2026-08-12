package issuessection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prompt"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
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

// newTestModel creates a minimal Model with the prompt confirmation box
// focused and a single issue row so that GetCurrRow returns non-nil.
func newTestModel(action string) Model {
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	if err != nil {
		panic(err)
	}
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}
	m := Model{
		BaseModel: section.BaseModel{
			Ctx:                       ctx,
			IsPromptConfirmationShown: true,
			PromptConfirmationAction:  action,
			PromptConfirmationBox:     prompt.NewModel(ctx),
		},
		Issues: []data.IssueData{
			{Number: 42},
		},
	}
	m.PromptConfirmationBox.Focus()
	m.Table.UpdateProgramContext(ctx)
	return m
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

func TestConfirmation_EmptyInputDoesNotConfirm(t *testing.T) {
	m := newTestModel("close")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"confirmation prompt should be dismissed")
}

func TestConfirmation_AcceptWithLowercaseY(t *testing.T) {
	m := newTestModel("close")
	m.PromptConfirmationBox.SetValue("y")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "lowercase y should execute the action")
}

func TestApplySnooze_ValidPresetSnoozesIssue(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("1")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.True(t, data.GetSnoozeStore().IsSnoozed("issue:#42"),
		"issue should be snoozed after confirming preset 1")
}

func TestApplySnooze_InvalidIndexIsIgnored(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
	m.Ctx.Config.Defaults.SnoozePresets = []config.SnoozePreset{{Label: "10m", After: "10m"}}
	m.PromptConfirmationBox.SetValue("99")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, data.GetSnoozeStore().IsSnoozed("issue:#42"),
		"invalid preset index should be silently ignored")
}

func TestApplySnooze_FiresSnoozeFeedback(t *testing.T) {
	withTestSnoozeStore(t)

	m := newTestModel("snooze")
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

	m := newTestModel("snooze")
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

func TestBuildRows_FiltersSnoozedIssues(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("issue:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Issues = []data.IssueData{
		{Number: 1, Repository: data.Repository{NameWithOwner: "owner/repo"}},
		{Number: 2, Repository: data.Repository{NameWithOwner: "owner/repo"}},
	}

	rows := m.BuildRows()
	require.Len(t, rows, 1, "snoozed issue should be filtered out of BuildRows")
}

func TestNumRows_ExcludesSnoozedIssues(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("issue:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Issues = []data.IssueData{
		{Number: 1, Repository: data.Repository{NameWithOwner: "owner/repo"}},
		{Number: 2, Repository: data.Repository{NameWithOwner: "owner/repo"}},
	}

	require.Equal(t, 1, m.NumRows(),
		"NumRows should exclude snoozed issues so pagination triggers correctly")
}

func TestGetTotalCount_AdjustsForSnoozedIssues(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("issue:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Issues = []data.IssueData{
		{Number: 1, Repository: data.Repository{NameWithOwner: "owner/repo"}},
		{Number: 2, Repository: data.Repository{NameWithOwner: "owner/repo"}},
	}
	m.TotalCount = 2

	require.Equal(t, 1, m.GetTotalCount(),
		"GetTotalCount should exclude snoozed issues so the tab badge matches visible rows")
}

func TestGetCurrRow_IndexesIntoVisibleIssuesOnly(t *testing.T) {
	store := withTestSnoozeStore(t)
	store.Snooze("issue:owner/repo#1", time.Now().Add(time.Hour))

	m := newTestModel("")
	m.Issues = []data.IssueData{
		{Number: 1, Repository: data.Repository{NameWithOwner: "owner/repo"}},
		{Number: 2, Repository: data.Repository{NameWithOwner: "owner/repo"}},
	}

	row := m.GetCurrRow()
	require.NotNil(t, row)
	require.Equal(t, 2, row.GetNumber())
}
