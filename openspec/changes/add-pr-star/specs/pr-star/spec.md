## Purpose

Lets a user mark any PR in the PRs list with a purely local, personal
"star" flag to make it visually stand out for later attention, independent
of the PR's actual GitHub state.

## ADDED Requirements

### Requirement: User can toggle a star on the selected PR

The PRs list SHALL provide a keybinding that toggles a starred flag on the
currently selected PR. Toggling SHALL NOT make any GitHub API call and
SHALL NOT alter any property of the PR itself.

#### Scenario: Starring an unstarred PR

- **WHEN** the user presses the star keybinding on a PR that is not
  currently starred
- **THEN** the PR becomes starred and no request is sent to GitHub

#### Scenario: Unstarring a starred PR

- **WHEN** the user presses the star keybinding on a PR that is currently
  starred
- **THEN** the PR becomes unstarred and no request is sent to GitHub

### Requirement: Starred state is shown in the PRs list

The PRs list SHALL render a dedicated column at the leftmost position of
the table that displays a star indicator for starred PRs and is blank for
unstarred PRs.

#### Scenario: Starred PR is visually marked

- **WHEN** a PR in the list is starred
- **THEN** its row's leftmost column shows the star indicator

#### Scenario: Unstarred PR shows no indicator

- **WHEN** a PR in the list is not starred
- **THEN** its row's leftmost column is blank

### Requirement: Starred state persists across restarts

Starred state SHALL be stored locally and SHALL survive restarting the
application. A star SHALL remain set indefinitely - unlike a snooze, it
SHALL NOT expire on its own and SHALL NOT be cleared by new activity on the
PR (comments, pushes, review state changes, etc.). It SHALL only be
cleared when the user explicitly unstars the PR.

#### Scenario: Star survives a restart

- **WHEN** a user stars a PR and then restarts the application
- **THEN** the PR still shows as starred

#### Scenario: New activity does not clear a star

- **WHEN** a starred PR receives a new comment, push, or review
- **THEN** the PR remains starred

### Requirement: Star column is configurable

The star column's visibility and width SHALL be configurable the same way
as other PR list columns (a per-column `hidden`/`width` setting, at both
the defaults and section level), and SHALL be visible by default.

#### Scenario: User hides the star column

- **WHEN** a user sets the star column's `hidden` option to `true` (at the
  defaults or section level)
- **THEN** the column does not appear in that PRs section's table, though
  starring still works via the keybinding

#### Scenario: Default visibility

- **WHEN** no configuration is provided for the star column
- **THEN** the column is visible in the PRs list view
