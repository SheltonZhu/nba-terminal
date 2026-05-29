package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type FileFavoriteStore struct {
	path string
}

func NewFileFavoriteStore() FileFavoriteStore {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		configDir = "."
	}
	return FileFavoriteStore{path: filepath.Join(configDir, "nba-terminal", "favorites.json")}
}

func (s FileFavoriteStore) Load() (map[string]bool, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}

	var teams []string
	if err := json.Unmarshal(data, &teams); err != nil {
		return nil, err
	}
	favorites := make(map[string]bool, len(teams))
	for _, team := range teams {
		if team != "" {
			favorites[team] = true
		}
	}
	return favorites, nil
}

func (s FileFavoriteStore) Save(favorites map[string]bool) error {
	teams := make([]string, 0, len(favorites))
	for team, favorite := range favorites {
		if favorite {
			teams = append(teams, team)
		}
	}
	sort.Strings(teams)

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(teams, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}

func cloneFavorites(favorites map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(favorites))
	for team, favorite := range favorites {
		cloned[team] = favorite
	}
	return cloned
}
