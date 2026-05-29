package main

import (
	"strings"
	"testing"
)

func TestParseMatchesParsesLivingPendingAndEndedGames(t *testing.T) {
	html := `
<div class="list_box">
  <div class="team_vs_a">
    <div class="team_vs_a_1"><div class="txt"><span>108</span><span>
      湖人
    </span></div></div>
    <div class="team_vs_a_2"><div class="txt"><span>105</span><span>凯尔特人</span></div></div>
  </div>
  <div class="team_vs_c"><div class="b"><p>Q4 02:31</p></div></div>
  <div class="table_choose">
    <a href="https://nba.hupu.com/games/boxscore/1">数据统计</a>
    <a href="https://nba.hupu.com/games/lives/1">文字直播</a>
  </div>
</div>
<div class="list_box">
  <div class="team_vs_a">
    <div class="team_vs_a_1"><div class="txt"><span></span><span>勇士</span></div></div>
    <div class="team_vs_a_2"><div class="txt"><span></span><span>灰熊</span></div></div>
  </div>
  <div class="team_vs_b"><div class="b">未开始 10:00</div></div>
</div>
<div class="list_box">
  <div class="team_vs_a">
    <div class="team_vs_a_1"><div class="txt"><span>99</span><span>掘金</span></div></div>
    <div class="team_vs_a_2"><div class="txt"><span>101</span><span>太阳</span></div></div>
  </div>
  <div class="team_vs_c"><div class="b"><p>已结束</p></div></div>
  <div class="table_choose">
    <a href="https://nba.hupu.com/games/boxscore/2">数据统计</a>
    <a href="https://nba.hupu.com/games/lives/2">文字直播</a>
  </div>
</div>`

	matches, err := ParseMatches(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseMatches returned error: %v", err)
	}

	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}

	living := matches[0]
	if living.Team1Name != "湖人" || living.Team2Name != "凯尔特人" {
		t.Fatalf("unexpected living teams: %#v", living)
	}
	if living.Team1Score != "108" || living.Team2Score != "105" {
		t.Fatalf("unexpected living score: %#v", living)
	}
	if living.Status != StatusLiving || living.MatchTime != "Q4 02:31" {
		t.Fatalf("unexpected living status/time: %#v", living)
	}
	if living.LiveURL != "https://nba.hupu.com/games/lives/1" || living.DataStatisticsURL != "https://nba.hupu.com/games/boxscore/1" {
		t.Fatalf("unexpected living urls: %#v", living)
	}

	pending := matches[1]
	if pending.Status != StatusPending || pending.ScoreText() != "-" {
		t.Fatalf("unexpected pending match: %#v", pending)
	}

	ended := matches[2]
	if ended.Status != StatusEnded || ended.ScoreText() != "99 - 101" {
		t.Fatalf("unexpected ended match: %#v", ended)
	}
}

func TestParseMatchesSkipsMalformedFinishedGamesWithoutLiveURL(t *testing.T) {
	html := `
<div class="list_box">
  <div class="team_vs_a">
    <div class="team_vs_a_1"><div class="txt"><span>90</span><span>热火</span></div></div>
    <div class="team_vs_a_2"><div class="txt"><span>88</span><span>公牛</span></div></div>
  </div>
  <div class="team_vs_c"><div class="b"><p>已结束</p></div></div>
</div>`

	matches, err := ParseMatches(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseMatches returned error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected malformed finished game to be skipped, got %#v", matches)
	}
}

func TestParseLiveDetailParsesRowsAndFinishState(t *testing.T) {
	html := `
<div class="yuece_num_b"><a>文字直播</a></div>
<div class="gamecenter_content_l">
  <div class="table_list_live"><table><tr><td>old</td></tr></table></div>
  <div class="table_list_live">
    <table>
      <tr><td>Q4 02:31</td><td>詹姆斯 三分命中</td><td>108-105</td></tr>
      <tr><td>Q4 02:45</td><td>塔图姆 罚球命中</td><td>105-105</td></tr>
    </table>
  </div>
</div>`

	detail, err := ParseLiveDetail(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseLiveDetail returned error: %v", err)
	}

	if detail.IsFinished {
		t.Fatalf("expected live detail to be unfinished: %#v", detail)
	}
	if len(detail.LiveTextRows) != 2 {
		t.Fatalf("expected 2 live rows, got %#v", detail.LiveTextRows)
	}
	if detail.LiveTextRows[0] != "Q4 02:31 | 詹姆斯 三分命中 | 108-105" {
		t.Fatalf("unexpected first row: %q", detail.LiveTextRows[0])
	}
}

func TestParseBoxScoreDetailParsesTeamAndPlayerStats(t *testing.T) {
	html := `
<div class="team_vs">
  <div class="team_a">湖人</div>
  <div class="team_b">凯尔特人</div>
</div>
<div class="gamecenter_content_l">
  <div class="table_list_live">
    <table>
      <tr><th>球员</th><th>时间</th><th>得分</th><th>篮板</th><th>助攻</th><th>抢断</th><th>盖帽</th><th>失误</th><th>犯规</th></tr>
      <tr><td>詹姆斯</td><td>35</td><td>28</td><td>7</td><td>9</td><td>2</td><td>1</td><td>3</td><td>2</td></tr>
      <tr><td>戴维斯</td><td>33</td><td>24</td><td>12</td><td>3</td><td>1</td><td>4</td><td>2</td><td>3</td></tr>
    </table>
  </div>
  <div class="table_list_live">
    <table>
      <tr><th>球员</th><th>时间</th><th>得分</th><th>篮板</th><th>助攻</th><th>抢断</th><th>盖帽</th><th>失误</th><th>犯规</th></tr>
      <tr><td>塔图姆</td><td>37</td><td>31</td><td>8</td><td>5</td><td>1</td><td>0</td><td>4</td><td>2</td></tr>
    </table>
  </div>
</div>`

	detail, err := ParseBoxScoreDetail(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseBoxScoreDetail returned error: %v", err)
	}

	if detail.Team1.Name != "湖人" || detail.Team2.Name != "凯尔特人" {
		t.Fatalf("unexpected team names: %#v", detail)
	}
	if len(detail.Team1.Players) != 2 || len(detail.Team2.Players) != 1 {
		t.Fatalf("unexpected players: %#v", detail)
	}

	player := detail.Team1.Players[0]
	if player.Name != "詹姆斯" || player.Minutes != "35" || player.Points != "28" || player.Rebounds != "7" || player.Assists != "9" {
		t.Fatalf("unexpected first player: %#v", player)
	}
	if player.Steals != "2" || player.Blocks != "1" || player.Turnovers != "3" || player.Fouls != "2" {
		t.Fatalf("unexpected defensive/misc stats: %#v", player)
	}
}
