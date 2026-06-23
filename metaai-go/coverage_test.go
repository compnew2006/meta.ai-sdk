package metaai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smart-studio/metaai-go/internal/transport"
)

type testRoundTrip func(*http.Request) (*http.Response, error)

func (f testRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
func testClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient(WithCookies(map[string]string{"datr": "d", "ecto_1_sess": "s"}), WithAccessToken("ecto1:token"), WithDotEnv(false, ""))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClientOptionsAccessorsAndSessionTransitions(t *testing.T) {
	mode := "learn"
	c, err := NewClient(WithCookies(map[string]string{"datr": "d"}), WithAccessToken("tok"), WithFBCredentials("e", "p"), WithDefaultMode(mode), WithDefaultThinking(true), WithHTTPTimeout(time.Second), WithUserAgent("ua"), WithDotEnv(false, ""))
	if err != nil {
		t.Fatal(err)
	}
	cookies := c.Cookies()
	cookies["datr"] = "changed"
	if c.Cookies()["datr"] != "d" || c.AccessToken() != "tok" || !c.IsAuthed() || !strings.Contains(c.CookieHeader(), "datr=d") {
		t.Fatal("client accessors/config invalid")
	}
	thinking, instant := true, false
	c.NewTopic("topic", &thinking, &instant, &mode)
	if c.CurrentTopic() != "topic" || len(c.ListTopics()) != 0 {
		t.Fatal("topic not registered")
	}
	c.SetTopicConfig("topic", nil, nil, nil)
	c.SetTopic("missing")
	ct, ci, cm := c.resolveChatConfig(&ChatOptions{Topic: "topic"})
	if !ct || ci || cm != mode {
		t.Fatalf("config=%v %v %q", ct, ci, cm)
	}
	c.SetTopic("topic")
	if c.resolveConversationID(&ChatOptions{NewConversation: true}) != "" {
		t.Fatal("new conversation should reset id")
	}
}

func TestEnsureAccessTokenHandlesPageOutcomes(t *testing.T) {
	cases := []struct {
		name, body string
		status     int
		network    bool
		want       error
	}{
		{"token", `<script>{\"accessToken\":\"ecto1:fresh\"}</script>`, 200, false, nil}, {"missing", "<html></html>", 200, false, ErrAccessTokenMissing}, {"challenge", "executeChallenge fetch('/__rd_verify/x')", 200, false, ErrRegionBlocked}, {"unauthorized", "denied", 401, false, ErrNotAuthenticated}, {"network", "", 0, true, errors.New("offline")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t)
			c.accessToken = ""
			c.http = &http.Client{Transport: testRoundTrip(func(*http.Request) (*http.Response, error) {
				if tc.network {
					return nil, errors.New("offline")
				}
				return response(tc.status, tc.body), nil
			})}
			err := c.EnsureAccessToken(context.Background())
			if tc.want == nil {
				if err != nil || c.AccessToken() != "ecto1:fresh" {
					t.Fatalf("token=%q err=%v", c.AccessToken(), err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want.Error()) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	c, _ := NewClient(WithDotEnv(false, ""))
	if !errors.Is(c.EnsureAccessToken(context.Background()), ErrNotAuthenticated) {
		t.Fatal("expected unauthenticated")
	}
}

func TestGraphQLBoundaryHeadersAndErrors(t *testing.T) {
	c := testClient(t)
	var auth string
	c.http = &http.Client{Transport: testRoundTrip(func(r *http.Request) (*http.Response, error) {
		auth = r.Header.Get("Authorization")
		return response(200, `{"data":{"ok":true}}`), nil
	})}
	if _, err := c.graphqlRequest(context.Background(), "doc", map[string]any{"x": 1}); err != nil || auth != "ecto1:token" {
		t.Fatalf("auth=%q err=%v", auth, err)
	}
	if _, err := c.graphqlRequestOAuth(context.Background(), "doc", nil); err != nil || auth != "OAuth ecto1:token" {
		t.Fatalf("auth=%q err=%v", auth, err)
	}
	responses := []struct{ name, body, want string }{{"validation", `{"errors":[{"message":"bad","extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}]}`, ErrGraphQLValidation.Error()}, {"graphql", `{"errors":[{"message":"bad"}]}`, "graphql error"}, {"malformed", "not json", "decode graphql"}}
	for _, tc := range responses {
		t.Run(tc.name, func(t *testing.T) {
			c.http = &http.Client{Transport: testRoundTrip(func(*http.Request) (*http.Response, error) { return response(200, tc.body), nil })}
			if _, err := c.graphqlRequest(context.Background(), "doc", nil); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	c.http = &http.Client{Transport: testRoundTrip(func(*http.Request) (*http.Response, error) { return response(200, `{"data":{}}`), nil })}
	if err := c.RegisterConversation(context.Background(), "c"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetConversationMode(context.Background(), "c", true); err != nil {
		t.Fatal(err)
	}
	if err := c.SetConversationMode(context.Background(), "c", false); err != nil {
		t.Fatal(err)
	}
	c.markConversationSeen(context.Background(), "c", "m")
	c.markConversationSeen(context.Background(), "", "")
}

func TestHistoryGenerationVibesAndAnalyzeUseBoundaryResponses(t *testing.T) {
	c := testClient(t)
	c.http = &http.Client{Transport: testRoundTrip(func(r *http.Request) (*http.Response, error) {
		return response(200, `{"data":{"data":{"viewer":{"conversations":{"count":0,"edges":[{"node":{"id":"1","name":"Fallback","last_updated":"now","last_message":{"text":"hi"}}}]}}}}}`), nil
	})}
	h, err := c.GetConversationHistory(context.Background(), 0, 2)
	if err != nil || h.Total != 1 || h.Limit != 20 || h.Conversations[0].Title != "Fallback" {
		t.Fatalf("history=%+v err=%v", h, err)
	}
	if _, err := c.SearchConversations(context.Background(), "q", 0); err == nil {
		t.Fatal("SearchConversations should return error (not yet implemented)")
	}
	// GenerateImage/GenerateVideo now go through WS chat, not HTTP SSE.
	// Skip testing them here (they need a live WS connection).
	for _, a := range []VibesAction{VibesList, VibesSet, VibesGet} {
		if _, err := c.Vibes(context.Background(), a, "calm"); err == nil {
			t.Fatal("Vibes should return error (doc IDs are placeholders)")
		}
	}
	if _, err := c.AnalyzeImage(context.Background(), "", "", "", nil); err == nil {
		t.Fatal("expected missing media error")
	}
}

func TestChatStreamsThroughLocalWebSocket(t *testing.T) {
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_, _, _ = ws.ReadMessage()
		_, _, _ = ws.ReadMessage()
		_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+`{"type":"patch","operations":[{"op":"delta","value":"hello"}]}`))
		body := []byte(`{"code":201}`)
		_ = ws.WriteMessage(websocket.BinaryMessage, append([]byte{0x0f, 0, 0, byte(len(body)), 0, 0}, body...))
	}))
	defer s.Close()
	tc, err := transport.Dial(context.Background(), transport.DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	c := testClient(t)
	c.ws = &clippyConn{tc: tc}
	c.wsConvID = "" // force reuse regardless of conversation ID
	got, err := c.Chat(context.Background(), "prompt", nil)
	if err != nil || got != "hello" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	c.closeClippy()
}

// TestChatDedupsDuplicatePatchDeltas guards against the garbled-stream bug where
// Meta AI's clippy transport sends every streaming delta twice (two consecutive
// frames with identical delta text). Without dedup the streamed text doubles up
// ("مسسمساء النور!"). The test sends each delta twice and asserts each appears
// once in the assembled Chat result.
func TestChatDedupsDuplicatePatchDeltas(t *testing.T) {
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_, _, _ = ws.ReadMessage() // CONNECT
		_, _, _ = ws.ReadMessage() // CHAT
		// Send each delta TWICE (simulating the live duplicate-frame behavior).
		for _, d := range []string{"مساء ", "النور"} {
			payload := `{"type":"patch","operations":[{"op":"delta","value":"` + d + `"}]}`
			_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+payload))
			_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+payload)) // duplicate
		}
		// Final full response + a duplicate to signal stream completion.
		full := `{"type":"full","response":{"sections":[{"view_model":{"primitive":{"text":"مساء النور"}}}]}}`
		_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+full))
		_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+full))
	}))
	defer s.Close()
	tc, err := transport.Dial(context.Background(), transport.DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	c := testClient(t)
	c.ws = &clippyConn{tc: tc}
	c.wsConvID = ""
	got, err := c.Chat(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	// The result must be exactly "مساء النور" — NOT the doubled "مساء مساء النورالنور".
	if got != "مساء النور" {
		t.Fatalf("dedup failed: got=%q (want %q — duplicate patch deltas should be collapsed)", got, "مساء النور")
	}
	c.closeClippy()
}

func localChatTransport(t *testing.T, text string) *transport.Conn {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_, _, _ = ws.ReadMessage()
		_, _, _ = ws.ReadMessage()
		_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+`{"type":"patch","operations":[{"op":"delta","value":"`+text+`"}]}`))
		body := []byte(`{"code":201}`)
		_ = ws.WriteMessage(websocket.BinaryMessage, append([]byte{0x0f, 0, 0, byte(len(body)), 0, 0}, body...))
	}))
	t.Cleanup(s.Close)
	tc, err := transport.Dial(context.Background(), transport.DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

func TestAnalyzeImageStreamsAttachmentThroughChat(t *testing.T) {
	c := testClient(t)
	c.ws = &clippyConn{tc: localChatTransport(t, "described")}
	c.wsConvID = "" // force reuse
	instant := true
	got, err := c.AnalyzeImage(context.Background(), "", "media-id", "question", &ChatOptions{Instant: &instant})
	if err != nil || got != "described" || c.attachMedia != nil {
		t.Fatalf("got=%q attachment=%+v err=%v", got, c.attachMedia, err)
	}
}


// delayedChatTransport is a fake clippy server that holds its first response
// frame for `delay` before sending it. It mimics the high first-token latency
// of vision/image turns (AnalyzeImage), where Meta AI can spend a long time
// running the model before emitting any frame.
func delayedChatTransport(t *testing.T, delay time.Duration, text string) *transport.Conn {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_, _, _ = ws.ReadMessage() // CONNECT
		_, _, _ = ws.ReadMessage() // CHAT
		time.Sleep(delay)          // simulate slow first-token (vision turn)
		_ = ws.WriteMessage(websocket.BinaryMessage, []byte("xx"+`{"type":"patch","operations":[{"op":"delta","value":"`+text+`"}]}`))
		body := []byte(`{"code":201}`)
		_ = ws.WriteMessage(websocket.BinaryMessage, append([]byte{0x0f, 0, 0, byte(len(body)), 0, 0}, body...))
	}))
	t.Cleanup(s.Close)
	tc, err := transport.Dial(context.Background(), transport.DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return tc
}

// TestAnalyzeImageSurvivesSlowVisionFirstToken guards the recvLoop read-timeout
// regression: a vision turn whose first frame arrives after the OLD 30s gap
// deadline must still return the answer, not "chat produced no response".
// The 35s delay exceeds the old 30s timeout and stays well under the 120s one.
func TestAnalyzeImageSurvivesSlowVisionFirstToken(t *testing.T) {
	c := testClient(t)
	c.ws = &clippyConn{tc: delayedChatTransport(t, 35*time.Second, "described")}
	c.wsConvID = ""
	instant := true
	got, err := c.AnalyzeImage(context.Background(), "", "media-id", "question", &ChatOptions{Instant: &instant})
	if err != nil || got != "described" {
		t.Fatalf("got=%q err=%v (want %q, nil — slow vision first token must not be treated as no-response)", got, err, "described")
	}
}

// TestChatSurfacesErrorWhenServerNeverResponds guards the recvLoop failure
// path: when no frame ever arrives and the socket times out, Chat() must
// surface the underlying error rather than the generic "no response" message.
func TestChatSurfacesErrorWhenServerNeverResponds(t *testing.T) {
	c := testClient(t)
	// A server that accepts CONNECT+CHAT then closes without replying.
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		_, _, _ = ws.ReadMessage()
		_, _, _ = ws.ReadMessage()
		// Immediately close: next RecvBinary gets an EOF.
	}))
	t.Cleanup(s.Close)
	tc, err := transport.Dial(context.Background(), transport.DialOptions{Endpoint: "ws" + strings.TrimPrefix(s.URL, "http"), AccessToken: "ecto1:x", DialTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	c.ws = &clippyConn{tc: tc}
	c.wsConvID = ""
	_, err = c.Chat(context.Background(), "hi", nil)
	if err == nil {
		t.Fatal("expected error when server never responds, got nil")
	}
	if !strings.Contains(err.Error(), "no response") && !strings.Contains(err.Error(), "stream ended") && !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected a stream/socket error, got %v", err)
	}
}

func TestUtilityAndSessionBoundaryCases(t *testing.T) {
	if extractValue("a[start]value[end]", "[start]", "[end]") != "value" || extractValue("none", "x", "y") != "" || extractValue("x", "x", "y") != "" {
		t.Fatal("extractValue contract")
	}
	if len(numericIDString(19)) != 19 || quotePath("a b") != `"a b"` || len(newConversationID()) != 36 {
		t.Fatal("identifier contract")
	}
	for _, m := range []string{ModeLearn, ModeAnalyze, ModeCreateImage, ModeCreateVideo} {
		if !isValidMode(m) {
			t.Fatalf("mode %q invalid", m)
		}
	}
	for _, v := range []any{int(1), int64(2), float64(3)} {
		if _, ok := toInt(v); !ok {
			t.Fatalf("cannot convert %T", v)
		}
	}
	if _, ok := toInt("x"); ok {
		t.Fatal("string converted")
	}
	c := testClient(t)
	c.NewTopic("", nil, nil, nil)
	c.topics["t"] = "id"
	c.SetTopic("t")
	if c.GetTopic("") != "id" || c.GetTopic("t") != "id" {
		t.Fatal("topic lookup")
	}
	thinking, instant, mode := true, true, ModeAnalyze
	c.SetTopicConfig("t", &thinking, &instant, &mode)
	if len(c.ListTopics()) != 1 {
		t.Fatal("topic copy")
	}
	c.accessToken = ""
	if _, err := c.dialClippy(context.Background()); !errors.Is(err, ErrAccessTokenMissing) {
		t.Fatalf("dial error=%v", err)
	}
}

func TestGenerationFinalStatesAndFailures(t *testing.T) {
	// GenerationResult zero values and field accessors
	ready := &GenerationResult{URLs: []string{"u"}, Status: "READY", Success: true}
	if !ready.Success || ready.Status != "READY" {
		t.Fatalf("ready=%+v", ready)
	}
	c := testClient(t)
	// GenerateImage/Video go through WS chat now; test auth rejection
	c.accessToken = ""
	c.cookies = nil
	if _, err := c.GenerateImage(context.Background(), "p", "SQUARE", 1); err == nil {
		t.Fatal("expected image auth error")
	}
	if _, err := c.GenerateVideo(context.Background(), "p"); err == nil {
		t.Fatal("expected video auth error")
	}
	if _, err := c.UploadImage(context.Background(), "missing"); err == nil {
		t.Fatal("expected upload auth error")
	}
}

func TestRemainingPublicErrorContracts(t *testing.T) {
	c, err := NewClient(WithCookies(map[string]string{"datr": "d"}), WithAccessToken("ecto1:x"), WithProxy("http://127.0.0.1:9"), WithDefaultInstant(true), WithDotEnv(false, ""))
	if err != nil || !c.cfg.DefaultInstant || c.http.Transport == nil {
		t.Fatalf("client=%+v err=%v", c, err)
	}
	if _, err := c.UploadImage(context.Background(), t.TempDir()+"/missing.jpg"); err == nil {
		t.Fatal("expected missing upload file")
	}
	if _, err := c.AnalyzeImage(context.Background(), t.TempDir()+"/missing.jpg", "", "q", nil); err == nil {
		t.Fatal("expected upload failure")
	}
	c.http = &http.Client{Transport: testRoundTrip(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })}
	if _, err := c.graphqlRequest(context.Background(), "doc", nil); err == nil {
		t.Fatal("expected graphql network error")
	}
	if _, err := c.GetConversationHistory(context.Background(), 1, 0); err == nil {
		t.Fatal("expected history network error")
	}
	if _, err := c.SearchConversations(context.Background(), "q", 1); err == nil {
		t.Fatal("expected search network error")
	}
	c.topics["topic"] = "conversation"
	c.SetTopic("topic")
	if c.GetTopic("") != "conversation" || c.GetTopic("topic") != "conversation" {
		t.Fatal("topic lookup")
	}
	c.ws = &clippyConn{tc: localChatTransport(t, "unused")}
	c.onWSFailure()
	if c.ws != nil {
		t.Fatal("failed websocket retained")
	}
}

func TestChatReportsEmptyServerResponse(t *testing.T) {
	c := testClient(t)
	c.ws = &clippyConn{tc: localChatTransport(t, "")}
	if got, err := c.Chat(context.Background(), "prompt", &ChatOptions{NewConversation: true, Topic: "fresh"}); err == nil || got != "" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if c.GetTopic("fresh") == "" {
		t.Fatal("new topic conversation was not retained")
	}
}

func TestCancelledStreamClosesWithoutTerminalChunk(t *testing.T) {
	c := testClient(t)
	c.ws = &clippyConn{tc: localChatTransport(t, "ignored")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for chunk := range c.StreamChat(ctx, "prompt", nil) {
		t.Fatalf("unexpected chunk: %+v", chunk)
	}
}

func TestAuthenticatedFeaturesRejectAnonymousClient(t *testing.T) {
	c, _ := NewClient(WithDotEnv(false, ""))
	ctx := context.Background()
	if r, err := c.GenerateImage(ctx, "p", "", 0); err == nil || r.Status != "FAILED" {
		t.Fatalf("image=%+v err=%v", r, err)
	}
	if r, err := c.GenerateVideo(ctx, "p"); err == nil || r.Status != "FAILED" {
		t.Fatalf("video=%+v err=%v", r, err)
	}
	if r, err := c.GetConversationHistory(ctx, 0, 0); err == nil || r.Error == "" {
		t.Fatalf("history=%+v err=%v", r, err)
	}
	if _, err := c.SearchConversations(ctx, "q", 0); err == nil {
		t.Fatal("expected search auth error")
	}
	if _, err := c.Vibes(ctx, VibesList, ""); err == nil {
		t.Fatal("expected vibes auth error")
	}
}

func TestChatRejectsInvalidConfigurationBeforeNetwork(t *testing.T) {
	c := testClient(t)
	yes := true
	for _, opts := range []*ChatOptions{{Thinking: &yes, Instant: &yes}, {Mode: func() *string { s := "bad"; return &s }()}} {
		ch := c.StreamChat(context.Background(), "x", opts)
		chunk := <-ch
		if chunk.Err == nil {
			t.Fatal("expected configuration error")
		}
	}
	c.accessToken = ""
	c.cookies = nil
	if _, err := c.Chat(context.Background(), "x", nil); err == nil {
		t.Fatal("expected auth error")
	}
}
