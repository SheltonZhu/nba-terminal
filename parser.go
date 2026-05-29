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

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	return strings.TrimSpace(text)
}
