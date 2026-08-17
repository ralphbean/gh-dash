## Purpose

Defines the set of columns available in the PRs section list view, including
what each column displays, where its data comes from, and how it can be
configured.

## ADDED Requirements

### Requirement: Review status is split into a human column and a bot column

The PRs list view SHALL display two separate review-status columns in place
of a single combined one: one summarizing the review status of human
(non-bot) reviewers, and one summarizing the review status of bot (GitHub
App) reviewers. The two columns SHALL appear adjacent to each other, in the
position previously occupied by the single `reviewStatus` column.

#### Scenario: PR has both human and bot reviews

- **WHEN** a PR has been approved by a human reviewer and also reviewed
  (in any state) by a bot reviewer
- **THEN** the human column shows the approved icon and the bot column
  shows its own, independently computed icon for the bot's review state

#### Scenario: PR has only bot reviews

- **WHEN** a PR has one or more reviews from bot reviewers and no reviews
  from human reviewers
- **THEN** the human column shows the waiting (no reviews yet) icon and the
  bot column reflects the bot reviewers' status

### Requirement: Each review-status column has a distinct header icon

The human review-status column and the bot review-status column SHALL use
different header icons, so the two columns are visually distinguishable
without relying on column order alone.

#### Scenario: Column headers are rendered

- **WHEN** the PRs list view's column headers are rendered
- **THEN** the human review-status column's header is a person icon and the
  bot review-status column's header is a robot icon, and the two icons are
  different from each other

### Requirement: Review status per column reflects only that group's reviews

Each review-status column SHALL compute its icon from only the reviews
authored by reviewers of its own group (human or bot), determined by each
review author's account type. Within a group, a reviewer who has both
approved (or requested changes) and separately commented SHALL count as
approved (or changes-requested), not commented.

#### Scenario: Group has a changes-requested review

- **WHEN** a group's most decisive review state is "changes requested"
  (i.e. at least one reviewer in that group requested changes and no
  higher-priority state applies)
- **THEN** that column shows the changes-requested icon

#### Scenario: Group has an approval and no changes-requested

- **WHEN** a group has at least one approval and no reviewer in that group
  has requested changes
- **THEN** that column shows the approved icon

#### Scenario: Group has only comments

- **WHEN** a group has at least one review but none are approvals or
  changes-requested
- **THEN** that column shows the "commented" icon

#### Scenario: Group has no reviews

- **WHEN** a group has no reviews at all
- **THEN** that column shows the waiting icon

### Requirement: Both review-status columns are independently configurable

Each review-status column SHALL support its own configurable `width` and
`hidden` flag under `Defaults.Layout.Prs` and a section's own
`Layout.Prs`, so a user can show/hide or resize the human and bot columns
independently. The existing `reviewStatus` config key continues to control
the human column; a new `reviewStatusBot` config key controls the bot
column.

#### Scenario: User hides only the bot column

- **WHEN** a user sets `layout.prs.reviewStatusBot.hidden: true` in their
  config (at the defaults or section level)
- **THEN** the bot review-status column does not appear in that PRs
  section's table, while the human review-status column continues to
  appear (unless separately hidden)

#### Scenario: Default visibility

- **WHEN** no configuration is provided for either column
- **THEN** both the human and bot review-status columns are visible in the
  PRs list view
