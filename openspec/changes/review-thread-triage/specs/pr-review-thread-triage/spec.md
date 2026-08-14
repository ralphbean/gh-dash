## Purpose

Lets a user work through a pull request's unresolved review threads one at
a time from the PR details view - reading each thread's diff context and
discussion, replying, resolving, or skipping it - so every thread can be
cleared before merge without leaving the TUI.

## ADDED Requirements

### Requirement: Entering the review thread triage workflow
The system SHALL let the user enter the review thread triage workflow only
while the full-window details view is active for a pull request selected
from the PRs section, by pressing a dedicated key.

#### Scenario: Enter triage from a PR's details view
- **WHEN** the user is in the full-window details view for a pull request
  opened from the PRs section and presses the triage key
- **THEN** the system SHALL fetch the pull request's current review threads
  and enter the triage workflow, showing the first unresolved thread (if
  any)

#### Scenario: Triage key is a no-op outside the details view
- **WHEN** the user presses the triage key while viewing the PRs list (not
  in the details view), or while in the details view for an issue, branch,
  or notification
- **THEN** the system SHALL take no action

#### Scenario: No unresolved threads
- **WHEN** the user enters the triage workflow for a pull request that has
  zero unresolved review threads
- **THEN** the system SHALL show an empty state indicating there is nothing
  to triage, and SHALL remain in the triage workflow until the user presses
  the exit key

### Requirement: Triage workflow fetches current thread state
Entering the triage workflow SHALL fetch the pull request's review threads
directly from GitHub rather than reusing any previously cached copy, so the
queue reflects threads added, replied to, or resolved by others since the
details view was last loaded.

#### Scenario: Refetch on entry
- **WHEN** the user enters the triage workflow for a pull request whose
  details were already loaded and cached earlier in the session
- **THEN** the system SHALL request the pull request's review threads
  again before building the queue of unresolved threads

### Requirement: Viewing a thread
While in the triage workflow, the system SHALL display the current
thread's file path and line number, the diff hunk associated with the
thread's anchoring comment, and the thread's full comment history,
rendered the same way comment bodies are rendered elsewhere in the app.

#### Scenario: Thread content is shown
- **WHEN** the triage workflow displays a thread
- **THEN** the system SHALL show that thread's file/line, its diff hunk
  snippet, and every comment in the thread in order

#### Scenario: Scrolling a long thread
- **WHEN** a thread's rendered content is taller than the available screen
  height
- **THEN** the system SHALL let the user scroll the content with the same
  line-scroll and jump-to-top/bottom keys used elsewhere in the details
  view, without changing which thread is current

#### Scenario: Thread's diff hunk is outdated
- **WHEN** the triage workflow displays a thread whose anchoring diff hunk
  is outdated relative to the pull request's current contents
- **THEN** the system SHALL show a prominent indicator that the hunk is
  outdated, visually distinct from the thread's normal file/line header

### Requirement: Moving between threads without changing resolution state
The system SHALL let the user move to the next or previous thread in the
queue without resolving, replying to, or otherwise altering the current
thread.

#### Scenario: Move to next thread
- **WHEN** the user presses the next-thread key
- **THEN** the system SHALL display the next thread in the queue, leaving
  the current thread's resolved state and comments unchanged

#### Scenario: Move to previous thread
- **WHEN** the user presses the previous-thread key
- **THEN** the system SHALL display the previous thread in the queue,
  leaving the current thread's resolved state and comments unchanged

#### Scenario: Wrapping past the ends of the queue
- **WHEN** the user presses the next-thread key while viewing the last
  thread in the queue, or the previous-thread key while viewing the first
- **THEN** the system SHALL wrap around to the first, respectively last,
  thread still in the queue

### Requirement: Replying to a thread
The system SHALL let the user reply to the current thread when the thread
allows it, posting the reply to that specific thread rather than as a
general pull request comment.

#### Scenario: Reply is allowed
- **WHEN** the user presses the reply key on a thread the user is allowed
  to reply to, types a message, and submits it
- **THEN** the system SHALL post the message as a reply on that thread and
  SHALL NOT resolve the thread or change which thread is current

#### Scenario: Reply is not allowed
- **WHEN** the user presses the reply key on a thread the user is not
  allowed to reply to
- **THEN** the system SHALL take no action

#### Scenario: Cancelling a reply
- **WHEN** the user opens the reply editor for a thread and cancels before
  submitting
- **THEN** the system SHALL discard the draft and remain on the same
  thread with its resolved state and comments unchanged

### Requirement: Resolving a thread
The system SHALL let the user resolve the current thread when the thread
allows it, removing it from the queue and advancing to the next remaining
unresolved thread.

#### Scenario: Resolve is allowed
- **WHEN** the user presses the resolve key on a thread the user is
  allowed to resolve
- **THEN** the system SHALL mark that thread resolved, remove it from the
  queue, and display the next remaining thread in the queue

#### Scenario: Resolve is not allowed
- **WHEN** the user presses the resolve key on a thread the user is not
  allowed to resolve
- **THEN** the system SHALL take no action

#### Scenario: Resolving the last thread in the queue
- **WHEN** the user resolves a thread and no unresolved threads remain in
  the queue
- **THEN** the system SHALL exit the triage workflow and return to the
  pull request's normal details view

#### Scenario: List view unresolved-thread count reflects resolution
- **WHEN** the user resolves a thread while triaging and later returns to
  the PRs list view
- **THEN** the list view's unresolved review-thread count for that pull
  request SHALL reflect the resolution without requiring a manual refresh

### Requirement: Exiting the triage workflow
Pressing the exit key while in the triage workflow SHALL return the user to
the pull request's normal details view, on the same tab that was showing
before the triage workflow was entered, without altering any thread beyond
the actions already taken during the session.

#### Scenario: Exit mid-queue
- **WHEN** the user presses the exit key while threads remain unresolved in
  the queue
- **THEN** the system SHALL leave those threads unresolved and return to
  the pull request's normal details view

#### Scenario: Exiting the details view exits triage first
- **WHEN** the user presses the exit key while in the triage workflow
- **THEN** the system SHALL exit only the triage workflow on that key
  press, requiring a second press of the same key (once back in the normal
  details view) to exit the details view itself

### Requirement: Item-level pull request actions are unavailable while triaging
While the triage workflow is active, keybindings for actions that are not
part of the triage workflow (for example: approve, assign, label, merge,
close, checkout) SHALL have no effect.

#### Scenario: Unrelated action key pressed during triage
- **WHEN** the user presses a keybinding for a pull-request action that is
  not one of the triage workflow's own keys (enter/exit, next/previous
  thread, reply, resolve, scroll) while the triage workflow is active
- **THEN** the system SHALL take no action for that key press

### Requirement: Triage keybindings are shown in the help footer
While the triage workflow is active, opening the full help footer SHALL
show only the triage workflow's own keybindings, not the normal
pull-request action bindings that the previous requirement makes inactive
during triage.

#### Scenario: Full help while triaging
- **WHEN** the user opens the full help footer while the triage workflow is
  active
- **THEN** the system SHALL list the triage workflow's keybindings (next
  thread, previous thread, reply, resolve, exit) and SHALL NOT list
  pull-request action bindings that are inactive during triage (for
  example: approve, assign, label, merge, close, checkout)

#### Scenario: Full help outside triage
- **WHEN** the user opens the full help footer for a pull request while the
  triage workflow is not active
- **THEN** the system SHALL list the normal pull-request action bindings
  and the key that enters triage, and SHALL NOT list the triage-only
  next-thread, previous-thread, or resolve bindings, since those keys have
  no effect (or, for a key also bound to an unrelated action outside
  triage, a different effect) until triage is active
