// Package metaai provides an idiomatic Go client for Meta AI.
//
// It provides programmatic access to Meta AI (https://meta.ai): chat over the
// clippy WebSocket binary protocol, image/video generation, image upload/analysis,
// conversation history, and topic-based session management.
//
// The primary entry point is Client, constructed with NewClient:
//
//	client, err := metaai.NewClient(
//	    metaai.WithCookies(map[string]string{"datr": "...", "ecto_1_sess": "..."}),
//	    metaai.WithAccessToken("ecto1:..."),
//	)
//	resp, err := client.Chat(ctx, "Hello", nil)
//
// Wire-format ground truth for the clippy protocol is documented in
// docs/protocol.md and derived from a live capture made on 2026-06-19.
package metaai

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/smart-studio/metaai-go/internal/upload"
)

// Client is the Meta AI SDK entry point. Construct one with NewClient.
//
// A Client is safe for concurrent use after construction; the internal WebSocket
// connection (used for chat) is guarded by a mutex and lazily established.
type Client struct {
	cfg  *Config
	http *http.Client

	// logger receives debug output for best-effort operations.
	logger Logger

	// Auth state (resolved at construction).
	cookies     map[string]string
	accessToken string
	isAuthed    bool

	// Chat session state.
	externalConversationID string
	offlineThreadingID     string
	topics                 map[string]string // topic → conversation id
	currentTopic           string

	// chat-default overrides per topic (topic → {thinking,instant,mode}).
	topicConfig   map[string]topicOverride
	topicConfigMu sync.RWMutex

	// Lazily-connected clippy WebSocket (guarded by wsMu).
	ws   *clippyConn
	wsMu sync.Mutex

	// wsConvID tracks which conversation the current WS connection belongs to.
	// Messages in the same conversation reuse the connection for context.
	wsConvID string

	// attachMedia holds optional image attachment for the next chat message
	// (set by AnalyzeImage, cleared after the message is sent).
	attachMedia *attachInfo
}

// topicOverride holds per-topic default overrides for chat config.
type topicOverride struct {
	Thinking *bool
	Instant  *bool
	Mode     *string
}

// NewClient constructs a Client from options. See the Option setters for the
// available configuration; env vars (META_AI_*) fill in anything left unset.
//
// NewClient does NOT perform any network I/O. Cookies and the access token are
// resolved from options/env at construction; scraping meta.ai for a missing
// access token happens lazily on first chat/upload (see EnsureAccessToken).
func NewClient(opts ...Option) (*Client, error) {
	cfg := newConfig(opts)
	c := &Client{
		cfg:         cfg,
		http:        cfg.httpClient(),
		logger:      cfg.Logger,
		cookies:     cfg.Cookies,
		accessToken: cfg.AccessToken,
		topics:      map[string]string{},
		topicConfig: map[string]topicOverride{},
	}
	if c.logger == nil {
		c.logger = NewDefaultLogger()
	}
	if c.cookies == nil {
		c.cookies = map[string]string{}
	}
	// Authentication can use account credentials or browser cookies.
	c.isAuthed = (cfg.FBEmail != "" && cfg.FBPassword != "") || len(c.cookies) > 0
	return c, nil
}

// Cookies returns a copy of the resolved cookie map.
func (c *Client) Cookies() map[string]string {
	out := make(map[string]string, len(c.cookies))
	for k, v := range c.cookies {
		out[k] = v
	}
	return out
}

// CookieHeader returns the cookies formatted as an HTTP Cookie header value.
func (c *Client) CookieHeader() string { return cookieHeader(c.cookies) }

// AccessToken returns the resolved ecto1: access token (may be empty until
// EnsureAccessToken succeeds).
func (c *Client) AccessToken() string { return c.accessToken }

// IsAuthed reports whether the client has enough credentials to attempt requests.
func (c *Client) IsAuthed() bool { return c.isAuthed }

// EnsureAccessToken fetches the ecto1: access token from meta.ai page HTML if it
// is not already set. Safe to call repeatedly; no-op once a token is present.
// Token discovery uses the authenticated home-page response.
func (c *Client) EnsureAccessToken(ctx context.Context) error {
	if c.accessToken != "" {
		return nil
	}
	if len(c.cookies) == 0 {
		return ErrNotAuthenticated
	}
	tok, err := c.scrapeAccessToken(ctx)
	if err != nil {
		return err
	}
	if tok == "" {
		return ErrAccessTokenMissing
	}
	c.accessToken = tok
	return nil
}

// UploadImage uploads an image file to Meta AI's rupload CDN and returns the
// media id usable in subsequent Chat/AnalyzeImage calls.
//
// Result/error contract: the returned error is authoritative. On the error
// path the *upload.Result is diagnostic metadata (Error/ErrorType describe the
// failure for logging) and Success is false; it is never a success when
// err != nil.
func (c *Client) UploadImage(ctx context.Context, path string) (*upload.Result, error) {
	if err := c.EnsureAccessToken(ctx); err != nil {
		return &upload.Result{Error: err.Error()}, err
	}
	// rupload works with OAuth + ecto_1_sess cookie (confirmed live 2026-06-19).
	// Use a plain http.Client (not the proxy-configured c.http) — the cloned
	// transport from c.http can interfere with rupload's auth.
	ruploadCookie := ""
	if sess, ok := c.cookies["ecto_1_sess"]; ok {
		ruploadCookie = "ecto_1_sess=" + sess
	}
	u := &upload.Uploader{
		HTTP:         &http.Client{Timeout: 60 * time.Second},
		AccessToken:  c.accessToken,
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		CookieHeader: ruploadCookie,
	}
	return u.Upload(ctx, path, 3)
}
