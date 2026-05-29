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
