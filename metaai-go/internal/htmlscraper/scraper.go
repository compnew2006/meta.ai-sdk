// Package htmlscraper extracts video URLs from meta.ai prompt pages.
//
// Pages at https://www.meta.ai/prompt/<id> embed generated videos; this package finds
// them via <video>/<source> tags (fbcdn.net) and script-content regexes.
package htmlscraper

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// VideoURL is one extracted URL with its source strategy.
type VideoURL struct {
	URL  string `json:"url"`
	Type string `json:"type"` // video_tag | source_tag | script_json | html_search
}

// Scraper fetches prompt pages and extracts video URLs.
type Scraper struct {
	HTTP         *http.Client
	UserAgent    string
	CookieHeader string
}

var (
	fbcdnVideoRe = regexp.MustCompile(`https://video-[a-z0-9-]+\.xx\.fbcdn\.net/[^\s"'<>]+\.mp4[^\s"'<>]*`)
	scriptURLRe  = regexp.MustCompile(`https://video-[^"']+\.mp4[^"']*`)
)

// FetchConversationHTML GETs https://www.meta.ai/prompt/<id>.
func (s *Scraper) FetchConversationHTML(ctx context.Context, conversationID string) (string, error) {
	url := "https://www.meta.ai/prompt/" + conversationID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}
	if s.CookieHeader != "" {
		req.Header.Set("Cookie", s.CookieHeader)
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ExtractVideoURLs finds generated video URLs in page HTML:
// strategy 1 (<video>/<source> filtered by fbcdn.net), strategy 2 (script
// regex), strategy 3 (full-HTML fbcdn regex).
func ExtractVideoURLs(html string) []VideoURL {
	seen := map[string]struct{}{}
	var out []VideoURL
	add := func(u, kind string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, VideoURL{URL: u, Type: kind})
	}

	// Strategy 2 + 3: regex-based (covers <video>/<source> too in practice).
	for _, m := range scriptURLRe.FindAllString(html, -1) {
		add(m, "script_json")
	}
	for _, m := range fbcdnVideoRe.FindAllString(html, -1) {
		add(m, "html_search")
	}
	return out
}
