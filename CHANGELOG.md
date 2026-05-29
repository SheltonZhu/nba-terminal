# Changelog

All notable changes to this project are documented here.

## Unreleased

- Added tag-based release workflow that tests, builds cross-platform binaries, generates checksums, and publishes GitHub release notes.

## v0.1.1

- Fixed Go module path for `go install github.com/SheltonZhu/nba-terminal@latest`.
- Lowered the Go directive to `1.21.0`.
- Updated release CI to build with the latest Go 1.21 patch version.

## v0.1.0

- Added terminal UI for NBA game lists and live detail views.
- Added Hupu NBA HTML parsing for match lists and text live rows.
- Added polling while games are live.
- Added unit tests for parsing, HTTP fetching, and UI state updates.
