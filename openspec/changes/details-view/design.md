## Context

See `proposal.md` - Why/What Changes for motivation and scope.

Relevant existing structure (`internal/tui/ui.go` and
`internal/tui/components/sidebar/sidebar.go`):

- The top-level `Model` in `ui.go` owns one `sidebar.Model` plus per-domain
  render models (`prView`, `issueSidebar`, `branchSidebar`,
  `notificationView`) that each know how to render their row's data as a
  string. `syncSidebar()` picks the right render model based on the type of
  `m.getCurrRowData()` and calls `m.sidebar.SetContent(...)`.
- `sidebar.Model` is a thin wrapper around a bubbles `viewport.Model`. It
  already renders either a right-column or bottom-strip layout based on
  `ctx.PreviewPosition`, and exposes `ScrollToTop`/`ScrollToBottom` (mapped
  to the viewport's `GotoTop`/`GotoBottom`) and handles `PageUp`/`PageDown`
  internally via `key.Matches` in its own `Update`.
- `Model.Update` in `ui.go` handles `tea.KeyMsg` through one large `switch`:
  a first tier of universal keys (`Up`, `Down`, `FirstLine`, `LastLine`,
  `TogglePreview`, section switching, etc.) that act on `currSection`, then
  a second tier `switch m.ctx.View` with per-view keybindings (PR actions,
  issue actions, notification flow).
- `enter` is unbound at the universal and PR/Issue/Repo level today; it is
  only bound as `NotificationKeys.View` (load notification content into the
  split preview).

## Goals / Non-Goals

**Goals:**
- Add a details-view mode that reuses the existing per-row render models and
  `sidebar.Model` viewport unchanged in how they produce content — only how
  much of the screen they occupy and how navigation keys are routed changes.
- Keep the key-routing change centralized (one guard near the top of
  `Update`), not duplicated across the four `m.ctx.View` branches.
- Make `esc` behavior compose with the existing notification
  `BackToNotification` (`esc`) semantics rather than replacing them.

**Non-Goals:**
- No changes to how preview content is fetched, rendered, or styled per
  domain (PR/Issue/Branch/Notification render models are untouched).
- No new persistent state (e.g., remembering details-view across restarts).
- No mouse support for the details view beyond what already exists for the
  sidebar.
- Not adding details-view entry points beyond `enter` on a list row (e.g.,
  no dedicated command palette entry).

## Decisions

### 1. Represent details-view as a single boolean-ish mode flag on `Model`, not a new state machine
Add `detailsView bool` (name TBD at implementation time, e.g.
`m.inDetailsView`) to `tui.Model`, plus a small struct capturing what to
restore on exit:
```go
type detailsViewState struct {
    active        bool
    prevOpen      bool   // m.sidebar.IsOpen before entering
    prevPosition  string // m.positionOverride before entering, if any
}
```
Alternatives considered: a generic `ViewMode` enum shared with other
possible future full-window modes. Rejected for now as speculative -
there's only one such mode today, and the boolean + saved-state struct is
enough to implement the restore-on-`esc` requirement. If a second
full-window mode appears later, this can be generalized then.

### 2. Intercept navigation keys with a guard placed before the existing `Up`/`Down`/`FirstLine`/`LastLine`/section-switch/`TogglePreview*` cases
In `Update`, add a check immediately after the existing early-exit guards
(search focused, prompt confirmation focused, text input focused, confirm
quit, pending notification action) and before the main `switch { ... }`:

```go
if m.detailsViewState.active {
    switch {
    case key.Matches(msg, m.keys.Down):
        m.sidebar.ScrollDown(1)
        return m, nil
    case key.Matches(msg, m.keys.Up):
        m.sidebar.ScrollUp(1)
        return m, nil
    case key.Matches(msg, m.keys.FirstLine):
        m.sidebar.ScrollToTop()
        return m, nil
    case key.Matches(msg, m.keys.LastLine):
        m.sidebar.ScrollToBottom()
        return m, nil
    case key.Matches(msg, m.keys.PrevSection), key.Matches(msg, m.keys.NextSection),
        key.Matches(msg, m.keys.TogglePreview), key.Matches(msg, m.keys.TogglePreviewPosition):
        return m, nil
    case key.Matches(msg, keys.Esc): // new universal binding, see Decision 4
        m.exitDetailsView()
        return m, nil
    }
    // fall through: everything else (item actions, page up/down, tab
    // switching, notification back-to-list, etc.) continues to the
    // existing per-view switch below unchanged.
}
```
This keeps the per-view `switch m.ctx.View` blocks (PR actions, issue
actions, notification flow) completely unchanged - they still receive and
handle their action keys normally, since the guard above only intercepts
the small set of keys whose meaning changes.

`sidebar.Model` needs two new one-line-scroll methods (`ScrollDown(n int)`/
`ScrollUp(n int)`) that call through to the underlying viewport's existing
`ScrollDown`/`ScrollUp`, mirroring the existing `ScrollToTop`/
`ScrollToBottom` wrappers.

Alternatives considered: pushing this logic into `sidebar.Model.Update` and
routing all key messages there while in details view. Rejected because item
actions (approve, comment, checkout, ...) live in `ui.go`'s per-view
switch and operate on `prView`/`issueSidebar`/etc., not on `sidebar.Model`
itself - routing everything through `sidebar.Update` first would require
duplicating that dispatch inside the sidebar package, inverting the
existing ownership (sidebar is dumb content display; `ui.go` owns
behavior).

### 3. Entering details view: new case in the universal key switch
Add one case to the main key switch (works for PRs/Issues/Repo, and for
Notifications only when a subject is already loaded — see Decision 5):
```go
case key.Matches(msg, keys.Enter):
    if currSection != nil && currRowData != nil && m.readyForDetailsView() {
        m.enterDetailsView()
    }
```
`m.readyForDetailsView()` returns true for PRs/Issues/Repo views
unconditionally, and for Notifications only when
`m.notificationView.GetSubjectPR() != nil || m.notificationView.GetSubjectIssue() != nil`
(content already loaded). `enterDetailsView()` snapshots
`detailsViewState{active: true, prevOpen: m.sidebar.IsOpen, prevPosition:
m.positionOverride}`, force-opens the sidebar if it was closed
(`m.sidebar.IsOpen = true`), and calls `m.syncMainContentDimensions()` so
the sidebar viewport is resized to full-window before the next render.

### 4. New universal `enter`/`esc` bindings, layered with existing notification `esc`
Add `Enter key.Binding` (`"enter"`) and reuse a new universal `Esc key.Binding`
(`"esc"`) to `keys.KeyMap`, rebindable like other universal keys. The
Notifications view's existing `NotificationKeys.BackToNotification` (`esc`,
"back to notification") remains a separate, per-view binding matched later
in the `m.ctx.View == config.NotificationsView` switch; it's simply
unreachable while `m.detailsViewState.active` is true, because the new
top-level `esc` case (Decision 2) returns early first. This gives the
layering the proposal asks for: first `esc` exits details view (if active),
next `esc` (now that details view is inactive) falls through to
`BackToNotification` as it does today.

Alternatives considered: keeping `esc` scoped only to
`NotificationKeys.BackToNotification` and adding a distinct key (e.g. `q`)
to exit details view. Rejected - `esc` is the conventional "back/cancel" key
already used elsewhere in the app (search, prompts, text input), and the
proposal explicitly calls for `esc` here.

### 5. `View()` full-window render path
Add a branch at the top of `Model.View()`: if `m.detailsViewState.active`,
render only `m.sidebar.View()` (sized to the full content area, same width
math already used for `PreviewPosition == "bottom"` at full height, reused
by temporarily treating the details view as a full-height single pane
regardless of `PreviewPosition`) plus the existing footer, skipping tabs and
the section list entirely. `syncMainContentDimensions()` gains a branch so
that when details view is active, `DynamicPreviewWidth`/`Height` are set to
the full main content area instead of the split fraction.

## Risks / Trade-offs

- **[Risk]** `enter` is a very commonly-expected key for "open/select" in
  TUIs; making it universal could collide with a user's custom keybinding
  that already maps another action to `enter` via `Keybindings.Universal`
  config. → Mitigation: `Rebind` already lets users remap any universal
  key including `Enter`; document the new default in the changelog/config
  docs so existing custom configs can be adjusted if needed.
- **[Risk]** Forcing `m.sidebar.IsOpen = true` on entry and restoring it on
  exit could interact oddly with `TogglePreviewPosition`'s
  `m.positionOverride` if a user had never opened the preview in the
  current session (no meaningful "position" to restore). → Mitigation:
  `positionOverride == ""` is already a valid "no override, use
  config/auto" state; restoring `""` is a no-op, which is the correct
  behavior (falls back to configured/auto position).
- **[Risk]** The Notifications two-step `enter` behavior (load, then
  details) adds a small amount of state-dependent branching
  (`readyForDetailsView`) that the other three views don't need. →
  Mitigation: scenario-level tests in `ui_test.go` cover both the
  "not yet loaded" and "already loaded" cases explicitly.
- **[Trade-off]** Reusing `sidebar.Model`'s existing viewport for full-window
  rendering instead of introducing a dedicated "details" component avoids
  duplicating render logic, but means the details view's line-wrapping/width
  is whatever the sidebar already produces at full width — any preview
  content that assumes a narrow column (if such assumptions exist) would
  need to already handle a wide viewport correctly. This should be verified
  during implementation by checking `prview`/`issueview` render code for
  hard-coded width assumptions rather than using `ctx`-provided width.
