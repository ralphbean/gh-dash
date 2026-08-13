## Context

See `proposal.md` for motivation and scope. Relevant existing structure:

- The full-window details view (`internal/tui/ui.go`, `detailsViewState`)
  already exists (see the archived-in-spirit `details-view` change under
  `openspec/changes/details-view/`). It's entered via a universal `enter`
  binding, tracked with `m.detailsViewState.active`, and `Update` has a
  single centralized guard (placed after the existing focus/prompt/confirm
  early-exit checks, before the main per-view key switch) that intercepts
  `Up`/`Down`/`FirstLine`/`LastLine` (routed to `m.sidebar.ScrollDown/Up/
  ScrollToTop/ScrollToBottom`), no-ops `PrevSection`/`NextSection`/
  `TogglePreview*`, and handles `Esc` (`m.exitDetailsView()`). Everything
  else falls through to the existing per-view switch unchanged.
- `internal/tui/components/prview/prview.go` (`prview.Model`) owns PR
  rendering and in-place edit flows. It already has a single `editor`
  field (`cmpcontroller.Controller`) reused across Comment/Approve/Assign/
  Label modes; `IsTextInputBoxFocused()` reports whether that editor has
  focus, and `ui.go` has an early-exit guard (before the details-view
  guard) that routes all keys straight to `m.prView.Update(msg)` whenever
  it does. Submission happens inside `prview.Model.Update`'s `ctrl+d`
  branch, which inspects `m.editor.Mode()` and returns the matching
  `tasks.XxxPR(...)` command directly - it does not go back through
  `ui.go`'s action dispatch.
  - Today the editor is only rendered inside `viewOverviewTab()`, which is
    why opening it forces the Overview tab first (`GoToFirstTab()` in
    `ui.go`'s `openSidebarForPRInput`). Triage mode does not need this: its
    view is its own top-level branch in `prview.Model.View()`, so it can
    render the editor inline itself, independent of the tab carousel.
- `internal/data/prapi.go`'s `EnrichedPullRequestData.ReviewThreads`
  (`ReviewThreadsWithComments`) already fetches, per thread, `Id`,
  `IsOutdated`, `OriginalLine`, `StartLine`, `Line`, `Path`, and up to 20
  comments (`ReviewComment`: author, body, updatedAt, startLine, line) -
  but not `IsResolved`, `ViewerCanReply`, `ViewerCanResolve`, or the
  comments' `DiffHunk`. `EnrichCurrRow()` only fetches enriched data once
  per row (guarded by `IsEnriched`); nothing in the app currently forces a
  refetch of already-enriched data.
- Every existing write action (approve, comment, label, assign, merge,
  close, checkout, snooze, ...) shells out to `gh` via
  `internal/tui/components/tasks`' `fireTask`/`GitHubTask`, which also
  drives the footer's task-in-progress/finished feedback. Nothing in the
  codebase calls the GraphQL client's `Mutate` method directly - `client`
  (in `internal/data`) is only ever used for `.Query(...)`.

## Goals / Non-Goals

**Goals:**
- Reuse the existing details-view key-routing guard's centralization
  approach for triage's own keys, rather than adding a second, separately
  maintained key-routing layer.
- Reuse the existing `editor`/`cmpcontroller` reply mechanism and its
  `IsTextInputBoxFocused()` early-exit routing unchanged, just with a new
  `Mode` and a different submit-time task call.
  - Keep all thread data (`Id`, `IsResolved`, `ViewerCanReply`,
    `ViewerCanResolve`, comments, `DiffHunk`) inside the existing enriched
    PR struct on `prview.Model`, rather than introducing a parallel data
    model, so applying `UpdatePRMsg`-style local updates after a mutation
    stays a matter of mutating the same struct fields other actions
    already mutate.

**Non-Goals:**
- No in-TUI full diff viewer. "Jump to the diff hunk" is satisfied by
  rendering the anchoring comment's existing `diffHunk` text inline in the
  thread panel (a few lines of context GitHub already computes and
  returns), not by adding diff parsing/rendering or by trying to scroll an
  external pager to a specific location.
- No combined "reply and resolve" action. The three actions map 1:1 to the
  three keys (`c` reply, `r` resolve, `n`/`N` move on); chaining them is a
  possible later enhancement, not part of this change.
- No triage entry point from Issues, Repo, or Notifications details views,
  and none from the PRs list view directly (only from a PR's details
  view). Review threads are a PR-only concept and the proposal explicitly
  scopes entry to "PRs section only."
- No persistence of triage progress across sessions or across re-entering
  the workflow for the same PR (each entry re-fetches and rebuilds the
  queue from current server state).

## Decisions

### 1. Triage state lives on `prview.Model`, not on the top-level `tui.Model`
Add a small unexported struct, e.g.:
```go
type threadTriageState struct {
    active       bool
    threads      []reviewThread // ordered queue, resolved ones removed as they're resolved
    currentIndex int
    prevTab      string // carousel item to restore on exit
}
```
as a field on `prview.Model`, plus methods `EnterThreadTriage() tea.Cmd`,
`ExitThreadTriage()`, `IsTriagingThreads() bool`, `NextThread()`,
`PrevThread()`, `ResolveCurrentThread() tea.Cmd`, `StartThreadReply() tea.Cmd`.

Alternatives considered: a new top-level field on `tui.Model` (mirroring
`detailsViewState`). Rejected - triage is entirely about *what a PR's
sidebar content shows and does*, the same responsibility `prview.Model`
already owns for Comment/Approve/Assign/Label; putting it on `tui.Model`
would split one cohesive piece of behavior across two files for no benefit,
and would require `ui.go` to reach into `prview` internals it doesn't
otherwise touch (e.g. thread lists) to render anything.

### 2. `prview.Model.View()` gets an early branch for triage, bypassing the tab carousel
```go
func (m Model) View() string {
    if !m.hasData() { return "" }
    if m.threadTriage.active {
        return m.viewThreadTriage() // header + diff hunk + comments + editor if replying
    }
    ... existing carousel switch ...
}
```
`viewThreadTriage()` renders the current thread's file/line header, its
diff hunk (styled similarly to `renderComment`'s existing file/line
header), its comments (reusing `renderComment`), and, when
`m.editor.Mode() == cmpcontroller.ModeThreadReply`, the editor view inline
at the bottom - mirroring exactly how `viewOverviewTab()` already appends
`m.editor.View()` today.

Alternatives considered: adding "Threads" as a sixth carousel tab.
Rejected - the proposal's triage workflow is a distinct, focused mode (it
takes over next/prev/reply/resolve keys and suppresses other PR actions),
not a passive read-only tab a user would casually flip to with `[`/`]`;
modeling it as a tab would blur that distinction and complicate the
tab-switching keys' existing behavior during triage (would `[`/`]` need to
special-case leaving the Threads tab differently than others?).

### 3. Key routing: extend the existing details-view guard in `ui.go`, gated on `m.prView.IsTriagingThreads()`
Add cases to the *existing* `if m.detailsViewState.active { switch { ... } }`
block (the same one from the `details-view` change), rather than a second,
separate guard:
```go
case m.prView.IsTriagingThreads() && key.Matches(msg, m.keys.Esc):
    m.prView.ExitThreadTriage()
    m.syncSidebar()
    return m, nil

case m.prView.IsTriagingThreads() && key.Matches(msg, keys.PRKeys.TriageNextThread):
    m.prView.NextThread()
    m.syncSidebar()
    return m, nil

case m.prView.IsTriagingThreads() && key.Matches(msg, keys.PRKeys.TriagePrevThread):
    m.prView.PrevThread()
    m.syncSidebar()
    return m, nil

case m.prView.IsTriagingThreads() && key.Matches(msg, keys.PRKeys.Comment):
    return m, m.prView.StartThreadReply()

case m.prView.IsTriagingThreads() && key.Matches(msg, keys.PRKeys.TriageResolve):
    return m, m.prView.ResolveCurrentThread(sectionIdentifier)

case m.prView.IsTriagingThreads():
    // catch-all: swallow anything else (approve, merge, label, ...) while triaging
    return m, nil
```
placed so the existing `Esc` case above it is checked as an ordinary case
too - Go evaluates `switch { case cond: }` arms in order, so putting the
triage-specific `Esc` case *before* the existing `key.Matches(msg,
m.keys.Esc): m.exitDetailsView()` case gives exactly the layering the spec
requires (first `Esc` exits triage, second exits details view), the same
pattern already used for `NotificationKeys.BackToNotification` layering
under the plain details-view guard.

The `PRKeys.Comment` case reuses the existing binding rather than adding a
new one for "reply" - while triaging, this whole guard block runs *before*
the main per-view switch (where `PRKeys.Comment` normally opens a
general-PR comment), so there's no ambiguity at runtime about which
behavior fires; it's the same "same key, different meaning depending on
mode" pattern the details-view change already established for `enter` in
Notifications.

New `keys.PRKeys` bindings needed: `TriageThreads` (`T`, to *enter* the
workflow - matched in the main per-view switch, gated on
`m.detailsViewState.active`, since triage can only start from within an
already-active details view), `TriageNextThread` (`n`), `TriagePrevThread`
(`N`), `TriageResolve` (`r`). These are freely assignable: while triaging,
this guard fully owns key handling before any other switch runs, so there
is no live conflict with `PRKeys.Refresh`/`RefreshAll` (`r`/`R`) or any
other existing binding sharing a letter - those bindings simply don't fire
while `IsTriagingThreads()` is true.

Alternatives considered: giving triage entirely new, currently-unclaimed
letters for every action (only `i` is free across every existing keymap
combined). Rejected as unnecessary given the guard's placement already
makes reuse conflict-free, and reusing `c` for "reply" and `r` for
"resolve" is more discoverable/mnemonic than the letters that would
otherwise be free.

### 4. Entering triage: refetch review threads, then build the queue
```go
func (m *Model) EnterThreadTriage() tea.Cmd {
    return tasks.FetchReviewThreads(m.ctx, m.pr.Data.Primary, func(threads []data.ReviewThread) tea.Msg {
        return ReviewThreadsFetchedMsg{Threads: threads}
    })
}
```
On `ReviewThreadsFetchedMsg`, `prview.Model.Update` replaces
`m.pr.Data.Enriched.ReviewThreads` with the fresh result, filters to
`IsResolved == false`, sorts by `(Path, Line)`, sets
`m.threadTriage = threadTriageState{active: true, threads: ..., prevTab: m.carousel.SelectedItem()}`.

Alternatives considered: reusing `EnrichCurrRow()` (which no-ops once
`IsEnriched` is true). Rejected - it's specifically designed to fetch once
and cache, which is right for the details view in general but wrong for a
workflow whose entire purpose is to be accurate against a hard merge gate;
a dedicated, always-fresh fetch avoids adding a "force" flag to the
general enrichment path for one caller.

### 5. Mutations via `gh api graphql`, following the existing `GitHubTask` pattern
```go
func ReplyToReviewThread(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, threadId, body string) tea.Cmd {
    return fireTask(ctx, GitHubTask{
        Id: buildTaskId("thread_reply", pr.GetNumber()),
        Args: []string{
            "api", "graphql",
            "-f", "query=mutation($threadId:ID!,$body:String!){addPullRequestReviewThreadReply(input:{pullRequestReviewThreadId:$threadId,body:$body}){comment{id}}}",
            "-f", "threadId=" + threadId,
            "-f", "body=" + body,
        },
        Section:      section,
        StartText:    "Replying to review thread",
        FinishedText: "Replied to review thread",
        Msg: func(c *exec.Cmd, err error) tea.Msg {
            return UpdatePRMsg{PrNumber: pr.GetNumber(), NewThreadReply: &ThreadReply{ThreadId: threadId, Comment: ...}}
        },
    })
}

func ResolveReviewThread(ctx *context.ProgramContext, section SectionIdentifier, pr data.RowData, threadId string) tea.Cmd {
    return fireTask(ctx, GitHubTask{
        Id: buildTaskId("thread_resolve", pr.GetNumber()),
        Args: []string{
            "api", "graphql",
            "-f", "query=mutation($threadId:ID!){resolveReviewThread(input:{threadId:$threadId}){thread{id}}}",
            "-f", "threadId=" + threadId,
        },
        Section:      section,
        StartText:    "Resolving review thread",
        FinishedText: "Review thread resolved",
        Msg: func(c *exec.Cmd, err error) tea.Msg {
            return UpdatePRMsg{PrNumber: pr.GetNumber(), ResolvedThreadId: &threadId}
        },
    })
}
```
`prssection.go`'s existing `case tasks.UpdatePRMsg:` switch gains two
branches: `ResolvedThreadId` marks the matching thread's `IsResolved` true
in `currPr.Enriched.ReviewThreads.Nodes` (which also updates what the
`pr-list-columns` unresolved-count column shows on the next render);
`NewThreadReply` appends the reply to that thread's `Comments.Nodes`.

Alternatives considered:
- Calling the GraphQL client's `.Mutate(...)` directly from `internal/data`
  (avoiding a subprocess). Rejected for consistency: every other mutation
  in this app - including ones that could equally well use the GraphQL
  client - goes through `gh` as a subprocess via `fireTask`, which is also
  what gives users the existing footer task-progress/finished/error
  feedback for free. Introducing a second mutation pathway for just these
  two actions would be a deeper architectural change than this feature
  warrants.
- Using `gh pr comment --reply-to <id>` or similar. Rejected - the `gh`
  CLI has no subcommand for review-thread replies or resolution; `gh api
  graphql` is the only way to reach these mutations through `gh`.

### 6. Diff hunk rendering
`viewThreadTriage()` renders the diff hunk of the thread's *first* comment
(the one that anchors the thread) using the same syntax as `gh`/GitHub
diffs (`+`/`-`/context lines), styled with the theme's existing
success/error colors for added/removed lines - reusing the coloring
already used for `renderFile`'s additions/deletions counts, applied
per-line instead of as a single count.

Alternatives considered: rendering every comment's own `diffHunk` (they
can differ slightly for older comments on an updated thread). Rejected -
GitHub's own UI shows one hunk per thread (from the original/anchoring
comment); showing multiple slightly-different hunks for one thread would
be confusing rather than clarifying, and the anchoring comment's hunk is
the one that best represents "where this thread is in the diff."

## Risks / Trade-offs

- **[Risk]** Reusing `c` (Comment) and `r` (Refresh) with different
  meanings depending on mode adds a small amount of "same key, different
  behavior" surface area to reason about, on top of the two cases the
  details-view change already introduced (`enter`, `esc`). → Mitigation:
  all triage-mode reuse is centralized in the one guard described in
  Decision 3, matching the codebase's existing precedent and keeping the
  "what does this key do right now" logic in a single, easily-audited
  place rather than scattered across per-view switches.
- **[Risk]** Always refetching review threads on triage entry adds an API
  call every time the workflow is opened, including re-opening it
  repeatedly for the same PR in one session. → Mitigation: this is a
  deliberate trade-off (see Decision 4) given the feature's purpose is
  accuracy against a merge gate; the call is a single GraphQL query, not
  materially more expensive than the enrichment fetch already performed
  the first time a PR's details view is opened.
- **[Risk]** `gh api graphql` subprocess calls are slower and have more
  failure modes (auth scopes, rate limiting, malformed `-f` escaping for
  reply bodies containing special characters) than a direct typed mutation
  call. → Mitigation: `fireTask` already surfaces subprocess errors through
  the existing footer error UI used by every other action; body values are
  passed via `-f` (not interpolated into the query string), which `gh api
  graphql` handles as a properly-escaped GraphQL variable.
- **[Trade-off]** Showing only the anchoring comment's diff hunk (Decision
  6) means a thread on a since-updated (`isOutdated`) line shows the hunk
  as it looked when the thread was created, not the current file content.
  This matches GitHub's own web UI behavior for outdated threads, so it
  should not surprise users familiar with GitHub, but it does mean the
  hunk can't be used to verify the *current* state of the line - only the
  state being discussed.
