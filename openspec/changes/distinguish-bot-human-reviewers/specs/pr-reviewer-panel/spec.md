## Purpose

Defines how the PR preview pane and full-window details view render the
list of reviewers and review requests for a pull request, since both views
share the same rendering.

## ADDED Requirements

### Requirement: Reviewers are grouped by human vs. bot

The PR preview/details reviewer list SHALL group entries (pending review
requests, completed reviews, and suggested/code-owner reviewers) into a
human group and a bot group, determined by each reviewer's account type,
instead of rendering all reviewers in one undifferentiated list.

#### Scenario: PR has both human and bot reviewers

- **WHEN** a PR has at least one human reviewer (requested, reviewed, or
  suggested) and at least one bot reviewer
- **THEN** the reviewer list shows a human group and a bot group, each
  listing only the reviewers belonging to that group

### Requirement: Reviewer groups are visually distinguished and empty groups are omitted

Each non-empty reviewer group SHALL be prefixed with an icon identifying it
as the human group or the bot group. A group with no reviewers in it SHALL
NOT be rendered.

#### Scenario: PR has only human reviewers

- **WHEN** a PR has one or more human reviewers and no bot reviewers
- **THEN** only the human group is rendered (prefixed with the person
  icon); no empty bot group placeholder appears

#### Scenario: PR has only bot reviewers

- **WHEN** a PR has one or more bot reviewers and no human reviewers
- **THEN** only the bot group is rendered (prefixed with the robot icon);
  no empty human group placeholder appears

### Requirement: Existing per-reviewer status rendering is preserved within each group

Within each group, existing per-reviewer indicators (waiting/commented/
approved/changes-requested state, the code-owner icon, and wrapping to
multiple lines when a group's entries exceed the available width) SHALL
continue to render exactly as they did before grouping was introduced.

#### Scenario: A group's reviewers wrap across multiple lines

- **WHEN** a single group (human or bot) has enough reviewers that they
  don't fit on one line at the current width
- **THEN** that group's reviewers wrap onto additional lines, the same way
  the combined list wrapped before this change

#### Scenario: Code owner indicator is preserved

- **WHEN** a reviewer (in either group) is a requested reviewer marked as a
  code owner, or is a suggested code-owner reviewer
- **THEN** that reviewer's entry still shows the code-owner icon, within
  its group
