## Context

See proposal.md - Why. Two independent rendering paths currently collapse
human and bot reviewers together, and both need the same new signal (a
review's author account type) that the current GraphQL queries don't fetch:

- **List view** (`internal/tui/components/prrow/prrow.go`,
  `renderReviewStatus`): reads `pr.Data.Primary.ReviewDecision` (GitHub's
  single, PR-wide computed decision) and `Reviews.TotalCount`. The list
  query's `PullRequestData.Reviews` field is typed `ReviewsNumber`
  (`{ totalCount }` only, from `internal/data/prapi.go`) - no per-review
  data at all, so there's nothing to group by author type today.
- **Preview/details view** (`internal/tui/components/prview/prview.go`,
  `renderRequestedReviewers`, shared by the sidebar preview and the
  full-window details view): reads the enriched query's
  `Reviews.Nodes []Review`, where `Review.Author` only has `Login` - no
  account-type field either.

`ReviewRequestNode` (pending review requests, both views) already
distinguishes bot from non-bot via its `... on Bot` GraphQL fragment
(`GetReviewerType() == "Bot"`), so no query change is needed for *pending*
reviewers - only for *completed* reviews.

## Goals / Non-Goals

**Goals:**
- Give both the list column and the preview panel a per-review "is this
  author a bot" signal, fetched once per review, reused by both.
- Compute a human/bot decision (approved / changes-requested / commented /
  waiting) using one shared, unit-testable rule, so the list and preview
  can't drift into inconsistent definitions of "approved" over time.
- Keep each list cell's rendering as narrow as today's single column (an
  icon, nothing else) - the human/bot distinction is carried by the column
  header and, in the preview, by a group-prefix icon, not by widening every
  cell.

**Non-Goals:**
- Matching GitHub's official `reviewDecision` field exactly. That field
  factors in branch protection (which reviewers/teams are *required*,
  CODEOWNERS enforcement, explicit review dismissals) that isn't visible
  from the plain `reviews` connection. This change computes its own,
  simpler approximation for both groups instead (see Decisions and Risks).
- A third "team" or "mixed identity" bucket - team review requests and
  suggested/code-owner reviewers (which are never bots) are bucketed with
  humans; see Decisions.
- Any change to how reviews are grouped/deduped *within* a single account
  (the existing "approved/changes-requested beats commented" priority rule
  is reused as-is, just applied per group instead of globally).

## Decisions

- **Fetch `__typename` on review authors, not a heuristic**: add
  `Typename graphql.String \`graphql:"__typename"\`` to a review's author
  struct (mirroring the existing pattern in
  `LastCommitWithStatusChecks.Contexts.Nodes` for
  `CheckRun`/`StatusContext`). A review's author type is `Bot` for GitHub
  Apps, `User` for people; other interface implementers (`Mannequin`,
  `Organization`) are treated as non-bot (see below).
  - Alternative considered: pattern-match `login` against `[bot]`/`-bot`
    suffixes. Rejected - it's a naming convention, not a guarantee, and
    the schema already gives us an authoritative type via `__typename`
    (the same approach `ReviewRequestNode` already uses for review
    *requests*).
- **New, lighter-weight struct for the list query** (`ReviewsWithAuthorType`
  or similar): the list's `PullRequestData.Reviews` field changes from
  `ReviewsNumber` to a new struct fetching `totalCount` and
  `nodes { state author { login __typename } }` - not the heavier existing
  `Reviews`/`Review` struct used by the enriched query, which also carries
  `body` and `updatedAt` that the list view doesn't need. This follows the
  same precedent as `ReviewThreadsWithResolution`
  (`add-unresolved-threads-column`): a new, narrower struct for the list
  context instead of reusing a heavier one designed for the single-PR view.
  The enriched query's existing `Review` struct just gains the `Typename`
  field in place.
- **One shared decision function, two call sites**: add a small,
  independent-of-either-struct type (e.g. `data.ReviewSummary{ Login,
  Typename, State string }`) and a pure function
  `data.ComputeReviewStatus(reviews []ReviewSummary) string` returning
  `"APPROVED" | "CHANGES_REQUESTED" | "COMMENTED" | ""`. Both
  `renderReviewStatus`'s replacement (list) and
  `renderRequestedReviewers` (preview) convert their respective review
  slices into `[]ReviewSummary`, partition by `Typename == "Bot"`, and call
  this function once per partition. This replaces
  `PullRequestData.ReviewDecision` as the list column's data source
  entirely (that field can't be split by author type) and reuses the exact
  per-author dedup rule (don't let a later `COMMENTED` review from the same
  author downgrade an earlier `APPROVED`/`CHANGES_REQUESTED`) that
  `renderRequestedReviewers` already implements today, without duplicating
  it.
  - Alternative considered: keep using `ReviewDecision` for the human
    column (on the theory that required reviewers are almost always
    human) and only compute the bot column from raw states. Rejected for
    consistency and simplicity - GitHub Apps with write access *can*
    contribute to `reviewDecision`, so the "humans-only" assumption isn't
    reliably true, and having two different computation strategies for
    what's conceptually the same column is harder to reason about and
    test than one function used symmetrically.
- **Non-bot interface types (`Mannequin`, `Organization`) and teams bucket
  as human**: `ReviewSummary.Typename == "Bot"` is the only bot check;
  everything else (including team review requests, which have no single
  author type) buckets as human/non-bot. Suggested reviewers
  (`SuggestedReviewer`, sourced from CODEOWNERS) also bucket as human,
  since GitHub Apps as CODEOWNERS are not a case this app currently
  distinguishes and are rare in practice.
- **Config key**: add `ReviewStatusBot ColumnConfig` to `PrsLayoutConfig`
  (`internal/config/parser.go`) alongside the existing `ReviewStatus`
  field, which keeps controlling the human column. Both columns default to
  the same width/visibility the single column had.
  - Alternative considered: one config key controlling both columns'
    shared visibility (e.g. hide/show the pair together). Rejected - a
    user who only cares about human review status (the common case, per
    proposal.md) should be able to hide the bot column alone without
    losing the human one.
- **New `RobotIcon` constant**: add one Nerd Font glyph to
  `internal/tui/constants/constants.go` alongside the existing
  `PersonIcon`, used both as the bot column's header and the bot group's
  prefix in the preview panel. Exact glyph choice is a visual detail to
  confirm by rendering it in a terminal during implementation, not a
  behavioral one.
- **Preview grouping reuses the existing wrap algorithm per group**:
  `renderRequestedReviewers`'s existing width-aware line-wrapping (used
  today for the single combined list) is run once per non-empty group
  instead of once overall, with each group's rows prefixed by its group
  icon. This is a smaller change than introducing a second wrapping
  strategy, and preserves the exact wrapping behavior scenarios already
  covered by `reviewers_test.go`.

## Risks / Trade-offs

- [Risk] The computed per-group decision won't always match what GitHub's
  UI shows via `reviewDecision` for the *combined* PR (e.g. a PR
  GitHub considers "changes requested" only because of a *dismissed*
  review this app doesn't know is dismissed, since dismissed reviews still
  appear in the `reviews` connection with state `DISMISSED` and are simply
  excluded from this app's priority rule like any other non-actionable
  state) → Mitigation: this is an accepted, documented approximation (see
  Non-Goals); it mirrors the existing single-column behavior, which was
  already an approximation in different ways (it used `ReviewDecision`
  when present but fell back to a raw `Reviews.TotalCount > 0` check
  otherwise).
- [Risk] Widening the list query's `reviews` field from `{ totalCount }` to
  `{ totalCount nodes { state author { login __typename } } }` increases
  per-PR response size in the list query → Mitigation: bounded by the same
  `last: 100` connection size already used elsewhere for list-scoped data
  (`ReviewThreadsWithResolution`); each node is three scalar fields, no
  nested connections.
- [Risk] Users with existing `layout.prs.reviewStatus.width` configuration
  tuned for a single column may find two narrower columns look cramped
  side by side → Mitigation: each column keeps the same default width the
  single column had; users who preferred a single combined view can hide
  the bot column via the new `reviewStatusBot.hidden` key.
