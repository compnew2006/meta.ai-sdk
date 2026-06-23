package rest

// types.go defines the request/response JSON shapes for the REST API server.
// Shapes are aligned with the canonical SDK REST API for cross-language
// consistency.

// HealthResponse is returned by /healthz.
type HealthResponse struct {
	Status string `json:"status"`
}

// ChatRequest is the body for POST /chat.
type ChatRequest struct {
	Message         string `json:"message"`
	NewConversation *bool  `json:"new_conversation,omitempty"`
	Stream          bool   `json:"stream,omitempty"`
	Thinking        *bool  `json:"thinking,omitempty"`
	Instant         *bool  `json:"instant,omitempty"`
	Mode            string `json:"mode,omitempty"`
	// ConversationID, when set, resumes an existing Meta AI conversation so
	// Meta AI reuses its prior context for the new turn. Omit (or send empty)
	// to start a new conversation; the assigned id is echoed back in the
	// response so the client can store and reuse it.
	ConversationID string `json:"conversation_id,omitempty"`
	// SystemInstruction, when set, is prepended to the message as a
	// "[System]\n...\n\n" block for this single turn (overrides any
	// server-wide default).
	SystemInstruction string `json:"system_instruction,omitempty"`
}

// ChatResponse is the body for POST /chat (non-streaming).
type ChatResponse struct {
	Success        bool             `json:"success"`
	Message        string           `json:"message"`
	Sources        []map[string]any `json:"sources,omitempty"`
	Media          []map[string]any `json:"media,omitempty"`
	ConversationID string           `json:"conversation_id,omitempty"`
	Error          string           `json:"error,omitempty"`
}

// UploadRequest is the body for POST /upload (multipart/form-data).
// File is carried in the "file" multipart part; the other fields are form
// values. They are read by the handler, not JSON-decoded.
type UploadResponse struct {
	Success  bool   `json:"success"`
	MediaID  string `json:"media_id"`
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ImageRequest is the body for POST /image.
type ImageRequest struct {
	Prompt      string `json:"prompt"`
	Orientation string `json:"orientation,omitempty"` // LANDSCAPE | VERTICAL | SQUARE
}

// ImageResponse is the body for POST /image.
type ImageResponse struct {
	Success        bool     `json:"success"`
	Prompt         string   `json:"prompt,omitempty"`
	ImageURLs      []string `json:"image_urls,omitempty"`
	MediaIDs       []string `json:"media_ids,omitempty"`
	Status         string   `json:"status,omitempty"`
	ConversationID string   `json:"conversation_id,omitempty"`
	Error          string   `json:"error,omitempty"`
}

// ImageFetchResponse is the body for GET /image/fetch. The server pulls the
// requested fbcdn URL server-side (dodging the CDN's inconsistent CORS on the
// browser) and returns the bytes as base64 so the SPA can keep its base64
// ImageFile shape without a cross-origin fetch.
type ImageFetchResponse struct {
	Success  bool   `json:"success"`
	Base64   string `json:"base64,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Error    string `json:"error,omitempty"`
}

// VideoRequest is the body for POST /video and POST /video/async.
type VideoRequest struct {
	Prompt   string `json:"prompt"`
	AutoPoll *bool  `json:"auto_poll,omitempty"` // /video/async only; default true for /video
}

// VideoResponse is the body for POST /video.
type VideoResponse struct {
	Success        bool     `json:"success"`
	Prompt         string   `json:"prompt,omitempty"`
	VideoURLs      []string `json:"video_urls,omitempty"`
	MediaIDs       []string `json:"media_ids,omitempty"`
	Status         string   `json:"status,omitempty"`
	ConversationID string   `json:"conversation_id,omitempty"`
	Error          string   `json:"error,omitempty"`
}

// ExtendVideoRequest is the body for POST /video/extend.
type ExtendVideoRequest struct {
	MediaID string `json:"media_id"`
}

// AsyncJobResponse is returned by POST /video/async.
type AsyncJobResponse struct {
	Success bool   `json:"success"`
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
}

// JobStatusResponse is returned by GET /video/jobs/{job_id}.
type JobStatusResponse struct {
	JobID  string         `json:"job_id"`
	Status string         `json:"status"` // queued | running | completed | failed
	Result *VideoResponse `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// AnalyzeRequest is the body for POST /analyze.
type AnalyzeRequest struct {
	MediaID  string `json:"media_id"`
	Question string `json:"question"`
	// Stream, when true, switches the response to text/event-stream with one
	// "data:" frame per ChatChunk (mirrors POST /chat streaming). Default
	// (false/omitted) keeps the single-JSON blocking response.
	Stream bool `json:"stream,omitempty"`
	// ConversationID, when set, makes this a follow-up question about the SAME
	// image in an existing conversation (text-only, no media_id needed). Omit
	// on the first turn (media_id required) to start a new conversation; the
	// assigned id is echoed back so the client can store and reuse it.
	ConversationID string `json:"conversation_id,omitempty"`
	// SystemInstruction, when set, is prepended to the question for this turn.
	SystemInstruction string `json:"system_instruction,omitempty"`
}

// AnalyzeResponse is the body for POST /analyze.
type AnalyzeResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// IndexResponse describes the service for GET /.
type IndexResponse struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}
