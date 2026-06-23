package metaai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/smart-studio/metaai-go/internal/uuid"
)

// generateOfflineThreadingID produces the threading identifier used by the web
// client. Delegates to internal/uuid, which panics on CSPRNG failure rather
// than silently producing a predictable id.
func generateOfflineThreadingID() string {
	return uuid.OfflineThreadingID()
}

// extractValue returns the substring of text between
// the first occurrence of startStr and the following endStr. Returns "" if not found.
func extractValue(text, startStr, endStr string) string {
	i := strings.Index(text, startStr)
	if i < 0 {
		return ""
	}
	rest := text[i+len(startStr):]
	j := strings.Index(rest, endStr)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// challengeURLRe finds the /__rd_verify path inside a challenge page.
var challengeURLRe = regexp.MustCompile(`fetch\('(/__rd_verify[^']+)`)

// detectChallengePage returns the
// verification URL when the HTML is a Meta challenge/verification page, else "".
func detectChallengePage(htmlText string) string {
	if !strings.Contains(htmlText, "executeChallenge") &&
		!strings.Contains(htmlText, "__rd_verify") {
		return ""
	}
	m := challengeURLRe.FindStringSubmatch(htmlText)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// accessTokenRe extracts the ecto1: OAuth access token from meta.ai page HTML.
// Matches both the escaped form accessToken\\":\\"(ecto1:...) and
// the plain accessToken":"(ecto1:...) form seen in some inline scripts.
var accessTokenRe = regexp.MustCompile(`accessToken\\"{0,1}:\\"{0,1}(ecto1:[A-Za-z0-9_-]+)`)

// extractAccessToken finds the ecto1: access token in page HTML. Returns "" if
// not present.
func extractAccessToken(htmlText string) string {
	m := accessTokenRe.FindStringSubmatch(htmlText)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// numericIDString returns a random n-digit decimal string. Used for message ids
// in the clippy frame. Delegates to internal/uuid, which panics on CSPRNG
// failure rather than silently returning a collision-prone zero.
func numericIDString(digits int) string {
	return uuid.NumericString(digits)
}

// quotePath returns a debug-friendly quoted representation.
func quotePath(s string) string { return fmt.Sprintf("%q", s) }

// newConversationID returns a fresh UUIDv4 string for a new chat conversation.
// (meta.ai conversation ids are the `/prompt/<uuid>` path component.) Delegates
// to internal/uuid, which panics on CSPRNG failure rather than silently
// returning a zero UUID that would collide across conversations.
func newConversationID() string {
	return uuid.V4()
}

// toInt coerces a JSON-decoded numeric value to int. Handles int, int64,
// and float64 — the three numeric types that appear in json.Unmarshal output.
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}
