package prview

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmpcontroller"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/fuzzyselect"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
)

// threadTriageState tracks the in-progress review-thread triage workflow for
// the currently viewed PR.
type threadTriageState struct {
	active       bool
	threads      []data.ReviewThreadWithComments // ordered queue, resolved ones removed as they're resolved
	currentIndex int
	prevTab      string // carousel item to restore on exit
}

// ReviewThreadsFetchedMsg is returned once EnterThreadTriage's refetch of the
// PR's current review threads completes.
type ReviewThreadsFetchedMsg struct {
	Threads []data.ReviewThreadWithComments
	Err     error
}

// EnterThreadTriage kicks off an always-fresh fetch of the PR's review
// threads so triage starts from current server state.
func (m *Model) EnterThreadTriage() tea.Cmd {
	if m == nil || m.pr == nil {
		return nil
	}
	url := m.pr.Data.Primary.Url
	return func() tea.Msg {
		threads, err := data.FetchReviewThreads(url)
		return ReviewThreadsFetchedMsg{Threads: threads, Err: err}
	}
}

// SetReviewThreads applies a completed ReviewThreadsFetchedMsg: it replaces
// the PR's enriched review threads with the fresh result and (re)builds the
// triage queue from the unresolved ones.
func (m *Model) SetReviewThreads(msg ReviewThreadsFetchedMsg) {
	if m.pr == nil || msg.Err != nil {
		return
	}
	m.pr.Data.Enriched.ReviewThreads = data.ReviewThreadsWithComments{Nodes: msg.Threads}

	var unresolved []data.ReviewThreadWithComments
	for _, thread := range msg.Threads {
		if !thread.IsResolved {
			unresolved = append(unresolved, thread)
		}
	}
	sort.Slice(unresolved, func(i, j int) bool {
		if unresolved[i].Path != unresolved[j].Path {
			return unresolved[i].Path < unresolved[j].Path
		}
		return unresolved[i].Line < unresolved[j].Line
	})

	m.threadTriage = threadTriageState{
		active:       true,
		threads:      unresolved,
		currentIndex: 0,
		prevTab:      m.carousel.SelectedItem(),
	}
}

// IsTriagingThreads reports whether the review-thread triage workflow is
// currently active.
func (m *Model) IsTriagingThreads() bool {
	return m.threadTriage.active
}

// NextThread moves to the next thread in the queue, wrapping past the end.
func (m *Model) NextThread() {
	n := len(m.threadTriage.threads)
	if n == 0 {
		return
	}
	m.threadTriage.currentIndex = (m.threadTriage.currentIndex + 1) % n
}

// PrevThread moves to the previous thread in the queue, wrapping past the start.
func (m *Model) PrevThread() {
	n := len(m.threadTriage.threads)
	if n == 0 {
		return
	}
	m.threadTriage.currentIndex = (m.threadTriage.currentIndex - 1 + n) % n
}

func (m *Model) currentThread() (data.ReviewThreadWithComments, bool) {
	if !m.threadTriage.active || len(m.threadTriage.threads) == 0 {
		return data.ReviewThreadWithComments{}, false
	}
	return m.threadTriage.threads[m.threadTriage.currentIndex], true
}

// ResolveCurrentThread resolves the thread currently being triaged. It's a
// no-op if there's no current thread or the viewer can't resolve it.
// Removal from the queue is optimistic: the thread is dropped and the
// queue advances immediately, rather than waiting for the mutation to
// complete. If no threads remain, triage exits.
func (m *Model) ResolveCurrentThread() tea.Cmd {
	thread, ok := m.currentThread()
	if !ok || !thread.ViewerCanResolve {
		return nil
	}
	sid := tasks.SectionIdentifier{Id: m.sectionId, Type: prssection.SectionType}
	cmd := tasks.ResolveReviewThread(m.ctx, sid, m.pr.Data.Primary, thread.Id)
	m.removeCurrentThreadFromQueue()
	return cmd
}

// removeCurrentThreadFromQueue drops the current thread from the triage
// queue and advances to the next remaining thread, exiting triage if the
// queue is now empty.
func (m *Model) removeCurrentThreadFromQueue() {
	threads := m.threadTriage.threads
	i := m.threadTriage.currentIndex
	m.threadTriage.threads = append(threads[:i], threads[i+1:]...)

	if len(m.threadTriage.threads) == 0 {
		m.ExitThreadTriage()
		return
	}
	if m.threadTriage.currentIndex >= len(m.threadTriage.threads) {
		m.threadTriage.currentIndex = 0
	}
}

// StartThreadReply opens the reply editor for the thread currently being
// triaged. It's a no-op if there's no current thread or the viewer can't
// reply to it.
func (m *Model) StartThreadReply() tea.Cmd {
	thread, ok := m.currentThread()
	if !ok || !thread.ViewerCanReply {
		return nil
	}
	m.editor.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	return m.editor.Enter(cmpcontroller.EnterOptions{
		Mode:                             cmpcontroller.ModeThreadReply,
		Prompt:                           constants.ReplyPrompt,
		Repo:                             m.repoRef(),
		EnterFetch:                       cmpcontroller.FetchSilent,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
}

// ExitThreadTriage restores the carousel to the tab that was selected before
// triage started and clears the triage state.
func (m *Model) ExitThreadTriage() {
	prevTab := m.threadTriage.prevTab
	m.threadTriage = threadTriageState{}
	for i, tab := range tabs {
		if tab == prevTab {
			m.carousel.SetCursor(i)
			break
		}
	}
}

func (m *Model) viewThreadTriage() string {
	thread, ok := m.currentThread()
	if !ok {
		return lipgloss.NewStyle().
			Padding(0, m.ctx.Styles.Sidebar.ContentPadding).
			Italic(true).
			Foreground(m.ctx.Theme.FaintText).
			Render("No unresolved review threads.")
	}

	width := m.getIndentedContentWidth()
	markdownRenderer := markdown.GetMarkdownRenderer(width, m.ctx)

	body := strings.Builder{}
	body.WriteString(m.renderThreadTriageHeader(thread))
	body.WriteString("\n\n")

	if len(thread.Comments.Nodes) > 0 {
		if hunk := m.renderDiffHunk(thread.Comments.Nodes[0].DiffHunk); hunk != "" {
			body.WriteString(hunk)
			body.WriteString("\n\n")
		}
	}

	var rendered []string
	for _, c := range thread.Comments.Nodes {
		renderedComment, err := m.renderComment(comment{
			Author:    c.Author.Login,
			Body:      c.Body,
			UpdatedAt: c.UpdatedAt,
		}, markdownRenderer)
		if err != nil {
			continue
		}
		rendered = append(rendered, renderedComment)
	}
	body.WriteString(lipgloss.JoinVertical(lipgloss.Left, rendered...))

	if m.editor.Mode() == cmpcontroller.ModeThreadReply {
		body.WriteString("\n\n")
		body.WriteString(m.ctx.Styles.Sidebar.InputBox.Render(m.editor.View()))
	}

	return lipgloss.NewStyle().Padding(0, m.ctx.Styles.Sidebar.ContentPadding).Render(body.String())
}

func (m *Model) renderThreadTriageHeader(thread data.ReviewThreadWithComments) string {
	width := m.getIndentedContentWidth()
	position := fmt.Sprintf("Thread %d of %d", m.threadTriage.currentIndex+1, len(m.threadTriage.threads))
	location := fmt.Sprintf("%s#L%d", thread.Path, thread.Line)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.ctx.Styles.Common.MainTextStyle.Bold(true).Width(width).Render(location),
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Width(width).Render(position),
	)
}

func (m *Model) renderDiffHunk(hunk string) string {
	if hunk == "" {
		return ""
	}
	successStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.SuccessText)
	errorStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText)
	faintStyle := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	lines := strings.Split(hunk, "\n")
	rendered := make([]string, len(lines))
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "@@"):
			rendered[i] = faintStyle.Render(line)
		case strings.HasPrefix(line, "+"):
			rendered[i] = successStyle.Render(line)
		case strings.HasPrefix(line, "-"):
			rendered[i] = errorStyle.Render(line)
		default:
			rendered[i] = line
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.ctx.Theme.FaintBorder).
		Width(m.getIndentedContentWidth()).
		Render(strings.Join(rendered, "\n"))
}
