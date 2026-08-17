## 1. Data layer

- [x] 1.1 In `internal/data/prapi.go`, add `Typename graphql.String
      \`graphql:"__typename"\`` to the enriched query's `Review.Author`
      struct.
- [x] 1.2 Add a new, lighter-weight struct (e.g. `ReviewsWithAuthorType`)
      with `TotalCount int` and `Nodes []struct{ State string; Author
      struct{ Login string; Typename string \`graphql:"__typename"\` } }`,
      and change `PullRequestData.Reviews` to use it with
      `graphql:"reviews(last: 100)"` (replacing `ReviewsNumber`).
- [x] 1.3 Add a small shared type, e.g. `ReviewSummary{ Login, Typename,
      State string }`, plus conversion helpers from both the enriched
      `[]Review` and the new list `[]ReviewsWithAuthorType` node type into
      `[]ReviewSummary`.
- [x] 1.4 Add `func ComputeReviewStatus(reviews []ReviewSummary) string`
      returning `"APPROVED"`, `"CHANGES_REQUESTED"`, `"COMMENTED"`, or `""`
      (waiting), implementing the per-author dedup + priority rule
      currently inlined in `renderRequestedReviewers` (don't let a later
      `COMMENTED` from the same author downgrade an earlier
      `APPROVED`/`CHANGES_REQUESTED`), per design.md's "One shared decision
      function" decision.
- [x] 1.5 Add a helper, e.g. `func PartitionByBotAuthor(reviews
      []ReviewSummary) (human, bot []ReviewSummary)`, splitting on
      `Typename == "Bot"`.
- [x] 1.6 Remove the now-unused `PullRequestData.ReviewDecision` field's
      read sites in the list rendering path (keep the field itself only if
      something else still reads it - confirm via grep before removing the
      struct field).

## 2. List view rendering

- [x] 2.1 In `internal/tui/components/prrow/prrow.go`, replace
      `renderReviewStatus` with two functions (e.g.
      `renderReviewStatusHuman`/`renderReviewStatusBot`), each: partition
      `pr.Data.Primary.Reviews.Nodes` (converted to `[]ReviewSummary`) via
      `PartitionByBotAuthor`, call `ComputeReviewStatus` on its own half,
      and render the existing icon/color rules
      (approved/changes-requested/commented/waiting) from that result
      instead of from `ReviewDecision`.
- [x] 2.2 Update both `ToTableRow` cases (compact and non-compact) to call
      the two new functions in place of the single `renderReviewStatus`
      call, in that order (human column before bot column).

## 3. List column definitions and config

- [x] 3.1 In `internal/config/parser.go`, add `ReviewStatusBot
      ColumnConfig \`yaml:"reviewStatusBot,omitempty"\`` to
      `PrsLayoutConfig`.
- [x] 3.2 In `internal/tui/components/prssection/prssection.go`,
      merge a `reviewStatusBotLayout` config (mirroring
      `reviewStatusLayout`) and replace the single `reviewStatus` column
      definition (both compact and non-compact layouts) with two adjacent
      column definitions: the existing one (human) unchanged in
      position/width, plus a new one for bot status using
      `reviewStatusBotLayout.Hidden` and the new `RobotIcon` constant as
      its title.

## 4. Constants

- [x] 4.1 In `internal/tui/constants/constants.go`, add a `RobotIcon`
      constant (Nerd Font glyph) alongside `PersonIcon`; verify it renders
      distinctly from `PersonIcon` in a terminal.

## 5. Preview / details panel rendering

- [x] 5.1 In `internal/tui/components/prview/prview.go`, refactor
      `renderRequestedReviewers` to build its `reviewerItems` per group
      (human, bot) instead of one combined slice - reuse the existing
      per-item construction (state icon, code-owner icon, team/user
      formatting) unchanged, just routed into two slices based on each
      entry's account type (`ReviewRequestNode.GetReviewerType() ==
      "Bot"` for pending requests; `Typename == "Bot"` for completed
      reviews; suggested reviewers always go to the human slice, per
      design.md).
- [x] 5.2 Run the existing width-aware wrapping logic once per non-empty
      group, prefix each group's wrapped rows with its group icon
      (`PersonIcon`/`RobotIcon`), and join non-empty groups vertically
      under the existing "Reviewers" header. Omit a group entirely when it
      has zero entries.

## 6. Tests

- [x] 6.1 Add `data.ComputeReviewStatus` unit tests: empty input, single
      approval, single changes-requested, single comment,
      approval-then-comment (stays approved), changes-requested-then-comment
      (stays changes-requested), multiple authors with mixed states.
- [x] 6.2 Add `data.PartitionByBotAuthor` unit tests: all human, all bot,
      mixed, empty.
- [x] 6.3 Add `prrow_test.go` cases for `renderReviewStatusHuman`/
      `renderReviewStatusBot` covering: human-only reviews, bot-only
      reviews, mixed reviews, no reviews, `Data.Primary == nil`.
- [x] 6.4 Update `reviewers_test.go` (`TestRenderRequestedReviewers` and
      the wrapping test) to cover: bot reviewer request, bot completed
      review, a mix of human and bot reviewers (asserting both group icons
      appear and reviewers land in the expected group), an all-bot PR
      (human group omitted), and an all-human PR (bot group omitted,
      current behavior preserved).

## 7. Docs

- [x] 7.1 Update the "PR Review Status Column" section in
      `docs/src/content/docs/configuration/layout/pr.mdx` to describe the
      two columns (human and bot), their independent `reviewStatus`/
      `reviewStatusBot` config keys, and their header icons.
