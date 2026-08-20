## 1. Search state on `prview.Model`

- [ ] 1.1 Add the `activitySearchState` struct and field to `prview.Model`
      per design.md Decision 1 (`active`, `term`, `matches` array,
      `currentIndex`, `direction`), along with helper types `matchPos` and
      `searchDir`.
- [ ] 1.2 Implement `StartSearch(direction searchDir) tea.Cmd`: sets
      `active: true`, `direction`, and calls `m.editor.Enter(...)` with a
      new `cmpcontroller.ModeSearch` mode (added to `cmpcontroller`'s `Mode`
      enum).
- [ ] 1.3 Implement `IsSearching() bool`: returns `m.activitySearch.active`.
- [ ] 1.4 Implement `ClearSearch()`: resets `activitySearch` to zero value
      and calls `m.editor.Exit()` to close the search input box.
- [ ] 1.5 Implement `ExecuteSearch(term string)`: strips ANSI codes from
      each `RenderedActivity.RenderedString` (using a library like
      `github.com/acarl005/stripansi`), searches for `term`
      (case-insensitive) in the plain text, records all match positions in
      `matches`, sets `currentIndex` to 0 (forward search) or
      `len(matches)-1` (backward search), and updates the editor's prompt to
      show match count.
- [ ] 1.6 Implement `NextMatch()` and `PrevMatch()`: increment/decrement
      `currentIndex` with wrapping, then call a new `sidebar.ScrollToLine`
      method (task 2.1) to auto-scroll to the current match.
- [ ] 1.7 Add unit tests for search state management (queue building,
      wrapping, case-insensitive matching).

## 2. Sidebar auto-scroll enhancement

- [ ] 2.1 Add `ScrollToLine(line int)` method to
      `internal/tui/components/sidebar/sidebar.go`: calculates the target
      Y offset to center the given line in the viewport and calls
      `m.viewport.SetYOffset(...)`, per design.md Decision 4.
- [ ] 2.2 Add unit test for `ScrollToLine`: verify it correctly centers a
      line, and correctly handles edge cases (lines near top/bottom where
      centering would go out of bounds).

## 3. Search rendering and highlighting

- [ ] 3.1 Extend `renderActivity()` (or wrap it with a new function) in
      `internal/tui/components/prview/activity.go` to accept search state
      and apply highlights when `m.activitySearch.active` is true.
- [ ] 3.2 Implement match highlighting: for each `matchPos`, split the line
      at the match boundaries, wrap the matched substring in a lipgloss
      style with a distinct background color (e.g., `theme.SelectedBackground`
      for non-current matches, brighter for the current match), and
      reassemble, per design.md Decision 3.
- [ ] 3.3 Map match positions from plain-text coordinates (where the search
      ran) back to styled-text coordinates (where highlights are applied) by
      counting characters while skipping ANSI codes, per design.md Decision 8.
- [ ] 3.4 Render the search input box at the bottom of the Activity view
      when `m.activitySearch.active` is true and `m.editor.Active()` is
      true, similar to how the comment editor is rendered in Overview.
- [ ] 3.5 Update the search input box prompt dynamically based on state:
      `/ ` or `? ` before Enter, `/ [3/7] ` after Enter with matches,
      `/ [no matches] ` after Enter with no matches, per design.md
      Decision 6.

## 4. Search input handling in `prview.Model.Update`

- [ ] 4.1 Handle Enter key in `cmpcontroller.ModeSearch`: extract the search
      term via `m.editor.Value()`, call `m.ExecuteSearch(term)`, and keep
      the editor open (don't exit) so the user can see match count and
      continue navigating with `n`/`N`.
- [ ] 4.2 Handle Esc key when `m.activitySearch.active`: call
      `m.ClearSearch()` and `m.syncSidebar()` to remove highlights and
      close the input box.
- [ ] 4.3 Clear search state when switching tabs: in `prview.Model.Update`,
      detect tab changes (carousel position changes) and call
      `m.ClearSearch()`.
- [ ] 4.4 Clear search state when `SetRow` is called with a different PR's
      data: detect PR change and call `m.ClearSearch()`.

## 5. Key bindings

- [ ] 5.1 Add `SearchForward`, `SearchBackward`, `NextSearchMatch`,
      `PrevSearchMatch` bindings to `internal/tui/keys/prKeys.go`, using
      `/`, `?`, `n`, `N` respectively.
- [ ] 5.2 Add rebinding support for the new keys in `rebindPRKeys` and
      `PRFullHelp` (or the new `PRSearchHelp` per task 6.x).
- [ ] 5.3 Update docs: add the new keybindings to
      `docs/src/content/docs/getting-started/keybindings/selected-pr.mdx`.

## 6. Key routing in `ui.go`

- [ ] 6.1 Add cases to the existing `if m.detailsViewState.active { switch
      { ... } }` block for `SearchForward` and `SearchBackward`, gated on
      `m.ctx.View == config.PRsView && m.prView.SelectedTab() == " Activity"
      && !m.prView.IsTriagingThreads()`, per design.md Decision 5.
- [ ] 6.2 Add cases for `NextSearchMatch` and `PrevSearchMatch`, gated on
      `m.prView.IsSearching()`.
- [ ] 6.3 Add a case for `Esc` when `m.prView.IsSearching()`, placed before
      the existing details-view `Esc` handler so search clears first, then
      (on a second Esc) the details view exits, following the same layering
      pattern used for thread triage.

## 7. Help footer integration

- [ ] 7.1 Add a package-level `searchActive bool` flag in
      `internal/tui/keys/keys.go`, with a `SetSearchActive(bool)` function,
      following the same pattern as `triageActive` and
      `notificationSubject`, per design.md Decision 7.
- [ ] 7.2 Call `keys.SetSearchActive(m.prView.IsSearching())` in
      `ui.go`'s `View()` method, alongside the existing `SetTriageActive`
      call.
- [ ] 7.3 Update `FullHelp()`'s `case config.PRsView:` branch to read
      `searchActive`: when true and viewing the Activity tab, include
      `SearchForward`, `SearchBackward`, `NextSearchMatch`,
      `PrevSearchMatch` in `additionalKeys`; when false, include only
      `SearchForward` and `SearchBackward` (since `n`/`N` are inactive
      until a search is executed).

## 8. Testing and edge cases

- [ ] 8.1 Add integration test: start search, execute, navigate with `n`/`N`,
      verify auto-scroll and current match highlighting.
- [ ] 8.2 Add test for case-insensitive matching: search for "Alice", verify
      matches "alice", "ALICE", "aLiCe".
- [ ] 8.3 Add test for search with no matches: verify "no matches" indicator
      and no highlights.
- [ ] 8.4 Add test for wrapping: navigate past last match with `n`, verify
      wraps to first; navigate past first match with `N`, verify wraps to
      last.
- [ ] 8.5 Add test for search state clearing: execute search, switch tabs,
      verify highlights cleared and input box closed.
- [ ] 8.6 Add test for search term persistence: search, clear, start new
      search, verify previous term is pre-populated.
- [ ] 8.7 Add test for highlighting preservation with existing ANSI styles:
      search for text within bold/italic/code-formatted markdown, verify
      highlights don't corrupt the original styling.

## 9. Polish and documentation

- [ ] 9.1 Ensure the search input box is styled consistently with other input
      boxes in the app (comment editor, prompt confirmation).
- [ ] 9.2 Update `docs/src/content/docs/getting-started/usage.mdx` with a
      section on searching within the Activity tab.
- [ ] 9.3 Consider adding a visual indicator (beyond highlighting) when no
      matches are found, e.g., a subtle message in the search input prompt.
