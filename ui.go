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

type detailTab int

const (
	liveTab detailTab = iota
	statsTab
)

type App struct {
	fetcher Fetcher

	matches      []Match
	selected     int
	detail       LiveDetail
	boxScore     BoxScoreDetail
	viewMode     viewMode
	detailTab    detailTab
	detailScroll int
	loading      bool
	err          string
	width        int
	height       int
}

type matchesLoadedMsg struct {
	matches []Match
	err     error
}

type liveDetailLoadedMsg struct {
	detail LiveDetail
	err    error
}

type boxScoreLoadedMsg struct {
	boxScore BoxScoreDetail
	err      error
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
	case boxScoreLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err.Error()
			return a, nil
		}
		a.err = ""
		a.boxScore = msg.boxScore
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
		if a.viewMode == detailView {
			a.detailScroll = max(0, a.detailScroll-1)
		} else if a.viewMode == listView && a.selected > 0 {
			a.selected--
		}
	case tea.KeyDown:
		if a.viewMode == detailView {
			a.detailScroll++
		} else if a.viewMode == listView && a.selected < len(a.matches)-1 {
			a.selected++
		}
	case tea.KeyTab:
		if a.viewMode == detailView {
			a.toggleDetailTab()
		}
	case tea.KeyEnter:
		if a.viewMode == listView && len(a.matches) > 0 {
			match := a.matches[a.selected]
			a.viewMode = detailView
			a.detail = LiveDetail{}
			a.boxScore = BoxScoreDetail{}
			a.detailTab = liveTab
			a.detailScroll = 0
			if match.LiveURL == "" {
				a.err = "该比赛暂无直播"
				return a, nil
			}
			a.loading = true
			a.err = ""
			return a, tea.Batch(a.fetchLiveDetailCmd(match), a.fetchBoxScoreDetailCmd(match))
		}
	case tea.KeyEsc:
		if a.viewMode == detailView {
			a.viewMode = listView
			a.err = ""
		}
	default:
		switch msg.String() {
		case "q":
			return a, tea.Quit
		case "j":
			if a.viewMode == detailView {
				a.detailScroll++
			}
		case "k":
			if a.viewMode == detailView {
				a.detailScroll = max(0, a.detailScroll-1)
			}
		}
	}
	return a, nil
}

func (a *App) toggleDetailTab() {
	if a.detailTab == liveTab {
		a.detailTab = statsTab
	} else {
		a.detailTab = liveTab
	}
	a.detailScroll = 0
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
	lines = append(lines, renderDetailTabs(a.detailTab), "")
	if a.err != "" {
		lines = append(lines, errorStyle.Render(a.err), "")
	}

	rows := a.detailRows()
	if len(rows) == 0 {
		rows = []string{"暂无文字直播"}
		if a.detailTab == statsTab {
			rows = []string{"暂无技术统计"}
		}
	}

	limit := a.height - 8
	if limit <= 0 {
		limit = len(rows)
	}
	start := clamp(a.detailScroll, 0, max(0, len(rows)-1))
	if start > len(rows) {
		start = len(rows)
	}
	end := start + limit
	if end > len(rows) {
		end = len(rows)
	}
	for _, row := range rows[start:end] {
		lines = append(lines, row)
	}

	if a.loading {
		lines = append(lines, "", mutedStyle.Render("刷新中..."))
	}
	lines = append(lines, "", mutedStyle.Render("Tab 切换  ↑↓/j/k 滚动  Esc 返回列表  q 退出"))
	return strings.Join(lines, "\n")
}

func (a App) detailRows() []string {
	if a.detailTab == statsTab {
		return renderBoxScoreRows(a.boxScore)
	}
	return a.detail.LiveTextRows
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

func (a App) fetchBoxScoreDetailCmd(match Match) tea.Cmd {
	if match.DataStatisticsURL == "" {
		return nil
	}
	return func() tea.Msg {
		boxScore, err := a.fetcher.FetchBoxScoreDetail(match)
		return boxScoreLoadedMsg{boxScore: boxScore, err: err}
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

func renderDetailTabs(active detailTab) string {
	live := "直播"
	stats := "统计"
	if active == liveTab {
		live = selectedTabStyle.Render(live)
		stats = mutedStyle.Render(stats)
	} else {
		live = mutedStyle.Render(live)
		stats = selectedTabStyle.Render(stats)
	}
	return live + "  " + stats
}

func renderBoxScoreRows(detail BoxScoreDetail) []string {
	if detail.IsEmpty() {
		return nil
	}

	widths := boxScoreColumnWidths(detail)
	rows := []string{}
	rows = appendTeamStatsRows(rows, detail.Team1, widths)
	if len(detail.Team1.Players) > 0 && len(detail.Team2.Players) > 0 {
		rows = append(rows, "")
	}
	rows = appendTeamStatsRows(rows, detail.Team2, widths)
	return rows
}

func appendTeamStatsRows(rows []string, team TeamStats, widths []int) []string {
	if len(team.Players) == 0 {
		return rows
	}
	if team.Name != "" {
		rows = append(rows, titleStyle.Render(team.Name))
	}
	rows = append(rows, renderBoxScoreLine(widths, []string{"球员", "时间", "得分", "篮板", "助攻", "抢断", "盖帽", "失误", "犯规"}))
	for _, player := range team.Players {
		rows = append(rows, renderBoxScoreLine(widths, []string{
			player.Name,
			player.Minutes,
			player.Points,
			player.Rebounds,
			player.Assists,
			player.Steals,
			player.Blocks,
			player.Turnovers,
			player.Fouls,
		}))
	}
	return rows
}

func boxScoreColumnWidths(detail BoxScoreDetail) []int {
	widths := []int{12, 4, 4, 4, 4, 4, 4, 4, 4}
	for index, header := range []string{"球员", "时间", "得分", "篮板", "助攻", "抢断", "盖帽", "失误", "犯规"} {
		updateColumnWidth(widths, index, header)
	}
	for _, team := range []TeamStats{detail.Team1, detail.Team2} {
		for _, player := range team.Players {
			for index, value := range []string{player.Name, player.Minutes, player.Points, player.Rebounds, player.Assists, player.Steals, player.Blocks, player.Turnovers, player.Fouls} {
				updateColumnWidth(widths, index, value)
			}
		}
	}
	return widths
}

func updateColumnWidth(widths []int, index int, value string) {
	if index >= len(widths) {
		return
	}
	if width := lipgloss.Width(value); width > widths[index] {
		widths[index] = width
	}
}

func renderBoxScoreLine(widths []int, values []string) string {
	parts := make([]string, 0, len(values))
	for index, value := range values {
		width := 0
		if index < len(widths) {
			width = widths[index]
		}
		if index == 0 {
			parts = append(parts, padRightDisplay(value, width))
		} else {
			parts = append(parts, padLeftDisplay(value, width))
		}
	}
	return strings.Join(parts, "  ")
}

func padRightDisplay(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func padLeftDisplay(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return strings.Repeat(" ", padding) + value
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	selectedTabStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("39")).Padding(0, 1)
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
