## 1. Data layer

- [x] 1.1 Add `IsResolved`, `ViewerCanReply`, `ViewerCanResolve` fields to
      the `ReviewThreadsWithComments.Nodes` struct in
      `internal/data/prapi.go`.
- [x] 1.2 Add a `DiffHunk` field to the `ReviewComment` struct in
      `internal/data/prapi.go`.
- [x] 1.3 Add a new exported function (e.g. `FetchReviewThreads(prUrl
      string) ([]ReviewThread, error)` or equivalent) that queries just a
      PR's review threads (id, path, line, isResolved, viewerCanReply,
      viewerCanResolve, comments with author/body/updatedAt/diffHunk),
      independent of the full `FetchPullRequest` enrichment query.
- [x] 1.4 Add/extend `prapi_test.go` coverage for the new fields and
      function, including a case with `IsResolved: true` threads correctly
      excluded by triage's queue-building logic (tested at the point that
      logic lives, see task 4.x).

## 2. Mutations (tasks package)

- [x] 2.1 Add `ReplyToReviewThread(ctx, section, pr, threadId, body)
      tea.Cmd` to `internal/tui/components/tasks/` (new file, e.g.
      `reviewthread.go`), following the `fireTask`/`GitHubTask` pattern
      used by `CommentOnPR`, invoking `gh api graphql` with the
      `addPullRequestReviewThreadReply` mutation per design.md Decision 5.
- [x] 2.2 Add `ResolveReviewThread(ctx, section, pr, threadId) tea.Cmd` to
      the same file, invoking `gh api graphql` with the
      `resolveReviewThread` mutation.
- [x] 2.3 Add `ResolvedThreadId *string` and `NewThreadReply
      *ThreadReply` (new small struct: `ThreadId string`, `Comment
      data.ReviewComment`) fields to `tasks.UpdatePRMsg`.
- [x] 2.4 Add unit tests for both new task functions (args built
      correctly, `UpdatePRMsg` populated correctly), mirroring
      `pr_test.go`'s existing coverage of `CommentOnPR`/`ApprovePR`.

## 3. Local state updates

- [x] 3.1 In `internal/tui/components/prssection/prssection.go`'s
      `case tasks.UpdatePRMsg:` switch, handle `ResolvedThreadId` by
      marking the matching node's `IsResolved = true` in
      `currPr.Enriched.ReviewThreads.Nodes`.
- [x] 3.2 In the same switch, handle `NewThreadReply` by appending the
      comment to the matching thread's `Comments.Nodes`.
- [x] 3.3 Add/extend `prssection_test.go` coverage for both new
      `UpdatePRMsg` branches.
- [x] 3.4 Fix: in the `ResolvedThreadId` branch, also flip one
      `IsResolved: false` node in `currPr.Primary.ReviewThreads.Nodes` to
      `true`, so the `pr-list-columns` unresolved-count column reflects
      the resolution on return to the list view, per design.md Decision
      5's correction.
- [x] 3.5 Add/extend `prssection_test.go` coverage: after
      `ResolvedThreadId`, `Primary.ReviewThreads`'s unresolved count (via
      `UnresolvedThreadsCount()`) has decremented by one.

## 4. Triage state and queue on `prview.Model`

- [x] 4.1 Add the `threadTriageState` struct and field to `prview.Model`
      per design.md Decision 1 (`active`, ordered `threads` queue,
      `currentIndex`, `prevTab`).
- [x] 4.2 Implement `EnterThreadTriage() tea.Cmd`: fires the new
      `tasks`/data fetch from task 1.3, wrapped in a message type (e.g.
      `ReviewThreadsFetchedMsg`).
- [x] 4.3 Handle `ReviewThreadsFetchedMsg`: replace
      `m.pr.Data.Enriched.ReviewThreads`, filter to `IsResolved == false`,
      sort by `(Path, Line)`, populate `threadTriage` with `active: true`,
      `currentIndex: 0`, `prevTab: m.carousel.SelectedItem()`. Applied via
      an explicit `SetReviewThreads` method called from a dedicated case
      in `ui.go`'s top-level `Update`, mirroring the existing
      `prview.EnrichedPrMsg`/`SetEnrichedPR` precedent (`ui.go` does not
      unconditionally forward all messages into `prview.Model.Update`).
- [x] 4.4 Implement `IsTriagingThreads() bool`, `NextThread()`,
      `PrevThread()` (wrapping per spec's "Wrapping past the ends of the
      queue" scenario).
- [x] 4.5 Implement `ResolveCurrentThread() tea.Cmd`: no-ops when
      the current thread's `ViewerCanResolve` is false; otherwise returns
      `tasks.ResolveReviewThread(...)` for the current thread's id and
      builds the `tasks.SectionIdentifier` internally (no external
      `section` param, per design.md Decision 1). Queue removal is
      optimistic (immediate): the thread is dropped from the queue and
      the queue advances to the next remaining thread right away, exiting
      triage if none remain, per spec's "Resolve is allowed" and
      "Resolving the last thread in the queue" scenarios.
- [x] 4.6 Implement `StartThreadReply() tea.Cmd`: no-ops when the current
      thread's `ViewerCanReply` is false; otherwise calls
      `m.editor.Enter(...)` with a new `cmpcontroller.ModeThreadReply`
      mode (added to `cmpcontroller`'s `Mode` enum and its
      `usesAutocomplete` switch, mirroring `ModeComment`).
- [x] 4.7 Handle `cmpcontroller.ModeThreadReply` in `prview.Model.Update`'s
      `ctrl+d` submit branch: return
      `tasks.ReplyToReviewThread(ctx, section, pr, currentThreadId,
      value)` instead of `tasks.CommentOnPR(...)`.
- [x] 4.8 Implement `ExitThreadTriage()`: restore `m.carousel` to
      `prevTab`, reset `threadTriage` to zero value.

## 5. Triage rendering

- [x] 5.1 Add `viewThreadTriage() string` to `prview.Model` (new file,
      e.g. `threadtriage.go`): renders the current thread's file/line
      header, diff hunk (colored per design.md Decision 6), and comments
      (reusing `renderComment`), plus the editor view inline when
      `m.editor.Mode() == cmpcontroller.ModeThreadReply`.
- [x] 5.2 Add the empty-state render for "no unresolved threads" (spec's
      "No unresolved threads" scenario).
- [x] 5.3 Branch `prview.Model.View()` to call `viewThreadTriage()` first
      when `m.threadTriage.active`, before the existing carousel switch.
- [x] 5.4 Confirm `getIndentedContentWidth()`-based width is used
      throughout (no hard-coded widths), consistent with the existing
      full-window details-view rendering.
- [x] 5.5 Render an "outdated" indicator in `renderThreadTriageHeader`
      when `thread.IsOutdated` is true, styled with `Theme.WarningText`,
      per design.md Decision 8.

## 6. Keybindings and routing in `ui.go`

- [x] 6.1 Add `TriageThreads` (`T`), `TriageNextThread` (`n`),
      `TriagePrevThread` (`N`), `TriageResolve` (`r`) bindings to
      `PRKeyMap`/`PRKeys` in `internal/tui/keys/prKeys.go`, including
      help text and the `rebindPR`-style switch so they're user-rebindable
      (mirroring existing entries like `Comment`).
- [x] 6.2 Add the `TriageThreads` case to the main per-view key switch in
      `ui.go`, gated on `m.detailsViewState.active && m.ctx.View ==
      config.PrsView`, calling `m.prView.EnterThreadTriage()`.
- [x] 6.3 Extend the existing `if m.detailsViewState.active { switch {
      ... } }` guard in `ui.go` per design.md Decision 3: add the
      triage-gated `Esc`, `TriageNextThread`, `TriagePrevThread`,
      `Comment` (reply), `TriageResolve`, and catch-all cases, placed so
      the triage `Esc` case is checked before the existing
      `m.exitDetailsView()` `Esc` case.
- [x] 6.4 Confirm `Down`/`Up`/`FirstLine`/`LastLine` (already handled by
      the existing guard, routed to `m.sidebar`) correctly scroll triage
      content once `m.sidebar.SetContent` is fed `m.prView.View()`'s
      triage output. Required adding triage-gated `Down`/`Up`/`FirstLine`/
      `LastLine` cases ahead of the triage catch-all in the guard switch,
      since the catch-all would otherwise shadow them - verify via test.
- [x] 6.5 Add a `triageActive`-style package-level flag and
      `SetTriageActive(bool)` to `internal/tui/keys/keys.go`, mirroring
      `notificationSubject`/`SetNotificationSubject`. In `KeyMap.FullHelp()`'s
      `case config.PRsView:` branch, use it to substitute a triage-only
      binding list (`TriageNextThread`, `TriagePrevThread`, `Comment`,
      `TriageResolve`) for `PRFullHelp()` when active, per design.md
      Decision 7. Call `keys.SetTriageActive(m.prView.IsTriagingThreads())`
      from `ui.go` alongside the existing per-render UI state sync.
- [x] 6.6 Remove `TriageNextThread`/`TriagePrevThread`/`TriageResolve` from
      `PRFullHelp()`'s normal (non-triaging) binding list, keeping
      `TriageThreads`, per design.md Decision 7's correction.

## 7. Tests

- [x] 7.1 `ui_test.go`: entering triage via `T` only works when
      `detailsViewState.active` and `m.ctx.View == config.PrsView`; no-op
      otherwise (list view, Issues/Repo/Notifications details views).
- [x] 7.2 `ui_test.go`: `n`/`N` move between threads without changing
      resolved state; wraps at both ends.
- [x] 7.3 `ui_test.go`: `esc` while triaging exits triage back to the
      prior tab, not the details view; a second `esc` then exits the
      details view.
- [x] 7.4 `ui_test.go`: item-level PR actions (approve, merge, label, ...)
      are no-ops while triaging.
- [x] 7.5 `prview` tests: `StartThreadReply`/submit posts via
      `tasks.ReplyToReviewThread` with the correct thread id and does not
      advance/resolve; no-ops when `ViewerCanReply` is false.
- [x] 7.6 `prview` tests: `ResolveCurrentThread` posts via
      `tasks.ResolveReviewThread`, removes the thread from the queue and
      advances (or exits triage when the queue empties); no-ops when
      `ViewerCanResolve` is false.
- [x] 7.7 `prview` tests: entering triage always re-fetches threads, even
      when `m.pr.Data.IsEnriched` is already true.
- [x] 7.8 `prview` tests: empty queue on entry shows the empty state and
      stays in triage until `esc`.
- [x] 7.9 `prview` tests: outdated indicator renders when `IsOutdated` is
      true and is absent when false.

## 8. Docs

- [x] 8.1 Document the new `T`/`n`/`N`/`r` bindings (and `c`'s
      triage-mode meaning) in the keybindings reference under
      `docs/src/content/...`, alongside the existing `enter`/`esc`
      details-view entries.
- [x] 8.2 Note in that same docs section that the in-app help footer (`?`)
      reflects these bindings while triaging, so the keybindings reference
      and the live help stay in sync as a discoverability path (not just
      the static docs page).
