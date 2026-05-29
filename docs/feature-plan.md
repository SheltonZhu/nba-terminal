# NBA Terminal Feature Plan

> This is the persistent next-work document for the project. When a feature is completed, update its checkbox, add the release or commit reference, and move to the next unchecked feature.

## Status Legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Completed

## Current Recommendation

Next recommended feature: **Feature 2: CLI Flags And Version Output**. Feature 1 is complete, and version/help flags are the smallest release-quality improvement left.

## Feature 1: Detail Tabs, Scrolling, And Box Score Stats

**Goal:** Let users switch between live text and player/team statistics in the match detail view.

**Why This First:** The app already captures `DataStatisticsURL`, but it does not use it. The detail view also truncates live rows without scroll support. This feature turns the existing detail screen into the complete game view.

**Files:**

- Modify: `model.go`
- Modify: `parser.go`
- Modify: `crawler.go`
- Modify: `ui.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Test: `parser_test.go`
- Test: `crawler_test.go`
- Test: `ui_test.go`

**Implementation Tasks:**

- [x] Add `BoxScoreDetail`, `TeamStats`, and `PlayerStat` structs to `model.go`.
- [x] Add `ParseBoxScoreDetail(io.Reader) (BoxScoreDetail, error)` to `parser.go`.
- [x] Add parser tests using static Hupu-like HTML with `.team_vs` and two `.table_list_live` stat tables.
- [x] Extend `Fetcher` with `FetchBoxScoreDetail(Match) (BoxScoreDetail, error)`.
- [x] Add HTTP fetcher tests for `DataStatisticsURL`.
- [x] Add detail view state: active tab, vertical scroll offset, live rows, stats rows.
- [x] Add keys: `tab` switches live/stats, `up/down` or `j/k` scroll, `esc` returns to list.
- [x] Render a compact stats table with columns for player, minutes, points, rebounds, assists, steals, blocks, turnovers, and fouls when present.
- [x] Update README controls and feature list.
- [x] Update CHANGELOG under `Unreleased`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go build ./...`.
- [x] Commit with `feat: add detail tabs and box score stats`.

**Completion Record:**

- Completed in commit: this commit
- Released in:
- Notes: Added live/stats tabs, detail scrolling, Hupu-like box score parsing, and compact terminal stats rendering.

## Feature 2: CLI Flags And Version Output

**Goal:** Add production-friendly command-line entry points.

**Why:** Release CI already injects `main.version`, but the app has no `--version` or `--help` output. This makes `go install` usage cleaner and gives users a quick way to verify installed versions.

**Files:**

- Modify: `main.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Test: add `main_test.go` if argument parsing is extracted into a testable function.

**Implementation Tasks:**

- [ ] Add package-level `version = "dev"` variable in `main.go`.
- [ ] Parse `--version` and `-v` before starting Bubble Tea.
- [ ] Parse `--help` and `-h` before starting Bubble Tea.
- [ ] Add optional `--no-alt-screen` for terminals where alt screen is undesirable.
- [ ] Extract argument handling into a small function so it can be tested without launching the TUI.
- [ ] Add tests for version/help/no-alt-screen parsing.
- [ ] Update README with examples.
- [ ] Update CHANGELOG under `Unreleased`.
- [ ] Run `go test -count=1 ./...`.
- [ ] Run `go build ./...`.
- [ ] Commit with `feat: add cli flags and version output`.

**Completion Record:**

- Completed in commit:
- Released in:
- Notes:

## Feature 3: Configurable Refresh And Network Resilience

**Goal:** Make polling and HTTP behavior easier to control and more reliable.

**Why:** The current 3-second refresh is hard-coded, and transient network failures are shown immediately without retry/backoff strategy.

**Files:**

- Modify: `main.go`
- Modify: `crawler.go`
- Modify: `ui.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Test: `crawler_test.go`
- Test: `ui_test.go`

**Implementation Tasks:**

- [ ] Add `--refresh 3s` CLI flag with `time.ParseDuration`.
- [ ] Pass refresh interval into `NewApp` instead of using a package constant.
- [ ] Add HTTP retry for temporary network failures with one retry after a short delay.
- [ ] Preserve last successful match list when a refresh fails, and show the error in a muted status line.
- [ ] Add tests for custom refresh interval scheduling through injected duration.
- [ ] Add tests for one retry on temporary HTTP failure using `httptest`.
- [ ] Update README with refresh flag.
- [ ] Update CHANGELOG under `Unreleased`.
- [ ] Run `go test -count=1 ./...`.
- [ ] Run `go build ./...`.
- [ ] Commit with `feat: add configurable refresh and retry`.

**Completion Record:**

- Completed in commit:
- Released in:
- Notes:

## Feature 4: Team Filter And Favorites

**Goal:** Let users focus on selected teams.

**Why:** A daily game list can be noisy. Filtering is useful without requiring account state or external services.

**Files:**

- Modify: `main.go`
- Modify: `ui.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Test: `ui_test.go`

**Implementation Tasks:**

- [ ] Add `--team <name>` flag that filters list rows by either team name.
- [ ] Show a visible status line when a filter is active.
- [ ] Show an empty-state message when no games match the filter.
- [ ] Add tests for filter matching team 1, team 2, and no matches.
- [ ] Update README with examples.
- [ ] Update CHANGELOG under `Unreleased`.
- [ ] Run `go test -count=1 ./...`.
- [ ] Run `go build ./...`.
- [ ] Commit with `feat: add team filter`.

**Completion Record:**

- Completed in commit:
- Released in:
- Notes:

## Feature 5: Parser Diagnostics

**Goal:** Make page structure breakage easier to debug.

**Why:** This project depends on Hupu HTML selectors. When the page changes, users should get actionable diagnostics instead of a vague empty list.

**Files:**

- Modify: `parser.go`
- Modify: `crawler.go`
- Modify: `main.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Test: `parser_test.go`
- Test: `crawler_test.go`

**Implementation Tasks:**

- [ ] Return a typed parse warning when the page has HTML but zero `.list_box` matches.
- [ ] Add `--debug-dump <path>` flag to write failed response HTML to disk.
- [ ] Add tests for zero-list diagnostics.
- [ ] Add tests for debug dump writing on parser failure.
- [ ] Update README troubleshooting section.
- [ ] Update CHANGELOG under `Unreleased`.
- [ ] Run `go test -count=1 ./...`.
- [ ] Run `go build ./...`.
- [ ] Commit with `feat: add parser diagnostics`.

**Completion Record:**

- Completed in commit:
- Released in:
- Notes:

## Release Checklist For Each Completed Feature

- [ ] Update this plan's feature checkbox and completion record.
- [ ] Update `CHANGELOG.md`.
- [ ] Run `go test -count=1 ./...`.
- [ ] Run `go build ./...`.
- [ ] Push `main`.
- [ ] If releasing, tag the next patch or minor version and push the tag.
- [ ] Verify the release workflow succeeds.
- [ ] Verify `go install github.com/SheltonZhu/nba-terminal@latest` after the tag is public.
