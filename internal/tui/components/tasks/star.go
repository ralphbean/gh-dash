package tasks

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// StarFeedback surfaces the same footer spinner/checkmark feedback used by
// other actions (close, merge, snooze, ...) for a star toggle that has
// already been applied synchronously. Since there's no external command to
// wait on, it starts the task and immediately reports it as finished.
func StarFeedback(
	ctx *context.ProgramContext,
	section SectionIdentifier,
	key, itemDescription string,
	starred bool,
) tea.Cmd {
	taskId := fmt.Sprintf("star_%s", key)

	startVerb := "Starring"
	finishedVerb := "starred"
	if !starred {
		startVerb = "Unstarring"
		finishedVerb = "unstarred"
	}

	start := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("%s %s", startVerb, itemDescription),
		FinishedText: fmt.Sprintf("%s has been %s", itemDescription, finishedVerb),
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
