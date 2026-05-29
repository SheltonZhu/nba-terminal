# NBA Terminal

A terminal NBA live score viewer for Hupu's NBA game pages.

## Features

- Shows today's NBA games with teams, scores, and status.
- Refreshes every 3 seconds while any game is live.
- Stops list polling automatically when no games are live.
- Opens a selected game's text live feed.
- Supports keyboard navigation in a Bubble Tea TUI.

## Requirements

- Go 1.24 or newer
- Network access to `https://nba.hupu.com/games`

## Usage

```bash
go run .
```

Keyboard controls:

- `up` / `down`: move selection
- `enter`: open live detail
- `esc`: return to game list
- `q` or `ctrl+c`: quit

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build ./...
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
