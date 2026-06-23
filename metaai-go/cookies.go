package metaai

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Cookie environment-variable name to request-cookie key.
var cookieEnvKeys = []struct{ env, cookie string }{
	{"META_AI_DATR", "datr"},
	{"META_AI_ABRA_SESS", "abra_sess"},
	{"META_AI_ECTO_1_SESS", "ecto_1_sess"},
	{"META_AI_DPR", "dpr"},
	{"META_AI_WD", "wd"},
	{"META_AI_JS_DATR", "_js_datr"},
	{"META_AI_ABRA_CSRF", "abra_csrf"},
	{"META_AI_RD_CHALLENGE", "rd_challenge"},
	{"META_AI_PS_L", "ps_l"},
	{"META_AI_PS_N", "ps_n"},
}

// loadCookiesFromEnv builds the cookie map from META_AI_* environment variables.
// Returns nil when META_AI_DATR (the only strictly-required cookie) is absent.
// Unknown or empty environment values are ignored.
func loadCookiesFromEnv() map[string]string {
	if os.Getenv("META_AI_DATR") == "" {
		return nil
	}
	cookies := map[string]string{}
	for _, m := range cookieEnvKeys {
		if v := strings.TrimSpace(os.Getenv(m.env)); v != "" {
			cookies[m.cookie] = v
		}
	}
	return cookies
}

// cookieHeader formats a cookie map as an HTTP Cookie header value
// ("k1=v1; k2=v2"), with keys in stable (sorted) order so tests are deterministic.
// Keys are sorted to keep the header deterministic.
func cookieHeader(cookies map[string]string) string {
	if len(cookies) == 0 {
		return ""
	}
	keys := make([]string, 0, len(cookies))
	for k := range cookies {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, cookies[k]))
	}
	return strings.Join(parts, "; ")
}

// attachCookies sets the cookie map on an http.Request as a Cookie header.
func attachCookies(req *http.Request, cookies map[string]string) {
	if len(cookies) == 0 {
		return
	}
	if h := cookieHeader(cookies); h != "" {
		req.Header.Set("Cookie", h)
	}
}
