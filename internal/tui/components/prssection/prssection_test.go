package prssection

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prompt"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/search"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

// newTestModel creates a minimal Model with the prompt confirmation box
// focused and a single PR row so that GetCurrRow returns non-nil.
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
			SearchBar:                 search.NewModel(ctx, search.SearchOptions{}),
		},
		Prs: []prrow.Data{
			{Primary: &data.PullRequestData{Number: 42}},
		},
	}
	m.PromptConfirmationBox.Focus()
	m.Table.UpdateProgramContext(ctx)
	return m
}

func TestConfirmation_EmptyInputDoesNotConfirm(t *testing.T) {
	// Pressing Enter without typing anything should NOT confirm, since the
	// prompt says (y/N) indicating N is the default.
	m := newTestModel("close")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"confirmation prompt should be dismissed")
}

func TestConfirmation_AcceptWithLowercaseY(t *testing.T) {
	m := newTestModel("merge")
	m.PromptConfirmationBox.SetValue("y")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "lowercase y should execute the action")
}

func TestConfirmation_AcceptWithUppercaseY(t *testing.T) {
	m := newTestModel("reopen")
	m.PromptConfirmationBox.SetValue("Y")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "uppercase Y should execute the action")
}

func TestConfirmation_RejectWithN(t *testing.T) {
	m := newTestModel("close")
	m.PromptConfirmationBox.SetValue("n")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	// cmd is a batch of (nil, blinkCmd) -- the nil means no action was taken.
	// We verify the prompt is dismissed regardless.
	require.False(t, m.IsPromptConfirmationShown,
		"confirmation prompt should be dismissed on rejection")
	_ = cmd
}

func TestConfirmation_CancelWithEsc(t *testing.T) {
	m := newTestModel("merge")

	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"Esc should dismiss the confirmation prompt")
	_ = cmd
}

func TestConfirmation_CancelWithCtrlC(t *testing.T) {
	m := newTestModel("update")

	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_, cmd := m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"Ctrl+C should dismiss the confirmation prompt")
	_ = cmd
}

func TestConfirmation_AllActions(t *testing.T) {
	actions := []string{"close", "reopen", "ready", "merge", "update", "approveWorkflows"}

	for _, action := range actions {
		t.Run(action+"_empty_input_does_not_confirm", func(t *testing.T) {
			m := newTestModel(action)

			msg := tea.KeyPressMsg{Code: tea.KeyEnter}
			_, _ = m.Update(msg)

			require.False(t, m.IsPromptConfirmationShown,
				"empty input should dismiss prompt for action %q", action)
		})

		t.Run(action+"_explicit_y", func(t *testing.T) {
			m := newTestModel(action)
			m.PromptConfirmationBox.SetValue("y")

			msg := tea.KeyPressMsg{Code: tea.KeyEnter}
			_, cmd := m.Update(msg)

			require.NotNil(t, cmd,
				"explicit y should confirm for action %q", action)
		})
	}
}

func TestUpdatePRMsg_ResolvedThreadIdMarksMatchingThreadResolved(t *testing.T) {
	m := newTestModel("")
	m.Prs[0].Enriched.ReviewThreads.Nodes = []data.ReviewThreadWithComments{
		{Id: "thread-1", IsResolved: false},
		{Id: "thread-2", IsResolved: false},
	}
	threadId := "thread-2"

	_, _ = m.Update(tasks.UpdatePRMsg{PrNumber: 42, ResolvedThreadId: &threadId})

	require.False(t, m.Prs[0].Enriched.ReviewThreads.Nodes[0].IsResolved)
	require.True(t, m.Prs[0].Enriched.ReviewThreads.Nodes[1].IsResolved)
}

func TestUpdatePRMsg_NewThreadReplyAppendsCommentToMatchingThread(t *testing.T) {
	m := newTestModel("")
	m.Prs[0].Enriched.ReviewThreads.Nodes = []data.ReviewThreadWithComments{
		{Id: "thread-1"},
		{Id: "thread-2"},
	}
	reply := tasks.ThreadReply{
		ThreadId: "thread-2",
		Comment:  data.ReviewComment{Body: "sounds good"},
	}

	_, _ = m.Update(tasks.UpdatePRMsg{PrNumber: 42, NewThreadReply: &reply})

	require.Empty(t, m.Prs[0].Enriched.ReviewThreads.Nodes[0].Comments.Nodes)
	require.Len(t, m.Prs[0].Enriched.ReviewThreads.Nodes[1].Comments.Nodes, 1)
	require.Equal(t, "sounds good", m.Prs[0].Enriched.ReviewThreads.Nodes[1].Comments.Nodes[0].Body)
}
