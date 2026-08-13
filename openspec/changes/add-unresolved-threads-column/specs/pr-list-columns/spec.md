## Purpose

Defines the set of columns available in the PRs section list view, including
what each column displays, where its data comes from, and how it can be
configured.

## ADDED Requirements

### Requirement: Number of comments column shows unresolved review threads

The PRs list view's `numComments` column SHALL display the number of
unresolved review threads on each pull request, rather than the total
number of issue comments and review threads.

#### Scenario: PR has unresolved review threads

- **WHEN** a PR has one or more review threads where `isResolved` is `false`
- **THEN** the `numComments` column shows the count of those unresolved
  threads

#### Scenario: PR has no unresolved review threads

- **WHEN** a PR has zero review threads, or all of its review threads have
  `isResolved` equal to `true`
- **THEN** the `numComments` column renders as blank (no `0` or other
  placeholder is shown), regardless of how many issue comments the PR has

### Requirement: Number of comments column remains configurable

The `numComments` column's configuration SHALL be unchanged by this
behavior change: it continues to support a configurable `width` and a
`hidden` flag under `Defaults.Layout.Prs` and a section's own
`Layout.Prs`, and remains visible by default.

#### Scenario: User hides the column

- **WHEN** a user sets `layout.prs.numComments.hidden: true` in their
  config (at the defaults or section level)
- **THEN** the column does not appear in that PRs section's table

#### Scenario: Default visibility

- **WHEN** no configuration is provided for the column
- **THEN** the column is visible in the PRs list view
