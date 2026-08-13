## 1. Keybindings

- [x] 1.1 Add a universal `Enter` binding (`"enter"`) to `keys.KeyMap` in
      `internal/tui/keys/keys.go`, including it in the `rebindUniversal`
      switch (builtin name `"enter"`) so it's user-rebindable.
- [x] 1.2 Add a universal `Esc` binding (`"esc"`) to `keys.KeyMap` the same
      way, with builtin name `"esc"`.
- [x] 1.3 Add both new bindings to `KeyMap.NavigationKeys()` or `AppKeys()`
      (whichever fits) so they show up in help text.
- [x] 1.4 Write/adjust `keys_test.go` coverage for the new bindings and
      their rebinding via `rebindUniversal`.

## 2. Sidebar scroll helpers

- [x] 2.1 Add `ScrollDown(n int)` and `ScrollUp(n int)` methods to
      `sidebar.Model` in `internal/tui/components/sidebar/sidebar.go`,
      delegating to the underlying `viewport.Model`'s `ScrollDown`/
      `ScrollUp`, mirroring the existing `ScrollToTop`/`ScrollToBottom`
      wrappers.
- [x] 2.2 Add/extend `listviewport` or `sidebar` tests covering line-level
      scroll behavior (scrolling down/up by one line, clamped at top/bottom).

## 3. Details-view state on `Model`

- [x] 3.1 Add a `detailsViewState` struct (`active bool`, `prevOpen bool`,
      `prevPosition string`) and a field on `tui.Model` in
      `internal/tui/ui.go`.
- [x] 3.2 Implement `m.enterDetailsView()`: snapshot `prevOpen`/
      `prevPosition`, force `m.sidebar.IsOpen = true`, set
      `detailsViewState.active = true`, call
      `m.syncMainContentDimensions()`.
- [x] 3.3 Implement `m.exitDetailsView()`: restore `m.sidebar.IsOpen` and
      `m.positionOverride` from the snapshot, set
      `detailsViewState.active = false`, call
      `m.syncMainContentDimensions()` and `m.syncProgramContext()` as
      needed.
- [x] 3.4 Implement `m.readyForDetailsView()`: `true` for PRs/Issues/Repo
      views; for Notifications, `true` only when
      `m.notificationView.GetSubjectPR() != nil ||
      m.notificationView.GetSubjectIssue() != nil`.

## 4. Key routing in `Update`

- [x] 4.1 Add the details-view key-interception guard described in
      design.md Decision 2, placed after the existing early-exit guards
      (search/prompt/text-input focus, confirm-quit, pending notification
      action) and before the main key `switch`. Route `Down`/`Up` to
      `sidebar.ScrollDown(1)`/`ScrollUp(1)`, `FirstLine`/`LastLine` to
      `sidebar.ScrollToTop()`/`ScrollToBottom()`, swallow
      `PrevSection`/`NextSection`/`TogglePreview`/`TogglePreviewPosition`,
      and handle the new `Esc` binding by calling `m.exitDetailsView()`.
- [x] 4.2 Add the `Enter` case to the main key switch: when not already in
      details view and `currSection != nil && currRowData != nil &&
      m.readyForDetailsView()`, call `m.enterDetailsView()`.
- [x] 4.3 Verify the existing `NotificationKeys.View` (`enter`, load
      notification content) and `NotificationKeys.BackToNotification`
      (`esc`) cases in the `config.NotificationsView` branch are unreachable
      while `detailsViewState.active` is true (guard in 4.1 returns early),
      and still reachable and functioning as today once details view is
      inactive.
- [x] 4.4 Confirm all per-view action keybindings (PR approve/assign/label/
      comment/checkout/diff/close/reopen/merge/ready/update/approve
      workflows/snooze/sidebar-tab-switch, Issue equivalents, notification
      PR/Issue action flows) are unaffected by the guard in 4.1 (i.e., they
      fall through to the existing per-view switch unchanged).

## 5. Full-window rendering

- [x] 5.1 Add a branch at the top of `Model.View()` in
      `internal/tui/ui.go`: when `detailsViewState.active`, render only
      `m.sidebar.View()` plus the footer, skipping tabs and the section
      list.
- [x] 5.2 Add a branch in `syncMainContentDimensions()` so that when details
      view is active, the sidebar's width/height are computed from the full
      main content area instead of the split-layout fraction.
- [x] 5.3 Check `prview`/`issueview`/`branchsidebar`/`notificationview`
      render code for any hard-coded width assumptions (vs. using the
      `ctx`-provided width) that would need adjusting to render correctly at
      full window width, per design.md's noted trade-off.

## 6. Tests

- [x] 6.1 Add `ui_test.go` cases: `enter` on a PR/Issue/Repo row enters
      details view; `j`/`k`/`g`/`G` scroll the preview instead of moving
      list selection while active; `esc` exits and restores prior
      open/closed preview state (both starting-open and starting-closed
      cases); section-switch and preview-toggle keys are no-ops while
      active.
- [x] 6.2 Add `ui_test.go` cases for Notifications: first `enter` loads
      content without entering details view; second `enter` (content
      loaded) enters details view; `esc` returns to the split preview with
      content still shown.
- [x] 6.3 Add a case confirming an item-level action (e.g., PR approve)
      still triggers correctly while in details view.
- [x] 6.4 Add a case confirming `enter` with no rows in the current section
      is a no-op.

## 7. Docs

- [x] 7.1 Update relevant docs (`docs/src/content/...` keybindings
      reference, if one exists) to document the new `enter` (enter details
      view) and `esc` (exit details view) universal bindings and their
      config keys (`enter`, `esc`) for `Keybindings.Universal`.
