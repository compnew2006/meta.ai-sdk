// Package transport implements the clippy WebSocket dialer and message streaming.
//
// It connects to wss://gateway.meta.ai/ws/clippy with the exact query-string
// parameters and request headers the live browser uses (Origin, User-Agent,
// Cookie; no Sec-WebSocket-Protocol). Precise browser headers are required to
// avoid pre-auth validation failures.
package transport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smart-studio/metaai-go/internal/uuid"
)

// DefaultEndpoint is the clippy WebSocket URL.
const DefaultEndpoint = "wss://gateway.meta.ai/ws/clippy"

// ChatChunk is one streamed text fragment yielded by Stream.
type ChatChunk struct {
	Text    string
	Sources []map[string]any
	Media   []map[string]any
	// Err is non-nil on a terminal error (connection lost, decode failure).
	Err error
	// Done is true on the final chunk of a response.
	Done bool
}

// DialOptions configures the WebSocket dial.
type DialOptions struct {
	Endpoint      string            // override wss URL (default DefaultEndpoint)
	AccessToken   string            // ecto1:… token (required)
	CookieHeader  string            // HTTP Cookie header value (auth cookies)
	UserAgent     string            // User-Agent header
	Origin        string            // Origin header (default https://meta.ai)
	Authorization string            // optional raw Authorization override
	ExtraHeaders  map[string]string // additional request headers
	DialTimeout   time.Duration     // default 30s
}

// Conn is a connected clippy WebSocket. Safe for concurrent Send/Recv? No —
// clippy frames are request/response on a single connection; callers should
// serialize sends (the Client guards this with wsMu).
type Conn struct {
	ws     *websocket.Conn
	mu     sync.Mutex
	done   chan struct{}
	failed bool // a prior read/write errored; further ops return ErrClosed
}

// Dial connects to the clippy endpoint and returns a Conn.
//
// The URL is built to match the browser capture exactly:
//
//	wss://gateway.meta.ai/ws/clippy
//	  ?x-dgw-appid=1522763855472543&x-dgw-appversion=1.0.0&x-dgw-authtype=15%3A0
//	  &x-dgw-version=5&x-dgw-uuid=0&x-dgw-tier=prod
//	  &Authorization=<urlencoded token>
//	  &x-dgw-app-origin=meta.ai&x-dgw-app-clippy-request-id=<uuid4>
func Dial(ctx context.Context, opts DialOptions) (*Conn, error) {
	if opts.AccessToken == "" {
		return nil, errors.New("transport: AccessToken is required")
	}
	endpoint := opts.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	origin := opts.Origin
	if origin == "" {
		origin = "https://meta.ai"
	}
	ua := opts.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"
	}
	timeout := opts.DialTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	reqID := uuid.V4()
	wsURL := buildClippyURL(endpoint, opts.AccessToken, reqID)

	header := http.Header{}
	header.Set("Origin", origin)
	header.Set("User-Agent", ua)
	if opts.CookieHeader != "" {
		header.Set("Cookie", opts.CookieHeader)
	}
	for k, v := range opts.ExtraHeaders {
		header.Set(k, v)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
		// No Subprotocols — the browser opens a plain WS.
		// gorilla sends a default User-Agent unless we override via header (done above).
	}
	conn, resp, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("transport: ws dial failed: %w (status %d)", err, resp.StatusCode)
		}
		return nil, fmt.Errorf("transport: ws dial failed: %w", err)
	}
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return &Conn{ws: conn, done: make(chan struct{})}, nil
}

// buildClippyURL assembles the clippy WebSocket URL with the captured query params.
func buildClippyURL(endpoint, token, requestID string) string {
	q := url.Values{}
	q.Set("x-dgw-appid", "1522763855472543")
	q.Set("x-dgw-appversion", "1.0.0")
	q.Set("x-dgw-authtype", "15:0")
	q.Set("x-dgw-version", "5")
	q.Set("x-dgw-uuid", "0")
	q.Set("x-dgw-tier", "prod")
	q.Set("Authorization", token)
	q.Set("x-dgw-app-origin", "meta.ai")
	q.Set("x-dgw-app-clippy-request-id", requestID)
	// url.Values.Encode percent-encodes ":" as "%3A" and spaces as "+"; the browser
	// uses "%3A" for authtype/Authorization and the raw token has no spaces. Matches.
	return endpoint + "?" + q.Encode()
}

// SendBinary writes a binary message (clippy frames are binary).
func (c *Conn) SendBinary(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failed {
		return ErrClosed
	}
	if err := c.ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
		c.failed = true
		return err
	}
	return nil
}

// ErrClosed is returned by RecvBinary/SendBinary after the connection has failed.
var ErrClosed = errors.New("transport: connection closed")

// RecvBinary reads the next message as raw bytes with a deadline. After a prior
// read or write errors, subsequent calls return ErrClosed without touching the
// underlying socket (gorilla panics on repeated reads of a failed connection).
func (c *Conn) RecvBinary(timeout time.Duration) ([]byte, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failed {
		return nil, 0, ErrClosed
	}
	if timeout > 0 {
		_ = c.ws.SetReadDeadline(time.Now().Add(timeout))
	}
	msgType, data, err := c.ws.ReadMessage()
	if err != nil {
		c.failed = true
	}
	return data, msgType, err
}

// Close terminates the connection.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	return c.ws.Close()
}

// Done returns a channel closed when Close is called.
func (c *Conn) Done() <-chan struct{} { return c.done }
