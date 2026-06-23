package metaai

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// Root package coverage tests.
// Mock only at HTTP boundary, test behavior from caller perspective.

func TestTruncateLongStringAddsEllipsis(t *testing.T) {
	long := "this is a very long string that exceeds the limit"
	got := truncate(long, 10)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncated should end with ellipsis, got %q", got)
	}
}

func TestTruncateShortStringUnchanged(t *testing.T) {
	short := "hi"
	got := truncate(short, 10)
	if got != short {
		t.Errorf("got %q, want %q", got, short)
	}
}

func TestResolveConversationIDTopicOverridden(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	c.topics["mytopic"] = "topic-conv-id"
	got := c.resolveConversationID(&ChatOptions{Topic: "mytopic"})
	if got != "topic-conv-id" {
		t.Errorf("got %q, want 'topic-conv-id'", got)
	}
}

func TestResolveConversationIDNewConversationReturnsEmpty(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	c.externalConversationID = "existing-conv"
	got := c.resolveConversationID(&ChatOptions{NewConversation: true})
	if got != "" {
		t.Errorf("got %q, want empty for new conversation", got)
	}
}

func TestResolveConversationIDUsesExternalWhenNoOpts(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	c.externalConversationID = "stored-conv"
	got := c.resolveConversationID(nil)
	if got != "stored-conv" {
		t.Errorf("got %q, want 'stored-conv'", got)
	}
}

func TestResolveChatConfigPriorityCallOverridesTopic(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	c.SetTopicConfig("t1", boolPtrCov(true), nil, strPtrCov("learn"))
	thinking, _, _ := c.resolveChatConfig(&ChatOptions{
		Topic:    "t1",
		Thinking: boolPtrCov(false),
	})
	if thinking {
		t.Error("call override should win over topic: thinking should be false")
	}
}

func TestResolveChatConfigDefaultsFromConstructor(t *testing.T) {
	c, _ := NewClient(
		WithCookies(map[string]string{"datr": "d"}),
		WithDefaultThinking(true),
		WithDefaultMode("analyze"),
	)
	thinking, _, mode := c.resolveChatConfig(nil)
	if !thinking {
		t.Error("default thinking should be true")
	}
	if mode != "analyze" {
		t.Errorf("mode = %q, want 'analyze'", mode)
	}
}

func TestIsValidModeAcceptsKnownModes(t *testing.T) {
	for _, m := range []string{ModeLearn, ModeAnalyze, ModeCreateImage, ModeCreateVideo} {
		if !isValidMode(m) {
			t.Errorf("mode %q should be valid", m)
		}
	}
}

func TestIsValidModeRejectsUnknown(t *testing.T) {
	if isValidMode("invalid") {
		t.Error("invalid mode should be rejected")
	}
}

func TestToIntConvertsNumbers(t *testing.T) {
	cases := []struct {
		input any
		want  int
		ok    bool
	}{
		{200, 200, true},
		{int64(200), 200, true},
		{float64(200), 200, true},
		{"200", 0, false},
	}
	for _, tc := range cases {
		got, ok := toInt(tc.input)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("toInt(%v) = %d, %v; want %d, %v", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAttachCookiesSetsHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	attachCookies(req, map[string]string{"datr": "d", "ecto_1_sess": "s"})
	if req.Header.Get("Cookie") == "" {
		t.Error("cookie header not set")
	}
}

func TestAttachCookiesSkipsWhenEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	attachCookies(req, nil)
	if req.Header.Get("Cookie") != "" {
		t.Error("cookie header should be empty for nil cookies")
	}
}

func TestNewClientWithProxySetsTransport(t *testing.T) {
	c, err := NewClient(
		WithCookies(map[string]string{"datr": "d"}),
		WithProxy("http://proxy.example.com:8080"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.http == nil {
		t.Error("HTTP client not set")
	}
}

func TestNewClientWithCustomTimeout(t *testing.T) {
	c, _ := NewClient(
		WithCookies(map[string]string{"datr": "d"}),
		WithHTTPTimeout(5000000000),
	)
	if c.cfg.HTTPTimeout != 5000000000 {
		t.Error("timeout not set")
	}
}

func TestNewClientWithDotEnvDisabled(t *testing.T) {
	c, _ := NewClient(
		WithCookies(map[string]string{"datr": "d"}),
		WithDotEnv(false, ""),
	)
	if c.cfg.LoadDotEnv {
		t.Error("dotenv should be disabled")
	}
}

func TestCloseClippyWhenNoConnection(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	c.closeClippy() // should not panic
}

func TestGenerateImageHandlesAuthError(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	_, err := c.GenerateImage(context.Background(), "test", "SQUARE", 1)
	if err == nil {
		t.Error("expected error without access token")
	}
}

func TestGenerateVideoHandlesAuthError(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	_, err := c.GenerateVideo(context.Background(), "test")
	if err == nil {
		t.Error("expected error without access token")
	}
}

func TestGetConversationHistoryHandlesAuthError(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	_, err := c.GetConversationHistory(context.Background(), 5, 0)
	if err == nil {
		t.Error("expected error without access token")
	}
}

func TestAnalyzeImageRejectsMissingMediaID(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	_, err := c.AnalyzeImage(context.Background(), "", "", "what?", nil)
	if err == nil {
		t.Error("expected error for missing media ID")
	}
}

func TestVibesHandlesAuthError(t *testing.T) {
	c, _ := NewClient(WithCookies(map[string]string{"datr": "d"}))
	_, err := c.Vibes(context.Background(), VibesList, "")
	if err == nil {
		t.Error("expected error without access token")
	}
}

func TestGraphQLErrorCodeExtraction(t *testing.T) {
	e := graphQLError{
		Message:    "test",
		Extensions: map[string]any{"code": "GRAPHQL_VALIDATION_FAILED"},
	}
	if e.code() != "GRAPHQL_VALIDATION_FAILED" {
		t.Errorf("code = %q", e.code())
	}
}

func TestGraphQLErrorCodeEmptyWhenNoExtensions(t *testing.T) {
	e := graphQLError{Message: "test"}
	if e.code() != "" {
		t.Errorf("code = %q, want empty", e.code())
	}
}

// helpers (prefixed to avoid collision with session_test.go)
func boolPtrCov(b bool) *bool    { return &b }
func strPtrCov(s string) *string { return &s }
