package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fetcher := NewHTTPFetcher(defaultGamesURL, &http.Client{Timeout: 10 * time.Second})
	program := tea.NewProgram(NewApp(fetcher, 80, 24), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nba-terminal: %v\n", err)
		os.Exit(1)
	}
}
