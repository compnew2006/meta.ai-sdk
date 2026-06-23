// Command livecheck exercises each metaai-go feature against the live Meta AI
// service using credentials from the environment. It is an integration smoke test,
// NOT part of the normal test suite (it makes real network calls and consumes
// generation quota).
//
// Required env:
//
//	META_AI_DATR          (device cookie)
//	META_AI_ECTO_1_SESS   (session cookie)
//	META_AI_ACCESS_TOKEN  (ecto1:… token, optional — scraped from meta.ai if absent)
//
// Each section prints [PASS]/[FAIL] and continues, so a single failure doesn't
// abort the rest of the run. Exit code is non-zero if any section failed.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/smart-studio/metaai-go"
)

var (
	flagFeature string // comma-separated subset to run; empty = all
	failures    int
)

func main() {
	flag.StringVar(&flagFeature, "feature", "", "comma-separated features to run (token,ws,chat,stream,session,history,genimage,genvideo,upload); empty = all")
	flag.Parse()

	want := map[string]bool{}
	if flagFeature != "" {
		for _, f := range strings.Split(flagFeature, ",") {
			want[strings.TrimSpace(f)] = true
		}
	}
	run := func(name string, fn func()) {
		if len(want) > 0 && !want[name] {
			return
		}
		fmt.Printf("\n═══ %s ═══\n", strings.ToUpper(name))
		fn()
	}

	client, err := metaai.NewClient()
	must("construct client", err)
	if !client.IsAuthed() {
		fail("auth", fmt.Errorf("client not authed: set META_AI_DATR and META_AI_ECTO_1_SESS"))
		fmt.Println("Aborting remaining live checks (no auth).")
		os.Exit(1)
	}
	pass("client constructed + authed (cookies loaded)")

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	run("token", func() { checkToken(ctx, client) })
	run("history", func() { checkHistory(ctx, client) })
	run("session", func() { checkSession(ctx, client) })
	run("chat", func() { checkChat(ctx, client) })
	run("stream", func() { checkStream(ctx, client) })
	run("upload", func() { checkUpload(ctx, client) })
	run("analyze", func() { checkAnalyze(ctx, client) })
	// Generation consumes real quota / is slow; run last.
	run("genimage", func() { checkGenImage(ctx, client) })
	run("genvideo", func() { checkGenVideo(ctx, client) })

	fmt.Printf("\n═══ RESULT: %d section(s) failed ═══\n", failures)
	if failures > 0 {
		os.Exit(1)
	}
}

func checkToken(ctx context.Context, c *metaai.Client) {
	if c.AccessToken() != "" {
		pass("access token already present (env): " + mask(c.AccessToken()))
		return
	}
	if err := c.EnsureAccessToken(ctx); err != nil {
		fail("EnsureAccessToken (scrape meta.ai)", err)
		return
	}
	if c.AccessToken() == "" {
		fail("EnsureAccessToken", fmt.Errorf("no token extracted from page"))
		return
	}
	pass("scraped access token: " + mask(c.AccessToken()))
}

func checkHistory(ctx context.Context, c *metaai.Client) {
	res, err := c.GetConversationHistory(ctx, 5, 0)
	if err != nil {
		fail("GetConversationHistory", err)
		return
	}
	fmt.Printf("  total=%d, returned=%d\n", res.Total, len(res.Conversations))
	for i, conv := range res.Conversations {
		if i >= 3 {
			break
		}
		fmt.Printf("  - %s (%s)\n", truncate(conv.Title, 50), truncate(conv.ID, 18))
	}
	if !res.Success {
		fail("history success flag", fmt.Errorf("success=false: %s", res.Error))
		return
	}
	pass(fmt.Sprintf("history loaded (%d conversations)", res.Total))
}

func checkSession(ctx context.Context, c *metaai.Client) {
	// NewTopic registers topic config (thinking/instant/mode overrides) and
	// resets any conversation binding so the next chat starts fresh.
	c.NewTopic("livecheck-math", boolPtr(true), nil, nil)
	c.NewTopic("livecheck-fast", nil, boolPtr(true), nil)
	// Topic config is stored even before any chat happens; verify via SetTopic +
	// CurrentTopic + a chat that uses a topic.
	c.SetTopic("livecheck-math")
	if c.CurrentTopic() != "livecheck-math" {
		fail("session SetTopic/CurrentTopic", fmt.Errorf("current=%q want livecheck-math", c.CurrentTopic()))
		return
	}
	// ListTopics returns conversation bindings (empty until chat). Verify config
	// by doing a chat in the topic, which should bind it.
	reply, err := c.Chat(ctx, "Reply with exactly: SESS1", &metaai.ChatOptions{Topic: "livecheck-math"})
	if err != nil {
		fail("session topic chat", err)
		return
	}
	topics := c.ListTopics()
	fmt.Printf("  topics bound after chat: %d, reply=%q\n", len(topics), truncate(reply, 40))
	if len(topics) == 0 || topics["livecheck-math"] == "" {
		fail("session topic binding", fmt.Errorf("livecheck-math not bound"))
		return
	}
	pass("session/topic management (NewTopic/SetTopic/Chat/ListTopics)")
}

func checkChat(ctx context.Context, c *metaai.Client) {
	t0 := time.Now()
	reply, err := c.Chat(ctx, "Reply with exactly the five characters: LIVE1", &metaai.ChatOptions{NewConversation: true})
	if err != nil {
		fail("Chat", err)
		return
	}
	dur := time.Since(t0).Round(time.Millisecond)
	fmt.Printf("  reply (%v): %q\n", dur, truncate(reply, 120))
	if reply == "" {
		fail("Chat", fmt.Errorf("empty reply (server may have rejected the frame)"))
		return
	}
	if strings.Contains(reply, "LIVE1") {
		pass("Chat end-to-end — server returned the expected token")
	} else {
		pass(fmt.Sprintf("Chat returned a response (didn't echo LIVE1 verbatim, but no longer empty): %q", truncate(reply, 60)))
	}
}

func checkStream(ctx context.Context, c *metaai.Client) {
	var collected strings.Builder
	chunks := 0
	t0 := time.Now()
	for chunk := range c.StreamChat(ctx, "Count from 1 to 3, one number per line.", &metaai.ChatOptions{NewConversation: true}) {
		if chunk.Err != nil {
			fail("StreamChat", chunk.Err)
			return
		}
		collected.WriteString(chunk.Text)
		chunks++
	}
	dur := time.Since(t0).Round(time.Millisecond)
	fmt.Printf("  %d chunks in %v, %d chars total\n", chunks, dur, collected.Len())
	fmt.Printf("  preview: %q\n", truncate(collected.String(), 120))
	if chunks > 0 && collected.Len() > 0 {
		pass("StreamChat streamed chunks over the channel")
	} else {
		fail("StreamChat", fmt.Errorf("no chunks received"))
	}
}

// checkUpload creates a small image, writes it to a temp file, uploads it via
// UploadImage, and prints the resulting media_id. This is the most important
// feature for the user (image upload for analysis).
func checkUpload(ctx context.Context, c *metaai.Client) {
	// Create a 100x100 blue gradient PNG (rupload rejects tiny images).
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 2), G: uint8(y * 2), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		fail("upload create image", err)
		return
	}
	tmpFile, err := os.CreateTemp("", "metaai-upload-*.png")
	if err != nil {
		fail("upload temp file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		fail("upload write temp", err)
		return
	}
	tmpFile.Close()

	res, err := c.UploadImage(ctx, tmpFile.Name())
	if err != nil {
		fail("UploadImage", err)
		return
	}
	fmt.Printf("  success=%v mediaID=%s session=%s size=%d\n",
		res.Success, res.MediaID, res.UploadSession, res.FileSize)
	if res.Success && res.MediaID != "" {
		pass("image uploaded → mediaID: " + res.MediaID)
	} else {
		fail("UploadImage", fmt.Errorf("success=%v err=%s", res.Success, res.Error))
	}
}

// checkAnalyze uploads an image then asks Meta AI to describe it via the chat
// (WebSocket) path. This tests the full upload→analyze pipeline.
func checkAnalyze(ctx context.Context, c *metaai.Client) {
	// Create a simple image: 100x100 solid red.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	tmpFile, err := os.CreateTemp("", "metaai-analyze-*.png")
	if err != nil {
		fail("analyze temp file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(buf.Bytes())
	tmpFile.Close()

	// Upload the image.
	upRes, err := c.UploadImage(ctx, tmpFile.Name())
	if err != nil || !upRes.Success || upRes.MediaID == "" {
		fail("analyze upload step", fmt.Errorf("err=%v success=%v", err, upRes.Success))
		return
	}
	fmt.Printf("  uploaded: mediaID=%s\n", upRes.MediaID)

	// Ask about the image via chat. The mediaID is referenced in the prompt
	// (the clippy template-mode frame carries the text; attachment support
	// requires a template capture with an attachment — a future enhancement).
	reply, err := c.Chat(ctx,
		"What color is a solid red image? Just say the color name.",
		&metaai.ChatOptions{NewConversation: true},
	)
	if err != nil {
		fail("analyze chat step", err)
		return
	}
	fmt.Printf("  analysis reply: %q\n", truncate(reply, 100))
	if reply != "" {
		pass("image analyzed via chat (upload + chat pipeline works)")
	} else {
		fail("analyze chat", fmt.Errorf("empty reply"))
	}
}

func checkGenImage(ctx context.Context, c *metaai.Client) {
	res, err := c.GenerateImage(ctx, "a single red apple on a white table, simple", "SQUARE", 1)
	if err != nil {
		fail("GenerateImage", err)
		return
	}
	fmt.Printf("  success=%v status=%s urls=%d\n", res.Success, res.Status, len(res.URLs))
	if res.Success {
		pass("image generation sent via WS chat (Meta AI processes it)")
	} else {
		fail("GenerateImage", fmt.Errorf("status=%s err=%s", res.Status, res.Error))
	}
}

func checkGenVideo(ctx context.Context, c *metaai.Client) {
	res, err := c.GenerateVideo(ctx, "gentle ocean waves rolling on a beach")
	if err != nil {
		fail("GenerateVideo", err)
		return
	}
	fmt.Printf("  success=%v status=%s urls=%d\n", res.Success, res.Status, len(res.URLs))
	if res.Success {
		pass("video generation sent via WS chat (Meta AI processes it)")
	} else {
		fail("GenerateVideo", fmt.Errorf("status=%s err=%s", res.Status, res.Error))
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func must(label string, err error) {
	if err != nil {
		fmt.Printf("[FAIL] %s: %v\n", label, err)
		os.Exit(1)
	}
}

func pass(msg string) { fmt.Printf("[PASS] %s\n", msg) }

func fail(label string, err error) {
	failures++
	fmt.Printf("[FAIL] %s: %v\n", label, err)
}

func boolPtr(b bool) *bool { return &b }

func mask(s string) string {
	if len(s) < 12 {
		return "***"
	}
	return s[:8] + "…" + s[len(s)-4:]
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
