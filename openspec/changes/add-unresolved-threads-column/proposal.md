## Why

PRs with unresolved review discussions are easy to miss in the PR list: the
existing `numComments` column mixes issue comments and review-thread comments
into one total count, and doesn't distinguish resolved from unresolved
threads. A reviewer scanning the list has no quick way to tell which PRs
still have open conversations blocking merge without opening each one.

## What Changes

- **BREAKING**: Change what the existing `numComments` column shows: instead
  of the sum of issue comments and total review threads, it now shows only
  the number of *unresolved* review threads on the PR. PRs that previously
  showed a nonzero count purely from resolved threads or issue comments will
  now show a blank cell.
- Render nothing (blank cell) when a PR has zero unresolved threads, so the
  column stays visually quiet for the common case.
- Fetch per-thread `isResolved` status in the PR list GraphQL query (the
  current list query only fetches `reviewThreads { totalCount }`) and count
  threads where `isResolved == false`.
- No new column or config key is introduced - the column keeps its existing
  `numComments` config key, header icon, and position; its existing `width`/
  `hidden` configuration continues to apply unchanged.

## Capabilities

### New Capabilities
- `pr-list-columns`: defines the set of columns available in the PRs list
  view. No existing spec currently documents this column set, so this
  change introduces the capability and documents the `numComments` column's
  new (unresolved-review-threads) behavior as part of it.

### Modified Capabilities
(none - `pr-list-columns` is a new capability; no existing spec currently
documents the PR list column set)

## Impact

- `internal/data/prapi.go`: `PullRequestData.ReviewThreads` query needs
  per-node `isResolved` data (currently `TotalCount` only). The now-unused
  `PullRequestData.Comments` field (issue comment count) can be dropped from
  the query, since it was only read by the column being replaced.
- `internal/tui/components/prrow/prrow.go`: modify `renderNumComments` in
  place to compute and render the unresolved review thread count instead of
  the comments+threads total.
- `internal/tui/components/prssection/prssection.go`: no change - the
  existing `numComments` column definition (both compact and non-compact
  layouts) is reused unchanged.
- `internal/config/parser.go`: no change - the existing `NumComments
  ColumnConfig` field continues to control this column's width/visibility.
- Docs: update (not add) the "PR Number of Comments Column" section under
  `docs/src/content/docs/configuration/layout/pr.mdx` to describe the new
  behavior.
