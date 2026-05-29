package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultGamesURL  = "https://nba.hupu.com/games"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"
)

type Fetcher interface {
	FetchMatches(time.Time) ([]Match, error)
	FetchLiveDetail(Match) (LiveDetail, error)
	FetchBoxScoreDetail(Match) (BoxScoreDetail, error)
}

type HTTPFetcher struct {
	gamesURL string
	client   *http.Client
}

func NewHTTPFetcher(gamesURL string, client *http.Client) *HTTPFetcher {
	if gamesURL == "" {
		gamesURL = defaultGamesURL
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &HTTPFetcher{
		gamesURL: gamesURL,
		client:   client,
	}
}

func (f *HTTPFetcher) FetchMatches(date time.Time) ([]Match, error) {
	body, err := f.get(f.gamesURLForDate(date))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return ParseMatches(body)
}

func (f *HTTPFetcher) gamesURLForDate(date time.Time) string {
	if date.IsZero() || sameDay(date, time.Now()) {
		return f.gamesURL
	}
	return strings.TrimRight(f.gamesURL, "/") + "/" + date.Format("2006-01-02")
}

func (f *HTTPFetcher) FetchLiveDetail(match Match) (LiveDetail, error) {
	if match.LiveURL == "" {
		return LiveDetail{}, fmt.Errorf("match has no live url")
	}

	body, err := f.get(match.LiveURL)
	if err != nil {
		return LiveDetail{}, err
	}
	defer body.Close()

	return ParseLiveDetail(body)
}

func (f *HTTPFetcher) FetchBoxScoreDetail(match Match) (BoxScoreDetail, error) {
	if match.DataStatisticsURL == "" {
		return BoxScoreDetail{}, fmt.Errorf("match has no data statistics url")
	}

	body, err := f.get(match.DataStatisticsURL)
	if err != nil {
		return BoxScoreDetail{}, err
	}
	defer body.Close()

	return ParseBoxScoreDetail(body)
}

func (f *HTTPFetcher) get(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("GET %s returned status %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}
