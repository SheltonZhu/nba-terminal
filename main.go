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
	favoriteStore := NewFileFavoriteStore()
	favorites, err := favoriteStore.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nba-terminal: %v\n", err)
		os.Exit(1)
	}
	program := tea.NewProgram(NewAppWithFavorites(fetcher, 80, 24, favorites, favoriteStore.Save), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nba-terminal: %v\n", err)
		os.Exit(1)
	}
}
