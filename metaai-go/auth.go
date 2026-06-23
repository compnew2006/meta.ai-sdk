package metaai

import (
	"context"
	"io"
	"net/http"
)

// scrapeAccessToken fetches https://meta.ai with the current cookies and extracts
// the ecto1: access token from the page HTML. Challenge handling is a TODO
// since the live capture did not encounter a challenge page).
//
// Returns ("", nil) when the page loads but no token is present.
func (c *Client) scrapeAccessToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, MetaAIHomeURL, nil)
	if err != nil {
		return "", err
	}
	attachCookies(req, c.cookies)
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	html := string(body)

	// If challenge-page detection succeeds, surface
	// ErrRegionBlocked — automated challenge resolution is not yet implemented.
	if cu := detectChallengePage(html); cu != "" {
		return "", ErrRegionBlocked
	}
	if resp.StatusCode >= 400 {
		return "", ErrNotAuthenticated
	}
	return extractAccessToken(html), nil
}
