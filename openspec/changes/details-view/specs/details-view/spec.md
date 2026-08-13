## Purpose

Lets a user focus on a single PR, issue, branch, or notification in a
full-window view so they can read and act on it without the list competing
for screen space, while keeping the navigation and action keys they already
know.

## ADDED Requirements

### Requirement: Entering the details view
The system SHALL let the user enter a full-window details view for the
currently selected row in the PRs, Issues, Repo (branches), or Notifications
list view by pressing `enter`.

#### Scenario: Enter from PRs, Issues, or Repo view
- **WHEN** the user presses `enter` while a row is selected in the PRs,
  Issues, or Repo (branches) view and no text input, search, or
  confirmation prompt is focused
- **THEN** the system SHALL replace the current list/preview layout with a
  full-window view of that row's preview content

#### Scenario: Enter from Notifications view before content is loaded
- **WHEN** the user presses `enter` on a notification row whose content has
  not yet been loaded into the preview
- **THEN** the system SHALL load the notification's content into the split
  preview (as it does today) and SHALL NOT enter the details view on this
  key press

#### Scenario: Enter from Notifications view after content is loaded
- **WHEN** the user presses `enter` on a notification row whose content is
  already loaded into the preview
- **THEN** the system SHALL enter the full-window details view showing that
  content

#### Scenario: No row selected
- **WHEN** the user presses `enter` and the current section has no rows (for
  example, an empty or still-loading list)
- **THEN** the system SHALL take no action and SHALL remain in the list view

### Requirement: Navigation keys scroll the preview while in the details view
While the details view is active, the system SHALL interpret the list's
navigation keys as scrolling/jumping within the displayed preview content
rather than changing the list selection.

#### Scenario: Line-by-line scroll
- **WHEN** the user presses `j`/down or `k`/up while in the details view
- **THEN** the system SHALL scroll the preview content down or up by one
  line respectively, and SHALL NOT change which row is selected in the
  underlying list

#### Scenario: Jump to top or bottom
- **WHEN** the user presses `g`/home or `G`/end while in the details view
- **THEN** the system SHALL scroll the preview content to its top or bottom
  respectively

#### Scenario: Page scrolling still works
- **WHEN** the user presses the existing page-up/page-down keys while in the
  details view
- **THEN** the system SHALL page the preview content up or down, the same
  as it does today when the preview is open in split mode

### Requirement: Item-level actions remain available in the details view
All keybindings that act on the currently focused row (for example: approve,
assign, unassign, label, comment, checkout, diff, close, reopen, merge,
ready, update, approve workflows, snooze, and switching sidebar tabs) SHALL
continue to function identically while the details view is active.

#### Scenario: Approve a PR from the details view
- **WHEN** the user is in the details view for a pull request and presses
  the approve key
- **THEN** the system SHALL start the approve flow exactly as it would from
  the split list+preview layout

#### Scenario: Switch preview tabs from the details view
- **WHEN** the user is in the details view for a pull request and presses
  the previous/next sidebar tab keys
- **THEN** the system SHALL switch the displayed tab within the full-window
  preview

### Requirement: Section switching and preview toggling are disabled in the details view
While the details view is active, the system SHALL ignore the
section-switching keys (previous/next section) and the preview toggle keys
(toggle preview, toggle preview position).

#### Scenario: Section switch key pressed in details view
- **WHEN** the user presses the previous-section or next-section key while
  in the details view
- **THEN** the system SHALL take no action and SHALL remain in the details
  view for the same row

#### Scenario: Preview toggle key pressed in details view
- **WHEN** the user presses the toggle-preview or toggle-preview-position
  key while in the details view
- **THEN** the system SHALL take no action and SHALL remain in the details
  view unchanged

### Requirement: Exiting the details view restores prior layout
Pressing `esc` while in the details view SHALL return the user to the list
view, restoring the preview pane's open/closed state and position to what
they were immediately before the details view was entered.

#### Scenario: Preview was open before entering details view
- **WHEN** the user enters the details view while the preview pane was
  already open (split list+preview layout) and then presses `esc`
- **THEN** the system SHALL return to the split list+preview layout, with
  the previously selected row still selected

#### Scenario: Preview was closed before entering details view
- **WHEN** the user enters the details view while the preview pane was
  closed (list-only layout) and then presses `esc`
- **THEN** the system SHALL return to the list-only layout with the preview
  pane closed again

#### Scenario: Exiting the details view for a loaded notification
- **WHEN** the user is in the details view for a notification whose content
  is loaded, and presses `esc`
- **THEN** the system SHALL return to the list with the notification's
  content still shown in the split preview, matching the layout that was in
  place before the details view was entered

### Requirement: List selection changes while the details view is closed still update the preview
Moving the list selection while the details view is not active SHALL
continue to update the preview content shown for the next time the details
view is entered, unchanged from today's behavior.

#### Scenario: Preview stays in sync with list navigation
- **WHEN** the user moves the list selection to a different row while the
  details view is not active
- **THEN** the system SHALL update the preview content for that row exactly
  as it does today, so that entering the details view afterward shows the
  newly selected row's content
