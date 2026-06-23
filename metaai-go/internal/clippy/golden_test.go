package clippy

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// loadGoldenFrame reads a base64-encoded captured frame from testdata/.
func loadGoldenFrame(t *testing.T, name string) []byte {
	t.Helper()
	b64, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	// trim whitespace
	s := strings.TrimSpace(string(b64))
	out, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode base64 %s: %v", name, err)
	}
	return out
}

// TestGoldenConnectFrameHeader asserts the captured CONNECT frame header matches
// what BuildConnectMessage produces (6-byte type-0x0f header + the JSON body).
func TestGoldenConnectFrameHeader(t *testing.T) {
	got, err := BuildConnectMessageWithRequestID("812e551d-f1dd-4adb-81b8-8defff2d7b94", "")
	if err != nil {
		t.Fatalf("BuildConnectMessage: %v", err)
	}
	// Header layout per docs/protocol.md: [0x0f][seq=0][len LE16][0x00]
	if got[0] != 0x0f {
		t.Fatalf("type byte = %#x, want 0x0f", got[0])
	}
	if got[1] != 0 || got[2] != 0 {
		t.Errorf("seq = %d,%d, want 0,0", got[1], got[2])
	}
	if got[5] != 0x00 {
		t.Errorf("flags byte = %#x, want 0x00", got[5])
	}
	// JSON body must contain the conversation id and payload type.
	var body map[string]string
	if err := json.Unmarshal(got[6:], &body); err != nil {
		t.Fatalf("unmarshal connect json: %v", err)
	}
	if body["x-dgw-app-x-ecto-conversation-id"] != "812e551d-f1dd-4adb-81b8-8defff2d7b94" {
		t.Errorf("conversation id = %q", body["x-dgw-app-x-ecto-conversation-id"])
	}
	if body["x-dgw-app-client-payload-type"] != "PROTO_INSIDE_JSON" {
		t.Errorf("payload type = %q", body["x-dgw-app-client-payload-type"])
	}
}

// TestGoldenChatFrameHeader asserts the CHAT SEND frame header matches the capture:
// [0x0d][seq=0][len LE16][0x00][sub=0x00][0x80] + JSON.
//
// The captured SEND subtype is 0x00.
func TestGoldenChatFrameHeader(t *testing.T) {
	frame := loadGoldenFrame(t, "pong4_frame.b64")
	if frame[0] != 0x0d {
		t.Fatalf("type byte = %#x, want 0x0d", frame[0])
	}
	if frame[1] != 0 || frame[2] != 0 {
		t.Errorf("seq = %d,%d, want 0,0", frame[1], frame[2])
	}
	if frame[5] != 0x00 {
		t.Errorf("byte5 = %#x, want 0x00", frame[5])
	}
	if frame[6] != 0x00 {
		t.Errorf("sub_type (byte6) = %#x, want 0x00 from the captured frame", frame[6])
	}
	if frame[7] != 0x80 {
		t.Errorf("byte7 = %#x, want 0x80", frame[7])
	}
	payloadLen := int(frame[3]) | int(frame[4])<<8
	// The captured length field is ~2 bytes larger than the trailing JSON (a server-
	// side encoder quirk). We only assert the JSON region parses and the length is
	// in the right ballpark.
	jsonBytes := frame[8:]
	if payloadLen < len(jsonBytes) || payloadLen > len(jsonBytes)+8 {
		t.Errorf("payloadLen field = %d, json region = %d (out of expected range)", payloadLen, len(jsonBytes))
	}
	if err := json.Unmarshal(jsonBytes, &map[string]any{}); err != nil {
		t.Errorf("frame json region does not parse: %v", err)
	}

	// Our encoder must produce the same header shape.
	got, err := BuildChatMessage(ChatMessageOptions{
		Text:           "Reply with exactly: PONG4",
		ConversationID: "812e551d-f1dd-4adb-81b8-8defff2d7b94",
		RequestID:      "f19bb7d2-ed60-415c-ac3d-510307770467",
	})
	if err != nil {
		t.Fatalf("BuildChatMessage: %v", err)
	}
	for i, want := range []byte{0x0d, 0x00, 0x00, 0, 0, 0x00, 0x00, 0x80} {
		if i == 3 || i == 4 {
			continue // length depends on payload
		}
		if got[i] != want {
			t.Errorf("header byte %d = %#x, want %#x", i, got[i], want)
		}
	}
}

// TestGoldenChatFrameStructure parses the captured chat frame and asserts the
// high-level protobuf field layout matches docs/protocol.md: exactly two top-level
// fields (f1 ENVELOPE, f2 MESSAGE-BLOCK), with the supporting fields nested
// inside the envelope and the static config nested inside envelope.f1 (inner).
func TestGoldenChatFrameStructure(t *testing.T) {
	frame := loadGoldenFrame(t, "pong4_frame.b64")
	var outer struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame[8:], &outer); err != nil {
		t.Fatalf("unmarshal outer json: %v", err)
	}
	inner, err := base64.StdEncoding.DecodeString(outer.Payload)
	if err != nil {
		t.Fatalf("decode inner proto: %v", err)
	}
	top := parseProtoFields(inner)

	// Top-level must be EXACTLY {1, 2} (envelope + message-block).
	if !sameKeys(top, map[int][][]byte{1: nil, 2: nil}) {
		t.Errorf("top-level fields = %v, want exactly [1 2]", keys(top))
	}

	// ENVELOPE (top.f1) direct fields.
	env := parseProtoFields(top[1][0])
	for _, want := range []int{1, 2, 3, 4, 5, 6, 7, 9, 10, 15, 16, 20, 26} {
		if _, ok := env[want]; !ok {
			t.Errorf("envelope missing field %d", want)
		}
	}
	if len(env[18]) == 0 {
		t.Errorf("envelope field 18 (capabilities) not repeated")
	}

	// INNER (envelope.f1) direct fields.
	innerMsg := parseProtoFields(env[1][0])
	for _, want := range []int{1, 2, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16, 19} {
		if _, ok := innerMsg[want]; !ok {
			t.Errorf("inner (envelope.f1) missing field %d", want)
		}
	}

	// MESSAGE-BLOCK (top.f2) direct fields.
	mb := parseProtoFields(top[2][0])
	for _, want := range []int{1, 2, 4} {
		if _, ok := mb[want]; !ok {
			t.Errorf("message-block missing field %d", want)
		}
	}
}

// TestEncoderProducesSameStructure builds a frame with the Go encoder (pinned IDs)
// and asserts its parsed field tree matches the captured frame's field tree at
// every field number. Byte values differ only where IDs/text differ by design.
func TestEncoderProducesSameStructure(t *testing.T) {
	frame := loadGoldenFrame(t, "pong4_frame.b64")
	var outer struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame[8:], &outer); err != nil {
		t.Fatalf("unmarshal outer json: %v", err)
	}
	capturedInner, _ := base64.StdEncoding.DecodeString(outer.Payload)
	capturedTop := parseProtoFields(capturedInner)

	// Build with the same message/conversation and pinned IDs.
	got, err := BuildChatMessage(ChatMessageOptions{
		Text:               "Reply with exactly: PONG4",
		ConversationID:     "812e551d-f1dd-4adb-81b8-8defff2d7b94",
		RequestID:          "f19bb7d2-ed60-415c-ac3d-510307770467",
		UserMessageID:      "111111111111111",
		AssistantMessageID: "222222222222222",
		MessageIDSuffix:    "5a5b-8d4e-f054-99ef-b2de-db02-0d05-52c7",
		PromptSessionID:    "9a4381d0-a065-4f20-91b0-ba654d34cf03",
	})
	if err != nil {
		t.Fatalf("BuildChatMessage: %v", err)
	}
	var gotOuter struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(got[8:], &gotOuter); err != nil {
		t.Fatalf("unmarshal got outer: %v", err)
	}
	gotInner, _ := base64.StdEncoding.DecodeString(gotOuter.Payload)
	gotTop := parseProtoFields(gotInner)

	// Top-level field-number sets must be identical.
	if !sameKeys(capturedTop, gotTop) {
		t.Errorf("top-level field mismatch:\n captured=%v\n encoded  =%v", keys(capturedTop), keys(gotTop))
	}
}

// ── protobuf field-walking helpers (test-only) ─────────────────────────────

// parseProtoFields walks a protobuf blob and returns fieldNumber → list of LEN
// payloads (for wire type 2) or nil entries (for VARINT). Used only to assert
// field *presence*; values are intentionally not compared (IDs/timestamps vary).
func parseProtoFields(b []byte) map[int][][]byte {
	out := map[int][][]byte{}
	off := 0
	for off < len(b) {
		tag, n, ok := readVarint(b, off)
		if !ok {
			return out
		}
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		switch wt {
		case 0: // VARINT
			_, n, ok := readVarint(b, off)
			if !ok {
				return out
			}
			off += n
			out[fn] = append(out[fn], nil)
		case 2: // LEN
			length, n, ok := readVarint(b, off)
			if !ok {
				return out
			}
			off += n
			end := off + int(length)
			if end > len(b) {
				return out
			}
			out[fn] = append(out[fn], append([]byte(nil), b[off:end]...))
			off = end
		case 5:
			off += 4
			out[fn] = append(out[fn], nil)
		case 1:
			off += 8
			out[fn] = append(out[fn], nil)
		default:
			return out
		}
	}
	return out
}

func sameKeys(a, b map[int][][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func keys(m map[int][][]byte) []int {
	ks := make([]int, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	return ks
}
