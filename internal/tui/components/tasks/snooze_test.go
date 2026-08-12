package tasks

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestSnoozeFeedback_TaskConfiguration(t *testing.T) {
	var capturedTask context.Task
	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			capturedTask = task
			return nil
		},
	}
	section := SectionIdentifier{Id: 2, Type: "pr"}

	_ = SnoozeFeedback(ctx, section, "pr:owner/repo#42", "PR #42")

	require.Equal(t, "snooze_pr:owner/repo#42", capturedTask.Id)
	require.Equal(t, "Snoozing PR #42", capturedTask.StartText)
	require.Equal(t, "PR #42 has been snoozed", capturedTask.FinishedText)
	require.Equal(t, context.TaskStart, capturedTask.State)
	require.Nil(t, capturedTask.Error)
}

func TestSnoozeFeedback_ReturnsNonNilCommand(t *testing.T) {
	ctx := &context.ProgramContext{StartTask: noopStartTask}
	cmd := SnoozeFeedback(ctx, SectionIdentifier{}, "some-key", "some item")
	require.NotNil(t, cmd, "SnoozeFeedback should return a non-nil command")
}

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

func TestSnoozeFeedback_ProducesTaskFinishedMsg(t *testing.T) {
	ctx := &context.ProgramContext{StartTask: noopStartTask}
	section := SectionIdentifier{Id: 3, Type: "issue"}

	cmd := SnoozeFeedback(ctx, section, "issue:owner/repo#7", "Issue #7")
	msgs := collectMsgs(t, cmd)

	var found bool
	for _, msg := range msgs {
		if finished, ok := msg.(constants.TaskFinishedMsg); ok {
			found = true
			require.Equal(t, "snooze_issue:owner/repo#7", finished.TaskId)
			require.Equal(t, section.Id, finished.SectionId)
			require.Equal(t, section.Type, finished.SectionType)
			require.NoError(t, finished.Err)
		}
	}
	require.True(t, found, "expected a constants.TaskFinishedMsg among the produced messages")
}
