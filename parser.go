package main

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseMatches(r io.Reader) ([]Match, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var matches []Match
	doc.Find(".list_box").Each(func(_ int, s *goquery.Selection) {
		match := Match{
			Team1Name: cleanText(s.Find(".team_vs_a .team_vs_a_1 .txt span:last-child").Text()),
			Team2Name: cleanText(s.Find(".team_vs_a .team_vs_a_2 .txt span:last-child").Text()),
		}

		match.MatchTime = cleanText(s.Find(".team_vs_c .b p").Text())
		if match.MatchTime == "" {
			match.MatchTime = cleanText(s.Find(".team_vs_b .b").Text())
		}

		switch {
		case strings.Contains(match.MatchTime, "未开始"):
			match.Status = StatusPending
		case strings.Contains(match.MatchTime, "已结束"):
			match.Status = StatusEnded
		default:
			match.Status = StatusLiving
		}

		if match.Status != StatusPending {
			match.Team1Score = cleanText(s.Find(".team_vs_a .team_vs_a_1 .txt span:first-child").Text())
			match.Team2Score = cleanText(s.Find(".team_vs_a .team_vs_a_2 .txt span:first-child").Text())
			match.DataStatisticsURL, _ = s.Find(".table_choose a:first-child").Attr("href")
			match.LiveURL, _ = s.Find(".table_choose a:last-child").Attr("href")
		}

		if match.Status == StatusPending || match.LiveURL != "" {
			matches = append(matches, match)
		}
	})

	return matches, nil
}

func ParseLiveDetail(r io.Reader) (LiveDetail, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return LiveDetail{}, err
	}

	finishText := cleanText(doc.Find(".yuece_num_b a:first-child").Text())
	detail := LiveDetail{
		IsFinished: !strings.Contains(finishText, "直播"),
	}

	doc.Find(".gamecenter_content_l .table_list_live:last-child table tr").Each(func(_ int, s *goquery.Selection) {
		var cells []string
		s.Find("td").Each(func(_ int, c *goquery.Selection) {
			text := cleanText(c.Text())
			if text != "" {
				cells = append(cells, text)
			}
		})
		if len(cells) > 0 {
			detail.LiveTextRows = append(detail.LiveTextRows, strings.Join(cells, " | "))
		}
	})

	return detail, nil
}

func ParseBoxScoreDetail(r io.Reader) (BoxScoreDetail, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return BoxScoreDetail{}, err
	}

	detail := BoxScoreDetail{
		Team1: TeamStats{Name: firstNonEmptyText(doc.Find(".team_vs .team_a"), doc.Find(".team_vs div").Eq(0))},
		Team2: TeamStats{Name: firstNonEmptyText(doc.Find(".team_vs .team_b"), doc.Find(".team_vs div").Eq(1))},
	}

	tables := doc.Find(".gamecenter_content_l .table_list_live table")
	detail.Team1.Players = parsePlayerStatsTable(tables.Eq(0))
	detail.Team2.Players = parsePlayerStatsTable(tables.Eq(1))

	return detail, nil
}

func parsePlayerStatsTable(table *goquery.Selection) []PlayerStat {
	if table.Length() == 0 {
		return nil
	}

	var headers []string
	var players []PlayerStat
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("th,td")
		values := make([]string, 0, cells.Length())
		cells.Each(func(_ int, cell *goquery.Selection) {
			values = append(values, cleanText(cell.Text()))
		})
		if len(values) == 0 {
			return
		}

		if i == 0 || row.Find("th").Length() > 0 {
			headers = values
			return
		}

		player := playerStatFromRow(headers, values)
		if player.Name != "" {
			players = append(players, player)
		}
	})

	return players
}

func playerStatFromRow(headers, values []string) PlayerStat {
	player := PlayerStat{}
	for i, value := range values {
		header := ""
		if i < len(headers) {
			header = headers[i]
		}
		switch normalizeStatHeader(header, i) {
		case "name":
			player.Name = value
		case "minutes":
			player.Minutes = value
		case "points":
			player.Points = value
		case "rebounds":
			player.Rebounds = value
		case "assists":
			player.Assists = value
		case "steals":
			player.Steals = value
		case "blocks":
			player.Blocks = value
		case "turnovers":
			player.Turnovers = value
		case "fouls":
			player.Fouls = value
		}
	}
	return player
}

func normalizeStatHeader(header string, index int) string {
	header = strings.ToLower(cleanText(header))
	switch {
	case header == "球员" || strings.Contains(header, "姓名") || strings.Contains(header, "player"):
		return "name"
	case header == "时间" || strings.Contains(header, "分钟") || strings.Contains(header, "min"):
		return "minutes"
	case header == "得分" || header == "分" || strings.Contains(header, "pts"):
		return "points"
	case header == "篮板" || strings.Contains(header, "reb"):
		return "rebounds"
	case header == "助攻" || strings.Contains(header, "ast"):
		return "assists"
	case header == "抢断" || strings.Contains(header, "stl"):
		return "steals"
	case header == "盖帽" || strings.Contains(header, "blk"):
		return "blocks"
	case header == "失误" || strings.Contains(header, "to"):
		return "turnovers"
	case header == "犯规" || strings.Contains(header, "pf"):
		return "fouls"
	}

	switch index {
	case 0:
		return "name"
	case 1:
		return "minutes"
	case 2:
		return "points"
	case 3:
		return "rebounds"
	case 4:
		return "assists"
	case 5:
		return "steals"
	case 6:
		return "blocks"
	case 7:
		return "turnovers"
	case 8:
		return "fouls"
	default:
		return ""
	}
}

func firstNonEmptyText(selections ...*goquery.Selection) string {
	for _, selection := range selections {
		if text := cleanText(selection.Text()); text != "" {
			return text
		}
	}
	return ""
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	return strings.TrimSpace(text)
}
