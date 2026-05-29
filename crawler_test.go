package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPFetcherFetchMatchesUsesConfiguredClientAndUserAgent(t *testing.T) {
	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.UserAgent()
		_, _ = w.Write([]byte(`
<div class="list_box">
  <div class="team_vs_a">
    <div class="team_vs_a_1"><div class="txt"><span>108</span><span>湖人</span></div></div>
    <div class="team_vs_a_2"><div class="txt"><span>105</span><span>凯尔特人</span></div></div>
  </div>
  <div class="team_vs_c"><div class="b"><p>Q4 02:31</p></div></div>
  <div class="table_choose">
    <a href="boxscore">数据统计</a>
    <a href="live">文字直播</a>
  </div>
</div>`))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(server.URL, &http.Client{Timeout: time.Second})

	matches, err := fetcher.FetchMatches()
	if err != nil {
		t.Fatalf("FetchMatches returned error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %#v", matches)
	}
	if gotUserAgent == "" {
		t.Fatal("expected User-Agent header to be set")
	}
}

func TestHTTPFetcherFetchLiveDetailParsesLiveURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
<div class="yuece_num_b"><a>赛后数据</a></div>
<div class="gamecenter_content_l">
  <div class="table_list_live">
    <table><tr><td>Q4 00:00</td><td>比赛结束</td></tr></table>
  </div>
</div>`))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher("unused", server.Client())
	detail, err := fetcher.FetchLiveDetail(Match{LiveURL: server.URL})
	if err != nil {
		t.Fatalf("FetchLiveDetail returned error: %v", err)
	}
	if !detail.IsFinished {
		t.Fatalf("expected finished live detail: %#v", detail)
	}
	if len(detail.LiveTextRows) != 1 || detail.LiveTextRows[0] != "Q4 00:00 | 比赛结束" {
		t.Fatalf("unexpected live rows: %#v", detail.LiveTextRows)
	}
}

func TestHTTPFetcherFetchLiveDetailRequiresLiveURL(t *testing.T) {
	fetcher := NewHTTPFetcher("unused", &http.Client{Timeout: time.Second})

	if _, err := fetcher.FetchLiveDetail(Match{}); err == nil {
		t.Fatal("expected error for match without live url")
	}
}
