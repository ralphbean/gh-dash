package tasks

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// SnoozeFeedback surfaces the same footer spinner/checkmark feedback used by
// other actions (close, merge, reopen, ...) for a snooze that has already
// been applied synchronously. Since there's no external command to wait on,
// it starts the task and immediately reports it as finished.
func SnoozeFeedback(
	ctx *context.ProgramContext,
	section SectionIdentifier,
	key, itemDescription string,
) tea.Cmd {
	taskId := fmt.Sprintf("snooze_%s", key)
	start := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Snoozing %s", itemDescription),
		FinishedText: fmt.Sprintf("%s has been snoozed", itemDescription),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := ctx.StartTask(start)
	return tea.Batch(startCmd, func() tea.Msg {
		return constants.TaskFinishedMsg{
			TaskId:      taskId,
			SectionId:   section.Id,
			SectionType: section.Type,
		}
	})
}
