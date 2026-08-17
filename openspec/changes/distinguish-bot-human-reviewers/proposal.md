## Why

The PRs list's review-status column and the PR preview/details reviewer list
both merge human and bot (GitHub App) reviewers into a single signal. A PR
that's been approved only by a bot (e.g. a linter or dependency-update app)
looks identical to one approved by a human teammate, and a bot that merely
left a comment can make the whole PR look "awaiting review" even though
every human reviewer has already approved. Splitting human and bot review
activity apart lets a reviewer tell at a glance, from the list alone,
whether a PR is actually waiting on a person.

## What Changes

- **BREAKING**: Replace the PRs list's single `reviewStatus` column with two
  adjacent columns: one summarizing human reviewers' status, one summarizing
  bot (GitHub App) reviewers' status. Each cell keeps today's icon set
  (waiting/commented/approved/changes-requested); the two columns are told
  apart by distinct header icons (a person icon and a robot icon).
- Compute each column's status by grouping the PR's individual reviews by
  author type (bot vs. non-bot) and applying the same
  approved/changes-requested-beats-commented priority rule per group,
  instead of relying on GitHub's single, ungrouped `reviewDecision` field.
- In the PR preview pane and full-window details view, split the "Reviewers"
  list into a human group and a bot group (each on its own wrapped line(s),
  prefixed with the same person/robot icon), instead of one mixed list.
  A group that has no reviewers in it is omitted entirely.
- Fetch per-review author type (`__typename`: `User`/`Bot`/etc.) in both the
  PRs list query and the enriched single-PR query, which currently only
  fetch the review's author login.
- No new column or config key is introduced for the bot column's visibility
  toggle beyond what's noted in Impact below; humans/bots are visually
  distinguished by header icon and (in the preview) by group icon, not by
  separate text labels.

## Capabilities

### New Capabilities
- `pr-list-columns`: defines the set of columns available in the PRs
  section list view. No existing spec currently documents this column set
  (an earlier, unarchived change touched it but only documented the
  `numComments` column), so this change introduces the capability and
  documents the new human/bot review-status columns as part of it.
- `pr-reviewer-panel`: defines how the PR preview pane and full-window
  details view render the list of reviewers/review-requests for a PR
  (shared by both, since they use the same rendering code).

### Modified Capabilities
(none - both capabilities above are new; no existing capability specs exist
in this repo's `openspec/specs/` yet to modify)

## Impact

- `internal/data/prapi.go`:
  - `PullRequestData.Reviews` (list query) changes from `ReviewsNumber`
    (`{ totalCount }`) to a new struct that also fetches each review's
    `state` and `author { login __typename }`, needed to bucket reviews by
    author type.
  - `Review.Author` (enriched query, used by the preview/details view)
    gains a `__typename` field.
  - A new shared helper computes the approved/changes-requested/commented/
    waiting decision for a set of reviews, used by both the list column and
    the preview panel, replacing `PullRequestData.ReviewDecision` as the
    list column's data source.
- `internal/tui/components/prrow/prrow.go`: replace `renderReviewStatus`
  with two functions (human/bot), added to `ToTableRow` in place of the
  single review-status cell.
- `internal/tui/components/prssection/prssection.go`: replace the single
  `reviewStatus` column definition (both compact and non-compact layouts)
  with two adjacent column definitions.
- `internal/tui/components/prview/prview.go`: `renderRequestedReviewers`
  groups its output into a human section and a bot section instead of one
  combined, comma-separated list.
- `internal/config/parser.go`: `PrsLayoutConfig.ReviewStatus` is
  supplemented with a new `ReviewStatusBot` field so the two columns'
  width/visibility can be configured independently; the existing
  `reviewStatus` config key continues to control the human column.
- `internal/tui/constants/constants.go`: add a robot icon constant
  alongside the existing person icon, used as both the bot column's header
  and the bot group's prefix in the preview panel.
- Docs: update the "PR Review Status Column" section under
  `docs/src/content/docs/configuration/layout/pr.mdx` to describe the two
  columns and the new `reviewStatusBot` config key.
