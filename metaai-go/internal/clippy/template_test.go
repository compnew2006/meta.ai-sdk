package clippy

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// TestBuildFromTemplateSubstitutesText verifies the template builder replaces
// the original text marker with the new text and patches the length varint.
func TestBuildFromTemplateSubstitutesText(t *testing.T) {
	tplB64, err := LoadTemplateB64("testdata/template_frame.b64")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "Hello world from test",
		ConversationID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		RequestID:      "11111111-2222-3333-4444-555555555555",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Decode the inner proto (the text/conv-id live base64-encoded inside).
	var outer struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame[8:], &outer); err != nil {
		t.Fatal(err)
	}
	inner, _ := base64.StdEncoding.DecodeString(outer.Payload)
	innerStr := string(inner)
	// The new text must be present in the inner proto.
	if !strings.Contains(innerStr, "Hello world from test") {
		t.Error("new text not found in inner proto")
	}
	// The old text must NOT be present.
	if strings.Contains(innerStr, DefaultTemplateText) {
		t.Error("old text marker still present in inner proto")
	}
	// The new conversation id must be present.
	if !strings.Contains(innerStr, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee") {
		t.Error("new conversation id not found in inner proto")
	}
	// The new request id must be present.
	if !strings.Contains(innerStr, "11111111-2222-3333-4444-555555555555") {
		t.Error("new request id not found in inner proto")
	}
}

// TestBuildFromTemplateRoundTrips parses the built frame to confirm it's still a
// valid type-0x0d frame with parseable outer JSON + inner proto.
func TestBuildFromTemplateRoundTrips(t *testing.T) {
	tplB64, _ := LoadTemplateB64("testdata/template_frame.b64")
	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "test",
		ConversationID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		RequestID:      "11111111-2222-3333-4444-555555555555",
	})
	if err != nil {
		t.Fatal(err)
	}
	if frame[0] != 0x0d {
		t.Fatalf("type byte = %#x, want 0x0d", frame[0])
	}
	var outer struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame[8:], &outer); err != nil {
		t.Fatalf("outer json unmarshal: %v", err)
	}
	if outer.ReqID != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("req-id = %q", outer.ReqID)
	}
	inner, err := base64.StdEncoding.DecodeString(outer.Payload)
	if err != nil {
		t.Fatalf("inner decode: %v", err)
	}
	if len(inner) == 0 {
		t.Fatal("empty inner proto")
	}
}

// TestBuildFromTemplateRequiresFields verifies input validation.
func TestBuildFromTemplateRequiresFields(t *testing.T) {
	_, err := BuildFromTemplate(TemplateOptions{})
	if err == nil {
		t.Error("expected error for empty options")
	}
}
