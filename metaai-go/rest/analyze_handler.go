package rest

import (
	"encoding/json"
	"net/http"

	"github.com/smart-studio/metaai-go"
)

// handleAnalyze handles POST /analyze.
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	// media_id is required on the FIRST turn (no conversation yet). On a
	// follow-up turn (conversation_id set) only the question is required.
	isFollowUp := req.ConversationID != ""
	if !isFollowUp && req.MediaID == "" {
		writeError(w, http.StatusBadRequest, "media_id is required (or send conversation_id for a follow-up)")
		return
	}
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	// Build opts so we can thread ConversationID (resume) + SystemInstruction.
	opts := &metaai.ChatOptions{ConversationID: req.ConversationID, SystemInstruction: req.SystemInstruction}

	// Streaming branch: Server-Sent Events, one "data:" frame per ChatChunk.
	// Mirrors handleChat's streaming path (handlers.go) byte-for-byte in shape.
	if req.Stream {
		s.chatMu.Lock()
		defer s.chatMu.Unlock()
		flusher, _ := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		ch := s.client.AnalyzeImageStream(r.Context(), "", req.MediaID, req.Question, opts)
		for chunk := range ch {
			if chunk.Err != nil {
				errPayload, _ := json.Marshal(map[string]any{"error": chunk.Err.Error()})
				_, _ = w.Write([]byte("data: " + string(errPayload) + "\n\n"))
				break
			}
			payload, _ := json.Marshal(AnalyzeResponse{Success: true, Message: chunk.Text, ConversationID: s.client.LastConversationID()})
			_, _ = w.Write([]byte("data: " + string(payload) + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
		return
	}

	s.chatMu.Lock()
	defer s.chatMu.Unlock()

	reply, err := s.client.AnalyzeImage(r.Context(), "", req.MediaID, req.Question, opts)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, AnalyzeResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, AnalyzeResponse{
		Success:        true,
		Message:        reply,
		ConversationID: s.client.LastConversationID(),
	})
}
