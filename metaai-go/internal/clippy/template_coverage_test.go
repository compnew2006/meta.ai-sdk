package clippy

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// Template + attachment coverage tests.

func TestInjectAttachmentIntoTextOnlyTemplate(t *testing.T) {
	tplB64, _ := LoadTemplateB64("testdata/template_frame.b64")
	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "analyze this",
		ConversationID: "conv-xyz",
		RequestID:      "req-xyz",
		MediaID:        "12345",
		MediaMime:      "image/png",
		MediaFilename:  "photo.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(frame) < 1000 {
		t.Fatalf("frame too small: %d", len(frame))
	}
}

func TestSubstituteMediaIDInAttachmentReplacesOldID(t *testing.T) {
	// Build a minimal message-block with an f3 attachment containing a media ID
	attInner := encodeVarintField(nil, 1, 9999)
	attachment := encodeMessage(nil, 1, attInner)
	attachment = encodeVarintField(attachment, 2, 1)
	attachment = encodeMessage(attachment, 3, nil)
	attachment = encodeString(attachment, 6, "image/png")
	attachment = encodeString(attachment, 7, "old.png")
	msgBlock := encodeMessage(nil, 3, attachment)
	msgBlock = encodeString(msgBlock, 2, "text")
	msgBlock = encodeMessage(msgBlock, 4, encodeString(nil, 1, "0"))

	result := substituteMediaIDInAttachment(msgBlock, "88888", "image/jpeg", "new.jpg")

	// Walk result, find f3, check it contains the new media ID
	found := false
	off := 0
	for off < len(result) {
		tag, n, _ := readVarint(result, off)
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		if wt == 2 {
			length, ln, _ := readVarint(result, off)
			payload := result[off+ln : off+ln+int(length)]
			off += ln + int(length)
			if fn == 3 {
				// This is the attachment — check for new media ID
				if contains(string(payload), "new.jpg") {
					found = true
				}
			}
		} else if wt == 0 {
			_, vn, _ := readVarint(result, off)
			off += vn
		} else {
			break
		}
	}
	if !found {
		t.Error("new filename not found in substituted attachment")
	}
}

func TestBytesReplaceAllSameLength(t *testing.T) {
	original := []byte("hello world hello")
	result := bytesReplaceAll(original, []byte("hello"), []byte("HELLO"))
	if string(result) != "HELLO world HELLO" {
		t.Errorf("got %q", string(result))
	}
}

func TestBytesReplaceAllDifferentLength(t *testing.T) {
	original := []byte("abc abc")
	result := bytesReplaceAll(original, []byte("abc"), []byte("XY"))
	if string(result) != "XY XY" {
		t.Errorf("got %q", string(result))
	}
}

func TestBytesReplaceAllEmptyOldReturnsOriginal(t *testing.T) {
	original := []byte("test")
	result := bytesReplaceAll(original, []byte(""), []byte("X"))
	if string(result) != "test" {
		t.Errorf("got %q", string(result))
	}
}

func TestIsUUIDCharsValidUUID(t *testing.T) {
	uuid := []byte("550e8400-e29b-41d4-a716-446655440000")
	if !isUUIDChars(uuid) {
		t.Error("valid UUID rejected")
	}
}

func TestIsUUIDCharsRejectsTooShort(t *testing.T) {
	if isUUIDChars([]byte("too-short")) {
		t.Error("short string accepted as UUID")
	}
}

func TestIsUUIDCharsRejectsBadChars(t *testing.T) {
	uuid := []byte("550e8400-e29b-41d4-a716-44665544000Z")
	if isUUIDChars(uuid) {
		t.Error("invalid UUID accepted")
	}
}

func TestScanForUUIDFindsUUIDInData(t *testing.T) {
	data := []byte("some prefix data 550e8400-e29b-41d4-a716-446655440000 suffix")
	result := scanForUUID(data)
	if result != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("got %q", result)
	}
}

func TestScanForUUIDReturnsEmptyWhenNotFound(t *testing.T) {
	result := scanForUUID([]byte("no uuid here"))
	if result != "" {
		t.Errorf("got %q, want empty", result)
	}
}

func TestLoadTemplateB64ReturnsErrorForMissingFile(t *testing.T) {
	_, err := LoadTemplateB64("testdata/nonexistent.b64")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestEncodeBytesProducesWireType2(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	encoded := encodeBytes(nil, 4, data)
	tag, off, _ := readVarint(encoded, 0)
	fn, wt := int(tag>>3), int(tag&7)
	if fn != 4 || wt != 2 {
		t.Fatalf("fn=%d wt=%d, want fn=4 wt=2", fn, wt)
	}
	length, _, _ := readVarint(encoded, off)
	if int(length) != 3 {
		t.Errorf("length=%d, want 3", length)
	}
}

func TestNumericIDReturnsCorrectLength(t *testing.T) {
	id := numericID(15)
	if len(id) != 15 {
		t.Errorf("len=%d, want 15", len(id))
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			t.Errorf("non-digit char: %c", c)
		}
	}
}

func TestRandIDSuffixProducesDashedFormat(t *testing.T) {
	s := randIDSuffix()
	if len(s) < 10 {
		t.Errorf("suffix too short: %q", s)
	}
	// Should contain dashes (8 groups of hex separated by dashes)
	if !contains(s, "-") {
		t.Errorf("suffix missing dashes: %q", s)
	}
}

func TestBuildFromTemplatePreservesHeaderLengthDelta(t *testing.T) {
	tplB64, _ := LoadTemplateB64("testdata/template_frame.b64")
	tpl, _ := base64.StdEncoding.DecodeString(tplB64)
	origLenField := int(tpl[3]) | int(tpl[4])<<8
	origBodyLen := len(tpl) - 8

	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "x", // different length than template text
		ConversationID: "c",
		RequestID:      "r",
	})
	if err != nil {
		t.Fatal(err)
	}
	newLenField := int(frame[3]) | int(frame[4])<<8
	newBodyLen := len(frame) - 8

	origDelta := origLenField - origBodyLen
	newDelta := newLenField - newBodyLen
	if origDelta != newDelta {
		t.Errorf("length delta changed: orig=%d new=%d", origDelta, newDelta)
	}
}

func TestBuildFromTemplateSubstitutesRequestID(t *testing.T) {
	tplB64, _ := LoadTemplateB64("testdata/template_frame.b64")
	frame, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    tplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "hi",
		ConversationID: "c",
		RequestID:      "my-req-id-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	var outer struct {
		ReqID string `json:"req-id"`
	}
	json.Unmarshal(frame[8:], &outer)
	if outer.ReqID != "my-req-id-123" {
		t.Errorf("req-id = %q", outer.ReqID)
	}
}

func countAttachments(protoBytes []byte) int {
	off := 0
	for off < len(protoBytes) {
		tag, n, ok := readVarint(protoBytes, off)
		if !ok {
			break
		}
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		if wt == 2 {
			length, ln, _ := readVarint(protoBytes, off)
			if off+ln+int(length) > len(protoBytes) {
				break
			}
			payload := protoBytes[off+ln : off+ln+int(length)]
			off += ln + int(length)
			if fn == 2 {
				count := 0
				po := 0
				for po < len(payload) {
					pt, pn, ok2 := readVarint(payload, po)
					if !ok2 {
						break
					}
					po += pn
					pfn, pwt := int(pt>>3), int(pt&7)
					if pfn == 3 {
						count++
					}
					if pwt == 2 {
						vl, pln, ok3 := readVarint(payload, po)
						if !ok3 {
							break
						}
						po += pln + int(vl)
					} else if pwt == 0 {
						_, vn, _ := readVarint(payload, po)
						po += vn
					} else if pwt == 5 {
						po += 4
					} else if pwt == 1 {
						po += 8
					} else {
						break
					}
				}
				return count
			}
		} else if wt == 0 {
			_, vn, _ := readVarint(protoBytes, off)
			off += vn
		} else if wt == 5 {
			off += 4
		} else if wt == 1 {
			off += 8
		} else {
			break
		}
	}
	return 0
}

func TestSingleAttachmentRegression(t *testing.T) {
	// 1. Check text-only template gains exactly one attachment
	txtTplB64 := DefaultTemplateB64()
	if txtTplB64 == "" {
		t.Fatal("empty default template")
	}
	frame1, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    txtTplB64,
		TemplateText:   DefaultTemplateText,
		NewText:        "analyze",
		ConversationID: "11111111-2222-3333-4444-555555555555",
		MediaID:        "98765432101",
		MediaMime:      "image/gif",
		MediaFilename:  "my_custom_image.gif",
	})
	if err != nil {
		t.Fatal(err)
	}
	var outer struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame1[8:], &outer); err != nil {
		t.Fatal(err)
	}
	inner1, _ := base64.StdEncoding.DecodeString(outer.Payload)
	if count := countAttachments(inner1); count != 1 {
		t.Errorf("expected text-only template with attachment to have exactly 1 attachment, got %d", count)
	}

	// 2. Load the real attachment_frame.b64 fixture
	attTplB64 := DefaultAttachmentTemplateB64()
	if attTplB64 == "" {
		t.Fatal("empty default attachment template")
	}
	tplBytes, _ := base64.StdEncoding.DecodeString(attTplB64)
	var attOuter struct {
		Payload string `json:"payload"`
	}
	json.Unmarshal(tplBytes[8:], &attOuter)
	innerTplBytes, _ := base64.StdEncoding.DecodeString(attOuter.Payload)

	// Ensure the source message block has exactly one attachment
	if count := countAttachments(innerTplBytes); count != 1 {
		t.Errorf("expected source attachment template to have exactly 1 attachment, got %d", count)
	}

	// Build a frame with a new media ID from the attachment template
	frame2, err := BuildFromTemplate(TemplateOptions{
		TemplateB64:    attTplB64,
		TemplateText:   "What do you see ?",
		NewText:        "analyze this",
		ConversationID: "11111111-2222-3333-4444-555555555555",
		MediaID:        "98765432101",
		MediaMime:      "image/gif",
		MediaFilename:  "my_custom_image.gif",
	})
	if err != nil {
		t.Fatal(err)
	}

	var outer2 struct {
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(frame2[8:], &outer2); err != nil {
		t.Fatal(err)
	}
	inner2, _ := base64.StdEncoding.DecodeString(outer2.Payload)

	// Resulting message block must still have exactly one attachment
	if count := countAttachments(inner2); count != 1 {
		t.Errorf("expected resulting template to have exactly 1 attachment, got %d", count)
	}

	// Assert new media ID, mime, and filename are present
	inner2Str := string(inner2)
	if !contains(inner2Str, "my_custom_image.gif") {
		t.Error("new filename missing")
	}
	if !contains(inner2Str, "image/gif") {
		t.Error("new mime missing")
	}
	if !contains(inner2Str, "98765432101") {
		t.Error("new media ID missing")
	}

	// Captured filename and media ID must be absent
	if contains(inner2Str, "student_card.png") {
		t.Error("captured student_card.png filename still present")
	}
	if contains(inner2Str, "867051314767696") {
		t.Error("captured media ID 867051314767696 still present")
	}

	// Malformed input does not panic
	_, _ = BuildFromTemplate(TemplateOptions{
		TemplateB64:    "malformed!!",
		TemplateText:   "nonexistent",
		NewText:        "x",
		ConversationID: "c",
	})
}
