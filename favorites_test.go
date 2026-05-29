package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileFavoriteStoreRoundTripsTeams(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	store := FileFavoriteStore{path: path}

	want := map[string]bool{"湖人": true, "凯尔特人": true}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected favorites %#v, got %#v", want, got)
	}
}

func TestFileFavoriteStoreReturnsEmptyMapForMissingFile(t *testing.T) {
	store := FileFavoriteStore{path: filepath.Join(t.TempDir(), "favorites.json")}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no favorites, got %#v", got)
	}
}
