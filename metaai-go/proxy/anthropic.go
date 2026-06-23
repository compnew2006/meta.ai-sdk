package proxy

// anthropic.go implements POST /v1/messages (and the /messages alias), mapping
// Anthropic Messages requests onto a single Meta AI chat turn, with the SSE
// event sequence Claude Code and other Anthropic clients expect.

import (
	"encoding/json"
	"net/http"

	"github.com/smart-studio/metaai-go"
)

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "invalid_request_error", "POST required")
		return
	}
	var req anthMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages is required")
		return
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 4096
	}

	s.chatMu.Lock()
	defer s.chatMu.Unlock()

	model := resolveModel(req.Model)
	prompt, images := assembleAnthTranscript(req.System, req.Messages)
	id := "msg_" + randID()

	if model.Generation != "" {
		text, err := s.doGeneration(r.Context(), model, prompt)
		if err != nil {
			writeError(w, http.StatusBadGateway, "upstream_error", err.Error())
			return
		}
		if req.Stream {
			s.streamAnthropicText(w, id, req.Model, prompt, text)
		} else {
			stop := "end_turn"
			writeJSON(w, http.StatusOK, anthMessagesResponse{
				ID: id, Type: "message", Role: "assistant", Model: req.Model,
				Content:    []anthContentPart{{Type: "text", Text: text}},
				StopReason: &stop,
				Usage:      anthUsage{InputTokens: approxTokens(prompt), OutputTokens: approxTokens(text)},
			})
		}
		return
	}

	opts := chatOptionsFor(model)

	if req.Stream {
		s.streamAnthropic(w, r, id, req.Model, prompt, images, opts)
		return
	}

	text, err := s.complete(r.Context(), prompt, images, opts)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	stop := "end_turn"
	writeJSON(w, http.StatusOK, anthMessagesResponse{
		ID: id, Type: "message", Role: "assistant", Model: req.Model,
		Content:    []anthContentPart{{Type: "text", Text: text}},
		StopReason: &stop,
		Usage:      anthUsage{InputTokens: approxTokens(prompt), OutputTokens: approxTokens(text)},
	})
}

func (s *Server) streamAnthropic(w http.ResponseWriter, r *http.Request, id, model, prompt string, images []imageRef, opts *metaai.ChatOptions) {
	sse := newSSE(w)
	emitAnthropicStart(sse, id, model, approxTokens(prompt))
	ch, cleanup := s.stream(r.Context(), prompt, images, opts)
	defer cleanup()
	for chunk := range ch {
		if chunk.Err != nil {
			_ = sse.errorEvent("upstream_error", chunk.Err.Error())
			break
		}
		if chunk.Text == "" {
			continue
		}
		_ = sse.event("content_block_delta", jsonMarshal(map[string]any{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]any{"type": "text_delta", "text": chunk.Text},
		}))
	}
	emitAnthropicEnd(sse, approxTokens(prompt))
}

// streamAnthropicText streams already-known text (e.g. generation output) in
// the Anthropic event format.
func (s *Server) streamAnthropicText(w http.ResponseWriter, id, model, prompt, text string) {
	sse := newSSE(w)
	emitAnthropicStart(sse, id, model, approxTokens(prompt))
	if text != "" {
		_ = sse.event("content_block_delta", jsonMarshal(map[string]any{
			"type": "content_block_delta", "index": 0,
			"delta": map[string]any{"type": "text_delta", "text": text},
		}))
	}
	emitAnthropicEnd(sse, approxTokens(prompt))
}

// emitAnthropicStart sends the message_start + content_block_start prefix.
func emitAnthropicStart(sse *sseWriter, id, model string, inputTokens int) {
	msg := anthMessagesResponse{
		ID: id, Type: "message", Role: "assistant", Model: model,
		Content: []anthContentPart{},
		Usage:   anthUsage{InputTokens: inputTokens},
	}
	_ = sse.event("message_start", jsonMarshal(map[string]any{
		"type":    "message_start",
		"message": msg,
	}))
	_ = sse.event("content_block_start", jsonMarshal(map[string]any{
		"type":          "content_block_start",
		"index":         0,
		"content_block": map[string]any{"type": "text", "text": ""},
	}))
}

// emitAnthropicEnd sends the content_block_stop + message_delta + message_stop
// trailer.
func emitAnthropicEnd(sse *sseWriter, outputTokens int) {
	_ = sse.event("content_block_stop", jsonMarshal(map[string]any{
		"type": "content_block_stop", "index": 0,
	}))
	_ = sse.event("message_delta", jsonMarshal(map[string]any{
		"type":  "message_delta",
		"delta": map[string]any{"stop_reason": "end_turn", "stop_sequence": nil},
		"usage": map[string]any{"output_tokens": outputTokens},
	}))
	_ = sse.event("message_stop", jsonMarshal(map[string]any{"type": "message_stop"}))
}
