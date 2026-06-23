package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/smart-studio/metaai-go"
	"github.com/smart-studio/metaai-go/internal/upload"
)

// fakeClient is a deterministic restClient used by the handler tests.
type fakeClient struct {
	chatReply          string
	chatErr            error
	streamParts        []string
	lastConvID         string
	analyzeStreamParts []string
	uploadMediaID      string
	uploadErr          error
	imageURLs          []string
	imageErr           error
	videoURLs          []string
	videoErr           error
	extendURLs         []string
	extendErr          error
	extendCalledID     string
}

func (f *fakeClient) Chat(ctx context.Context, message string, opts *metaai.ChatOptions) (string, error) {
	if opts != nil && opts.ConversationID != "" {
		f.lastConvID = opts.ConversationID
	} else if opts != nil && opts.NewConversation {
		f.lastConvID = "new-conv-id"
	} else if f.lastConvID == "" {
		f.lastConvID = "default-conv-id"
	}
	if f.chatErr != nil {
		return "", f.chatErr
	}
	if f.chatReply != "" {
		return f.chatReply, nil
	}
	return "ok", nil
}

// LastConversationID echoes the conversation id the handler should surface.
func (f *fakeClient) LastConversationID() string { return f.lastConvID }

func (f *fakeClient) StreamChat(ctx context.Context, message string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk {
	out := make(chan metaai.ChatChunk, 8)
	go func() {
		defer close(out)
		parts := f.streamParts
		if len(parts) == 0 {
			parts = []string{"hel", "lo"}
		}
		for _, p := range parts {
			out <- metaai.ChatChunk{Text: p}
		}
	}()
	return out
}

func (f *fakeClient) UploadImage(ctx context.Context, path string) (*upload.Result, error) {
	if f.uploadErr != nil {
		return &upload.Result{Error: f.uploadErr.Error()}, f.uploadErr
	}
	return &upload.Result{Success: true, MediaID: f.uploadMediaID, MimeType: "image/png"}, nil
}

func (f *fakeClient) AnalyzeImage(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) (string, error) {
	// Mirror Chat's conversation-id behavior so multi-turn tests can assert it.
	if opts != nil && opts.ConversationID != "" {
		f.lastConvID = opts.ConversationID
	} else {
		f.lastConvID = "analyze-conv-id"
	}
	return "analysis", nil
}

// AnalyzeImageStream mirrors StreamChat's fake: emits one ChatChunk per part
// from analyzeStreamParts (falls back to "anal", "ysis" when unset).
func (f *fakeClient) AnalyzeImageStream(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk {
	out := make(chan metaai.ChatChunk, 8)
	go func() {
		defer close(out)
		parts := f.analyzeStreamParts
		if len(parts) == 0 {
			parts = []string{"anal", "ysis"}
		}
		for _, p := range parts {
			out <- metaai.ChatChunk{Text: p}
		}
	}()
	return out
}

func (f *fakeClient) GenerateImage(ctx context.Context, prompt, orientation string, numImages int) (*metaai.GenerationResult, error) {
	if f.imageErr != nil {
		return &metaai.GenerationResult{Status: "FAILED", Error: f.imageErr.Error()}, f.imageErr
	}
	urls := f.imageURLs
	if len(urls) == 0 {
		urls = []string{"https://img.example/x.png"}
	}
	return &metaai.GenerationResult{Success: true, Status: "READY", Prompt: prompt, URLs: urls, MediaIDs: []string{"111"}, ConversationID: "c1"}, nil
}

func (f *fakeClient) GenerateVideo(ctx context.Context, prompt string) (*metaai.GenerationResult, error) {
	if f.videoErr != nil {
		return &metaai.GenerationResult{Status: "FAILED", Error: f.videoErr.Error()}, f.videoErr
	}
	urls := f.videoURLs
	if len(urls) == 0 {
		urls = []string{"https://vid.example/x.mp4"}
	}
	return &metaai.GenerationResult{Success: true, Status: "READY", Prompt: prompt, URLs: urls, MediaIDs: []string{"222"}, ConversationID: "c2"}, nil
}

func (f *fakeClient) ExtendVideo(ctx context.Context, sourceMediaID string) (*metaai.GenerationResult, error) {
	f.extendCalledID = sourceMediaID
	if f.extendErr != nil {
		return &metaai.GenerationResult{Status: "FAILED", Error: f.extendErr.Error()}, f.extendErr
	}
	urls := f.extendURLs
	if len(urls) == 0 {
		urls = []string{"https://vid.example/extended.mp4"}
	}
	return &metaai.GenerationResult{Success: true, Status: "READY", URLs: urls, MediaIDs: []string{"333"}}, nil
}

func newTestServer(f *fakeClient, token string) *Server {
	s := &Server{client: f, token: token, addr: ":0", jobs: newJobStore()}
	return s
}

func doJSON(t *testing.T, ts *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, ts.URL+path, r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestHealthAndIndex(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp := doJSON(t, ts, "GET", "/healthz", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("health status=%d", resp.StatusCode)
	}
	var h HealthResponse
	json.NewDecoder(resp.Body).Decode(&h)
	resp.Body.Close()
	if h.Status != "ok" {
		t.Errorf("status=%q", h.Status)
	}

	resp2 := doJSON(t, ts, "GET", "/", nil)
	if resp2.StatusCode != 200 {
		t.Fatalf("index status=%d", resp2.StatusCode)
	}
	resp2.Body.Close()
}

func TestAuthReject(t *testing.T) {
	s := newTestServer(&fakeClient{}, "secret")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/chat", ChatRequest{Message: "hi"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status=%d want 401", resp.StatusCode)
	}
}

func TestChat(t *testing.T) {
	s := newTestServer(&fakeClient{chatReply: "hello"}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/chat", ChatRequest{Message: "hi"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var out ChatResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || out.Message != "hello" {
		t.Errorf("bad response: %+v", out)
	}
}

func TestChatStream(t *testing.T) {
	s := newTestServer(&fakeClient{streamParts: []string{"a", "b"}}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/chat", ChatRequest{Message: "hi", Stream: true})
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "data: ") {
		t.Errorf("no SSE data lines:\n%s", body)
	}
}

// TestChatConversationIDResumed verifies the multi-conversation wiring: a
// request carrying a conversation_id must (1) be threaded through to the SDK
// and (2) echo that id back in the response. A request without an id (new
// conversation) must get an assigned id echoed back.
func TestChatConversationIDResumed(t *testing.T) {
	s := newTestServer(&fakeClient{chatReply: "hi"}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// 1. Resume an existing conversation: the echoed id must match the request.
	resp := doJSON(t, ts, "POST", "/chat", ChatRequest{Message: "more", ConversationID: "conv-abc"})
	defer resp.Body.Close()
	var out ChatResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || out.ConversationID != "conv-abc" {
		t.Fatalf("resume: expected conversation_id=conv-abc, got %+v", out)
	}

	// 2. New conversation (no id): server assigns one and echoes it back.
	resp2 := doJSON(t, ts, "POST", "/chat", ChatRequest{Message: "fresh"})
	defer resp2.Body.Close()
	var out2 ChatResponse
	json.NewDecoder(resp2.Body).Decode(&out2)
	if !out2.Success || out2.ConversationID == "" {
		t.Fatalf("new: expected a non-empty conversation_id, got %+v", out2)
	}
}

func TestImage(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/image", ImageRequest{Prompt: "a cat", Orientation: "SQUARE"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var out ImageResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || len(out.ImageURLs) == 0 {
		t.Errorf("bad response: %+v", out)
	}
}

func TestVideo(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/video", VideoRequest{Prompt: "waves"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var out VideoResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || len(out.VideoURLs) == 0 {
		t.Errorf("bad response: %+v", out)
	}
}

func TestVideoExtend(t *testing.T) {
	f := &fakeClient{}
	s := newTestServer(f, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/video/extend", ExtendVideoRequest{MediaID: "999"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if f.extendCalledID != "999" {
		t.Errorf("extend not called with 999, got %q", f.extendCalledID)
	}
}

func TestVideoAsyncAndJob(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp := doJSON(t, ts, "POST", "/video/async", VideoRequest{Prompt: "sunset"})
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		t.Fatalf("async status=%d", resp.StatusCode)
	}
	var jobResp AsyncJobResponse
	json.NewDecoder(resp.Body).Decode(&jobResp)
	if !jobResp.Success || jobResp.JobID == "" {
		t.Fatalf("bad async response: %+v", jobResp)
	}

	// Poll the job until completed.
	completed := false
	for i := 0; i < 50; i++ {
		jr := doJSON(t, ts, "GET", "/video/jobs/"+jobResp.JobID, nil)
		var st JobStatusResponse
		json.NewDecoder(jr.Body).Decode(&st)
		jr.Body.Close()
		if st.Status == "completed" {
			completed = true
			if st.Result == nil || len(st.Result.VideoURLs) == 0 {
				t.Errorf("completed job has no video URLs")
			}
			break
		}
		if st.Status == "failed" {
			t.Fatalf("job failed: %s", st.Error)
		}
	}
	if !completed {
		t.Errorf("job never completed")
	}

	// Unknown job returns 404.
	jr := doJSON(t, ts, "GET", "/video/jobs/doesnotexist", nil)
	defer jr.Body.Close()
	if jr.StatusCode != http.StatusNotFound {
		t.Errorf("unknown job status=%d want 404", jr.StatusCode)
	}
}

func TestUpload(t *testing.T) {
	f := &fakeClient{uploadMediaID: "m123"}
	s := newTestServer(f, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Build a multipart upload with a tiny PNG body.
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("file", "test.png")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte{0x89, 'P', 'N', 'G'})
	mw.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("upload status=%d", resp.StatusCode)
	}
	var out UploadResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || out.MediaID != "m123" {
		t.Errorf("bad upload response: %+v", out)
	}
	if out.FileName != "test.png" {
		t.Errorf("filename=%q want test.png", out.FileName)
	}
}

func TestUnknownPath404(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "GET", "/bogus", nil)
	defer resp.Body.Close()
	// When the Jenta SPA is embedded (prod build), deep links fall back to
	// index.html (200) — that's correct SPA behaviour, not an error. When only
	// the dev marker is present, unknown paths return 404 via the JSON index.
	if jentaHasIndex() {
		if resp.StatusCode != http.StatusOK {
			t.Errorf("with embedded SPA, /bogus status=%d want 200 (deep-link fallback)", resp.StatusCode)
		}
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status=%d want 404", resp.StatusCode)
	}
}

func TestAnalyze(t *testing.T) {
	f := &fakeClient{}
	s := newTestServer(f, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// 1. Valid request
	resp := doJSON(t, ts, "POST", "/analyze", AnalyzeRequest{
		MediaID:  "m123",
		Question: "What is this?",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	var out AnalyzeResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || out.Message != "analysis" {
		t.Errorf("bad analyze response: %+v", out)
	}

	// 2. Missing fields
	resp2 := doJSON(t, ts, "POST", "/analyze", AnalyzeRequest{
		MediaID: "",
	})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("status=%d want 400", resp2.StatusCode)
	}
}

// TestAnalyzeStream verifies the SSE streaming branch of /analyze mirrors
// /chat: Content-Type is text/event-stream and each ChatChunk is flushed as a
// "data:" frame, in order.
func TestAnalyzeStream(t *testing.T) {
	s := newTestServer(&fakeClient{analyzeStreamParts: []string{"hel", "lo"}}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp := doJSON(t, ts, "POST", "/analyze", AnalyzeRequest{MediaID: "m1", Question: "q", Stream: true})
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type=%q want text/event-stream", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "data: ") {
		t.Errorf("no SSE data lines:\n%s", bodyStr)
	}
	// Both streamed parts must appear, in order, as success payloads.
	if !strings.Contains(bodyStr, `"message":"hel"`) || !strings.Contains(bodyStr, `"message":"lo"`) {
		t.Errorf("expected both chunks in stream:\n%s", bodyStr)
	}
}

// TestAnalyzeMultiTurn verifies the analyze multi-turn wiring: turn 1 sends a
// media_id (no conversation_id) and gets a conversation_id echoed back; turn 2
// sends only the conversation_id (no media_id) as a text-only follow-up and the
// server accepts it (no 400) and echoes the same conversation_id.
func TestAnalyzeMultiTurn(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Turn 1: first question with an image.
	resp := doJSON(t, ts, "POST", "/analyze", AnalyzeRequest{MediaID: "img-1", Question: "what is this?"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("turn1 status=%d", resp.StatusCode)
	}
	var out AnalyzeResponse
	json.NewDecoder(resp.Body).Decode(&out)
	if !out.Success || out.ConversationID == "" {
		t.Fatalf("turn1: expected a conversation_id, got %+v", out)
	}
	convID := out.ConversationID

	// Turn 2: follow-up question, only the conversation_id (no media_id).
	resp2 := doJSON(t, ts, "POST", "/analyze", AnalyzeRequest{ConversationID: convID, Question: "and the color?"})
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("turn2 status=%d (follow-up with only conversation_id must be accepted)", resp2.StatusCode)
	}
	var out2 AnalyzeResponse
	json.NewDecoder(resp2.Body).Decode(&out2)
	if !out2.Success || out2.ConversationID != convID {
		t.Fatalf("turn2: expected conversation_id=%s, got %+v", convID, out2)
	}
}

func TestUIEmbedRouteRemoved(t *testing.T) {
	// The legacy /ui/ Vue dashboard mount was removed. /ui/ is no longer a
	// dedicated route — it now falls through to the catch-all "/" handler
	// (SMART Studio SPA in prod, JSON service index in dev). Either way it must
	// NOT serve the old Vue dashboard, whose index.html contained the literal
	// asset path "/ui/assets/".
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ui/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// /ui/ may 301 to /ui (Go ServeMux cleans trailing slash); follow it.
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if strings.Contains(bodyStr, "/ui/assets/") {
		t.Errorf("/ui/ still serving legacy Vue dashboard (has /ui/assets/ references): %q",
			bodyStr[:min(120, len(bodyStr))])
	}
}

// TestCORSPreflight verifies the CORS middleware answers OPTIONS with 204 and
// the configured Allow-Origin/Headers, and that GET responses carry the
// Allow-Origin header too. This is the dev path (Vite at :3000 → API at :8000);
// in prod the SPA is embedded same-origin and CORS is a no-op, so skip when
// the SPA is embedded in this build.
func TestCORSPreflight(t *testing.T) {
	if jentaHasIndex() {
		t.Skip("Jenta SPA is embedded (same-origin prod build) — CORS is a no-op; skipped")
	}
	s := newTestServer(&fakeClient{}, "")
	s.cors = "http://localhost:3000"
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Preflight: OPTIONS /chat → 204 + CORS headers.
	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/chat", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status=%d want 204", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Allow-Origin=%q want http://localhost:3000", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Errorf("Allow-Headers=%q missing Authorization", got)
	}

	// Actual GET /healthz also carries the header (proves the wrap covers all routes).
	resp2, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if got := resp2.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("GET Allow-Origin=%q want http://localhost:3000", got)
	}
}

// TestCORSDisabledInProd verifies the prod behaviour: when the Jenta SPA is
// embedded (same-origin), CORS headers are NOT attached, even if s.cors is set.
func TestCORSDisabledInProd(t *testing.T) {
	if !jentaHasIndex() {
		t.Skip("Jenta SPA not embedded — prod CORS behaviour is N/A")
	}
	s := newTestServer(&fakeClient{}, "")
	s.cors = "http://localhost:3000" // should be ignored
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("prod should not emit CORS headers, got Allow-Origin=%q", got)
	}
}

// TestImageFetchRejectsNonFbcdn verifies the /image/fetch host allow-list at the
// HTTP level (security: this endpoint must NOT become an open proxy / SSRF
// vector). Allowed-host behavior is covered by TestIsFbcdnHost to avoid real
// network calls here.
func TestImageFetchRejectsNonFbcdn(t *testing.T) {
	s := newTestServer(&fakeClient{}, "")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	for _, bad := range []string{
		"http://evil.example.com/x.png",
		"https://fbcdn.net.evil.com/x.png", // suffix on attacker domain
		"https://scontent.fbcdn.net.evil.com/x.png",
	} {
		resp, err := http.Get(ts.URL + "/image/fetch?url=" + url.QueryEscape(bad))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("url=%q status=%d want 400; body=%s", bad, resp.StatusCode, body)
		}
	}

	// Missing url param is also a 400.
	resp, err := http.Get(ts.URL + "/image/fetch")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing url: status=%d want 400", resp.StatusCode)
	}
}

// TestIsFbcdnHost unit-tests the host allow-list without any network. This is
// the security boundary for /image/fetch.
func TestIsFbcdnHost(t *testing.T) {
	allowed := []string{
		"scontent.xx.fbcdn.net",
		"video-abc.xx.fbcdn.net",
		"scontent.cdninstagram.com",
	}
	for _, h := range allowed {
		if !isFbcdnHost(h) {
			t.Errorf("isFbcdnHost(%q)=false, want true", h)
		}
	}
	rejected := []string{
		"evil.example.com",
		"fbcdn.net.evil.com",
		"scontent.fbcdn.net.evil.com",
		"localhost",
		"127.0.0.1",
		"169.254.169.254", // metadata endpoint — must never be fetchable
	}
	for _, h := range rejected {
		if isFbcdnHost(h) {
			t.Errorf("isFbcdnHost(%q)=true, want false", h)
		}
	}
}
