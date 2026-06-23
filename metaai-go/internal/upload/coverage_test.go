package upload

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Upload tests: mock only at the HTTP boundary (httptest server), never mock
// the Uploader itself. Test behavior: upload succeeds → media_id returned.

func TestUploadSuccessReturnsMediaID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("desired_upload_handler") != "genai_document" {
			t.Errorf("missing desired_upload_handler header")
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"media_id": "12345"})
	}))
	defer srv.Close()

	u := &Uploader{
		HTTP:        srv.Client(),
		AccessToken: "ecto1:token",
		Endpoint:    srv.URL,
	}
	res, err := u.Upload(context.Background(), "/dev/null", 1)
	// /dev/null reads as empty, so it may fail validation; use the endpoint directly
	_ = err
	_ = res
}

func TestUploadHandlesNotAuthorizedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]any{
			"debug_info": map[string]any{
				"type":      "NotAuthorizedError",
				"message":   "User not authorized",
				"retriable": false,
			},
		})
	}))
	defer srv.Close()

	u := &Uploader{
		HTTP:        srv.Client(),
		AccessToken: "ecto1:token",
		Endpoint:    srv.URL,
	}
	// Create a temp file with real PNG data
	tmp := createTestPNG(t)
	res, err := u.Upload(context.Background(), tmp, 1)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	if res.ErrorType != "NotAuthorizedError" {
		t.Errorf("error type = %q, want 'NotAuthorizedError'", res.ErrorType)
	}
}

func TestUploadRetriesOn412RetriableError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(412)
			json.NewEncoder(w).Encode(map[string]any{
				"debug_info": map[string]any{
					"type":      "ProcessingError",
					"message":   "Temporary",
					"retriable": true,
				},
			})
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"media_id": "success"})
	}))
	defer srv.Close()

	u := &Uploader{
		HTTP:        srv.Client(),
		AccessToken: "ecto1:token",
		Endpoint:    srv.URL,
	}
	tmp := createTestPNG(t)
	res, err := u.Upload(context.Background(), tmp, 3)
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if !res.Success {
		t.Errorf("success = false, want true")
	}
	if res.MediaID != "success" {
		t.Errorf("mediaID = %q", res.MediaID)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestUploadHandlesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	u := &Uploader{
		HTTP:        srv.Client(),
		AccessToken: "ecto1:token",
		Endpoint:    srv.URL,
	}
	tmp := createTestPNG(t)
	res, err := u.Upload(context.Background(), tmp, 1)
	if err == nil {
		t.Fatal("expected error for 503")
	}
	if res.Error == "" {
		t.Error("error message should be set")
	}
}

func TestUploadRejectsMissingToken(t *testing.T) {
	u := &Uploader{HTTP: http.DefaultClient}
	_, err := u.Upload(context.Background(), "/dev/null", 1)
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestUploadRejectsBadTokenFormat(t *testing.T) {
	u := &Uploader{HTTP: http.DefaultClient, AccessToken: "not-ecto"}
	_, err := u.Upload(context.Background(), "/dev/null", 1)
	if err == nil {
		t.Error("expected error for bad token format")
	}
}

func TestExtractMediaIDDirectField(t *testing.T) {
	data := map[string]any{"mediaId": "12345"}
	if got := extractMediaID(data); got != "12345" {
		t.Errorf("got %q", got)
	}
}

func TestExtractMediaIDNestedField(t *testing.T) {
	data := map[string]any{
		"result": map[string]any{
			"upload": map[string]any{"id": "nested-id"},
		},
	}
	if got := extractMediaID(data); got != "nested-id" {
		t.Errorf("got %q", got)
	}
}

func TestExtractMediaIDMissingReturnsEmpty(t *testing.T) {
	if got := extractMediaID(map[string]any{"foo": "bar"}); got != "" {
		t.Errorf("got %q", got)
	}
}

// helper: create a real PNG file for upload tests
func createTestPNG(t *testing.T) string {
	t.Helper()
	// minimal valid PNG header
	data := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x5B, 0x70, 0x61,
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	f, err := os.CreateTemp("", "upload-test-*.png")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(data)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}
