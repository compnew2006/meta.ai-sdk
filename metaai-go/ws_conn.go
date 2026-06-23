package metaai

import (
	"context"
	"time"

	"github.com/smart-studio/metaai-go/internal/transport"
)

// clippyConn is the SDK's handle on a connected clippy WebSocket.
// It wraps internal/transport.Conn so the root package stays decoupled from
// gorilla/websocket.
type clippyConn struct {
	tc *transport.Conn
}

// dialClippy opens a clippy WebSocket using the client's credentials.
// Returns ErrAccessTokenMissing when there is no ecto1: token to auth with.
func (c *Client) dialClippy(ctx context.Context) (*clippyConn, error) {
	if c.accessToken == "" {
		return nil, ErrAccessTokenMissing
	}
	c.logger.Debugf("clippy: dialing clippy WebSocket...")
	tc, err := transport.Dial(ctx, transport.DialOptions{
		AccessToken:  c.accessToken,
		CookieHeader: c.CookieHeader(),
		UserAgent:    c.cfg.UserAgent,
		Origin:       MetaAIOrigin,
		DialTimeout:  30 * time.Second,
	})
	if err != nil {
		c.logger.Debugf("clippy: dial failure: %v", err)
		return nil, err
	}
	c.logger.Debugf("clippy: successfully connected clippy WebSocket")
	return &clippyConn{tc: tc}, nil
}

// closeClippy closes the cached clippy connection if any.
func (c *Client) closeClippy() {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	if c.ws != nil {
		c.logger.Debugf("clippy: closing clippy connection")
		_ = c.ws.tc.Close()
		c.ws = nil
	}
}
