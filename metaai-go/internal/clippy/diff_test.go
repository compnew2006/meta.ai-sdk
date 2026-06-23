package clippy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestDiffBuiltVsGolden is a diagnostic test that prints the top-level + nested
// field tree of both the golden captured frame and a freshly-built frame, so the
// two can be eyeballed for the field that makes the server reject the built one.
//
// Skipped unless META_AI_DIFF=1 is set (it's a dev aid, not a regression test).
func TestDiffBuiltVsGolden(t *testing.T) {
	if os.Getenv("META_AI_DIFF") == "" {
		t.Skip("set META_AI_DIFF=1 to run the built-vs-golden diff diagnostic")
	}
	golden := loadGoldenFrame(t, "pong4_frame.b64")

	built, err := BuildChatMessage(ChatMessageOptions{
		Text:            "Reply with exactly: PONG4",
		ConversationID:  "812e551d-f1dd-4adb-81b8-8defff2d7b94",
		RequestID:       "f19bb7d2-ed60-415c-ac3d-510307770467",
		PromptSessionID: "9a4381d0-a065-4f20-91b0-ba654d34cf03",
		SessionID:       "c8653f37-5999-4059-804e-3e31db74a177",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== GOLDEN ===")
	t.Log(treeString(innerProtoBytes(t, golden)))
	t.Log("=== BUILT ===")
	t.Log(treeString(innerProtoBytes(t, built)))

	// Persist both inner protos as hex for offline payload-level diffing.
	gp := innerProtoBytes(t, golden)
	bp := innerProtoBytes(t, built)
	_ = os.WriteFile("/tmp/golden_inner.hex", []byte(hexEncode(gp)), 0o644)
	_ = os.WriteFile("/tmp/built_inner.hex", []byte(hexEncode(bp)), 0o644)
	t.Logf("wrote /tmp/golden_inner.hex (%d bytes) and /tmp/built_inner.hex (%d bytes)", len(gp), len(bp))
}

func hexEncode(b []byte) string {
	const h = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, x := range b {
		out[i*2] = h[x>>4]
		out[i*2+1] = h[x&0xf]
	}
	return string(out)
}

func innerProtoBytes(t *testing.T, frame []byte) []byte {
	t.Helper()
	var o struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame[8:], &o); err != nil {
		t.Fatal(err)
	}
	p, err := base64.StdEncoding.DecodeString(o.Payload)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// treeString returns a multi-line indented tree of field numbers / values for a
// protobuf blob, recursing up to 4 levels into non-printable LEN fields.
func treeString(b []byte) string {
	var sb strings.Builder
	walkTree(&sb, b, "", 4)
	return sb.String()
}

func walkTree(sb *strings.Builder, b []byte, prefix string, depth int) {
	off := 0
	for off < len(b) {
		tag, n, ok := readVarint(b, off)
		if !ok {
			return
		}
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		switch wt {
		case 0:
			v, n, _ := readVarint(b, off)
			off += n
			fmt.Fprintf(sb, "%sf%d VARINT=%d\n", prefix, fn, v)
		case 2:
			l, n, ok := readVarint(b, off)
			if !ok {
				return
			}
			off += n
			if off+int(l) > len(b) {
				fmt.Fprintf(sb, "%sf%d OVERFLOW\n", prefix, fn)
				return
			}
			chunk := b[off : off+int(l)]
			off += int(l)
			extra := ""
			if isASCIIPrint(chunk) {
				extra = " = \"" + string(chunk) + "\""
			}
			fmt.Fprintf(sb, "%sf%d LEN[%d]%s\n", prefix, fn, l, extra)
			if !isASCIIPrint(chunk) && l > 1 && depth > 0 {
				walkTree(sb, chunk, prefix+"  ", depth-1)
			}
		case 5:
			off += 4
		case 1:
			off += 8
		default:
			return
		}
	}
}

func isASCIIPrint(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c == 9 || c == 10 || c == 13 {
			continue
		}
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}
