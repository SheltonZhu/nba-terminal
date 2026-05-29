package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type fakeFetcher struct {
	matches []Match
	detail  LiveDetail
	stats   BoxScoreDetail
	err     error
}

func (f fakeFetcher) FetchMatches() ([]Match, error) {
	return f.matches, f.err
}

func (f fakeFetcher) FetchLiveDetail(Match) (LiveDetail, error) {
	return f.detail, f.err
}

func (f fakeFetcher) FetchBoxScoreDetail(Match) (BoxScoreDetail, error) {
	return f.stats, f.err
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

func TestAppDetailTabsScrollingAndStatsRendering(t *testing.T) {
	app := NewApp(fakeFetcher{}, 80, 16)
	app.viewMode = detailView
	app.matches = []Match{{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusLiving, MatchTime: "Q4 02:31", Team1Score: "108", Team2Score: "105"}}
	app.detail = LiveDetail{LiveTextRows: []string{
		"row 1",
		"row 2",
		"row 3",
		"row 4",
		"row 5",
		"row 6",
	}}
	app.boxScore = BoxScoreDetail{
		Team1: TeamStats{Name: "湖人", Players: []PlayerStat{{Name: "詹姆斯", Minutes: "35", Points: "28", Rebounds: "7", Assists: "9", Steals: "2", Blocks: "1", Turnovers: "3", Fouls: "2"}}},
		Team2: TeamStats{Name: "凯尔特人", Players: []PlayerStat{{Name: "塔图姆", Minutes: "37", Points: "31", Rebounds: "8", Assists: "5"}}},
	}

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app = model.(App)
	if app.detailScroll != 1 {
		t.Fatalf("expected detail scroll 1, got %d", app.detailScroll)
	}

	view := app.View()
	if strings.Contains(view, "row 1") {
		t.Fatalf("expected row 1 to scroll out of view, got:\n%s", view)
	}
	if !strings.Contains(view, "row 2") {
		t.Fatalf("expected scrolled live row in view, got:\n%s", view)
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app = model.(App)
	if app.detailTab != statsTab {
		t.Fatalf("expected stats tab, got %v", app.detailTab)
	}

	view = app.View()
	for _, want := range []string{"统计", "詹姆斯", "28", "塔图姆", "31"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected stats view to contain %q, got:\n%s", want, view)
		}
	}
}

func TestRenderBoxScoreRowsAlignsWidePlayerNames(t *testing.T) {
	rows := renderBoxScoreRows(BoxScoreDetail{
		Team1: TeamStats{Name: "雷霆", Players: []PlayerStat{
			{Name: "谢伊-吉尔杰斯-亚历山大", Minutes: "28", Points: "15", Rebounds: "1", Assists: "4", Steals: "0", Blocks: "0", Turnovers: "2", Fouls: "1"},
			{Name: "贾里德·麦凯恩", Minutes: "27", Points: "13", Rebounds: "2", Assists: "6", Steals: "2", Blocks: "0", Turnovers: "2", Fouls: "1"},
			{Name: "Chet Holmgren", Minutes: "24", Points: "10", Rebounds: "11", Assists: "1", Steals: "1", Blocks: "4", Turnovers: "0", Fouls: "2"},
		}},
	})
	if len(rows) < 5 {
		t.Fatalf("expected title, header, and player rows, got %#v", rows)
	}

	header := rows[1]
	cases := []struct {
		row    string
		minute string
		point  string
	}{
		{row: rows[2], minute: "28", point: "15"},
		{row: rows[3], minute: "27", point: "13"},
		{row: rows[4], minute: "24", point: "10"},
	}
	for _, tc := range cases {
		for headerName, value := range map[string]string{"时间": tc.minute, "得分": tc.point} {
			wantColumn := displayEndColumn(header, headerName)
			gotColumn := displayEndColumn(tc.row, value)
			if gotColumn != wantColumn {
				t.Fatalf("expected %q column to end at display column %d, got %d\nheader: %q\nrow:    %q", headerName, wantColumn, gotColumn, header, tc.row)
			}
		}
	}
}

func displayColumn(row, value string) int {
	byteIndex := strings.Index(row, value)
	if byteIndex < 0 {
		return -1
	}
	return lipgloss.Width(row[:byteIndex])
}

func displayEndColumn(row, value string) int {
	column := displayColumn(row, value)
	if column < 0 {
		return column
	}
	return column + lipgloss.Width(value)
}
