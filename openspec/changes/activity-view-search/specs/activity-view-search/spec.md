## Purpose

Lets a user search for and navigate to specific text (usernames, keywords,
phrases) within a pull request's Activity tab without manually scrolling
through the entire chronological comment feed.

## ADDED Requirements

### Requirement: Starting a search
The system SHALL let the user start a forward or backward search when viewing
the Activity tab in the full-window details view for a pull request, by
pressing a dedicated key.

#### Scenario: Start forward search from Activity tab
- **WHEN** the user is viewing the Activity tab in the full-window details
  view for a pull request and presses the forward search key
- **THEN** the system SHALL show a search input box at the bottom of the
  sidebar with a forward-search indicator prompt

#### Scenario: Start backward search from Activity tab
- **WHEN** the user is viewing the Activity tab in the full-window details
  view for a pull request and presses the backward search key
- **THEN** the system SHALL show a search input box at the bottom of the
  sidebar with a backward-search indicator prompt

#### Scenario: Search keys are inactive outside Activity tab
- **WHEN** the user presses the forward or backward search key while viewing
  a tab other than Activity (for example: Overview, Commits, Checks, Files
  Changed), or while in the thread triage workflow
- **THEN** the system SHALL take no action

#### Scenario: Search keys are inactive outside details view
- **WHEN** the user presses the forward or backward search key while viewing
  the PRs list (not in the details view), or while in the details view for an
  issue, branch, or notification
- **THEN** the system SHALL take no action

#### Scenario: Search restores previous term
- **WHEN** the user starts a new search after having previously searched in
  the same Activity view
- **THEN** the system SHALL pre-populate the search input box with the
  previous search term

### Requirement: Executing a search
Pressing Enter in the search input box SHALL find all occurrences of the
search term in the Activity tab's rendered content, highlight them, and move
to the first match (for forward search) or last match (for backward search).

#### Scenario: Search with matches
- **WHEN** the user types a search term and presses Enter, and the term
  appears one or more times in the Activity tab's content
- **THEN** the system SHALL highlight all occurrences of the term, visually
  distinguish the current match from other matches, auto-scroll the sidebar
  to show the current match, and display a match count (for example: "3/7")
  in or near the search input box

#### Scenario: Forward search starts at first match
- **WHEN** the user executes a forward search
- **THEN** the system SHALL make the first occurrence of the term (from the
  top of the Activity feed) the current match

#### Scenario: Backward search starts at last match
- **WHEN** the user executes a backward search
- **THEN** the system SHALL make the last occurrence of the term (from the
  bottom of the Activity feed) the current match

#### Scenario: Search with no matches
- **WHEN** the user types a search term and presses Enter, and the term does
  not appear anywhere in the Activity tab's content
- **THEN** the system SHALL display a "no matches" indicator in or near the
  search input box and SHALL NOT highlight any content

#### Scenario: Search is case-insensitive
- **WHEN** the user searches for a term containing uppercase letters, and the
  Activity content contains that term in a different case (for example:
  searching for "Alice" when the content contains "alice", "ALICE", "aLiCe")
- **THEN** the system SHALL match all case variations and highlight them

#### Scenario: Search matches plain text, not markdown syntax
- **WHEN** the user searches for text that appears in the rendered output of
  a comment's markdown body
- **THEN** the system SHALL match the rendered text as it appears on screen,
  ignoring markdown formatting characters (for example: searching for "bold"
  matches "bold" rendered from `**bold**`, not the literal string
  "**bold**")

### Requirement: Navigating between matches
The system SHALL let the user move to the next or previous match after
executing a search, auto-scrolling the sidebar to show each match as the user
navigates.

#### Scenario: Move to next match
- **WHEN** the user presses the next-match key after executing a search that
  found matches
- **THEN** the system SHALL make the next match in the Activity feed (moving
  down) the current match, visually distinguish it from other matches, update
  the match count indicator, and auto-scroll the sidebar to show it

#### Scenario: Move to previous match
- **WHEN** the user presses the previous-match key after executing a search
  that found matches
- **THEN** the system SHALL make the previous match in the Activity feed
  (moving up) the current match, visually distinguish it from other matches,
  update the match count indicator, and auto-scroll the sidebar to show it

#### Scenario: Wrapping past the last match
- **WHEN** the user presses the next-match key while the current match is the
  last match in the Activity feed
- **THEN** the system SHALL wrap around to the first match, making it the
  current match

#### Scenario: Wrapping past the first match
- **WHEN** the user presses the previous-match key while the current match is
  the first match in the Activity feed
- **THEN** the system SHALL wrap around to the last match, making it the
  current match

#### Scenario: Next/previous keys are inactive before search
- **WHEN** the user presses the next-match or previous-match key before
  executing a search, or after a search that found no matches
- **THEN** the system SHALL take no action

### Requirement: Clearing a search
Pressing the exit key while a search is active (either in the search input
box or after navigating matches) SHALL remove all search highlights and close
the search input box, returning to normal Activity tab viewing.

#### Scenario: Cancel search input before executing
- **WHEN** the user opens the search input box, types (or does not type) a
  term, and presses the exit key before pressing Enter
- **THEN** the system SHALL close the search input box without highlighting
  anything, returning to the Activity tab in its normal state

#### Scenario: Clear search after executing
- **WHEN** the user executes a search that highlights matches and presses the
  exit key
- **THEN** the system SHALL remove all highlights, close the search input
  box, and return to the Activity tab in its normal state

#### Scenario: Cleared search preserves scroll position
- **WHEN** the user clears a search after having navigated to a match
- **THEN** the system SHALL keep the sidebar scrolled to the position of the
  last current match, not jump back to the top or bottom

### Requirement: Search state is scoped to the current PR and Activity tab
Search highlights and the current match SHALL be cleared when the user
switches to a different pull request, switches to a different tab, or enters
the thread triage workflow, without requiring an explicit clear action.

#### Scenario: Switching tabs clears search
- **WHEN** the user executes a search in the Activity tab and switches to a
  different tab (for example: Overview, Commits, Checks, Files Changed)
- **THEN** the system SHALL clear the search highlights and close the search
  input box

#### Scenario: Returning to Activity tab does not restore search
- **WHEN** the user executes a search in the Activity tab, switches to
  another tab (which clears the search), and switches back to the Activity
  tab
- **THEN** the system SHALL show the Activity tab in its normal state without
  highlights, requiring the user to start a new search to restore them

#### Scenario: Switching PRs clears search
- **WHEN** the user executes a search for one pull request and navigates to a
  different pull request (in the same session)
- **THEN** the system SHALL clear the search highlights and close the search
  input box for the new pull request

#### Scenario: Entering thread triage clears search
- **WHEN** the user executes a search in the Activity tab and enters the
  thread triage workflow
- **THEN** the system SHALL clear the search highlights and close the search
  input box

### Requirement: Normal Activity tab actions work during search
While a search is active and matches are highlighted, the user SHALL still be
able to scroll the Activity tab content, switch tabs, open the help footer,
and exit the details view, using the same keys as when no search is active.

#### Scenario: Scrolling with search active
- **WHEN** the user has executed a search and presses a scroll key (for
  example: up, down, page up, page down, jump to top, jump to bottom)
- **THEN** the system SHALL scroll the Activity tab content and SHALL keep
  the search highlights visible and the search input box open

#### Scenario: Help footer shows search keys
- **WHEN** the user opens the full help footer while a search is active in
  the Activity tab
- **THEN** the system SHALL list the search keybindings (forward search,
  backward search, next match, previous match, exit search) alongside the
  other Activity tab actions

#### Scenario: Help footer hides search keys when inactive
- **WHEN** the user opens the full help footer for the Activity tab while no
  search is active
- **THEN** the system SHALL list only the forward-search and backward-search
  keys, and SHALL NOT list the next-match or previous-match keys (since those
  have no effect until a search has been executed)

### Requirement: Auto-scroll centers the current match in the viewport
When navigating to a match (either by executing a search or by pressing
next/previous match), the system SHALL auto-scroll the sidebar viewport to
show the current match, centering it vertically in the visible area when
possible.

#### Scenario: Match is below the visible area
- **WHEN** the user navigates to a match that is below the currently visible
  portion of the Activity tab
- **THEN** the system SHALL scroll down to bring that match into view,
  centering it vertically

#### Scenario: Match is above the visible area
- **WHEN** the user navigates to a match that is above the currently visible
  portion of the Activity tab
- **THEN** the system SHALL scroll up to bring that match into view,
  centering it vertically

#### Scenario: Match is already visible
- **WHEN** the user navigates to a match that is already visible in the
  sidebar viewport
- **THEN** the system SHALL update the visual distinction to show which match
  is current, without scrolling

#### Scenario: Centering is constrained by content bounds
- **WHEN** the user navigates to a match near the top or bottom of the
  Activity tab, where centering it would scroll past the start or end of the
  content
- **THEN** the system SHALL scroll as close to centered as possible without
  showing blank space above the first line or below the last line
