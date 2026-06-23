package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"
)

// randID returns a short hex id used for response/chunk ids.
func randID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// approxTokens returns a rough token estimate (~4 chars/token). Meta AI does
// not report usage, so the proxy fills the OpenAI/Anthropic usage fields with
// an estimate derived from text length.
func approxTokens(s string) int {
	if s == "" {
		return 0
	}
	return utf8.RuneCountInString(s) / 4
}

// jsonMarshal returns the JSON encoding of v, ignoring errors (used only for
// already-serializable response shapes).
func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
