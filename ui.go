package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const refreshInterval = 3 * time.Second

type viewMode int

const (
	listView viewMode = iota
	detailView
)

type App struct {
	fetcher Fetcher

	matches  []Match
	selected int
	detail   LiveDetail
	viewMode viewMode
	loading  bool
	err      string
	width    int
	height   int
}

type matchesLoadedMsg struct {
	matches []Match
	err     error
}

type liveDetailLoadedMsg struct {
	detail LiveDetail
	err    error
}

type refreshMatchesMsg struct{}
type refreshLiveDetailMsg struct{}

func NewApp(fetcher Fetcher, width, height int) App {
	return App{
		fetcher:  fetcher,
		viewMode: listView,
		width:    width,
		height:   height,
	}
}

func (a App) Init() tea.Cmd {
	return a.fetchMatchesCmd()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case tea.KeyMsg:
		return a.updateKey(msg)
	case matchesLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err.Error()
			return a, scheduleMatchesRefresh()
		}
		a.err = ""
		a.matches = msg.matches
		if a.selected >= len(a.matches) {
			a.selected = max(0, len(a.matches)-1)
		}
		if hasLiveMatch(a.matches) {
			return a, scheduleMatchesRefresh()
		}
		return a, nil
	case liveDetailLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err.Error()
			return a, nil
		}
		a.err = ""
		a.detail = msg.detail
		if !msg.detail.IsFinished {
			return a, scheduleLiveDetailRefresh()
		}
		return a, nil
	case refreshMatchesMsg:
		if a.viewMode != listView {
			return a, nil
		}
		a.loading = true
		return a, a.fetchMatchesCmd()
	case refreshLiveDetailMsg:
		if a.viewMode != detailView || len(a.matches) == 0 {
			return a, nil
		}
		a.loading = true
		return a, a.fetchLiveDetailCmd(a.matches[a.selected])
	default:
		return a, nil
	}
}

func (a App) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return a, tea.Quit
	case tea.KeyUp:
		if a.viewMode == listView && a.selected > 0 {
			a.selected--
		}
	case tea.KeyDown:
		if a.viewMode == listView && a.selected < len(a.matches)-1 {
			a.selected++
		}
	case tea.KeyEnter:
		if a.viewMode == listView && len(a.matches) > 0 {
			match := a.matches[a.selected]
			a.viewMode = detailView
			a.detail = LiveDetail{}
			if match.LiveURL == "" {
				a.err = "该比赛暂无直播"
				return a, nil
			}
			a.loading = true
			a.err = ""
			return a, a.fetchLiveDetailCmd(match)
		}
	case tea.KeyEsc:
		if a.viewMode == detailView {
			a.viewMode = listView
			a.err = ""
		}
	default:
		if msg.String() == "q" {
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) View() string {
	if a.viewMode == detailView {
		return a.detailView()
	}
	return a.listView()
}

func (a App) listView() string {
	title := titleStyle.Render("NBA 实时比分")
	var lines []string
	lines = append(lines, title, "")

	if a.err != "" {
		lines = append(lines, errorStyle.Render(a.err), "")
	}
	if len(a.matches) == 0 {
		lines = append(lines, "今日没有比赛")
	} else {
		for i, match := range a.matches {
			cursor := " "
			if i == a.selected {
				cursor = ">"
			}
			lines = append(lines, fmt.Sprintf("%s %s", cursor, renderMatchRow(match)))
		}
	}

	if a.loading {
		lines = append(lines, "", mutedStyle.Render("刷新中..."))
	}
	lines = append(lines, "", mutedStyle.Render("↑↓ 选择  Enter 详情  q 退出"))
	return strings.Join(lines, "\n")
}

func (a App) detailView() string {
	match := Match{}
	if len(a.matches) > 0 && a.selected < len(a.matches) {
		match = a.matches[a.selected]
	}

	var lines []string
	lines = append(lines, titleStyle.Render(fmt.Sprintf("%s  %s  %s", match.Title(), match.ScoreText(), match.MatchTime)), "")
	if a.err != "" {
		lines = append(lines, errorStyle.Render(a.err), "")
	}
	if len(a.detail.LiveTextRows) == 0 {
		lines = append(lines, "暂无文字直播")
	} else {
		limit := a.height - 6
		if limit <= 0 || limit > len(a.detail.LiveTextRows) {
			limit = len(a.detail.LiveTextRows)
		}
		for _, row := range a.detail.LiveTextRows[:limit] {
			lines = append(lines, row)
		}
	}
	if a.loading {
		lines = append(lines, "", mutedStyle.Render("刷新中..."))
	}
	lines = append(lines, "", mutedStyle.Render("Esc 返回列表  q 退出"))
	return strings.Join(lines, "\n")
}

func (a App) fetchMatchesCmd() tea.Cmd {
	return func() tea.Msg {
		matches, err := a.fetcher.FetchMatches()
		return matchesLoadedMsg{matches: matches, err: err}
	}
}

func (a App) fetchLiveDetailCmd(match Match) tea.Cmd {
	return func() tea.Msg {
		detail, err := a.fetcher.FetchLiveDetail(match)
		return liveDetailLoadedMsg{detail: detail, err: err}
	}
}

func scheduleMatchesRefresh() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return refreshMatchesMsg{}
	})
}

func scheduleLiveDetailRefresh() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return refreshLiveDetailMsg{}
	})
}

func hasLiveMatch(matches []Match) bool {
	for _, match := range matches {
		if match.IsLive() {
			return true
		}
	}
	return false
}

func renderMatchRow(match Match) string {
	return fmt.Sprintf("%-8s %-9s %-8s %s", match.Team1Name, match.ScoreText(), match.Team2Name, match.MatchTime)
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
