## Context

See proposal.md - Why. The PRs list query (`PullRequestData` in
`internal/data/prapi.go`) currently fetches
`ReviewThreads ReviewThreads \`graphql:"reviewThreads"\`` where `ReviewThreads`
is just `{ TotalCount int }` - no per-thread data. GitHub's GraphQL API has
no aggregate "unresolved thread count" field on `PullRequest`; the only way
to get it is to fetch the `reviewThreads` connection's nodes and inspect each
one's `isResolved` field, same pattern already used for
`EnrichedPullRequestData.ReviewThreads` (`reviewThreads(last: 50)`, used for
the PR detail/diff view).

`PullRequestData.Comments` (issue comment count) and `renderNumComments`'s
`Comments.TotalCount + ReviewThreads.TotalCount` sum are the only places
that field is read in the list-row context; `EnrichedPullRequestData`'s
separate `Comments` field (used by the PR detail view) is untouched by this
change.

## Goals / Non-Goals

**Goals:**
- Compute an accurate-in-the-common-case unresolved review thread count for
  each PR row in the list view, at acceptable query cost.
- Reuse the existing `numComments` column and its configuration as-is - no
  new column, config key, or icon.

**Non-Goals:**
- Exact correctness for PRs with pathologically large numbers of review
  threads (see Risks below) - this mirrors an existing, accepted limitation
  of the detail-view query.
- Any change to the Issues or Notifications list views (review threads are a
  PR-only concept; the Issues section has its own, separate
  `renderNumComments` that is unaffected).
- A dedicated action to jump to/resolve threads from the list view - this
  change is display-only.
- Preserving a way to see total comment count in the list view - that signal
  is intentionally dropped per proposal.md (marked **BREAKING**).

## Decisions

- **Query shape**: add `reviewThreads(last: 100) { totalCount nodes { isResolved } }`
  to `PullRequestData` (a new struct, not reusing `ReviewThreadsWithComments`,
  since the list view doesn't need thread comments - just resolution state).
  Compute the unresolved count client-side as the number of fetched nodes
  with `IsResolved == false`.
  - Alternative considered: request `reviewThreads` twice with a
    `(states: ...)`-style filter to get resolved/unresolved counts directly.
    Rejected - the GitHub GraphQL API's `reviewThreads` connection has no
    such filter argument; only pull-request-level search filters exist
    (e.g. `review:` in search qualifiers), not a per-connection resolved
    count.
  - `last: 100` mirrors the existing `last: 50` used for
    `ReviewThreadsWithComments` in the enriched (single-PR) query, sized up
    since these nodes are much cheaper (one boolean field vs. nested
    comments).
- **Rendering, in place**: modify `renderNumComments` in
  `internal/tui/components/prrow/prrow.go` directly - drop the
  `Comments.TotalCount + ReviewThreads.TotalCount` sum, replace it with the
  unresolved-thread count, and return `""` when that count is zero
  (matching the existing blank-when-empty pattern used by
  `renderLabels`/`renderAssignees`). No new render function, column, icon,
  or `ColumnConfig` field is introduced - the column keeps its existing
  identity (key, header, position) and its behavior changes underneath it.
- **Drop unused query field**: remove `PullRequestData.Comments` (the issue
  comment count) from the list query, since nothing in the list-row context
  reads it once `renderNumComments` no longer sums it in.

## Risks / Trade-offs

- [Risk] A PR with more than 100 review threads will undercount unresolved
  threads (only the last 100 fetched are inspected) → Mitigation: this
  matches the existing accepted limitation of the enriched PR query
  (`last: 50` there); 100+ review threads on a single PR is rare enough that
  a perfect count isn't worth a more expensive query strategy. Documented as
  a known limitation, not a bug.
- [Risk] Widening the list query (currently fetching many PRs per search
  page) with a nested 100-node connection per PR increases response size and
  GraphQL query cost → Mitigation: nodes only carry a single boolean field
  each (`isResolved`), unlike the enriched query's threads-with-comments; the
  incremental cost per PR is small relative to existing nested connections
  already fetched per row (e.g. `Commits`, `Labels`, `Assignees`).
- [Risk] Users who relied on `numComments` as a proxy for total discussion
  volume (issue comments + all review threads) lose that signal, since the
  column now only reflects unresolved review threads → Mitigation: this is
  an intentional, deliberate trade-off (see proposal.md, marked
  **BREAKING**), and the column's docs are updated in place to describe the
  new meaning so the change is discoverable, not silent.
