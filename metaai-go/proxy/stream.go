package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// sseWriter wraps an http.ResponseWriter for Server-Sent Events output. It is
// used by both the OpenAI (data:-only) and Anthropic (typed event:) streams.
type sseWriter struct {
	w  http.ResponseWriter
	fl http.Flusher
}

func newSSE(w http.ResponseWriter) *sseWriter {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	fl, _ := w.(http.Flusher)
	return &sseWriter{w: w, fl: fl}
}

// event writes a named SSE event: "event: <name>\ndata: <payload>\n\n".
func (s *sseWriter) event(name, payload string) error {
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", name, payload)
	if s.fl != nil {
		s.fl.Flush()
	}
	return err
}

// data writes an unnamed SSE data line: "data: <payload>\n\n" (OpenAI format).
func (s *sseWriter) data(payload string) error {
	_, err := fmt.Fprintf(s.w, "data: %s\n\n", payload)
	if s.fl != nil {
		s.fl.Flush()
	}
	return err
}

// errorEvent writes an Anthropic-style error event (also harmless for OpenAI
// clients, which surface it as a stream error).
func (s *sseWriter) errorEvent(errType, message string) error {
	payload, _ := json.Marshal(map[string]any{
		"type":  "error",
		"error": map[string]any{"type": errType, "message": message},
	})
	return s.event("error", string(payload))
}
