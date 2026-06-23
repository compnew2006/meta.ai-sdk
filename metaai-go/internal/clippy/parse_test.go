package clippy

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseFrameHandlesCapturedFrameShapes(t *testing.T) {
	connectBody := []byte(`{"code":200}`)
	connect := append([]byte{0x0f, 1, 0, byte(len(connectBody)), 0, 0}, connectBody...)
	inner := []byte{1, 2, 3}
	outer, _ := json.Marshal(map[string]string{"payload": base64.StdEncoding.EncodeToString(inner)})
	proto := append([]byte{0x0d, 2, 0, byte(len(outer)), 0, 0, 0, 0x80}, outer...)
	cases := []struct {
		name     string
		data     []byte
		wantType FrameType
		wantRaw  bool
	}{
		{"connect", connect, TypeConnect, false}, {"proto", proto, TypeProto, false}, {"unknown", []byte{0x01, 0, 0, 0, 0, 0, 'x'}, FrameType(1), true}, {"raw connect", []byte{0x0f, 0, 0, 1, 0, 0, 'x'}, TypeConnect, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := ParseFrame(tc.data)
			if err != nil || f.Type != tc.wantType {
				t.Fatalf("f=%+v err=%v", f, err)
			}
			if tc.name == "proto" && string(f.InnerProto) != string(inner) {
				t.Fatalf("inner=%v", f.InnerProto)
			}
			if tc.wantRaw && f.Payload["_raw"] == nil {
				t.Fatalf("payload=%v", f.Payload)
			}
		})
	}
	if _, err := ParseFrame([]byte{1}); err == nil {
		t.Fatal("expected short frame error")
	}
}

func TestParseFrameDecodesSubtype25JSON(t *testing.T) {
	js := []byte(`{"type":"full"}`)
	field := append([]byte{0x2a, byte(len(js))}, js...)
	frame := append([]byte{0x0d, 0, 0, byte(len(field)), 0, 0, 0x25, 0}, field...)
	f, err := ParseFrame(frame)
	if err != nil || f.Payload["type"] != "full" {
		t.Fatalf("f=%+v err=%v", f, err)
	}
}

func TestResponseTextExtractsFullAndPatchPayloads(t *testing.T) {
	full := []byte("xx" + `{"type":"full","response":{"sections":[{"view_model":{"primitive":{"text":"answer","thought_text":"thought"}}}]}}` + "yy")
	patch := []byte("xx" + `{"type":"patch","operations":[{"op":"replace","value":"no"},{"op":"delta","value":"delta"}]}`)
	if got, ok := ResponseText(full); !ok || got != "answer" {
		t.Fatalf("full=%q %v (want 'answer' only, thought_text should be skipped)", got, ok)
	}
	if got, ok := ResponseText(patch); !ok || got != "delta" {
		t.Fatalf("patch=%q %v", got, ok)
	}
	for _, bad := range [][]byte{nil, {0x0f}, []byte("no-json"), []byte(`xx{"type":"full"`)} {
		if got, ok := ResponseText(bad); ok || got != "" {
			t.Fatalf("bad=%q => %q %v", bad, got, ok)
		}
	}
	if s, ok := extractEmbeddedJSON([]byte(`x{"a":"}"}tail`)); !ok || !strings.Contains(s, `"}"`) {
		t.Fatalf("json=%q %v", s, ok)
	}
}

// TestIsFullResponseIgnoresEmbeddedThinkingFullInPatchFrame guards against the
// regression where IsFullResponse used a substring scan and returned true for a
// patch frame whose protobuf also embedded an earlier thinking-status
// "type":"full" section. That false positive made recvLoop terminate the stream
// mid-answer. IsFullResponse MUST agree with ResponseTextInfo: a frame is "full"
// only when its actionable field-5 payload has top-level type:"full".
func TestIsFullResponseIgnoresEmbeddedThinkingFullInPatchFrame(t *testing.T) {
	// Build a TypeProto (0x0d) frame whose field-5 payload is a PATCH delta,
	// but whose protobuf bytes also contain a thinking-status "full" JSON
	// earlier in the buffer (simulating the live capture layout).
	patchJSON := []byte(`{"seq":1,"type":"patch","operations":[{"op":"delta","value":" for asking"}]}`)
	// field 5 (tag 0x2a) LEN <patchJSON>
	f5 := append([]byte{0x2a, byte(len(patchJSON))}, patchJSON...)
	// Prepend a thinking-status "full" JSON as field 1 (tag 0x0a) so the raw
	// bytes contain `"type":"full"` before the patch payload — the substring
	// trap. readField(proto,5) still finds only the patch. Kept under 128 bytes
	// so the single-byte length varint is valid.
	thinkingJSON := []byte(`{"seq":0,"type":"full","response":{"sections":[]}}`)
	f1 := append([]byte{0x0a, byte(len(thinkingJSON))}, thinkingJSON...)
	proto := append(f1, f5...)

	header := []byte{
		0x0d,                                  // TypeProto
		0, 0,                                  // Seq
		byte(len(proto) & 0xff), byte((len(proto) >> 8) & 0xff), // Length
		0,    // flag
		0x01, // subtype
		0,    // flag
	}
	frame := append(header, proto...)

	if IsFullResponse(frame) {
		t.Fatal("IsFullResponse=true for a patch frame with embedded thinking 'full'; this truncates the stream mid-answer")
	}
	got, ok, isFull := ResponseTextInfo(frame)
	if !ok || got != " for asking" {
		t.Fatalf("ResponseTextInfo: got=%q ok=%v (want the patch delta)", got, ok)
	}
	if isFull {
		t.Fatal("isFull=true for a patch frame")
	}
}

func TestResponseTextExtractsProtobufDeltaText(t *testing.T) {
	// 1. Test Delta Text (Field 1 of Field 4)
	deltaText := []byte("Hello protobuf!")
	f1Inner := append([]byte{0x0a, byte(len(deltaText))}, deltaText...)
	f4Inner := append([]byte{0x22, byte(len(f1Inner))}, f1Inner...)
	f1Outer := append([]byte{0x0a, byte(len(f4Inner))}, f4Inner...)

	length := len(f1Outer)
	header := []byte{
		0x0d, // TypeProto
		0, 0, // Seq
		byte(length & 0xff), byte((length >> 8) & 0xff), // Length
		0,    // Unused
		0x46, // Subtype
		0,    // Unused
	}
	frame := append(header, f1Outer...)

	gotText, ok, isFull := ResponseTextInfo(frame)
	if !ok {
		t.Fatal("expected ResponseTextInfo to extract protobuf delta, but it failed")
	}
	if gotText != "Hello protobuf!" {
		t.Errorf("got text %q, want 'Hello protobuf!'", gotText)
	}
	if isFull {
		t.Error("expected isFull to be false for delta text frame")
	}

	// 2. Test Full Text (Field 2 of Field 4)
	fullText := []byte("Hello full protobuf!")
	f2Inner := append([]byte{0x12, byte(len(fullText))}, fullText...)
	f4InnerFull := append([]byte{0x22, byte(len(f2Inner))}, f2Inner...)
	f1OuterFull := append([]byte{0x0a, byte(len(f4InnerFull))}, f4InnerFull...)

	lengthFull := len(f1OuterFull)
	headerFull := []byte{
		0x0d,
		0, 0,
		byte(lengthFull & 0xff), byte((lengthFull >> 8) & 0xff),
		0,
		0x46,
		0,
	}
	frameFull := append(headerFull, f1OuterFull...)

	gotTextFull, okFull, isFullFull := ResponseTextInfo(frameFull)
	if !okFull {
		t.Fatal("expected ResponseTextInfo to extract protobuf full text, but it failed")
	}
	if gotTextFull != "Hello full protobuf!" {
		t.Errorf("got text %q, want 'Hello full protobuf!'", gotTextFull)
	}
	if !isFullFull {
		t.Error("expected isFull to be true for full text frame")
	}
	// IsFullResponse MUST agree with ResponseTextInfo here: a genuine full
	// response (detected via the raw-protobuf fallback path, field 4 -> field 2)
	// must classify as full so recvLoop can terminate on its duplicate.
	if !IsFullResponse(frameFull) {
		t.Error("IsFullResponse=false for a genuine full-response frame; recvLoop would never detect stream completion")
	}

	// 3. Test Clamped/Truncated Protobuf (where length fields specify a size larger than the actual bytes)
	f1InnerClamped := append([]byte{0x0a, byte(len(deltaText))}, deltaText...)
	hugeLenVarint := []byte{0xc0, 0x9a, 0x0c} // 200000 in varint
	f4InnerClamped := append(append([]byte{0x22}, hugeLenVarint...), f1InnerClamped...)
	f1OuterClamped := append(append([]byte{0x0a}, hugeLenVarint...), f4InnerClamped...)

	lengthClamped := len(f1OuterClamped)
	headerClamped := []byte{
		0x0d,
		0, 0,
		byte(lengthClamped & 0xff), byte((lengthClamped >> 8) & 0xff),
		0,
		0x46,
		0,
	}
	frameClamped := append(headerClamped, f1OuterClamped...)

	gotTextClamped, okClamped, isFullClamped := ResponseTextInfo(frameClamped)
	if !okClamped {
		t.Fatal("expected ResponseTextInfo to extract text from clamped/truncated protobuf, but it failed")
	}
	if gotTextClamped != "Hello protobuf!" {
		t.Errorf("got text %q, want 'Hello protobuf!'", gotTextClamped)
	}
	if isFullClamped {
		t.Error("expected isFull to be false for delta text frame")
	}
}
