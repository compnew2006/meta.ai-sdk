// Package generation implements Meta AI image/video generation via the
// persisted-query GraphQL endpoint (https://meta.ai/api/graphql, SSE).
//
// This implementation follows captured generation requests: document-ID resolution,
// imagineOperationRequest variable shape, SSE response parsing, media-URL
// extraction, and media-ready polling.
package generation

import (
	"os"
	"strings"
	"time"
)

// Operation is the imagineOperationRequest.operation value.
type Operation string

// Generation operations. The first two are text-input; the middle two are
// image-input (triggered when MediaIDs is non-empty); OpExtendVideo extends an
// existing generated clip.
const (
	OpTextToImage  Operation = "TEXT_TO_IMAGE"
	OpTextToVideo  Operation = "TEXT_TO_VIDEO"
	OpImageToImage Operation = "IMAGE_TO_IMAGE"
	OpImageToVideo Operation = "IMAGE_TO_VIDEO"
	OpExtendVideo  Operation = "EXTEND_VIDEO"
)

// DocID defaults and their environment-variable override keys.
var (
	defaultDocIDs = map[Operation]string{
		OpTextToImage: "2f707e4a86f4b01adba97e1376cbdc14",
		OpTextToVideo: "2f707e4a86f4b01adba97e1376cbdc14",
		OpExtendVideo: "865d6fe804a7ea98fbce7e562b1d61ce",
	}
	envKeys = map[Operation][]string{
		OpTextToImage: {"META_AI_DOC_ID_TEXT_TO_IMAGE", "META_AI_DOC_ID"},
		OpTextToVideo: {"META_AI_DOC_ID_TEXT_TO_VIDEO", "META_AI_DOC_ID"},
		OpExtendVideo: {"META_AI_DOC_ID_EXTEND_VIDEO"},
	}
)

// DocID resolves the active doc_id for an operation: env override (first
// non-empty alphanumeric value) > default.
//
// Not cached — the env may change between calls (e.g. in tests via t.Setenv),
// and the lookup is cheap.
func DocID(op Operation) string {
	for _, k := range envKeys[op] {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" && isAlnum(v) {
			return v
		}
	}
	return defaultDocIDs[op]
}

// FetchMediaDocID is the doc_id for fetch-media-by-id (= META_AI_DOC_ID_FETCH_MEDIA).
const FetchMediaDocID = "ecc43cc5adc3443611ed22bd8608a371"

func isAlnum(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return len(s) > 0
}

// uniqueID returns a 19-digit timestamp-derived identifier.
func uniqueID() string {
	return time.Now().Format("20060102150405") + "00000"
}

// DefaultUserAgent identifies the browser profile used for generation requests.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"
