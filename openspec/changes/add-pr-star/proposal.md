## Why

The PRs list has no way to flag a specific PR for later attention without
acting on it (commenting, checking out, snoozing it away). A PR you want to
keep visible and easy to spot - "come back to this one" - looks identical to
every other row. Gmail's star is the familiar model for this: a purely
personal, local marker that doesn't touch the underlying item, just makes it
stand out in the list.

## What Changes

- Add a `*` keybinding (`PRKeys.Star`) that toggles a "starred" flag on the
  currently selected PR in the PRs list. This is local-only state - it never
  calls the GitHub API and has no effect on the PR itself.
- Starred state persists across restarts in a local state file (new
  `StarStore` in `internal/data`, following the existing `SnoozeStore`/
  `DoneStore` pattern), keyed by repo + PR number. Unlike snooze, a star has
  no expiry - it stays until the user toggles it off.
- Add a new, dedicated column at the leftmost position of the PRs list
  (before the existing state/merge-status column) that renders a star glyph
  when the PR is starred, and is blank otherwise.
- The new column follows the same configuration pattern as other PR list
  columns: a `layout.prs.star` `ColumnConfig` (width/hidden), visible by
  default.
- Footer feedback (start/finished text, same mechanism `SnoozeFeedback`
  uses) confirms the toggle, since it's an instant, no-API-call action.

## Capabilities

### New Capabilities
- `pr-star`: a local, non-destructive "star" flag a user can toggle on any
  PR in the PRs list to visually flag it for attention, persisted locally
  and shown via a dedicated leftmost column.

## Impact

- `internal/data/starstore.go` (new): persisted set of starred PR keys,
  modeled on `internal/data/snoozestore.go` but without a wake-time/
  expiry - a plain toggleable set.
- `internal/tui/keys/prKeys.go`: new `Star` binding (default `*`) plus
  `rebindPRKeys`/`PRFullHelp` wiring, matching the existing `Snooze` entry.
- `internal/tui/components/prssection/`: new `star.go` (star-key helper,
  mirroring `snooze.go`'s `snoozeKey`), a key handler in `Update` that
  toggles the store and fires footer feedback, and a new leftmost column in
  `GetSectionColumns` (both compact and non-compact layouts).
- `internal/tui/components/prrow/prrow.go`: new `renderStar()`, added as
  the first entry in `ToTableRow`'s row slice (both layouts).
- `internal/tui/components/tasks/`: a `StarFeedback` helper alongside the
  existing `SnoozeFeedback`, for the footer's task-in-progress/finished UI.
- `internal/config/parser.go`: new `Star ColumnConfig` field on
  `PrsLayoutConfig`.
- Docs: add the new column and keybinding to
  `docs/src/content/docs/configuration/layout/pr.mdx` and
  `docs/src/content/docs/getting-started/keybindings/selected-pr.mdx`.
