# NBA Terminal

A terminal NBA live score viewer for Hupu's NBA game pages.

## Features

- Shows today's NBA games with teams, scores, and status.
- Supports previous/next day schedule navigation.
- Supports favorite teams with local persistence and pinned games.
- Refreshes every 3 seconds while any game is live.
- Stops list polling automatically when no games are live.
- Opens a selected game's text live feed.
- Shows box score stats in the match detail view.
- Adapts box score columns to narrower terminal widths.
- Supports live/stats tabs and scrolling inside match details.
- Supports keyboard navigation in a Bubble Tea TUI.

## Requirements

- Go 1.21 or newer
- Network access to `https://nba.hupu.com/games`

## Installation

Install the latest version from GitHub:

```bash
go install github.com/SheltonZhu/nba-terminal@latest
```

Then run:

```bash
nba-terminal
```

Make sure your Go binary directory is in `PATH`. By default, this is `$(go env GOPATH)/bin`, or `GOBIN` if you have set it.

## Usage

```bash
go run .
```

Or:

```bash
make run
```

Keyboard controls:

- `up` / `down`: move selection
- `enter`: open live detail
- `tab`: switch live and stats tabs in detail view
- `up` / `down` or `j` / `k`: scroll detail rows
- `[` / `]`: previous or next date
- `t`: jump back to today
- `f`: favorite a team from the selected game, then press `1` for the away team or `2` for the home team
- `esc`: return to game list
- `q` or `ctrl+c`: quit

Favorites are stored in `nba-terminal/favorites.json` under your user config directory.

## Development

Run tests:

```bash
go test ./...
```

Or:

```bash
make test
```

Build:

```bash
go build ./...
```

Or:

```bash
make build
```

## Release

Create and push a version tag to publish a GitHub release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs tests, builds macOS/Linux/Windows binaries for amd64 and arm64, uploads checksums, and generates release notes from git history.

## Notes

This tool parses Hupu HTML with CSS selectors. If Hupu changes the page structure, parsing may need to be updated in `parser.go`.
