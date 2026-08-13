## Context

See proposal.md - Why. Two existing local-state stores in `internal/data`
are the direct precedent:

- `SnoozeStore` (`snoozestore.go`): `map[string]time.Time` (id → wake-at),
  entries expire and are pruned once `time.Now()` passes the stored time.
- `DoneStore` (`donestore.go`): `map[string]time.Time` (id → the item's
  `updatedAt` when marked done), entries are invalidated when the item's
  current `updatedAt` moves past the stored value, and pruned after 90 days
  regardless.

Both persist via the same pattern: an in-memory map guarded by a
`sync.RWMutex`, atomic save (temp file + rename) to a path from
`getStateFilePath(filename)`, loaded once at startup, singleton accessor.
A star has no such invalidation condition - it's a plain boolean the user
sets and clears explicitly - so it doesn't need either store's time-based
logic.

The PRs list row-building path (`prssection.Model.visiblePRs`/`BuildRows`,
`prrow.PullRequest.ToTableRow`) and column definitions
(`prssection.GetSectionColumns`, `config.PrsLayoutConfig`) are the existing
machinery `numComments` and the snooze filter already extend; this change
follows the same shape rather than introducing a new one.

## Goals / Non-Goals

**Goals:**
- Reuse the existing store/singleton pattern (`SnoozeStore`/`DoneStore`)
  for the new `StarStore`, minus the time-based invalidation neither store
  needs here.
- Reuse the existing column configuration pattern (`ColumnConfig`,
  `PrsLayoutConfig`, `GetSectionColumns`) unchanged in shape, just adding
  one new entry.
- Keep starring a synchronous, local, no-API-call action, consistent with
  how `SnoozeFeedback` already models "instant" actions in the footer task
  UI.

**Non-Goals:**
- Reordering the list to float starred PRs to the top - this change is
  purely a visual marker, matching the proposal's explicit scope.
- Extending starring to the Issues or Notifications sections - the request
  scopes this to PRs; the same pattern could be replicated per-section
  later the same way snooze already is, but that's a separate change.
- Filtering/searching by starred state - not requested; the column is
  purely for visual scanning.
- Any interaction between starring and snoozing (e.g. preventing a starred
  PR from being snoozed) - they're independent, orthogonal flags; a PR can
  be both starred and snoozed at once, and a snoozed+starred PR is still
  hidden by the existing snooze filter (starring doesn't override it).

## Decisions

- **New `StarStore`, not a boolean extension of `SnoozeStore`/`DoneStore`**:
  a star has no wake-time and no "resurface on new activity" rule, so
  bolting it onto either existing store would mean carrying an unused
  `time.Time` and dead invalidation branches. `StarStore` is a minimal
  `map[string]struct{}` (or `map[string]bool`, decided during
  implementation for the cleanest JSON shape) persisted as a JSON array of
  IDs, with `Star(id)`, `Unstar(id)`, `IsStarred(id) bool`, and
  `Toggle(id) bool` (returns the resulting state, so the caller can drive
  footer feedback text without a second lookup). Persisted to its own file
  (e.g. `starred.json`) via the same atomic-save helper shape as the other
  two stores - no expiry, no pruning.
- **Key format matches `snoozeKey`**: `fmt.Sprintf("pr:%s#%d", ...)` in a
  new `prssection/star.go`, mirroring `prssection/snooze.go`'s
  `snoozeKey`. Same shape, different store/namespace - a PR's star key and
  snooze key happen to look identical today, but each store is keyed and
  persisted independently, so nothing but the format is shared.
- **New leftmost column, not folded into the existing state column**: the
  proposal calls for the star at the "farmost left" of the row. The
  existing first column (`renderState`, width 3) already renders a single
  merge-state glyph (open/closed/merged/draft/queued) and has no spare
  width for a second glyph without cramping both. A new, separate,
  narrow (`ColumnConfig`-backed, default width ~2) column placed before it
  keeps each column single-purpose, consistent with how `numComments` and
  review-status already get their own dedicated columns rather than
  overloading `renderState`.
  - Alternative considered: prefix the star glyph onto the title text
    (`renderExtendedTitle`/`renderTitle`). Rejected - the proposal asks for
    a farthest-left-or-right position "to draw attention", and a column
    that's blank/present is easier to scan down a list than text prefixed
    onto a variable-width, already-busy title cell.
- **`renderStar()` blank-when-unset, matching existing blank-cell
  convention**: mirrors `renderNumComments`/`renderLabels`/`renderAssignees`,
  which all return `""` for the "nothing to show" case rather than a
  placeholder like `-`. Keeps the common (unstarred) case visually quiet.
- **Synchronous toggle + `StarFeedback`, no confirmation prompt**: unlike
  `snooze` (which opens `PromptConfirmationBox` to pick a preset duration),
  starring has only one parameter (on/off) and is trivially reversible, so
  it toggles immediately on keypress - closer to `ToggleSmartFiltering`
  than to `snooze`/`close`/`merge`'s confirm-or-input flows. A new
  `tasks.StarFeedback(ctx, section, key, itemDescription, starred bool)`
  (modeled on `tasks.SnoozeFeedback`) drives the footer's instant start/
  finished text, with wording depending on the resulting state ("Starred
  PR #123" / "Unstarred PR #123").
- **Default keybinding `*`, new `PRKeys.Star`**: every unused single letter
  already spoken for by `PRKeys` plus the universal `KeyMap` is consumed by
  the app's ~22 existing single-letter bindings (see
  `memory/architecture.md`'s keymap letter budget). `*` is unbound
  anywhere in the app today, is visually evocative of a star, and needs no
  modifier - added as `Star key.Binding` in `PRKeyMap`, wired into
  `rebindPRKeys`'s builtin-name switch (`"star"`) and `PRFullHelp()`,
  exactly like the existing `Snooze` entry.

## Risks / Trade-offs

- [Risk] A new required column shifts every existing column's position by
  one in both the compact and non-compact `GetSectionColumns` layouts,
  which touches more rendering code (and tests) than a pure behavior change
  would → Mitigation: columns are addressed by name in config
  (`layout.prs.<name>`), not index, so no user config breaks; only the
  Go-level column slice literals and their paired `ToTableRow` slices need
  updating in lockstep, the same mechanical change `numComments`'s
  original addition already required.
- [Risk] Star state can silently accumulate forever with no cleanup (no
  expiry, no invalidation) → Mitigation: this is the explicit intent (a
  Gmail-style star persists until manually unstarred); the store is a
  small set of short string keys, negligible file size even at hundreds of
  entries.
- [Risk] A starred PR that gets merged/closed, or is filtered out of the
  section's search query entirely, disappears from view with its star
  still recorded in the store, potentially orphaning entries over time →
  Mitigation: same accepted trade-off `SnoozeStore`/`DoneStore` already
  make for their own entries; no automatic cleanup is introduced here
  either. A future change could add pruning against closed/merged PRs if
  it proves to matter in practice.
