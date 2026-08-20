## Why

The Activity tab in the PR details view shows a chronological feed of all
comments, review comments, and reviews. For PRs with many comments, finding
specific information - a particular user's feedback, a keyword, a phrase from
a discussion - requires manually scrolling through the entire feed. Users need
a way to quickly search and navigate to relevant comments without reading
everything sequentially.

## What Changes

- Add `/` keybinding to start a forward search (search from current position
  toward the end of the activity feed)
- Add `?` keybinding to start a backward search (search from current position
  toward the beginning of the activity feed)
- When `/` or `?` is pressed, show a search input box at the bottom of the
  sidebar (similar to vim's command line)
- As the user types, highlight all matches in the rendered activity content
- `Enter` in the search box moves to the first match and auto-scrolls the
  sidebar viewport to show it
- `n` moves to the next match (wrapping from last to first), auto-scrolling
  to show it
- `N` moves to the previous match (wrapping from first to last),
  auto-scrolling to show it
- Display search status (e.g., "match 3 of 7" or "no matches") in or near the
  search box
- `Esc` clears the search highlights and closes the search box, returning to
  normal activity view navigation
- Search is case-insensitive by default
- Search matches against the plain text content of comments (author names,
  comment bodies), not markdown formatting or timestamps
- Search state is scoped to the current PR - switching to a different PR or
  tab clears the search

## Capabilities

### New Capabilities
- `activity-view-search`: in-content search for the Activity tab that lets
  users quickly find and navigate to specific usernames, keywords, or phrases
  within a PR's comment history, with vim-style keybindings and visual
  feedback showing match positions.

### Modified Capabilities
(none - this is purely additive to the existing `pr-details-view` capability)

## Impact

- `internal/tui/components/prview/`: new search state (`searchTerm`,
  `searchMatches`, `currentMatchIndex`, `searchDirection`) and rendering
  logic to:
  - Track match positions within the rendered activity content
  - Highlight matching text (likely using lipgloss background/foreground
    styling to make matches stand out)
  - Render the search input box and match status
  - Auto-scroll the sidebar to the current match position
- `internal/tui/components/prview/activity.go`: modifications to
  `renderActivity` or a new wrapper to support highlighting search matches in
  the rendered output. May need to search the raw comment text and map
  positions to the rendered markdown to apply highlights correctly.
- `internal/tui/keys/prKeys.go`: new bindings for forward search (`/`),
  backward search (`?`), next match (`n`), and previous match (`N`). These
  are only active when viewing the Activity tab in details mode.
- `internal/tui/ui.go`: the existing details-view key routing adds cases for
  the new search keys, gated on being in the Activity tab. The search input
  box uses a pattern similar to the existing comment editor
  (`cmpcontroller.Controller`) but simpler since it's just text input with no
  autocomplete.
- `internal/tui/components/sidebar/`: the sidebar's viewport scrolling
  already supports `ScrollToTop`/`ScrollToBottom`/`ScrollUp`/`ScrollDown`;
  may need a `ScrollToLine(int)` or `ScrollToOffset(int)` method to jump
  directly to a match position.
- Search functionality is specific to the Activity tab and does not apply to
  other tabs (Overview, Commits, Checks, Files Changed) or to the thread
  triage workflow.
- No changes to config schema, data fetching, or the Issues/Notifications/
  Repo views.
