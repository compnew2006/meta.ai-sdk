package generation

// sse.go implements SSE response parsing + media-URL extraction for generation.
// It parses captured generation events, extracts media URLs, and detects readiness.

import (
	"encoding/json"
	"strings"
)

// SSEResult is the parsed outcome of a generation SSE stream.
type SSEResult struct {
	StatusCode       int              `json:"status_code"`
	StreamingState   string           `json:"streaming_state"` // OVERALL_DONE | STREAMING | FAILED
	Images           []string         `json:"images"`
	Videos           []string         `json:"videos"`
	ImageObjects     []map[string]any `json:"image_objects"`
	VideoObjects     []map[string]any `json:"video_objects"`
	ConversationID   string           `json:"conversation_id"`
	Message          string           `json:"message"`
	HasGraphQLErrors bool             `json:"has_graphql_errors"`
	GraphQLErrors    []GQLError       `json:"graphql_errors"`
}

// GQLError is a normalized GraphQL error from an SSE event.
type GQLError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ParseSSE walks the `data:` lines of a generation SSE response body and
// aggregates into an SSEResult.
func ParseSSE(body string) *SSEResult {
	r := &SSEResult{StatusCode: 200}
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" || payload == "null" || payload == "" {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue // skip invalid JSON events
		}
		// GraphQL errors at top level → FAILED.
		if errs, ok := ev["errors"].([]any); ok {
			r.HasGraphQLErrors = true
			r.StreamingState = "FAILED"
			for _, e := range errs {
				if em, ok := e.(map[string]any); ok {
					ge := GQLError{Message: asString(em["message"])}
					if ext, ok := em["extensions"].(map[string]any); ok {
						ge.Code = asString(ext["code"])
					}
					r.GraphQLErrors = append(r.GraphQLErrors, ge)
				}
			}
			continue
		}
		// sendMessageStream payload under data.sendMessageStream.
		data, _ := ev["data"].(map[string]any)
		if data != nil {
			if sms, ok := data["sendMessageStream"].(map[string]any); ok {
				if state, ok := sms["streamingState"].(string); ok {
					r.StreamingState = state
				}
				if id, ok := sms["conversationId"].(string); ok {
					r.ConversationID = id
				}
				if imgs, ok := sms["images"].([]any); ok {
					for _, ig := range imgs {
						if m, ok := ig.(map[string]any); ok {
							if u := asString(m["url"]); u != "" {
								r.Images = append(r.Images, u)
							}
							r.ImageObjects = append(r.ImageObjects, m)
						}
					}
				}
				if vids, ok := sms["videos"].([]any); ok {
					for _, v := range vids {
						if m, ok := v.(map[string]any); ok {
							if u := asString(m["url"]); u != "" {
								r.Videos = append(r.Videos, u)
							}
							r.VideoObjects = append(r.VideoObjects, m)
						}
					}
				}
				if c, ok := sms["content"].(string); ok && c != "" {
					r.Message += c
				}
			}
		}
	}
	return r
}

// ExtractMediaURLs pulls image/video URLs out of a fetch-media or generation
// response. It handles the
// xfb_kadabra_send_message / xfb_imagine_send_message / xfb_genai_fetch_post
// nesting shapes).
func ExtractMediaURLs(data map[string]any) []string {
	urls := map[string]struct{}{}
	collect := func(u string) {
		if u != "" {
			urls[u] = struct{}{}
		}
	}
	// Recursive walk: find uri/video_url/progressive_url keys.
	walk(data, collect)
	out := make([]string, 0, len(urls))
	for u := range urls {
		out = append(out, u)
	}
	return out
}

func walk(v any, collect func(string)) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			switch k {
			case "uri", "video_url", "progressive_url", "url", "fallbackUrl":
				if s, ok := val.(string); ok {
					collect(s)
				}
			}
			walk(val, collect)
		}
	case []any:
		for _, e := range t {
			walk(e, collect)
		}
	}
}

// IsMediaReady reports whether a generation result contains usable media.
func IsMediaReady(data map[string]any) bool {
	if len(data) == 0 {
		return false
	}
	if _, has := data["error"]; has {
		return false
	}
	if status, ok := data["status"].(string); ok {
		switch strings.ToUpper(status) {
		case "READY", "COMPLETE":
			return true
		}
	}
	if len(ExtractMediaURLs(data)) > 0 {
		return true
	}
	return false
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
