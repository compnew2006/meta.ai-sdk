package rest

// handlers.go implements the REST endpoint handlers.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/smart-studio/metaai-go"
)

// handleIndex returns service info on GET /.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "unknown endpoint")
		return
	}
	writeJSON(w, http.StatusOK, IndexResponse{
		Name: "metaai-rest",
		Endpoints: []string{
			"/healthz", "/chat", "/upload", "/analyze", "/image",
			"/video", "/video/extend", "/video/async", "/video/jobs/{job_id}",
		},
	})
}

// handleHealth is the liveness probe.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// handleChat handles POST /chat.
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	opts := &metaai.ChatOptions{}
	if req.NewConversation != nil {
		opts.NewConversation = *req.NewConversation
	}
	opts.Thinking = req.Thinking
	opts.Instant = req.Instant
	if req.Mode != "" {
		m := req.Mode
		opts.Mode = &m
	}
	// Resume an existing conversation when the client provides its id. The
	// SDK sends it in the CONNECT/CHAT frames so Meta AI reuses prior context.
	if req.ConversationID != "" {
		opts.ConversationID = req.ConversationID
	}
	// Per-call system instruction (overrides the client-wide default).
	if req.SystemInstruction != "" {
		opts.SystemInstruction = req.SystemInstruction
	}

	if req.Stream {
		s.chatMu.Lock()
		defer s.chatMu.Unlock()
		flusher, _ := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		ch := s.client.StreamChat(r.Context(), req.Message, opts)
		for chunk := range ch {
			if chunk.Err != nil {
				errPayload, _ := json.Marshal(map[string]any{"error": chunk.Err.Error()})
				_, _ = w.Write([]byte("data: " + string(errPayload) + "\n\n"))
				break
			}
			// Echo the conversation_id on every chunk so the UI learns it
			// immediately (even for a brand-new conversation assigned by Meta AI).
			payload, _ := json.Marshal(ChatResponse{Success: true, Message: chunk.Text, ConversationID: s.client.LastConversationID()})
			_, _ = w.Write([]byte("data: " + string(payload) + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
		return
	}

	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	reply, err := s.client.Chat(r.Context(), req.Message, opts)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ChatResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ChatResponse{Success: true, Message: reply, ConversationID: s.client.LastConversationID()})
}

// handleUpload handles POST /upload (multipart/form-data with "file" field).
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	// Limit upload size to 32 MB.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, `missing "file" field`)
		return
	}
	defer file.Close()

	// Persist to a temp file (UploadImage needs a path).
	tmp, err := os.CreateTemp("", "metaai-upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create temp file: "+err.Error())
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	size, err := io.Copy(tmp, file)
	if err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, "write temp file: "+err.Error())
		return
	}
	tmp.Close()

	s.chatMu.Lock()
	res, err := s.client.UploadImage(r.Context(), tmpPath)
	s.chatMu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, UploadResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, UploadResponse{
		Success:  res.Success,
		MediaID:  res.MediaID,
		FileName: header.Filename,
		FileSize: size,
		MimeType: res.MimeType,
	})
}

// handleImage handles POST /image.
func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req ImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	orientation := strings.ToUpper(strings.TrimSpace(req.Orientation))
	if orientation == "" {
		orientation = "VERTICAL"
	}

	// Image generation does not go through the WS chat lock (it uses the feed
	// poll), so no chatMu needed.
	res, err := s.client.GenerateImage(r.Context(), req.Prompt, orientation, 1)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ImageResponse{
			Prompt: req.Prompt, Status: "FAILED", Error: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, ImageResponse{
		Success:        res.Success,
		Prompt:         res.Prompt,
		ImageURLs:      res.URLs,
		MediaIDs:       res.MediaIDs,
		Status:         res.Status,
		ConversationID: res.ConversationID,
		Error:          res.Error,
	})
}

// handleImageFetch handles GET /image/fetch?url=... . It pulls a generated
// image URL (hosted on *.fbcdn.net) server-side and returns the bytes as
// base64, so the SPA can avoid the CDN's inconsistent browser CORS without
// needing to relax any same-origin policy. Only fbcdn hosts are allowed — this
// is a strict allow-list, not an open proxy.
func (s *Server) handleImageFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" {
		writeError(w, http.StatusBadRequest, "url query parameter is required")
		return
	}
	u, err := url.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid url: "+err.Error())
		return
	}
	host := strings.ToLower(u.Hostname())
	// Strict allow-list: only Facebook's CDN, which is where Meta AI surfaces
	// generated images and videos. Reject anything else to avoid turning this
	// endpoint into an open proxy / SSRF vector.
	if !isFbcdnHost(host) {
		writeError(w, http.StatusBadRequest, "url must be on a *.fbcdn.net host")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u.String(), nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "build request: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "metaai-rest/1.0")
	resp, err := s.imageHTTP.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ImageFetchResponse{Error: err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, ImageFetchResponse{Error: "fbcdn returned status " + resp.Status})
		return
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20)) // 32 MB cap
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ImageFetchResponse{Error: "read body: " + err.Error()})
		return
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = http.DetectContentType(body)
	}
	writeJSON(w, http.StatusOK, ImageFetchResponse{
		Success:  true,
		Base64:   base64.StdEncoding.EncodeToString(body),
		MimeType: mime,
	})
}

// isFbcdnHost reports whether host is a Facebook CDN host (the canonical home
// of Meta AI generated media), e.g. "scontent.xx.fbcdn.net".
func isFbcdnHost(host string) bool {
	return strings.HasSuffix(host, ".fbcdn.net") || strings.HasSuffix(host, ".cdninstagram.com")
}
func (s *Server) handleVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	res, err := s.client.GenerateVideo(r.Context(), req.Prompt)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, VideoResponse{
			Prompt: req.Prompt, Status: "FAILED", Error: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, VideoResponse{
		Success:        res.Success,
		Prompt:         res.Prompt,
		VideoURLs:      res.URLs,
		MediaIDs:       res.MediaIDs,
		Status:         res.Status,
		ConversationID: res.ConversationID,
		Error:          res.Error,
	})
}

// handleVideoExtend handles POST /video/extend.
func (s *Server) handleVideoExtend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req ExtendVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.MediaID == "" {
		writeError(w, http.StatusBadRequest, "media_id is required")
		return
	}
	res, err := s.client.ExtendVideo(r.Context(), req.MediaID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, VideoResponse{
			Status: "FAILED", Error: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, VideoResponse{
		Success:   res.Success,
		VideoURLs: res.URLs,
		MediaIDs:  res.MediaIDs,
		Status:    res.Status,
		Error:     res.Error,
	})
}

// handleVideoAsync handles POST /video/async — starts a background job.
func (s *Server) handleVideoAsync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	job := s.jobs.create()
	go s.runVideoJob(job, req.Prompt)

	status, _, _ := job.getInfo()
	writeJSON(w, http.StatusAccepted, AsyncJobResponse{
		Success: true,
		JobID:   job.id,
		Status:  status,
	})
}

// runVideoJob runs the video generation in the background and updates the job.
func (s *Server) runVideoJob(j *job, prompt string) {
	j.markRunning()
	res, err := s.client.GenerateVideo(context.Background(), prompt)
	if err != nil {
		j.fail(err.Error())
		return
	}
	j.complete(&VideoResponse{
		Success:        res.Success,
		Prompt:         res.Prompt,
		VideoURLs:      res.URLs,
		MediaIDs:       res.MediaIDs,
		Status:         res.Status,
		ConversationID: res.ConversationID,
		Error:          res.Error,
	})
}

// handleVideoJob handles GET /video/jobs/{job_id}.
func (s *Server) handleVideoJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET required")
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/video/jobs/")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job_id is required")
		return
	}
	j, ok := s.jobs.get(jobID)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	status, result, errStr := j.getInfo()
	resp := JobStatusResponse{JobID: j.id, Status: status}
	if errStr != "" {
		resp.Error = errStr
	}
	if result != nil {
		resp.Result = result
	}
	writeJSON(w, http.StatusOK, resp)
}

