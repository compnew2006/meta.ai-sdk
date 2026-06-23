package htmlscraper

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchConversationHTMLSendsBrowserContext(t *testing.T) {
	s := &Scraper{UserAgent: "ua", CookieHeader: "a=b", HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/prompt/conv" || r.Header.Get("User-Agent") != "ua" || r.Header.Get("Cookie") != "a=b" {
			t.Fatalf("unexpected request: %s %#v", r.URL.Path, r.Header)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("<html>ok</html>")), Header: make(http.Header)}, nil
	})}}
	got, err := s.FetchConversationHTML(context.Background(), "conv")
	if err != nil || got != "<html>ok</html>" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestFetchConversationHTMLPropagatesNetworkFailure(t *testing.T) {
	s := &Scraper{HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })}}
	if _, err := s.FetchConversationHTML(context.Background(), "conv"); err == nil {
		t.Fatal("expected network error")
	}
}

func TestExtractVideoURLsDeduplicatesCapturedURLs(t *testing.T) {
	u := "https://video-abc.xx.fbcdn.net/path/movie.mp4?x=1"
	got := ExtractVideoURLs(`<video src="` + u + `"></video><script>"` + u + `"</script>`)
	if len(got) != 1 || got[0].URL != u {
		t.Fatalf("got %#v", got)
	}
	if got := ExtractVideoURLs("no video"); len(got) != 0 {
		t.Fatalf("got %#v", got)
	}
}
