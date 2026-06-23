// Package video implements the lower-level video-generation SSE parsing path.
//
// This package handles the captured document IDs, SSE event parsing, and
// media-ID/video-URL extraction. The primary video path is
// generation.GenerateVideo; this package exposes the lower-level helpers.
package video

import (
	"encoding/json"
	"strings"
)

// Document-ID constants used by video generation requests.
const (
	VideoCardDocID     = "666834feb70769370072c294c87ebd23"
	VideoGenerateDocID = "a3d873304cb1411ba7f056e47060ad1d"
	VideoFetchDocID    = "10b7bd5aa8b7537e573e49d701a5b21b"
)

// Event is one parsed SSE data payload (a JSON object).
type Event map[string]any

// ParseSSEResponse yields each
// non-empty `data:` line parsed as JSON; invalid JSON lines are skipped.
func ParseSSEResponse(body string) []Event {
	var out []Event
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue // skip invalid JSON
		}
		out = append(out, ev)
	}
	return out
}

// ExtractMediaIDs returns media identifiers from a generation response.
// Walks the response for mediaId fields under messages.edges[].node.content.
func ExtractMediaIDs(resp map[string]any) []string {
	seen := map[string]struct{}{}
	var out []string
	walk(resp, func(key string, val any) {
		if key == "mediaId" {
			if s, ok := val.(string); ok && s != "" {
				if _, ok := seen[s]; !ok {
					seen[s] = struct{}{}
					out = append(out, s)
				}
			}
		}
	})
	return out
}

// ExtractVideoURLs returns playable URLs from media objects.
// Walks for video_url / uri / progressive_url fields.
func ExtractVideoURLs(resp map[string]any) []string {
	seen := map[string]struct{}{}
	var out []string
	walk(resp, func(key string, val any) {
		if key == "video_url" || key == "uri" || key == "progressive_url" {
			if s, ok := val.(string); ok && s != "" {
				if _, ok := seen[s]; !ok {
					seen[s] = struct{}{}
					out = append(out, s)
				}
			}
		}
	})
	return out
}

func walk(v any, visit func(string, any)) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			visit(k, val)
			walk(val, visit)
		}
	case []any:
		for _, e := range t {
			walk(e, visit)
		}
	}
}
