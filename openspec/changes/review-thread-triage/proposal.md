## Why

Some projects require every review thread on a PR to be resolved before it
can merge. Today there is no way to work through those threads from
gh-dash: the "Activity" tab in the PR details view flattens issue comments,
review-thread comments, and reviews into one chronological feed, with no
per-thread resolved/unresolved status, no way to reply to a specific
thread, and no way to resolve a thread at all. A user with this merge rule
has to leave the TUI and go to github.com to find and clear every open
thread before merging.

## What Changes

- Add a new "review thread triage" workflow, reachable only from the PR
  details view (`enter` on a row in the PRs section), via a new `T`
  keybinding.
- Entering the workflow fetches the PR's current review threads fresh
  (bypassing the existing enriched-PR-data cache) and builds a queue of its
  currently unresolved threads, ordered by file path then line number.
- For each thread in the queue, show: the file/line it's anchored to, the
  diff hunk snippet GitHub already attaches to the anchoring review
  comment, and the thread's full comment history - reusing the app's
  existing markdown rendering.
- `n` / `N` move to the next/previous thread in the queue without changing
  its resolved state (the "ignore, move on" action). The queue wraps
  around from the last thread back to the first.
- `c` opens the existing comment editor scoped to the current thread;
  submitting posts a reply on that thread (not a general PR comment).
- `r` resolves the current thread, removes it from the queue, and
  auto-advances to the next unresolved thread (or exits the workflow, back
  to the normal PR details view, if the queue is now empty).
- `esc` exits the workflow and returns to the normal PR details view
  (whichever tab was showing before triage started), leaving any threads
  already resolved or replied to as-is.
- Reply and resolve are independent actions - a user can reply without
  resolving, resolve without replying, or do both in either order, matching
  the three actions described in the request (respond, resolve, ignore).
- The reply and resolve actions are only offered when the thread's
  `viewerCanReply`/`viewerCanResolve` fields (from the GitHub API) are
  true; when false, the corresponding key is a no-op for that thread.

## Capabilities

### New Capabilities
- `pr-review-thread-triage`: a full-window, per-thread workflow reachable
  from the PR details view that lets a user page through a PR's unresolved
  review threads, see each thread's diff context and discussion, reply to
  it, resolve it, or skip it, so all threads can be cleared before merge
  without leaving the TUI.

### Modified Capabilities
(none - no existing capability specs cover PR review threads; the existing
`pr-list-columns` capability's unresolved-thread *count* is unaffected by
this change, though its number will naturally change as threads referenced
there get resolved through this workflow)

## Impact

- `internal/data/prapi.go`: `ReviewThreadsWithComments` (the enriched-PR
  query's thread struct) needs `IsResolved`, `ViewerCanReply`, and
  `ViewerCanResolve` added; `ReviewComment` needs `DiffHunk` added. A new
  exported function to re-fetch just a PR's review threads (rather than the
  full enriched PR) is needed for the "always refetch on entry" behavior.
- New GitHub mutations, invoked via `gh api graphql` (matching the existing
  `gh` CLI subprocess pattern used for every other write action - approve,
  comment, label, assign, etc. - rather than introducing direct use of the
  GraphQL client's `Mutate` method, which nothing in this codebase uses
  today): `addPullRequestReviewThreadReply` and `resolveReviewThread`.
- `internal/tui/components/tasks/`: new task functions for the two
  mutations above, following the existing `fireTask`/`GitHubTask` pattern
  (`CommentOnPR`, `ApprovePR`), including `UpdatePRMsg` fields so the PRs
  section's local PR data (and its unresolved-thread count) update
  immediately without waiting for a list refresh.
- `internal/tui/components/prview/`: new state and rendering for the triage
  workflow (current queue, current index, per-thread view), reusing the
  existing `editor` (`cmpcontroller.Controller`) for replies via a new mode
  and the existing markdown renderer for comment bodies and diff hunks.
- `internal/tui/keys/prKeys.go`: new bindings for entering triage (`T`),
  next/previous thread (`n`/`N`), and resolve (`r`); the existing `c`
  (Comment) binding is reused for thread replies while triaging.
- `internal/tui/ui.go`: the existing centralized details-view key-routing
  guard (added for the full-window details view) gains additional cases,
  gated on the PR view being in triage mode, so the change stays consistent
  with that guard's existing "one guard, not duplicated per view" approach;
  `esc` while triaging is layered to exit triage first, then fall through
  to the existing details-view `esc` handling on a second press.
- No changes to config schema or to the Issues/Notifications/Repo views.
