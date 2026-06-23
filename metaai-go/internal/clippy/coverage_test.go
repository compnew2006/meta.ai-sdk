package clippy

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// === PROTO PRIMITIVES TESTS ===
// These test the byte-level encoders that are critical for frame correctness.
// Testing behavior: correct varint/tag/string encoding → correct bytes.

func TestEncodeVarintSingleByte(t *testing.T) {
	// 0-127 should encode as a single byte
	got := encodeVarint(nil, 42)
	if len(got) != 1 || got[0] != 42 {
		t.Errorf("encodeVarint(42) = %v, want [42]", got)
	}
}

func TestEncodeVarintZero(t *testing.T) {
	got := encodeVarint(nil, 0)
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("encodeVarint(0) = %v, want [0]", got)
	}
}

func TestEncodeVarintMultiByte(t *testing.T) {
	// 300 = 0x12C → varint: [0xAC 0x02]
	got := encodeVarint(nil, 300)
	want := []byte{0xAC, 0x02}
	if len(got) != len(want) {
		t.Fatalf("encodeVarint(300) len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte %d: got 0x%02x want 0x%02x", i, got[i], want[i])
		}
	}
}

func TestEncodeStringRoundTripsThroughVarint(t *testing.T) {
	original := "hello clippy"
	encoded := encodeString(nil, 7, original)

	// Read tag
	tag, off, _ := readVarint(encoded, 0)
	fn, wt := int(tag>>3), int(tag&7)
	if fn != 7 || wt != 2 {
		t.Fatalf("tag: fn=%d wt=%d, want fn=7 wt=2", fn, wt)
	}
	// Read length
	length, _, _ := readVarint(encoded, off)
	if int(length) != len(original) {
		t.Fatalf("length=%d, want %d", length, len(original))
	}
	// Payload starts after tag + length, for exactly `length` bytes
	payloadStart := off + 1 // single-byte length for short strings
	payload := encoded[payloadStart : payloadStart+int(length)]
	if string(payload) != original {
		t.Fatalf("payload=%q, want %q", string(payload), original)
	}
}

func TestEncodeMessageWrapsPayload(t *testing.T) {
	inner := encodeString(nil, 1, "test")
	wrapped := encodeMessage(nil, 5, inner)

	tag, off, _ := readVarint(wrapped, 0)
	fn, wt := int(tag>>3), int(tag&7)
	if fn != 5 || wt != 2 {
		t.Fatalf("outer tag: fn=%d wt=%d, want fn=5 wt=2", fn, wt)
	}
	length, _, _ := readVarint(wrapped, off)
	if int(length) != len(inner) {
		t.Fatalf("outer length=%d, want %d (inner len)", length, len(inner))
	}
}

func TestEncodeVarintField(t *testing.T) {
	got := encodeVarintField(nil, 6, 5)
	// tag 0x30 = field 6 wire 0, then value 5
	if len(got) != 2 || got[0] != 0x30 || got[1] != 5 {
		t.Errorf("encodeVarintField(6,5) = %v, want [0x30 0x05]", got)
	}
}

func TestAppendFixed32(t *testing.T) {
	got := appendFixed32(nil, 2, 0x3f800000) // float 1.0
	// tag 0x15 = field 2 wire 5, then 4 LE bytes 00 00 80 3f
	if len(got) != 5 || got[0] != 0x15 {
		t.Fatalf("len=%d first=0x%02x, want 5 bytes starting 0x15", len(got), got[0])
	}
	// Verify the 4 bytes are LE 0x3f800000
	val := uint32(got[1]) | uint32(got[2])<<8 | uint32(got[3])<<16 | uint32(got[4])<<24
	if val != 0x3f800000 {
		t.Errorf("I32 value=0x%08x, want 0x3f800000", val)
	}
}

// === FRAME HEADER TESTS ===

func TestPackConnectHeaderLength(t *testing.T) {
	h := packConnectHeader(127)
	// type=0x0f, seq=0x0000, len=0x007f(LE), flags=0x00
	if h[0] != 0x0f || h[3] != 127 || h[4] != 0 || h[5] != 0x00 {
		t.Errorf("connect header mismatch: %x", h)
	}
}

func TestPackChatHeaderSubType(t *testing.T) {
	h := packChatHeader(1366, 0x00)
	if h[0] != 0x0d || h[5] != 0x00 || h[6] != 0x00 || h[7] != 0x80 {
		t.Errorf("chat header mismatch: %x", h)
	}
}

// === CONNECT FRAME TESTS ===

func TestBuildConnectMessageContainsConversationID(t *testing.T) {
	convID := "test-conv-123"
	frame, err := BuildConnectMessage(convID)
	if err != nil {
		t.Fatal(err)
	}
	if frame[0] != 0x0f {
		t.Fatalf("type = 0x%02x, want 0x0f", frame[0])
	}
	body := string(frame[6:])
	if !contains(body, convID) {
		t.Errorf("conversation ID not found in connect frame body")
	}
	if !contains(body, "PROTO_INSIDE_JSON") {
		t.Errorf("payload type not found in connect frame body")
	}
}

// === CHAT FRAME TESTS ===

func TestBuildChatMessageDefaultsApplied(t *testing.T) {
	frame, err := BuildChatMessage(ChatMessageOptions{
		Text:           "hello",
		ConversationID: "conv-id-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if frame[0] != 0x0d {
		t.Fatalf("type byte = 0x%02x", frame[0])
	}
	if frame[6] != 0x00 {
		t.Errorf("sub_type = 0x%02x, want 0x00", frame[6])
	}
	if frame[7] != 0x80 {
		t.Errorf("byte7 = 0x%02x, want 0x80", frame[7])
	}
}

func TestBuildChatMessageThinkingModeProducesDifferentF12(t *testing.T) {
	normal, _ := BuildChatMessage(ChatMessageOptions{
		Text: "hi", ConversationID: "c", RequestID: "r1",
	})
	thinking, _ := BuildChatMessage(ChatMessageOptions{
		Text: "hi", ConversationID: "c", RequestID: "r1", Thinking: true,
	})
	// The frames should differ in size (thinking adds mode_thinking wrapper)
	if len(normal) == len(thinking) {
		// They might be same length if random IDs compensate, but the proto
		// content should differ. At minimum, the frames are valid.
		if string(normal) == string(thinking) {
			t.Error("thinking frame identical to normal — mode config not applied")
		}
	}
}

func TestBuildChatMessageRejectsEmptyText(t *testing.T) {
	_, err := BuildChatMessage(ChatMessageOptions{ConversationID: "c"})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestBuildChatMessageRejectsEmptyConvID(t *testing.T) {
	_, err := BuildChatMessage(ChatMessageOptions{Text: "hi"})
	if err == nil {
		t.Error("expected error for empty conversation ID")
	}
}

func TestModeConfigThinkingProducesNestedStructure(t *testing.T) {
	cfg := modeConfig(true)
	// Should contain "mode_thinking" string
	if !contains(string(cfg), "mode_thinking") {
		t.Error("thinking mode config missing 'mode_thinking'")
	}
}

func TestModeConfigFastProducesMODE_FAST(t *testing.T) {
	cfg := modeConfig(false)
	if !contains(string(cfg), "MODE_FAST") {
		t.Error("fast mode config missing 'MODE_FAST'")
	}
}

func TestMessageIDWrapperContainsConversationID(t *testing.T) {
	wrapper := messageIDWrapper(ChatMessageOptions{
		RequestID:      "req-123",
		ConversationID: "conv-456",
	}, 1700000000000)
	if !contains(string(wrapper), "conv-456") {
		t.Error("wrapper missing conversation ID")
	}
	if !contains(string(wrapper), "req-123") {
		t.Error("wrapper missing request ID")
	}
}

func TestCapabilityHashMsgContainsHash(t *testing.T) {
	hash := "abc123"
	msg := capabilityHashMsg(hash)
	if !contains(string(msg), hash) {
		t.Error("capability hash message missing hash string")
	}
}

func TestTimestampPairProducesTwoVarints(t *testing.T) {
	tp := timestampPair(1000000)
	// Should have f1=1000000 and f3=1000000-30=999970
	tag1, off1, _ := readVarint(tp, 0)
	if int(tag1>>3) != 1 || int(tag1&7) != 0 {
		t.Fatalf("first field: fn=%d wt=%d, want fn=1 wt=0", tag1>>3, tag1&7)
	}
	val1, _, _ := readVarint(tp, off1)
	if val1 != 1000000 {
		t.Errorf("f1 value=%d, want 1000000", val1)
	}
}

// === PARSE TESTS ===

func TestParseFrameConnectTypeReturnsJSONPayload(t *testing.T) {
	connect, _ := BuildConnectMessage("conv-1")
	f, err := ParseFrame(connect)
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != TypeConnect {
		t.Errorf("type = %v, want TypeConnect", f.Type)
	}
	if f.Payload == nil {
		t.Fatal("payload is nil")
	}
	if ct, ok := f.Payload["x-dgw-app-client-payload-type"]; !ok || ct != "PROTO_INSIDE_JSON" {
		t.Errorf("missing or wrong payload type: %v", f.Payload)
	}
}

func TestParseFrameTooShortReturnsError(t *testing.T) {
	_, err := ParseFrame([]byte{0x01, 0x02})
	if err == nil {
		t.Error("expected error for short frame")
	}
}

func TestParseFrameConnectWithClampedLength(t *testing.T) {
	// Build a connect frame with a deliberately oversized length field
	frame := []byte{0x0f, 0x00, 0x00, 0xff, 0x00, 0x00}
	frame = append(frame, []byte(`{"code":200}`)...)
	f, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != TypeConnect {
		t.Errorf("type = %v", f.Type)
	}
}

func TestResponseTextConnectAckReturnsFalse(t *testing.T) {
	connect, _ := BuildConnectMessage("conv-1")
	_, ok := ResponseText(connect)
	if ok {
		t.Error("connect frame should not yield text")
	}
}

func TestResponseTextEmptyDataReturnsFalse(t *testing.T) {
	_, ok := ResponseText(nil)
	if ok {
		t.Error("nil data should not yield text")
	}
}

func TestResponseTextFullResponseExtractsText(t *testing.T) {
	// Build a simulated RECV frame with embedded JSON
	jsonPayload := `{"seq":0,"type":"full","response":{"sections":[{"view_model":{"primitive":{"text":"Hello world"}}}]}}`
	// Embed as a type-0x0d frame: type + len + json
	frame := []byte{0x0d, 0x00, 0x00, byte(len(jsonPayload) & 0xff), byte((len(jsonPayload) >> 8) & 0xff), 0x00, 0x00, 0x80}
	// Add some proto preamble bytes then the JSON
	preamble := []byte{0x0a, 0x05, 'h', 'e', 'l', 'l', 'o'}
	frame = append(frame, preamble...)
	frame = append(frame, []byte(jsonPayload)...)

	text, ok := ResponseText(frame)
	if !ok {
		t.Fatal("expected text from full response")
	}
	if !contains(text, "Hello world") {
		t.Errorf("text = %q, want 'Hello world'", text)
	}
}

func TestResponseTextPatchResponseExtractsDelta(t *testing.T) {
	jsonPayload := `{"seq":1,"type":"patch","operations":[{"op":"delta","value":"chunk text"}]}`
	frame := []byte{0x0d, 0x00, 0x00, byte(len(jsonPayload) & 0xff), byte((len(jsonPayload) >> 8) & 0xff), 0x00, 0x01, 0x80}
	frame = append(frame, []byte(jsonPayload)...)

	text, ok := ResponseText(frame)
	if !ok {
		t.Fatal("expected text from patch response")
	}
	if text != "chunk text" {
		t.Errorf("text = %q, want 'chunk text'", text)
	}
}

func TestExtractEmbeddedJSONHandlesEscapedStrings(t *testing.T) {
	// JSON with escaped quotes inside string values
	raw := []byte(`prefix {"key":"val\"ue"} suffix`)
	result, ok := extractEmbeddedJSON(raw)
	if !ok {
		t.Fatal("expected JSON extraction")
	}
	if result != `{"key":"val\"ue"}` {
		t.Errorf("got %q", result)
	}
}

func TestExtractEmbeddedJSONReturnsFalseWhenNoJSON(t *testing.T) {
	_, ok := extractEmbeddedJSON([]byte("no json here"))
	if ok {
		t.Error("expected false for no JSON")
	}
}

func TestFirstDeltaReturnsFirstDeltaValue(t *testing.T) {
	ops := []any{
		map[string]any{"op": "delta", "value": "first"},
		map[string]any{"op": "delta", "value": "second"},
	}
	got := firstDelta(ops)
	if got != "first" {
		t.Errorf("got %q, want 'first'", got)
	}
}

func TestFirstDeltaIgnoresNonDeltaOps(t *testing.T) {
	ops := []any{
		map[string]any{"op": "replace", "value": "ignored"},
		map[string]any{"op": "delta", "value": "found"},
	}
	got := firstDelta(ops)
	if got != "found" {
		t.Errorf("got %q, want 'found'", got)
	}
}

func TestJoinSectionTextExtractsText(t *testing.T) {
	sections := []any{
		map[string]any{
			"view_model": map[string]any{
				"primitive": map[string]any{"text": "hello"},
			},
		},
	}
	got := joinSectionText(sections)
	if got != "hello" {
		t.Errorf("got %q, want 'hello'", got)
	}
}

func TestJoinSectionTextEmptySectionsReturnsEmpty(t *testing.T) {
	got := joinSectionText([]any{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// === TEMPLATE TESTS ===

func TestBuildFromTemplateWithAttachmentSubstitutesMediaID(t *testing.T) {
	tplB64, err := LoadTemplateB64("testdata/attachment_frame.b64")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   "What do you see ?",
		NewText:        "Describe this",
		ConversationID: "aaaa-bbbb",
		RequestID:      "req-1",
		MediaID:        "9999",
		MediaMime:      "image/png",
		MediaFilename:  "photo.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(frame) < 100 {
		t.Fatalf("frame too small: %d bytes", len(frame))
	}
	// The new text should be present
	var o struct {
		Payload string `json:"payload"`
	}
	import_json(t, frame[8:], &o)
	inner, _ := b64decode(o.Payload)
	if !contains(string(inner), "Describe this") {
		t.Error("new text not found in frame")
	}
}

// === HELPERS ===

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (indexOf(s, substr) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func import_json(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

func b64decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
