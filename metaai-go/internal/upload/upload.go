// Package upload implements image upload to Meta AI's rupload service.
//
// Grounded in observed browser upload requests. The endpoint is
// https://rupload.meta.ai/gen_ai_document_gen_ai_tenant/<session-uuid> and the
// raw image bytes are POSTed with OAuth (ecto1:) auth + x-entity-* headers.
package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/smart-studio/metaai-go/internal/uuid"
)

// Uploader uploads images to Meta AI's rupload CDN.
type Uploader struct {
	HTTP        *http.Client
	AccessToken string // ecto1:…
	UserAgent   string
	// CookieHeader is sent as the Cookie header. The rupload endpoint requires
	// the ecto_1_sess cookie alongside the OAuth Authorization header — the
	// token alone gets NotAuthorizedError (confirmed live 2026-06-19).
	CookieHeader string
	// Endpoint lets tests override the upload URL.
	Endpoint string
}

// Result is the outcome of an upload.
type Result struct {
	Success       bool   `json:"success"`
	MediaID       string `json:"media_id"`
	UploadSession string `json:"upload_session_id"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	Error         string `json:"error,omitempty"`
	ErrorType     string `json:"error_type,omitempty"`
	RawResponse   any    `json:"response,omitempty"`
}

// Upload uploads the image at path with exponential-backoff retries on 412/5xx.
// maxRetries <= 0 uses the package default.
func (u *Uploader) Upload(ctx context.Context, path string, maxRetries int) (*Result, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if u.AccessToken == "" {
		return &Result{Error: "missing access token"}, ErrMissingToken
	}
	if !strings.HasPrefix(u.AccessToken, "ecto1:") {
		return &Result{Error: "invalid access token format (want ecto1:…)"}, ErrBadToken
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Result{Error: fmt.Sprintf("file not found: %s", path)}, err
	}
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return &Result{Error: fmt.Sprintf("invalid file type: %s", mimeType)}, fmt.Errorf("upload: not an image")
	}
	filename := filepath.Base(path)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		res, retryable, err := u.uploadOnce(ctx, data, filename, mimeType)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if !retryable || attempt == maxRetries {
			return res, err
		}
		// Exponential backoff: 1s, 2s, 4s, …
		wait := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-time.After(wait):
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("upload: failed after %d attempts", maxRetries)
	}
	return &Result{Error: lastErr.Error()}, lastErr
}

func (u *Uploader) uploadOnce(ctx context.Context, data []byte, filename, mimeType string) (*Result, bool, error) {
	sessionID := uuid.V4()
	endpoint := u.Endpoint
	if endpoint == "" {
		endpoint = "https://rupload.meta.ai/gen_ai_document_gen_ai_tenant/" + sessionID
	}

	req, err := u.buildUploadRequest(ctx, endpoint, data, filename, mimeType)
	if err != nil {
		return nil, false, err
	}

	resp, err := u.HTTP.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("upload: POST: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	res := &Result{
		UploadSession: sessionID,
		FileName:      filename,
		FileSize:      int64(len(data)),
		MimeType:      mimeType,
	}

	// 412 Precondition Failed: parse debug_info for retriable flag.
	if resp.StatusCode == 412 {
		dbg := parseDebugInfo(body)
		res.Error = fmt.Sprintf("%s: %s", dbg.Type, dbg.Message)
		res.ErrorType = dbg.Type
		return res, dbg.Retriable, fmt.Errorf("upload: 412 %s", dbg.Type)
	}
	if resp.StatusCode >= 500 {
		res.Error = fmt.Sprintf("server error %d", resp.StatusCode)
		return res, true, fmt.Errorf("upload: %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		dbg := parseDebugInfo(body)
		if dbg.Type != "" {
			res.Error = fmt.Sprintf("%s: %s", dbg.Type, dbg.Message)
			res.ErrorType = dbg.Type
			return res, dbg.Retriable, fmt.Errorf("upload: %d %s: %s", resp.StatusCode, dbg.Type, dbg.Message)
		}
		res.Error = fmt.Sprintf("upload failed: status %d body=%s", resp.StatusCode, truncateBody(body))
		return res, false, fmt.Errorf("upload: status %d", resp.StatusCode)
	}

	// Parse success body → extract media id.
	var parsed any
	if jerr := json.Unmarshal(body, &parsed); jerr == nil {
		res.RawResponse = parsed
		res.MediaID = extractMediaID(parsed)
	} else {
		res.RawResponse = map[string]string{"raw": string(body)}
	}
	res.Success = res.MediaID != ""
	if !res.Success {
		res.Error = "upload succeeded but no media_id in response"
	}
	return res, false, nil
}

// buildUploadRequest constructs the rupload POST request. rupload.meta.ai is
// HEADER-CASE SENSITIVE: Go's http.Header.Set canonicalizes names
// (e.g. "x-entity-name" → "X-Entity-Name") causing NotAuthorizedError, so
// headers are assigned directly via the raw map to keep them lowercase.
func (u *Uploader) buildUploadRequest(ctx context.Context, endpoint string, data []byte, filename, mimeType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	setH := func(k, v string) { req.Header[k] = []string{v} }
	ua := u.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"
	}
	setH("accept", "*/*")
	setH("authorization", "OAuth "+u.AccessToken)
	setH("desired_upload_handler", "genai_document")
	setH("ecto_auth_token", "true")
	setH("is_abra_user", "true")
	setH("offset", "0")
	setH("origin", "https://meta.ai")
	setH("referer", "https://meta.ai/")
	setH("user-agent", ua)
	setH("x-entity-length", strconv.FormatInt(int64(len(data)), 10))
	setH("x-entity-name", filename)
	setH("x-entity-type", mimeType)
	if u.CookieHeader != "" {
		setH("cookie", u.CookieHeader)
	}
	return req, nil
}

// debugInfo mirrors the debug_info object returned by the rupload endpoint on
// error responses (412 / non-200).
type debugInfo struct {
	Retriable bool   `json:"retriable"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

// parseDebugInfo decodes the debug_info object from an error response body.
// Returns a zero debugInfo when the body does not contain one.
func parseDebugInfo(body []byte) debugInfo {
	var wrap struct {
		DebugInfo debugInfo `json:"debug_info"`
	}
	_ = json.Unmarshal(body, &wrap)
	return wrap.DebugInfo
}

func truncateBody(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "…"
	}
	return string(b)
}

// extractMediaID searches common response shapes for a media id.
// It accepts the response shapes observed from the upload endpoint.
func extractMediaID(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range []string{"mediaId", "media_id", "id", "uploadId", "entityId"} {
			if s, ok := t[k].(string); ok && s != "" {
				return s
			}
		}
		// recurse one level into nested objects
		for _, v := range t {
			if s := extractMediaID(v); s != "" {
				return s
			}
		}
	}
	return ""
}
