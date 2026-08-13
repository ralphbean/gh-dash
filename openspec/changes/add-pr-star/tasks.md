## 1. Data layer: `StarStore`

- [x] 1.1 Write `internal/data/starstore_test.go` first (TDD): cover
      `Star`/`IsStarred` (true after starring), `Unstar`/`IsStarred` (false
      after unstarring), toggling an unstarred id starrs it, toggling a
      starred id unstars it, persistence across a `load()` on a fresh store
      pointed at the same file path, and that an unknown id is not
      starred. Confirm these fail (package/type doesn't exist yet).
- [x] 1.2 Add `internal/data/starstore.go`: `StarStore` struct (mutex +
      `filePath` + in-memory set), `load`/`save` (JSON array of ids, atomic
      temp-file-then-rename, same shape as `snoozestore.go`'s), `Star(id
      string)`, `Unstar(id string)`, `IsStarred(id string) bool`,
      `Toggle(id string) bool` (returns the resulting starred state),
      `Flush() error`, and the `starStore`/`starStoreOnce`/`GetStarStore()`
      singleton, persisted to `starred.json`.
- [x] 1.3 Add `internal/data/starstore_testing.go`:
      `NewStarStoreForTesting(filePath string) *StarStore` and
      `OverrideStarStoreForTesting(store *StarStore) func()`, mirroring
      `snoozestore_testing.go`.
- [x] 1.4 Run the tests from 1.1 and confirm they pass.

## 2. Keybinding

- [x] 2.1 Add `Star key.Binding` to `PRKeyMap` in
      `internal/tui/keys/prKeys.go`, default key `*`, help text `"toggle
      star"`.
- [x] 2.2 Add `PRKeys.Star` to `PRFullHelp()`.
- [x] 2.3 Add a `case "star":` arm to `rebindPRKeys`'s builtin-name switch.
- [x] 2.4 Add `star` to the built-in PR keybinding docs/config schema
      wherever `snooze`/other builtins are enumerated for user-facing
      config validation (check `internal/config` for a builtin-name
      allowlist beyond the switch in 2.3, if one exists). No separate
      config-level allowlist exists beyond the `rebindPRKeys` switch; added
      `star` to the "Built-in Commands" table in
      `docs/src/content/docs/configuration/keybindings/index.mdx`.

## 3. Section wiring (`internal/tui/components/prssection/`)

- [x] 3.1 Write `star_test.go` first (TDD): `starKey(pr)` produces
      `pr:<repoNameWithOwner>#<number>`, matching `snoozeKey`'s format but
      independently (don't just alias the two).
- [x] 3.2 Add `star.go`: `starKey(pr data.RowData) string`.
- [x] 3.3 In `prssection.go`'s `Update`, add a `case
      key.Matches(msg, keys.PRKeys.Star):` arm (in the main key switch,
      not the confirmation-prompt branch) that calls
      `data.GetStarStore().Toggle(starKey(pr))` on `m.GetCurrRow()` and
      returns a `tasks.StarFeedback(...)` cmd (see Task 5) using the
      resulting boolean to pick start/finished wording. No-op if
      `GetCurrRow()` is nil.

## 4. Column and rendering

- [x] 4.1 Add `Star ColumnConfig` to `PrsLayoutConfig` in
      `internal/config/parser.go` (`yaml:"star,omitempty"`), following the
      existing field pattern.
- [x] 4.2 In `prssection.go`'s `GetSectionColumns`, compute
      `starLayout := config.MergeColumnConfigs(dLayout.Star, sLayout.Star)`
      and add a new leftmost `table.Column` (before the existing state
      column) using a star glyph header, a narrow fixed width, and
      `Hidden: starLayout.Hidden`, in both the compact and non-compact
      branches.
- [x] 4.3 In `internal/tui/components/prrow/prrow.go`, add
      `renderStar() string`: returns the star glyph (pick an unused
      nerd-font star icon in `internal/tui/constants/constants.go`, e.g.
      `StarIcon`) if `data.GetStarStore().IsStarred(starKey-equivalent for
      this row)` else `""`. Note: `prrow` currently has no direct
      dependency on `prssection`'s `starKey` - either export a shared key
      helper both packages can call, or duplicate the one-line key format
      locally in `prrow` the way `data.RowData` methods already expose
      `GetRepoNameWithOwner()`/`GetNumber()` for this exact purpose.
      Resolve during implementation; keep it consistent with whatever
      `star.go` (Task 3.2) settles on. Resolved: duplicated the one-line
      key format directly in `renderStar()` to avoid a `prrow`->`prssection`
      import cycle.
- [x] 4.4 Add `renderStar()` as the first entry in `ToTableRow`'s returned
      slice, in both the compact and non-compact branches.

## 5. Footer feedback

- [x] 5.1 Add `tasks.StarFeedback(ctx *context.ProgramContext, section
      SectionIdentifier, key, itemDescription string, starred bool)
      tea.Cmd` in `internal/tui/components/tasks/star.go`, modeled on
      `SnoozeFeedback`: synchronous start+finished text, wording branches
      on `starred` ("Starring "/"is now starred" vs. "Unstarring "/"is now
      unstarred").

## 6. Tests

- [x] 6.1 Update `prrow_test.go`: `renderStar`/`ToTableRow` cases for
      starred (glyph shown) and unstarred (blank) rows, using
      `NewStarStoreForTesting`/`OverrideStarStoreForTesting`.
- [x] 6.2 Update `prssection_test.go`: pressing `PRKeys.Star` toggles the
      selected row's starred state via `GetStarStore().IsStarred`, and
      does nothing when there's no current row.
- [x] 6.3 Add a `parser_test.go` case (or extend an existing table-driven
      one) confirming `layout.prs.star.hidden`/`.width` parse into
      `PrsLayoutConfig.Star`, mirroring existing column config tests.
- [x] 6.4 Run the full test suite and linters; fix anything broken by the
      new leftmost column shifting existing column-index assumptions in
      other tests. `go build ./...`, `go test ./...`, `gofmt -l`, `go vet
      ./...`, and `golangci-lint run ./internal/...` all pass (the 4
      pre-existing golines findings in threadtriage.go/reviewthread.go are
      unrelated to this change).

## 7. Docs

- [x] 7.1 Add the star column to
      `docs/src/content/docs/configuration/layout/pr.mdx` (config key,
      default width, default visibility), following the existing
      per-column section format.
- [x] 7.2 Add a `` `*` - Star PR `` section to
      `docs/src/content/docs/getting-started/keybindings/selected-pr.mdx`,
      following the existing `z` - Snooze PR section's format.
