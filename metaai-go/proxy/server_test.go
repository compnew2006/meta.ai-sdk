package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smart-studio/metaai-go"
)

// fakeClient is a deterministic chatClient used by the handler tests.
type fakeClient struct {
	chatReply   string
	chatErr     error
	gotMessage  string
	gotOpts     *metaai.ChatOptions
	streamParts []string
	genImage    *metaai.GenerationResult
	genVideo    *metaai.GenerationResult
}

func (f *fakeClient) Chat(ctx context.Context, message string, opts *metaai.ChatOptions) (string, error) {
	f.gotMessage = message
	f.gotOpts = opts
	if f.chatErr != nil {
		return "", f.chatErr
	}
	if f.chatReply != "" {
		return f.chatReply, nil
	}
	return "pong", nil
}

func (f *fakeClient) StreamChat(ctx context.Context, message string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk {
	f.gotMessage = message
	f.gotOpts = opts
	out := make(chan metaai.ChatChunk, 8)
	go func() {
		defer close(out)
		parts := f.streamParts
		if len(parts) == 0 {
			parts = []string{"hel", "lo", " world"}
		}
		for _, p := range parts {
			out <- metaai.ChatChunk{Text: p}
		}
	}()
	return out
}

func (f *fakeClient) AnalyzeImage(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) (string, error) {
	f.gotMessage = question
	f.gotOpts = opts
	return "image-analysis", nil
}

func (f *fakeClient) GenerateImage(ctx context.Context, prompt, orientation string, numImages int) (*metaai.GenerationResult, error) {
	if f.genImage != nil {
		return f.genImage, nil
	}
	return &metaai.GenerationResult{Success: true, Status: "READY", URLs: []string{"https://img.example/x.png"}}, nil
}

func (f *fakeClient) GenerateVideo(ctx context.Context, prompt string) (*metaai.GenerationResult, error) {
	if f.genVideo != nil {
		return f.genVideo, nil
	}
	return &metaai.GenerationResult{Success: true, Status: "READY", URLs: []string{"https://vid.example/x.mp4"}}, nil
}

func newTestServer(f *fakeClient, token string) *Server {
	return &Server{
		client: f,
		token:  token,
		addr:   ":0",
		http:   &http.Client{},
	}
}

func TestResolveModel(t *testing.T) {
	cases := map[string]string{
		"meta-ai":           "meta-ai",
		"meta-ai-think":     "meta-ai-think",
		"gpt-4o":            "meta-ai",
		"claude-3-5-sonnet": "meta-ai",
		"claude-3-opus":     "meta-ai-think",
		"something-fast":    "meta-ai-fast",
		"any-image-model":   "meta-ai-image",
		"my-video-gen":      "meta-ai-video",
		"":                  "meta-ai",
		"  META-AI-THINK  ": "meta-ai-think",
	}
	for in, want := range cases {
		got := resolveModel(in).Name
		if got != want {
			t.Errorf("resolveModel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAssembleOaiTranscript(t *testing.T) {
	msgs := []oaiMessage{
		{Role: "system", Content: "Be terse."},
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
		{Role: "user", Content: "bye"},
	}
	text, imgs := assembleOaiTranscript(nil, msgs)
	if !strings.Contains(text, "[System]") || !strings.Contains(text, "[User]") || !strings.Contains(text, "[Assistant]") {
		t.Errorf("transcript missing role labels: %q", text)
	}
	if !strings.Contains(text, "Be terse.") || !strings.Contains(text, "bye") {
		t.Errorf("transcript missing content: %q", text)
	}
	if len(imgs) != 0 {
		t.Errorf("expected no images, got %d", len(imgs))
	}
	// Order: system content should appear before the last user content.
	if strings.Index(text, "Be terse.") > strings.Index(text, "bye") {
		t.Errorf("system should precede last user: %q", text)
	}
}

func TestAssembleOaiTranscriptImages(t *testing.T) {
	// Image in an earlier user message should be ignored; only the last user
	// message's images count.
	msgs := []oaiMessage{
		{Role: "user", Content: []interface{}{
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://e.example/a.png"}},
			map[string]any{"type": "text", "text": "first"},
		}},
		{Role: "assistant", Content: "ok"},
		{Role: "user", Content: []interface{}{
			map[string]any{"type": "text", "text": "describe this"},
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64,aGVsbG8="}},
		}},
	}
	text, imgs := assembleOaiTranscript(nil, msgs)
	if !strings.Contains(text, "describe this") {
		t.Errorf("last user text missing: %q", text)
	}
	if len(imgs) != 1 {
		t.Fatalf("expected 1 image from last user msg, got %d", len(imgs))
	}
	if len(imgs[0].Data) == 0 {
		t.Errorf("expected data: URL image to decode bytes")
	}
	if imgs[0].MediaType != "image/png" {
		t.Errorf("mime = %q, want image/png", imgs[0].MediaType)
	}
}

func TestAssembleAnthTranscriptSystem(t *testing.T) {
	msgs := []anthMessage{{Role: "user", Content: "ping"}}
	text, imgs := assembleAnthTranscript("You are helpful.", msgs)
	if !strings.Contains(text, "[System]") || !strings.Contains(text, "You are helpful.") {
		t.Errorf("system not folded in: %q", text)
	}
	if len(imgs) != 0 {
		t.Errorf("expected no images, got %d", len(imgs))
	}
}

func TestBearerToken(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/messages", nil)
	r.Header.Set("Authorization", "Bearer secret")
	if got := bearerToken(r); got != "secret" {
		t.Errorf("bearer = %q", got)
	}
	r2 := httptest.NewRequest("POST", "/v1/messages", nil)
	r2.Header.Set("X-Api-Key", "key123")
	if got := bearerToken(r2); got != "key123" {
		t.Errorf("apikey = %q", got)
	}
}

func TestAuthReject(t *testing.T) {
	f := &fakeClient{}
	srv := newTestServer(f, "required-token")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"meta-ai","messages":[{"role":"user","content":"hi"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", resp.StatusCode)
	}
}

func TestChatCompletionsNonStream(t *testing.T) {
	f := &fakeClient{chatReply: "hello there"}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"model":"meta-ai","messages":[{"role":"user","content":"ping"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out oaiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Choices) != 1 || out.Choices[0].Message.Content != "hello there" {
		t.Errorf("unexpected response: %+v", out)
	}
	if out.Choices[0].FinishReason != "stop" {
		t.Errorf("finish_reason = %q", out.Choices[0].FinishReason)
	}
	if f.gotMessage == "" {
		t.Errorf("client did not receive a message")
	}
	if f.gotOpts == nil || !f.gotOpts.NewConversation {
		t.Errorf("expected NewConversation=true, got %+v", f.gotOpts)
	}
}

func TestChatCompletionsStream(t *testing.T) {
	f := &fakeClient{streamParts: []string{"abc", "def"}}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"model":"meta-ai","stream":true,"messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	s := string(raw)
	if !strings.Contains(s, "data: ") {
		t.Errorf("no SSE data lines:\n%s", s)
	}
	if !strings.Contains(s, "[DONE]") {
		t.Errorf("missing [DONE] terminator:\n%s", s)
	}
	if !strings.Contains(s, "\"role\":\"assistant\"") {
		t.Errorf("missing role delta chunk:\n%s", s)
	}
}

func TestMessagesNonStream(t *testing.T) {
	f := &fakeClient{chatReply: "hi from meta"}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"model":"claude-3","max_tokens":100,"system":"be brief","messages":[{"role":"user","content":"hello"}]}`
	resp, err := http.Post(ts.URL+"/v1/messages", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out anthMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Role != "assistant" || out.Type != "message" {
		t.Errorf("bad envelope: %+v", out)
	}
	if len(out.Content) != 1 || out.Content[0].Text != "hi from meta" {
		t.Errorf("bad content: %+v", out.Content)
	}
	if out.StopReason == nil || *out.StopReason != "end_turn" {
		t.Errorf("stop_reason = %v", out.StopReason)
	}
	if !strings.Contains(f.gotMessage, "be brief") {
		t.Errorf("system prompt not folded into upstream message: %q", f.gotMessage)
	}
}

func TestMessagesStream(t *testing.T) {
	f := &fakeClient{streamParts: []string{"x", "y"}}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"model":"meta-ai-think","max_tokens":50,"stream":true,"messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(ts.URL+"/v1/messages", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	s := string(raw)
	for _, ev := range []string{"message_start", "content_block_start", "content_block_delta", "content_block_stop", "message_delta", "message_stop"} {
		if !strings.Contains(s, "event: "+ev) {
			t.Errorf("missing event %q in stream:\n%s", ev, s)
		}
	}
}

func TestChatCompletionsImageGenerationModel(t *testing.T) {
	f := &fakeClient{}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"model":"meta-ai-image","messages":[{"role":"user","content":"a red apple"}]}`
	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out oaiChatResponse
	json.NewDecoder(resp.Body).Decode(&out)
	content, _ := out.Choices[0].Message.Content.(string)
	if len(out.Choices) != 1 || !strings.Contains(content, "https://img.example/x.png") {
		t.Errorf("expected generated image URL in content, got %+v", out)
	}
}

func TestModelsEndpoint(t *testing.T) {
	f := &fakeClient{}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out oaiModelsResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out.Data) != len(modelRegistry) {
		t.Errorf("got %d models, want %d", len(out.Data), len(modelRegistry))
	}
}

func TestHealthAndIndex(t *testing.T) {
	f := &fakeClient{}
	srv := newTestServer(f, "")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp2, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("index status = %d", resp2.StatusCode)
	}
	resp2.Body.Close()
}
