## 1. Data layer

- [x] 1.1 In `internal/data/prapi.go`, add a new struct (e.g.
      `ReviewThreadsWithResolution`) with `TotalCount int` and
      `Nodes []struct{ IsResolved bool }`.
- [x] 1.2 Change `PullRequestData.ReviewThreads` to use the new struct with
      `graphql:"reviewThreads(last: 100)"`, per design.md's "Query shape"
      decision.
- [x] 1.3 Add a helper (e.g. `PullRequestData.UnresolvedThreadsCount() int`)
      that counts nodes with `IsResolved == false`.
- [x] 1.4 Remove the now-unused `Comments Comments \`graphql:"comments"\``
      field from `PullRequestData` (confirm via grep that nothing else in
      the list-row context reads `PullRequestData.Comments` before
      removing).

## 2. Rendering

- [x] 2.1 Modify `renderNumComments()` in
      `internal/tui/components/prrow/prrow.go` in place: drop the
      `Comments.TotalCount + ReviewThreads.TotalCount` sum, replace it with
      `pr.Data.Primary.UnresolvedThreadsCount()`, and return `""` when that
      count is `0` (in addition to the existing `Data.Primary == nil`
      blank case). Keep the function name, its call sites in `ToTableRow`,
      and the column's header/icon/position unchanged.

## 3. Tests

- [x] 3.1 Update `prrow_test.go` cases for `renderNumComments`/`ToTableRow`:
      blank for zero unresolved threads, shows the count for one or more
      unresolved threads, blank when `Data.Primary` is nil, and confirm a
      PR with only resolved threads or only issue comments (no unresolved
      threads) now renders blank instead of a nonzero count.
- [x] 3.2 Add a `PullRequestData.UnresolvedThreadsCount()` unit test in
      `internal/data` covering: no threads, all resolved, some unresolved,
      all unresolved.

## 4. Docs

- [x] 4.1 Update the existing "PR Number of Comments Column" section in
      `docs/src/content/docs/configuration/layout/pr.mdx` in place to
      describe the new behavior (unresolved review thread count, blank at
      zero) instead of the old comments+threads total. Config key, default
      width, and heading icon stay the same, so no changes are needed to
      the "By default, PR views display..." list or the link reference
      list.
