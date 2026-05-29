package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type fakeFetcher struct {
	matches      []Match
	detail       LiveDetail
	stats        BoxScoreDetail
	err          error
	dates        []time.Time
	liveFetches  int
	statsFetches int
}

func (f *fakeFetcher) FetchMatches(date time.Time) ([]Match, error) {
	f.dates = append(f.dates, date)
	return f.matches, f.err
}

func (f *fakeFetcher) FetchLiveDetail(Match) (LiveDetail, error) {
	f.liveFetches++
	return f.detail, f.err
}

func (f *fakeFetcher) FetchBoxScoreDetail(Match) (BoxScoreDetail, error) {
	f.statsFetches++
	return f.stats, f.err
}

func TestAppAppliesMatchesAndSchedulesRefreshOnlyWhenLive(t *testing.T) {
	app := NewApp(&fakeFetcher{}, 80, 24)

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
	app := NewApp(&fakeFetcher{}, 80, 24)
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
	app := NewApp(&fakeFetcher{}, 80, 24)
	app.matches = []Match{{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusLiving, MatchTime: "Q4 02:31", Team1Score: "108", Team2Score: "105"}}

	view := app.View()
	for _, want := range []string{"NBA 实时比分", "湖人", "108 - 105", "Q4 02:31"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected view to contain %q, got:\n%s", want, view)
		}
	}
}

func TestAppDetailTabsScrollingAndStatsRendering(t *testing.T) {
	app := NewApp(&fakeFetcher{}, 80, 16)
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

func TestAppDateNavigationFetchesSelectedDate(t *testing.T) {
	fetcher := &fakeFetcher{}
	app := NewApp(fetcher, 80, 24)
	app.currentDate = time.Date(2026, 5, 29, 0, 0, 0, 0, time.Local)

	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	app = model.(App)
	if cmd == nil {
		t.Fatal("expected date navigation to fetch matches")
	}
	if want := "2026-05-30"; app.currentDate.Format("2006-01-02") != want {
		t.Fatalf("expected current date %s, got %s", want, app.currentDate.Format("2006-01-02"))
	}
	runCommand(cmd)
	if len(fetcher.dates) != 1 || fetcher.dates[0].Format("2006-01-02") != "2026-05-30" {
		t.Fatalf("expected fetch for next date, got %#v", fetcher.dates)
	}

	model, cmd = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	app = model.(App)
	if cmd == nil {
		t.Fatal("expected previous date to fetch matches")
	}
	if want := "2026-05-29"; app.currentDate.Format("2006-01-02") != want {
		t.Fatalf("expected current date %s, got %s", want, app.currentDate.Format("2006-01-02"))
	}
}

func TestAppFavoriteSelectionTogglesTeamAndMarksRows(t *testing.T) {
	var saved map[string]bool
	app := NewAppWithFavorites(&fakeFetcher{}, 80, 24, map[string]bool{}, func(favorites map[string]bool) error {
		saved = cloneFavorites(favorites)
		return nil
	})
	app.matches = []Match{
		{Team1Name: "雷霆", Team2Name: "马刺", Status: StatusLiving, MatchTime: "Q4", Team1Score: "91", Team2Score: "118"},
	}

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	app = model.(App)
	if !app.choosingFavorite {
		t.Fatal("expected favorite chooser to open")
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	app = model.(App)
	if !app.favoriteTeams["马刺"] || !saved["马刺"] {
		t.Fatalf("expected home team to be favorited, app=%#v saved=%#v", app.favoriteTeams, saved)
	}
	if !strings.Contains(app.View(), "*") {
		t.Fatalf("expected favorite marker in view:\n%s", app.View())
	}
}

func TestFavoriteMatchesRenderBeforeOtherMatches(t *testing.T) {
	app := NewAppWithFavorites(&fakeFetcher{}, 80, 24, map[string]bool{"马刺": true}, nil)
	app.matches = []Match{
		{Team1Name: "湖人", Team2Name: "凯尔特人", Status: StatusPending, MatchTime: "未开始 10:00"},
		{Team1Name: "雷霆", Team2Name: "马刺", Status: StatusLiving, MatchTime: "Q4", Team1Score: "91", Team2Score: "118"},
	}

	view := app.View()
	spursIndex := strings.Index(view, "马刺")
	lakersIndex := strings.Index(view, "湖人")
	if spursIndex < 0 || lakersIndex < 0 || spursIndex > lakersIndex {
		t.Fatalf("expected favorite match before other matches:\n%s", view)
	}
}

func TestRenderBoxScoreRowsAlignsWidePlayerNames(t *testing.T) {
	rows := renderBoxScoreRows(BoxScoreDetail{
		Team1: TeamStats{Name: "雷霆", Players: []PlayerStat{
			{Name: "谢伊-吉尔杰斯-亚历山大", Minutes: "28", Points: "15", Rebounds: "1", Assists: "4", Steals: "0", Blocks: "0", Turnovers: "2", Fouls: "1"},
			{Name: "贾里德·麦凯恩", Minutes: "27", Points: "13", Rebounds: "2", Assists: "6", Steals: "2", Blocks: "0", Turnovers: "2", Fouls: "1"},
			{Name: "Chet Holmgren", Minutes: "24", Points: "10", Rebounds: "11", Assists: "1", Steals: "1", Blocks: "4", Turnovers: "0", Fouls: "2"},
		}},
	}, 120)
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

func TestRenderBoxScoreRowsDropsLowPriorityColumnsWhenNarrow(t *testing.T) {
	rows := renderBoxScoreRows(BoxScoreDetail{
		Team1: TeamStats{Name: "雷霆", Players: []PlayerStat{
			{Name: "谢伊-吉尔杰斯-亚历山大", Minutes: "28", Points: "15", Rebounds: "1", Assists: "4", Steals: "0", Blocks: "0", Turnovers: "2", Fouls: "1"},
		}},
	}, 42)
	if len(rows) < 2 {
		t.Fatalf("expected stats rows, got %#v", rows)
	}
	header := rows[1]
	for _, want := range []string{"球员", "时间", "得分", "篮板", "助攻"} {
		if !strings.Contains(header, want) {
			t.Fatalf("expected narrow header to keep %q: %q", want, header)
		}
	}
	for _, dropped := range []string{"抢断", "盖帽", "失误", "犯规"} {
		if strings.Contains(header, dropped) {
			t.Fatalf("expected narrow header to drop %q: %q", dropped, header)
		}
	}
}

func TestRefreshLiveDetailAlsoRefreshesBoxScore(t *testing.T) {
	fetcher := &fakeFetcher{}
	app := NewApp(fetcher, 80, 24)
	app.viewMode = detailView
	app.matches = []Match{{Team1Name: "雷霆", Team2Name: "马刺", LiveURL: "live", DataStatisticsURL: "stats"}}

	model, cmd := app.Update(refreshLiveDetailMsg{})
	app = model.(App)
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
	runCommand(cmd)
	if fetcher.liveFetches != 1 || fetcher.statsFetches != 1 {
		t.Fatalf("expected live and stats refresh, got live=%d stats=%d", fetcher.liveFetches, fetcher.statsFetches)
	}
}

func runCommand(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, child := range batch {
			runCommand(child)
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
