package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeFetcher struct {
	matches []Match
	detail  LiveDetail
	err     error
}

func (f fakeFetcher) FetchMatches() ([]Match, error) {
	return f.matches, f.err
}

func (f fakeFetcher) FetchLiveDetail(Match) (LiveDetail, error) {
	return f.detail, f.err
}

func TestAppAppliesMatchesAndSchedulesRefreshOnlyWhenLive(t *testing.T) {
	app := NewApp(fakeFetcher{}, 80, 24)

	liveMatches := []Match{{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusLiving, MatchTime: "Q4 02:31", Team1Score: "108", Team2Score: "105", LiveURL: "live"}}
	model, cmd := app.Update(matchesLoadedMsg{matches: liveMatches})
	app = model.(App)

	if len(app.matches) != 1 || app.matches[0].Status != StatusLiving {
		t.Fatalf("matches were not applied: %#v", app.matches)
	}
	if cmd == nil {
		t.Fatal("expected refresh command while a match is live")
	}

	endedMatches := []Match{{Team1Name: "掘金", Team2Name: "太阳", Status: StatusEnded, MatchTime: "已结束", Team1Score: "99", Team2Score: "101", LiveURL: "live"}}
	model, cmd = app.Update(matchesLoadedMsg{matches: endedMatches})
	app = model.(App)
	if cmd != nil {
		t.Fatal("expected no refresh command when no match is live")
	}
}

func TestAppSelectionAndDetailNavigation(t *testing.T) {
	app := NewApp(fakeFetcher{}, 80, 24)
	app.matches = []Match{
		{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusLiving, LiveURL: "live-1"},
		{Team1Name: "勇士", Team2Name: "灰熊", Status: StatusPending},
	}

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app = model.(App)
	if app.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", app.selected)
	}

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)
	if app.viewMode != detailView {
		t.Fatalf("expected detail view, got %v", app.viewMode)
	}
	if cmd != nil {
		t.Fatal("expected no fetch command for pending match without live URL")
	}
	if app.err == "" {
		t.Fatal("expected user-facing error for match without live URL")
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = model.(App)
	if app.viewMode != listView {
		t.Fatalf("expected list view after escape, got %v", app.viewMode)
	}
}

func TestAppViewRendersMatchRows(t *testing.T) {
	app := NewApp(fakeFetcher{}, 80, 24)
	app.matches = []Match{{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusLiving, MatchTime: "Q4 02:31", Team1Score: "108", Team2Score: "105"}}

	view := app.View()
	for _, want := range []string{"NBA 实时比分", "湖人", "108 - 105", "Q4 02:31"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got:\n%s", want, view)
		}
	}
}
