package main

type MatchStatus string

const (
	StatusPending MatchStatus = "PENDING"
	StatusLiving  MatchStatus = "ING"
	StatusEnded   MatchStatus = "END"
)

type Match struct {
	Team1Name         string
	Team2Name         string
	Team1Score        string
	Team2Score        string
	Status            MatchStatus
	MatchTime         string
	LiveURL           string
	DataStatisticsURL string
}

func (m Match) Title() string {
	if m.Team1Name == "" && m.Team2Name == "" {
		return ""
	}
	return m.Team1Name + " - " + m.Team2Name
}

func (m Match) ScoreText() string {
	if m.Team1Score == "" || m.Team2Score == "" {
		return "-"
	}
	return m.Team1Score + " - " + m.Team2Score
}

func (m Match) IsLive() bool {
	return m.Status == StatusLiving
}

type LiveDetail struct {
	IsFinished   bool
	LiveTextRows []string
}

type BoxScoreDetail struct {
	Team1 TeamStats
	Team2 TeamStats
}

func (d BoxScoreDetail) IsEmpty() bool {
	return len(d.Team1.Players) == 0 && len(d.Team2.Players) == 0
}

type TeamStats struct {
	Name    string
	Players []PlayerStat
}

type PlayerStat struct {
	Name      string
	Minutes   string
	Points    string
	Rebounds  string
	Assists   string
	Steals    string
	Blocks    string
	Turnovers string
	Fouls     string
}
