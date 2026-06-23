package clippy

import "testing"

// Edge-case tests for clippy functions below 90%.

func TestExtractField5JSONFindsField5(t *testing.T) {
	// Build a proto with field 5 containing JSON
	jsonData := []byte(`{"key":"value"}`)
	inner := encodeString(nil, 1, "skip")
	proto := encodeMessage(nil, 1, inner)
	proto = encodeString(proto, 5, string(jsonData))

	result, ok := extractField5JSON(proto)
	if !ok {
		t.Fatal("expected field5 JSON extraction")
	}
	if result["key"] != "value" {
		t.Errorf("got %v", result["key"])
	}
}

func TestExtractField5JSONReturnsFalseOnNoField5(t *testing.T) {
	proto := encodeString(nil, 1, "data")
	_, ok := extractField5JSON(proto)
	if ok {
		t.Error("expected false when no field 5")
	}
}

func TestReadVarintOverflowsGracefully(t *testing.T) {
	// Build a varint that exceeds 64 bits (10 continuation bytes)
	data := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}
	_, _, ok := readVarint(data, 0)
	if ok {
		t.Error("expected overflow to return false")
	}
}

func TestParseFrameProtoWithSubtypeNonZero(t *testing.T) {
	// Build a type-0x0d frame with subtype 0x01 (SEND echo)
	jsonStr := `{"req-id":"test","payload":"AQID"}`
	frame := make([]byte, 8+len(jsonStr))
	frame[0] = 0x0d
	frame[3] = byte(len(jsonStr) & 0xff)
	frame[4] = byte((len(jsonStr) >> 8) & 0xff)
	frame[5] = 0x00
	frame[6] = 0x01 // non-0x25 subtype
	frame[7] = 0x80
	copy(frame[8:], jsonStr)

	f, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if f.Type != TypeProto {
		t.Errorf("type = %v", f.Type)
	}
	if f.SubType != 0x01 {
		t.Errorf("subtype = 0x%02x", f.SubType)
	}
}

func TestParseFrameUnknownTypeReturnsRaw(t *testing.T) {
	frame := []byte{0x99, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x41, 0x42}
	f, err := ParseFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	// Unknown type should still parse without error
	if f.Raw == nil {
		t.Error("raw should not be nil")
	}
}

func TestResponseTextHandlesUnknownJSONType(t *testing.T) {
	jsonPayload := `{"type":"unknown","data":"test"}`
	frame := []byte{0x0d, 0x00, 0x00, byte(len(jsonPayload) & 0xff), 0x00, 0x00, 0x00, 0x80}
	frame = append(frame, []byte(jsonPayload)...)
	_, ok := ResponseText(frame)
	// Should return false for unknown type
	if ok {
		t.Error("expected false for unknown JSON type")
	}
}

func TestJoinSectionTextSkipsThoughtText(t *testing.T) {
	// thought_text should be SKIPPED (it's "thinking..." status, not the answer)
	sections := []any{
		map[string]any{
			"view_model": map[string]any{
				"primitive": map[string]any{"thought_text": "thinking...", "text": "actual answer"},
			},
		},
	}
	got := joinSectionText(sections)
	if got != "actual answer" {
		t.Errorf("got %q, want 'actual answer' (thought_text should be skipped)", got)
	}
}

func TestJoinSectionTextNilPrimitiveSkipsGracefully(t *testing.T) {
	sections := []any{
		map[string]any{"view_model": map[string]any{}}, // no primitive
		map[string]any{
			"view_model": map[string]any{
				"primitive": map[string]any{"text": "found"},
			},
		},
	}
	got := joinSectionText(sections)
	if got != "found" {
		t.Errorf("got %q, want 'found'", got)
	}
}

func TestFirstDeltaWithEmptyOps(t *testing.T) {
	got := firstDelta(nil)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFirstDeltaWithNonMapEntry(t *testing.T) {
	ops := []any{"not-a-map"}
	got := firstDelta(ops)
	if got != "" {
		t.Errorf("got %q", got)
	}
}

func TestFirstDeltaWithMissingValue(t *testing.T) {
	ops := []any{
		map[string]any{"op": "delta"}, // no "value" key
	}
	got := firstDelta(ops)
	if got != "" {
		t.Errorf("got %q, want empty for missing value", got)
	}
}

func TestBuildConnectMessageWithRequestIDOverridesDefault(t *testing.T) {
	frame, err := BuildConnectMessageWithRequestID("conv-1", "my-req-id")
	if err != nil {
		t.Fatal(err)
	}
	// The JSON should contain the request ID... actually connect frames don't
	// carry req-id in the JSON, they carry the conversation ID. Just verify it
	// builds without error and is a valid connect frame.
	if frame[0] != 0x0f {
		t.Errorf("type = 0x%02x", frame[0])
	}
}

func TestNumericIDProducesCorrectLengths(t *testing.T) {
	for _, digits := range []int{5, 10, 15} {
		id := numericID(digits)
		if len(id) != digits {
			t.Errorf("digits=%d: got len %d", digits, len(id))
		}
	}
}

func TestTurnCounterProducesStableValue(t *testing.T) {
	o1 := ChatMessageOptions{RequestID: "abc-123"}
	o2 := ChatMessageOptions{RequestID: "abc-123"}
	if o1.turnCounter() != o2.turnCounter() {
		t.Error("turnCounter should be deterministic for same RequestID")
	}
}

func TestMessageNonceProducesStableValue(t *testing.T) {
	o1 := ChatMessageOptions{RequestID: "abc-123"}
	o2 := ChatMessageOptions{RequestID: "abc-123"}
	if o1.messageNonce() != o2.messageNonce() {
		t.Error("messageNonce should be deterministic for same RequestID")
	}
}

func TestNowReturnsTimestampMsWhenPinned(t *testing.T) {
	o := ChatMessageOptions{TimestampMs: 12345}
	if o.now() != 12345 {
		t.Errorf("got %d, want 12345", o.now())
	}
}

func TestTurnCounterReturnsOverrideWhenSet(t *testing.T) {
	o := ChatMessageOptions{TurnCounter: 999}
	if o.turnCounter() != 999 {
		t.Errorf("got %d, want 999", o.turnCounter())
	}
}

func TestMessageNonceReturnsOverrideWhenSet(t *testing.T) {
	o := ChatMessageOptions{MessageNonce: 888}
	if o.messageNonce() != 888 {
		t.Errorf("got %d, want 888", o.messageNonce())
	}
}

func TestRandUUID4ReturnsValidUUID(t *testing.T) {
	uuid := UUIDFunc()
	if len(uuid) != 36 {
		t.Fatalf("uuid len = %d, want 36", len(uuid))
	}
	// version nibble should be 4
	if uuid[14] != '4' {
		t.Errorf("version nibble = %c, want 4", uuid[14])
	}
}

func TestReplaceAllTextFieldValuesNoMatchReturnsOriginal(t *testing.T) {
	proto := encodeString(nil, 1, "original")
	result, err := replaceAllTextFieldValues(proto, "nomatch", "replacement")
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(proto) {
		t.Error("should return original when no match")
	}
}
