package metaai

// These tests lock down cookie loading, cookie-header formatting,
// offline-threading-id shape, challenge detection, and access-token parsing.

import (
	"strings"
	"testing"
)

func TestLoadCookiesFromEnv(t *testing.T) {
	t.Setenv("META_AI_DATR", "test_datr")
	t.Setenv("META_AI_ABRA_SESS", "test_abra_sess")
	t.Setenv("META_AI_ECTO_1_SESS", "test_ecto")
	t.Setenv("META_AI_DPR", "1.25")
	cookies := loadCookiesFromEnv()
	if cookies["datr"] != "test_datr" {
		t.Errorf("datr = %q", cookies["datr"])
	}
	if cookies["abra_sess"] != "test_abra_sess" {
		t.Errorf("abra_sess = %q", cookies["abra_sess"])
	}
	if cookies["ecto_1_sess"] != "test_ecto" {
		t.Errorf("ecto_1_sess = %q", cookies["ecto_1_sess"])
	}
	if cookies["dpr"] != "1.25" {
		t.Errorf("dpr = %q", cookies["dpr"])
	}
}

func TestLoadCookiesFromEnvRequiresDatr(t *testing.T) {
	// Clear datr → nil.
	t.Setenv("META_AI_DATR", "")
	t.Setenv("META_AI_ABRA_SESS", "x")
	if got := loadCookiesFromEnv(); got != nil {
		t.Errorf("expected nil when DATR absent, got %v", got)
	}
}

func TestCookieHeader(t *testing.T) {
	h := cookieHeader(map[string]string{
		"datr":      "test_datr",
		"abra_sess": "test_abra_sess",
	})
	if !strings.Contains(h, "datr=test_datr") {
		t.Errorf("missing datr pair: %q", h)
	}
	if !strings.Contains(h, "abra_sess=test_abra_sess") {
		t.Errorf("missing abra_sess pair: %q", h)
	}
	if !strings.Contains(h, "; ") {
		t.Errorf("pairs not '; '-separated: %q", h)
	}
}

func TestGenerateOfflineThreadingID(t *testing.T) {
	id := generateOfflineThreadingID()
	if id == "" {
		t.Fatal("empty id")
	}
	// Should be a numeric string.
	for _, r := range id {
		if r < '0' || r > '9' {
			t.Errorf("non-digit in id: %q", id)
			break
		}
	}
}

func TestDetectChallengePage(t *testing.T) {
	cases := []struct {
		html string
		want string
	}{
		{"<html>normal</html>", ""},
		{"<html>executeChallenge fetch('/__rd_verify/abc123')</html>", "/__rd_verify/abc123"},
		{"<html>__rd_verify present but no fetch</html>", ""},
	}
	for _, tc := range cases {
		if got := detectChallengePage(tc.html); got != tc.want {
			t.Errorf("detectChallengePage(%q) = %q, want %q", tc.html, got, tc.want)
		}
	}
}

func TestExtractAccessToken(t *testing.T) {
	html := `<script>window.__x={\"accessToken\":\"ecto1:ABCdef123_-456XYZ\"}</script>`
	got := extractAccessToken(html)
	if got != "ecto1:ABCdef123_-456XYZ" {
		t.Errorf("extractAccessToken = %q", got)
	}
	if got := extractAccessToken("<html>no token</html>"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestNewClientConfigPriority(t *testing.T) {
	t.Setenv("META_AI_DATR", "")
	t.Setenv("META_AI_DEFAULT_THINKING", "true")
	t.Setenv("META_AI_DEFAULT_INSTANT", "true")
	c, err := NewClient(WithCookies(map[string]string{"datr": "d"}))
	if err != nil {
		t.Fatal(err)
	}
	// Both thinking and instant from env → mutual exclusion forces thinking off.
	if c.cfg.DefaultThinking {
		t.Error("DefaultThinking should be false when both set")
	}
	if !c.cfg.DefaultInstant {
		t.Error("DefaultInstant should be true")
	}
	if !c.IsAuthed() {
		t.Error("cookies present → should be authed")
	}
}
