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

	matches          []Match
	selected         int
	detail           LiveDetail
	boxScore         BoxScoreDetail
	viewMode         viewMode
	detailTab        detailTab
	detailScroll     int
	loading          bool
	err              string
	width            int
	height           int
	currentDate      time.Time
	favoriteTeams    map[string]bool
	saveFavorites    func(map[string]bool) error
	choosingFavorite bool
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
	return NewAppWithFavorites(fetcher, width, height, nil, nil)
}

func NewAppWithFavorites(fetcher Fetcher, width, height int, favorites map[string]bool, saveFavorites func(map[string]bool) error) App {
	if favorites == nil {
		favorites = map[string]bool{}
	}
	return App{
		fetcher:       fetcher,
		viewMode:      listView,
		width:         width,
		height:        height,
		currentDate:   startOfDay(time.Now()),
		favoriteTeams: cloneFavorites(favorites),
		saveFavorites: saveFavorites,
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
		match := a.selectedMatch()
		return a, tea.Batch(a.fetchLiveDetailCmd(match), a.fetchBoxScoreDetailCmd(match))
	default:
		return a, nil
	}
}

func (a App) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.choosingFavorite {
		return a.updateFavoriteChoice(msg)
	}
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
			match := a.selectedMatch()
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
		case "f":
			if a.viewMode == listView && len(a.matches) > 0 {
				a.choosingFavorite = true
			}
		case "[":
			a = a.changeDate(-1)
			return a, a.fetchMatchesCmd()
		case "]":
			a = a.changeDate(1)
			return a, a.fetchMatchesCmd()
		case "t":
			a.currentDate = startOfDay(time.Now())
			a.selected = 0
			a.matches = nil
			a.loading = true
			a.err = ""
			return a, a.fetchMatchesCmd()
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

func (a App) updateFavoriteChoice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc {
		a.choosingFavorite = false
		return a, nil
	}
	if msg.Type != tea.KeyRunes || len(a.matches) == 0 {
		return a, nil
	}
	match := a.selectedMatch()
	switch msg.String() {
	case "1":
		a.toggleFavorite(match.Team1Name)
	case "2":
		a.toggleFavorite(match.Team2Name)
	default:
		return a, nil
	}
	a.choosingFavorite = false
	if a.saveFavorites != nil {
		if err := a.saveFavorites(cloneFavorites(a.favoriteTeams)); err != nil {
			a.err = err.Error()
		}
	}
	return a, nil
}

func (a *App) toggleFavorite(team string) {
	if team == "" {
		return
	}
	if a.favoriteTeams == nil {
		a.favoriteTeams = map[string]bool{}
	}
	if a.favoriteTeams[team] {
		delete(a.favoriteTeams, team)
		return
	}
	a.favoriteTeams[team] = true
}

func (a App) changeDate(days int) App {
	a.currentDate = startOfDay(a.currentDate.AddDate(0, 0, days))
	a.selected = 0
	a.matches = nil
	a.loading = true
	a.err = ""
	return a
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
	title := titleStyle.Render(fmt.Sprintf("NBA 实时比分  %s", a.currentDate.Format("2006-01-02")))
	var lines []string
	lines = append(lines, title, "")

	if a.err != "" {
		lines = append(lines, errorStyle.Render(a.err), "")
	}
	if len(a.matches) == 0 {
		lines = append(lines, "今日没有比赛")
	} else {
		for i, match := range a.visibleMatches() {
			cursor := " "
			if i == a.selected {
				cursor = ">"
			}
			lines = append(lines, fmt.Sprintf("%s %s", cursor, renderMatchRow(match, a.favoriteTeams)))
		}
	}

	if a.choosingFavorite && len(a.matches) > 0 {
		match := a.selectedMatch()
		lines = append(lines, "", mutedStyle.Render(fmt.Sprintf("收藏球队: 1 %s  2 %s  Esc 取消", match.Team1Name, match.Team2Name)))
	}

	if a.loading {
		lines = append(lines, "", mutedStyle.Render("刷新中..."))
	}
	lines = append(lines, "", mutedStyle.Render("↑↓ 选择  Enter 详情  [/] 日期  t 今天  q 退出"))
	return strings.Join(lines, "\n")
}

func (a App) visibleMatches() []Match {
	matches := make([]Match, 0, len(a.matches))
	for _, match := range a.matches {
		if isFavoriteMatch(match, a.favoriteTeams) {
			matches = append(matches, match)
		}
	}
	for _, match := range a.matches {
		if !isFavoriteMatch(match, a.favoriteTeams) {
			matches = append(matches, match)
		}
	}
	return matches
}

func (a App) selectedMatch() Match {
	matches := a.visibleMatches()
	if len(matches) == 0 || a.selected >= len(matches) {
		return Match{}
	}
	return matches[a.selected]
}

func (a App) detailView() string {
	match := Match{}
	if len(a.matches) > 0 && a.selected < len(a.matches) {
		match = a.selectedMatch()
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
		return renderBoxScoreRows(a.boxScore, a.width)
	}
	return a.detail.LiveTextRows
}

func (a App) fetchMatchesCmd() tea.Cmd {
	return func() tea.Msg {
		matches, err := a.fetcher.FetchMatches(a.currentDate)
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

func renderMatchRow(match Match, favorites map[string]bool) string {
	marker := " "
	if isFavoriteMatch(match, favorites) {
		marker = "*"
	}
	return fmt.Sprintf("%s %-4s %-8s %-9s %-8s %s", marker, match.StatusLabel(), match.Team1Name, match.ScoreText(), match.Team2Name, match.MatchTime)
}

func isFavoriteMatch(match Match, favorites map[string]bool) bool {
	return favorites[match.Team1Name] || favorites[match.Team2Name]
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

func renderBoxScoreRows(detail BoxScoreDetail, maxWidth int) []string {
	if detail.IsEmpty() {
		return nil
	}

	columns := boxScoreColumnsForWidth(detail, maxWidth)
	widths := boxScoreColumnWidths(detail, columns)
	rows := []string{}
	rows = appendTeamStatsRows(rows, detail.Team1, columns, widths)
	if len(detail.Team1.Players) > 0 && len(detail.Team2.Players) > 0 {
		rows = append(rows, "")
	}
	rows = appendTeamStatsRows(rows, detail.Team2, columns, widths)
	return rows
}

func appendTeamStatsRows(rows []string, team TeamStats, columns []boxScoreColumn, widths []int) []string {
	if len(team.Players) == 0 {
		return rows
	}
	if team.Name != "" {
		rows = append(rows, titleStyle.Render(team.Name))
	}
	rows = append(rows, renderBoxScoreLine(widths, boxScoreHeaderValues(columns)))
	for _, player := range team.Players {
		rows = append(rows, renderBoxScoreLine(widths, boxScorePlayerValues(player, columns)))
	}
	return rows
}

type boxScoreColumn struct {
	header string
	value  func(PlayerStat) string
}

func allBoxScoreColumns() []boxScoreColumn {
	return []boxScoreColumn{
		{header: "球员", value: func(p PlayerStat) string { return p.Name }},
		{header: "时间", value: func(p PlayerStat) string { return p.Minutes }},
		{header: "得分", value: func(p PlayerStat) string { return p.Points }},
		{header: "篮板", value: func(p PlayerStat) string { return p.Rebounds }},
		{header: "助攻", value: func(p PlayerStat) string { return p.Assists }},
		{header: "抢断", value: func(p PlayerStat) string { return p.Steals }},
		{header: "盖帽", value: func(p PlayerStat) string { return p.Blocks }},
		{header: "失误", value: func(p PlayerStat) string { return p.Turnovers }},
		{header: "犯规", value: func(p PlayerStat) string { return p.Fouls }},
	}
}

func boxScoreColumnsForWidth(detail BoxScoreDetail, maxWidth int) []boxScoreColumn {
	columns := allBoxScoreColumns()
	if maxWidth <= 0 {
		return columns
	}
	for len(columns) > 5 {
		widths := boxScoreColumnWidths(detail, columns)
		if boxScoreLineWidth(widths) <= maxWidth {
			break
		}
		columns = columns[:len(columns)-1]
	}
	return columns
}

func boxScoreColumnWidths(detail BoxScoreDetail, columns []boxScoreColumn) []int {
	widths := make([]int, len(columns))
	for index := range widths {
		widths[index] = 4
	}
	if len(widths) > 0 {
		widths[0] = 12
	}
	for index, column := range columns {
		updateColumnWidth(widths, index, column.header)
	}
	for _, team := range []TeamStats{detail.Team1, detail.Team2} {
		for _, player := range team.Players {
			for index, column := range columns {
				updateColumnWidth(widths, index, column.value(player))
			}
		}
	}
	return widths
}

func boxScoreHeaderValues(columns []boxScoreColumn) []string {
	values := make([]string, 0, len(columns))
	for _, column := range columns {
		values = append(values, column.header)
	}
	return values
}

func boxScorePlayerValues(player PlayerStat, columns []boxScoreColumn) []string {
	values := make([]string, 0, len(columns))
	for _, column := range columns {
		values = append(values, column.value(player))
	}
	return values
}

func boxScoreLineWidth(widths []int) int {
	width := 0
	for index, columnWidth := range widths {
		width += columnWidth
		if index > 0 {
			width += 2
		}
	}
	return width
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
