package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smart-studio/metaai-go/internal/uuid"
)

func TestDialExchangesBinaryFramesWithBrowserHeaders(t *testing.T) {
	requests := make(chan *http.Request, 1)
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Clone(context.Background())
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		mt, data, err := ws.ReadMessage()
		if err == nil {
			_ = ws.WriteMessage(mt, append([]byte("reply:"), data...))
		}
	}))
	defer s.Close()
	endpoint := "ws" + strings.TrimPrefix(s.URL, "http")
	c, err := Dial(context.Background(), DialOptions{Endpoint: endpoint, AccessToken: "ecto1:t", CookieHeader: "a=b", UserAgent: "ua", Origin: "https://meta.ai", ExtraHeaders: map[string]string{"X-Test": "yes"}, DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	r := <-requests
	if r.Header.Get("Cookie") != "a=b" || r.Header.Get("User-Agent") != "ua" || r.Header.Get("X-Test") != "yes" {
		t.Fatalf("headers=%#v", r.Header)
	}
	q := r.URL.Query()
	if q.Get("Authorization") != "ecto1:t" || q.Get("x-dgw-appid") == "" || q.Get("x-dgw-app-clippy-request-id") == "" {
		t.Fatalf("query=%v", q)
	}
	if err := c.SendBinary([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	data, mt, err := c.RecvBinary(time.Second)
	if err != nil || mt != websocket.BinaryMessage || string(data) != "reply:ping" {
		t.Fatalf("data=%q type=%d err=%v", data, mt, err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-c.Done():
	default:
		t.Fatal("done channel open")
	}
}

func TestDialRejectsMissingTokenAndHandshakeFailure(t *testing.T) {
	if _, err := Dial(context.Background(), DialOptions{}); err == nil {
		t.Fatal("expected token error")
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "denied", http.StatusUnauthorized) }))
	defer s.Close()
	_, err := Dial(context.Background(), DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err == nil || !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("err=%v", err)
	}
}

func TestFailedConnectionRejectsFurtherOperations(t *testing.T) {
	c := &Conn{failed: true, done: make(chan struct{})}
	if !errors.Is(c.SendBinary(nil), ErrClosed) {
		t.Fatal("send should be closed")
	}
	if _, _, err := c.RecvBinary(0); !errors.Is(err, ErrClosed) {
		t.Fatal("recv should be closed")
	}
	u, _ := url.Parse(buildClippyURL("wss://example.test/ws", "ecto1:a", "rid"))
	if u.Query().Get("x-dgw-authtype") != "15:0" {
		t.Fatal(u.String())
	}
	if id := uuid.V4(); len(id) != 36 {
		t.Fatalf("id=%q", id)
	}
}
