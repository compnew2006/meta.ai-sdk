package proxy

// openai.go implements POST /v1/chat/completions (and the /chat/completions
// alias), mapping OpenAI Chat Completions requests onto a single Meta AI chat
// turn, with SSE streaming when stream:true.

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/smart-studio/metaai-go"
)

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "invalid_request_error", "POST required")
		return
	}
	var req oaiChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages is required")
		return
	}

	// Serialize: the SDK shares one WS connection across turns on a Client.
	s.chatMu.Lock()
	defer s.chatMu.Unlock()

	model := resolveModel(req.Model)
	prompt, images := assembleOaiTranscript(nil, req.Messages)

	id := "chatcmpl-" + randID()
	created := time.Now().Unix()

	if model.Generation != "" {
		text, err := s.doGeneration(r.Context(), model, prompt)
		if err != nil {
			writeError(w, http.StatusBadGateway, "upstream_error", err.Error())
			return
		}
		if req.Stream {
			s.streamOAICompletion(w, id, created, req.Model, text)
		} else {
			ctok := approxTokens(text)
			writeJSON(w, http.StatusOK, oaiChatResponse{
				ID: id, Object: "chat.completion", Created: created, Model: req.Model,
				Choices: []oaiChoice{{Index: 0, Message: oaiMessage{Role: "assistant", Content: text}, FinishReason: "stop"}},
				Usage:   oaiUsage{CompletionTokens: ctok, TotalTokens: ctok},
			})
		}
		return
	}

	opts := chatOptionsFor(model)

	if req.Stream {
		s.streamOAI(w, r, id, created, req.Model, prompt, images, opts)
		return
	}

	text, err := s.complete(r.Context(), prompt, images, opts)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	ptok := approxTokens(prompt)
	ctok := approxTokens(text)
	writeJSON(w, http.StatusOK, oaiChatResponse{
		ID: id, Object: "chat.completion", Created: created, Model: req.Model,
		Choices: []oaiChoice{{Index: 0, Message: oaiMessage{Role: "assistant", Content: text}, FinishReason: "stop"}},
		Usage:   oaiUsage{PromptTokens: ptok, CompletionTokens: ctok, TotalTokens: ptok + ctok},
	})
}

// chatOptionsFor builds a NewConversation ChatOptions from a resolved model.
func chatOptionsFor(model modelInfo) *metaai.ChatOptions {
	opts := &metaai.ChatOptions{NewConversation: true}
	if model.Mode != "" {
		m := model.Mode
		opts.Mode = &m
	}
	thinking := model.Thinking
	instant := model.Instant
	opts.Thinking = &thinking
	opts.Instant = &instant
	return opts
}

func (s *Server) streamOAI(w http.ResponseWriter, r *http.Request, id string, created int64, model, prompt string, images []imageRef, opts *metaai.ChatOptions) {
	sse := newSSE(w)
	_ = sse.data(jsonMarshal(oaiChatChunk{
		ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
		Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{Role: "assistant"}}},
	}))

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
		_ = sse.data(jsonMarshal(oaiChatChunk{
			ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
			Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{Content: chunk.Text}}},
		}))
	}
	finish := "stop"
	_ = sse.data(jsonMarshal(oaiChatChunk{
		ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
		Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{}, FinishReason: &finish}},
	}))
	_ = sse.data("[DONE]")
}

// streamOAICompletion streams a fully-known text (e.g. generation output) in
// the OpenAI chunk format.
func (s *Server) streamOAICompletion(w http.ResponseWriter, id string, created int64, model, text string) {
	sse := newSSE(w)
	_ = sse.data(jsonMarshal(oaiChatChunk{
		ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
		Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{Role: "assistant"}}},
	}))
	if text != "" {
		_ = sse.data(jsonMarshal(oaiChatChunk{
			ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
			Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{Content: text}}},
		}))
	}
	finish := "stop"
	_ = sse.data(jsonMarshal(oaiChatChunk{
		ID: id, Object: "chat.completion.chunk", Created: created, Model: model,
		Choices: []oaiChunkChoice{{Index: 0, Delta: oaiDelta{}, FinishReason: &finish}},
	}))
	_ = sse.data("[DONE]")
}
