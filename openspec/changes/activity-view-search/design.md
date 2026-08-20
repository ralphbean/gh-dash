## Context

See `proposal.md` for motivation and scope. Relevant existing structure:

- The Activity tab (`internal/tui/components/prview/activity.go`) renders all
  PR comments and reviews chronologically. It calls `renderActivity()` which
  builds an array of `RenderedActivity` structs (each containing `UpdatedAt`
  and `RenderedString`), sorts them by time, and joins them vertically with
  lipgloss. The rendering uses glamour/markdown for comment bodies, which
  means the final output is styled, ANSI-escaped text.
- The sidebar (`internal/tui/components/sidebar/sidebar.go`) already has
  viewport scrolling methods: `ScrollToTop()`, `ScrollToBottom()`,
  `ScrollUp(lines int)`, `ScrollDown(lines int)`. The viewport tracks the
  current scroll offset and visible height.
- The PR view (`internal/tui/components/prview/prview.go`) already uses a
  `cmpcontroller.Controller` for the reply/approve/assign/label editor. That
  controller has modes (`ModeComment`, `ModeApprove`, etc.) and handles input
  focus, Enter/Esc, and rendering the input box.
- Key routing in `ui.go` already has a centralized details-view guard (from
  the `details-view` change) that intercepts keys when
  `m.detailsViewState.active` is true. Thread triage adds its own cases to
  that same guard, gated on `m.prView.IsTriagingThreads()`.
- The footer help system uses package-level flags in `keys` (e.g.,
  `notificationSubject`, `triageActive`) to swap in context-specific
  keybindings. `FullHelp()` reads these flags to decide what to show.

## Goals / Non-Goals

**Goals:**
- Reuse the existing sidebar viewport scrolling infrastructure - no need for
  a separate scroll mechanism.
- Search state lives on `prview.Model`, scoped to the Activity tab (cleared
  when switching tabs or PRs).
- Keep the search interaction vim-like: `/` for forward search, `?` for
  backward, `n`/`N` for next/previous, `Esc` to clear.
- Highlight matches in the rendered content without re-rendering from
  markdown - apply highlights to the already-rendered ANSI strings.
- Auto-scroll the viewport to show the current match when navigating with
  `n`/`N`.

**Non-Goals:**
- No regex search - plain literal substring matching is sufficient for the
  stated use cases (usernames, keywords).
- No case-sensitive mode - case-insensitive is the default and only mode.
- Search does not extend beyond the Activity tab (not in Overview, Commits,
  Checks, Files Changed, or thread triage).
- No persistent search history across sessions or PRs.
- No incremental search-as-you-type - matches are found and highlighted only
  after pressing Enter in the search box.

## Decisions

### 1. Search state lives on `prview.Model`, scoped to Activity tab
Add a small unexported struct:
```go
type activitySearchState struct {
    active         bool
    term           string       // the search term (case-insensitive)
    matches        []matchPos   // positions of all matches in rendered content
    currentIndex   int          // index into matches array
    direction      searchDir    // forward or backward
}

type matchPos struct {
    activityIndex int // which RenderedActivity this match is in
    lineOffset    int // which line within that activity's RenderedString
    charOffset    int // character offset within that line
}

type searchDir int
const (
    searchForward searchDir = iota
    searchBackward
)
```
as a field on `prview.Model`, plus methods `StartSearch(direction searchDir)
tea.Cmd`, `ClearSearch()`, `IsSearching() bool`, `NextMatch()`, `PrevMatch()`,
`ExecuteSearch(term string)`.

When the user switches tabs (carousel changes), `prview.Model.Update` clears
the search state. When `syncSidebar()` populates a different PR's row, search
state is also cleared.

Alternatives considered: putting search state on `tui.Model`. Rejected for
the same reason thread triage lives on `prview.Model` - it's entirely about
what the PR sidebar shows and does, and `prview.Model` already owns that
responsibility.

### 2. Search input uses a new `cmpcontroller.ModeSearch` on the existing editor
The existing `m.editor` (`cmpcontroller.Controller`) field on `prview.Model`
is reused for search input, adding a new `ModeSearch` constant. When `/` or
`?` is pressed:
```go
func (m *Model) StartSearch(direction searchDir) tea.Cmd {
    m.activitySearch.direction = direction
    m.activitySearch.active = true
    return m.editor.Enter(cmpcontroller.EnterOptions{
        Mode:                             cmpcontroller.ModeSearch,
        Prompt:                           "/", // or "?" depending on direction
        Repo:                             m.repoRef(),
        EnterFetch:                       cmpcontroller.FetchNone,
        ConfirmDiscardOnCancel:           false,
        HideAutocompleteWhenContextEmpty: true,
        InitialValue:                     m.activitySearch.term, // restore previous term if any
    })
}
```
The input box is rendered at the bottom of the Activity view, similar to how
the comment editor is rendered in Overview. On `Enter`, the search term is
extracted via `m.editor.Value()` and `m.ExecuteSearch(term)` is called. On
`Esc`, `m.editor.Exit()` clears the input and `m.ClearSearch()` removes
highlights.

Alternatives considered:
- Creating a separate, simpler text input component just for search. Rejected
  - `cmpcontroller.Controller` already handles focus, input, Esc/Enter, and
  rendering; reusing it keeps the codebase consistent and avoids duplicating
  that logic.
- Using the existing `search.Model` component from
  `internal/tui/components/search/`. Rejected - that component is designed
  for the section-level search bars (filtering PR/issue lists) and has
  autocomplete integration for `repo:`, `author:`, etc. that doesn't apply to
  in-content text search. The mode-based `cmpcontroller` is a better fit.

### 3. Highlighting matches: styled wrapper around matched substrings
`ExecuteSearch(term)` walks through the array of `RenderedActivity` structs
(the same ones `renderActivity()` already built) and searches each activity's
`RenderedString` for occurrences of `term` (case-insensitive). For each
match, it records a `matchPos` with the activity index, line offset, and
character offset.

To highlight matches, `renderActivity()` is extended (or wrapped by a new
function) to apply a highlight style to matched regions. Since
`RenderedString` contains ANSI-styled markdown output from glamour, the
highlight needs to work with ANSI escape codes already present. Two
approaches:
1. Strip ANSI codes, find match positions in plain text, re-apply original
   styles plus highlight background.
2. Use lipgloss's `PlaceHorizontal` or a custom ANSI-aware substring wrapper
   to insert highlight styling around matched text without breaking existing
   styles.

Decision: approach 2 is simpler. For each line containing a match, split the
line at the match boundaries, wrap the matched substring in a lipgloss style
with a distinct background color (e.g., `theme.SelectedBackground`), and
reassemble. Lipgloss already handles nested styles correctly.

The current match (the one being navigated to) gets a different background
color (e.g., brighter or a different hue) to distinguish it from other
matches.

Alternatives considered: using ANSI `\e[7m` reverse video directly. Rejected
- lipgloss styles are more maintainable and theme-aware than raw escape
codes.

### 4. Auto-scrolling to the current match
`NextMatch()` and `PrevMatch()` increment/decrement `m.activitySearch.currentIndex`
(wrapping at the ends). After updating the index, they calculate the line
number of the current match within the entire rendered Activity content and
call a new sidebar method `ScrollToLine(line int)`.

`ScrollToLine(line)` sets the viewport's Y offset to center the given line in
the visible area (or as close as possible if near the start/end of content).

Implementation sketch for `ScrollToLine`:
```go
func (m *Model) ScrollToLine(line int) {
    contentHeight := lipgloss.Height(m.content)
    visibleHeight := m.height
    targetOffset := max(0, min(line - visibleHeight/2, contentHeight - visibleHeight))
    m.viewport.SetYOffset(targetOffset)
}
```

Alternatives considered: scrolling by a fixed number of lines up/down from
the current position. Rejected - jumping directly to the match line gives
more predictable, less disorienting navigation, especially when matches are
far apart.

### 5. Key routing: extend the existing details-view guard in `ui.go`, gated on Activity tab + not triaging
Add cases to the existing `if m.detailsViewState.active { switch { ... } }`
block, gated on the Activity tab being selected and thread triage not being
active:
```go
case m.ctx.View == config.PRsView &&
     m.prView.SelectedTab() == " Activity" &&
     !m.prView.IsTriagingThreads() &&
     key.Matches(msg, keys.PRKeys.SearchForward):
    return m, m.prView.StartSearch(searchForward)

case m.ctx.View == config.PRsView &&
     m.prView.SelectedTab() == " Activity" &&
     !m.prView.IsTriagingThreads() &&
     key.Matches(msg, keys.PRKeys.SearchBackward):
    return m, m.prView.StartSearch(searchBackward)

case m.prView.IsSearching() && key.Matches(msg, keys.PRKeys.NextSearchMatch):
    m.prView.NextMatch()
    m.syncSidebar() // refresh rendered content with updated current match highlight
    return m, nil

case m.prView.IsSearching() && key.Matches(msg, keys.PRKeys.PrevSearchMatch):
    m.prView.PrevMatch()
    m.syncSidebar()
    return m, nil

case m.prView.IsSearching() && key.Matches(msg, m.keys.Esc):
    m.prView.ClearSearch()
    m.syncSidebar()
    return m, nil
```

New `keys.PRKeys` bindings needed: `SearchForward` (`/`), `SearchBackward`
(`?`), `NextSearchMatch` (`n`), `PrevSearchMatch` (`N`). These are only
active when viewing the Activity tab in details mode and not triaging threads.

`n` and `N` are freely assignable here because thread triage already reuses
them for next/previous thread (and the `!IsTriagingThreads()` guard ensures
no conflict), and outside details mode these letters have no binding.

Alternatives considered: making search a mode that consumes *all* keys like
thread triage does. Rejected - search is lighter-weight: it only needs to
intercept `/`, `?`, `n`, `N`, and `Esc` when active; other keys (scroll,
help, tab switching) should continue to work normally.

### 6. Search status display: inline in the search input box prompt
The search input box's prompt changes based on state:
- Before Enter: `/ ` or `? ` (depending on direction)
- After Enter, with matches: `/ [3/7] ` (current match / total matches)
- After Enter, no matches: `/ [no matches] `

This is implemented by updating the `cmpcontroller.Controller`'s prompt text
after `ExecuteSearch` runs. The controller already supports dynamic prompts
(it's just a string field that gets re-rendered on each `View()` call).

Alternatives considered:
- Showing match status in the footer. Rejected - the footer is already busy
  with help keys and task feedback; keeping search status near the search
  input box is more discoverable.
- Showing match status in a separate status line below the input box.
  Rejected - adds UI complexity for marginal benefit; inline in the prompt is
  sufficient and more compact.

### 7. Help footer: add a `searchActive` flag in the `keys` package
Following the same pattern as `triageActive` and `notificationSubject`, add a
package-level `searchActive bool` in `keys`, set via
`keys.SetSearchActive(bool)`. `ui.go` calls
`keys.SetSearchActive(m.prView.IsSearching())` alongside the existing
`SetTriageActive` call.

`FullHelp()`'s `case config.PRsView:` branch reads `searchActive`. When true
and viewing the Activity tab, `additionalKeys` includes `SearchForward`,
`SearchBackward`, `NextSearchMatch`, `PrevSearchMatch`. When false, those
keys are not shown (since they have no effect outside an active search).

Alternatives considered: same as Decision 7 in the thread triage design.md -
this is the established pattern, and following it keeps help footer rendering
centralized.

### 8. Searching strips ANSI codes for matching, then highlights in the original styled content
To avoid having to parse ANSI escape sequences during search, the search
algorithm:
1. For each `RenderedString`, strip ANSI codes to get plain text (using a
   library like `github.com/acarl005/stripansi` or a simple regex).
2. Search the plain text for occurrences of the search term
   (case-insensitive).
3. Record the match positions (line, char offset) in the plain text.
4. When rendering, map those plain-text positions back to positions in the
   original styled string by counting characters while skipping ANSI codes,
   and insert highlight styles at those positions.

This avoids the complexity of searching inside ANSI-styled strings while
still preserving the original styling.

Alternatives considered: searching the ANSI-styled string directly. Rejected
- ANSI escape codes can split a word (e.g., `\e[1mhel\e[0mlo` for "hello"
with bold on "hel"), making substring matching fragile and error-prone.

## Risks / Trade-offs

- **[Risk]** Highlighting matches in already-rendered, ANSI-styled glamour
  output is non-trivial and could break existing styling if not done
  carefully. → Mitigation: strip ANSI for searching, then apply highlights on
  top of the original styled output using lipgloss, which already handles
  layered styles correctly. Test with markdown containing inline code, bold,
  links, etc. to ensure highlights don't corrupt formatting.
- **[Risk]** Searching large activity feeds (PRs with hundreds of comments)
  could be slow if done naively on every keystroke. → Mitigation: search is
  only triggered on Enter, not incrementally; once matches are found, `n`/`N`
  navigation is just array indexing. Renders are already cached in
  `RenderedActivity` structs, so searching doesn't require re-rendering
  markdown.
- **[Trade-off]** Search only matches plain text (stripped of markdown), so
  searching for `**bold**` won't match "**bold**" in the source markdown -
  it'll match "bold" in the rendered output. This is intentional and matches
  user expectations (searching for what they see, not the raw markdown), but
  means users can't search for markdown syntax itself.
- **[Trade-off]** No search history or saved searches. Each search session
  starts from scratch. This keeps the implementation simple and matches the
  use case (ad-hoc searching for a username or keyword), but power users who
  search for the same terms repeatedly might prefer persistent history. That
  can be added later if needed.
