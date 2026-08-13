## Why

Today the preview/sidebar pane is always secondary to the list: it's either a
narrow column or bottom strip next to the list, or hidden entirely. When you
want to actually read and act on a single PR, issue, or notification (review
a long description, scroll through activity, comment), the list keeps taking
up screen space you don't need. There's no way to focus on just one item
full-window while still being able to navigate within it using the same keys
(`j`/`k`, `g`/`G`) you'd use in the list.

## What Changes

- Add a new full-window "details" view that renders only the preview content
  (what's currently shown in the sidebar) for the currently selected row,
  using the entire terminal window.
- Pressing `enter` on a list item in the PRs, Issues, Repo (branches), or
  Notifications view enters the details view for that item.
  - In the Notifications view, `enter` keeps its current meaning of loading
    the notification's content into the preview on the first press; a second
    `enter` (once content has loaded) enters the details view.
- While in the details view, `j`/`k` (and arrow keys) scroll the preview
  content up/down, and `g`/`G` jump to the top/bottom of the preview content,
  instead of moving the list selection.
- All existing item-level action keybindings (approve, comment, label,
  assign, checkout, diff, close/reopen/merge, snooze, sidebar tab switching
  with `[`/`]`, etc.) continue to work unchanged while in the details view.
- Section-switching keys (`h`/`l`, left/right) and the preview toggle keys
  (`p`/`P`) are no-ops while in the details view, since there is no list or
  alternate section visible to switch to.
- Pressing `esc` exits the details view and returns to the list, restoring
  the preview pane to whatever state (open/closed, and position) it was in
  before `enter` was pressed. In the Notifications view, this first `esc`
  returns to the list with the notification's content still loaded in the
  split preview, matching today's `esc` ("back to notification") semantics
  when the preview was already open before entering the details view.

## Capabilities

### New Capabilities
- `details-view`: A full-window, single-item focus mode entered from any list
  view, reusing each view's existing preview rendering and item-level
  actions, with list-navigation keys remapped to scroll/jump within the
  preview and `esc` returning to the prior list/preview layout.

### Modified Capabilities
(none — no existing capability specs exist in this repo yet to modify)

## Impact

- `internal/tui/ui.go`: top-level `Model` gains details-view state; `Update`
  needs a new branch that, when details view is active, intercepts
  navigation keys (`Up`/`Down`/`FirstLine`/`LastLine`) before they reach the
  current per-section list-navigation handling and routes them to the
  sidebar viewport instead; `esc` handling needs to check details-view state
  before falling through to existing per-view `esc` handling; `View()` needs
  a new full-window render path for the sidebar/preview content.
- `internal/tui/components/sidebar/sidebar.go`: needs scroll-line and
  goto-top/bottom methods usable from the top-level model (page up/down
  already exist; line-level and top/bottom already exist via
  `ScrollToTop`/`ScrollToBottom`, need line-level `j`/`k` step methods).
- `internal/tui/keys/keys.go`: `enter` needs to become a universal binding
  (currently only bound in `NotificationKeys.View`); need to confirm `enter`
  is not already claimed by a per-view action in PRs/Issues/Repo (it isn't
  today) and reserve it going forward.
- `internal/tui/keys/notificationKeys.go`: `BackToNotification` (`esc`)
  semantics need to be layered with the new details-view `esc` handling.
- No changes to data fetching, GitHub API usage, or config schema are
  anticipated.
