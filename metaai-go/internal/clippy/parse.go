package clippy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// FrameType is the first byte of a clippy frame.
type FrameType byte

// Frame type discriminators. TypeConnect (0x0f) carries a JSON handshake or
// streaming-patch payload; TypeProto (0x0d) carries a PROTO_INSIDE_JSON message
// (base64-wrapped protobuf inside an outer JSON {req-id,payload} envelope).
const (
	TypeConnect FrameType = 0x0f // JSON handshake / patch
	TypeProto   FrameType = 0x0d // PROTO_INSIDE_JSON
)

// Frame is a parsed clippy WebSocket frame.
type Frame struct {
	Type    FrameType
	Seq     int
	SubType byte // only for TypeProto
	// Payload is the decoded JSON body as a generic map (for TypeConnect and for
	// TypeProto non-0x25 subtypes), or nil when the frame carried raw bytes.
	Payload map[string]any
	// InnerProto is the base64-decoded inner protobuf for TypeProto frames whose
	// outer JSON had a "payload" key (SEND echoes). nil otherwise.
	InnerProto []byte
	// Raw is the original frame bytes.
	Raw []byte
}

// typeProtoPayload extracts the inner protobuf bytes from a TypeProto (0x0d)
// frame: reads the 2-byte LE length at data[3:5], clamps it to the real buffer
// size (the server's length field can be up to 2 bytes larger than the actual
// payload — a known quirk), and returns data[8:8+payloadLen]. Returns
// (nil, false) when data is not a TypeProto frame or is too short. Shared by
// ResponseTextInfo, ResponseSeq, and IsFullResponse to avoid duplicating the
// header-parse + clamp logic (DRY).
func typeProtoPayload(data []byte) ([]byte, bool) {
	if len(data) == 0 || data[0] != byte(TypeProto) || len(data) < 8 {
		return nil, false
	}
	payloadLen := int(data[3]) | int(data[4])<<8
	if 8+payloadLen > len(data) {
		payloadLen = len(data) - 8
		if payloadLen < 0 {
			payloadLen = 0
		}
	}
	return data[8 : 8+payloadLen], true
}

// ParseFrame decodes a received clippy binary frame.
func ParseFrame(data []byte) (*Frame, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("clippy: frame too short (%d bytes)", len(data))
	}
	f := &Frame{Raw: data, Type: FrameType(data[0]), Seq: int(data[1]) | int(data[2])<<8}

	switch f.Type {
	case TypeConnect:
		// The one-byte payload length is stored at offset 3.
		payloadLen := int(data[3])
		if 6+payloadLen > len(data) {
			payloadLen = len(data) - 6
			if payloadLen < 0 {
				payloadLen = 0
			}
		}
		body := data[6 : 6+payloadLen]
		if err := json.Unmarshal(body, &f.Payload); err != nil {
			// Fall back to a raw-string payload for non-JSON acknowledgements.
			f.Payload = map[string]any{"_raw": string(body)}
		}
		return f, nil

	case TypeProto:
		// 2-byte LE length at bytes 3-4. NOTE: the server's length field can be
		// up to 2 bytes LARGER than the actual payload (a quirk); clamp to the
		// real buffer size to avoid out-of-range panics.
		payloadLen := int(data[3]) | int(data[4])<<8
		if 8+payloadLen > len(data) {
			payloadLen = len(data) - 8
			if payloadLen < 0 {
				payloadLen = 0
			}
		}
		if len(data) > 6 {
			f.SubType = data[6]
		}
		if f.SubType == 0x25 {
			// RECV full response: protobuf-wrapped JSON. Field 5 (tag 0x2a) carries JSON.
			proto := data[8 : 8+payloadLen]
			if js, ok := extractField5JSON(proto); ok {
				f.Payload = js
				return f, nil
			}
			return f, nil
		}
		// SEND echo or other: raw JSON at byte 8.
		body := data[8 : 8+payloadLen]
		if err := json.Unmarshal(body, &f.Payload); err != nil {
			f.Payload = map[string]any{"_raw": string(body)}
			return f, nil
		}
		if b64, ok := f.Payload["payload"].(string); ok {
			if inner, err := base64.StdEncoding.DecodeString(b64); err == nil {
				f.InnerProto = inner
			}
		}
		return f, nil
	}

	// Unknown type: expose the trailing bytes after the common header.
	f.Payload = map[string]any{"_raw": string(data[6:])}
	return f, nil
}

// extractField5JSON walks a protobuf blob looking for the first field-5 LEN entry
// (tag 0x2a) and JSON-decodes its content.
func extractField5JSON(proto []byte) (map[string]any, bool) {
	off := 0
	for off < len(proto) {
		tag, n, ok := readVarint(proto, off)
		if !ok {
			return nil, false
		}
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		if wt != 2 {
			return nil, false // parser only knows wire type 2 here
		}
		length, n2, ok := readVarint(proto, off)
		if !ok {
			return nil, false
		}
		off += n2
		if off+int(length) > len(proto) {
			return nil, false
		}
		chunk := proto[off : off+int(length)]
		off += int(length)
		if fn == 5 {
			var js map[string]any
			if err := json.Unmarshal(chunk, &js); err == nil {
				return js, true
			}
			return map[string]any{"_raw": string(chunk)}, true
		}
		// fn == 1: skip (already did via length advance).
	}
	return nil, false
}

// readVarint reads a base-128 varint, returning (value, bytesConsumed, ok).
func readVarint(b []byte, off int) (uint64, int, bool) {
	var v uint64
	var shift uint
	start := off
	for off < len(b) {
		by := b[off]
		off++
		v |= uint64(by&0x7f) << shift
		if by&0x80 == 0 {
			return v, off - start, true
		}
		shift += 7
		if shift >= 64 {
			return 0, 0, false
		}
	}
	return 0, 0, false
}

// ResponseText extracts streaming text from a RECV frame.
//
// The clippy RECV frames embed a JSON payload ({"seq":N,"type":"full"|"patch",
// "response":{...}}) at a variable offset inside the protobuf-wrapped frame.
// Rather than relying on ParseFrame's subtype dispatch (which only handles
// 0x25), we scan the raw bytes for the JSON payload and parse it directly.
//
// For {"type":"full","response":{"sections":[...]}} → joins thought_text/text.
// For {"type":"patch","operations":[{"op":"delta","value":…}]} → delta value.
// For {"code":N} connect acks → no text.
//
// Returns ("", false) when no text is present in the frame.
// IsFullResponse returns true if the frame's actionable payload represents a
// complete ("full") response — i.e. ResponseTextInfo classifies it with
// isFull=true. Used by recvLoop to suppress/terminate on duplicate full
// responses.
//
// This MUST agree exactly with ResponseTextInfo's isFull classification, so it
// delegates to ResponseTextInfo rather than doing its own extraction. An earlier
// implementation used strings.Contains(json, `"type":"full"`) which was a false
// positive on patch frames whose protobuf also embedded a thinking-status
// "full" section — truncating the stream mid-answer. Another tried field-5
// alone but missed fulls detected only via the raw-protobuf fallback path.
func IsFullResponse(data []byte) bool {
	_, _, isFull := ResponseTextInfo(data)
	return isFull
}

// ResponseSeq extracts the "seq" field from the frame's actionable JSON payload
// (the field-5 JSON for TypeProto frames). Returns (-1, false) when no seq is
// present. recvLoop uses it to dedup duplicate patch deltas: Meta AI's clippy
// transport sends every streaming delta twice (two frames with the same seq but
// different wire sizes), and without dedup the streamed text doubles up into a
// garbled mess ("مسسمساء النور!").
func ResponseSeq(data []byte) (int, bool) {
	proto, ok := typeProtoPayload(data)
	if !ok {
		return -1, false
	}
	f5, ok := readField(proto, 5)
	if !ok {
		return -1, false
	}
	var payload map[string]any
	if err := json.Unmarshal(f5, &payload); err != nil {
		return -1, false
	}
	if seq, ok := payload["seq"].(float64); ok {
		return int(seq), true
	}
	return -1, false
}

func ResponseText(data []byte) (string, bool) {
	txt, ok, _ := ResponseTextInfo(data)
	return txt, ok
}

// ResponseTextInfo extracts streaming text and returns its type (delta vs full).
func ResponseTextInfo(data []byte) (txt string, ok bool, isFull bool) {
	if len(data) == 0 {
		return "", false, false
	}
	// Fast path: type 0x0f connect acks carry {"code":N} — no text.
	if data[0] == byte(TypeConnect) {
		return "", false, false
	}

	// 1. Structured Protobuf Parsing (for TypeProto)
	if proto, ok := typeProtoPayload(data); ok {
		// Try Field 5 (tag 5, wire type 2 = 0x2a) JSON first
		if f5Bytes, okF5 := readField(proto, 5); okF5 {
			var payload map[string]any
			if err := json.Unmarshal(f5Bytes, &payload); err == nil {
				if t, _ := payload["type"].(string); t == "full" {
					if resp, ok := payload["response"].(map[string]any); ok {
						if txt := joinSectionText(resp["sections"]); txt != "" {
							return txt, true, true
						}
					}
					if tt, ok := payload["thought_text"].(string); ok && tt != "" {
						return tt, true, true
					}
					return "", false, false
				}
				if t, _ := payload["type"].(string); t == "patch" {
					if v := firstDelta(payload["operations"]); v != "" {
						return v, true, false
					}
				}
			}
		}

		// Try raw protobuf text fallback (Field 1 -> Field 4 -> Field 1 or 2)
		if txt, ok, isFull := extractProtoDeltaText(proto); ok && txt != "" {
			return txt, true, isFull
		}
	}

	// 2. Fuzzy JSON Scan Fallback (primarily for backward compatibility in tests)
	jsonStr, okJson := extractEmbeddedJSON(data)
	if okJson {
		var payload map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
			return "", false, false
		}
		if t, _ := payload["type"].(string); t == "full" {
			if resp, ok := payload["response"].(map[string]any); ok {
				if txt := joinSectionText(resp["sections"]); txt != "" {
					return txt, true, true
				}
			}
			if tt, ok := payload["thought_text"].(string); ok && tt != "" {
				return tt, true, true
			}
			return "", false, false
		}
		if t, _ := payload["type"].(string); t == "patch" {
			if v := firstDelta(payload["operations"]); v != "" {
				return v, true, false
			}
		}
	}

	return "", false, false
}

// extractProtoDeltaText decodes standard delta text fields directly from the
// raw protobuf message bytes in Field 1 -> Field 4 -> Field 1 (delta) or Field 2 (full).
func extractProtoDeltaText(proto []byte) (string, bool, bool) {
	f1, ok := readField(proto, 1)
	if !ok {
		return "", false, false
	}
	f4, ok := readField(f1, 4)
	if !ok {
		return "", false, false
	}
	delta, ok := readFieldString(f4, 1)
	if ok && delta != "" {
		return delta, true, false
	}
	full, ok := readFieldString(f4, 2)
	if ok && full != "" {
		return full, true, true
	}
	return "", false, false
}

// readField returns the payload of the first length-delimited field matching fieldNum.
func readField(proto []byte, fieldNum int) ([]byte, bool) {
	off := 0
	for off < len(proto) {
		tag, n, ok := readVarint(proto, off)
		if !ok || tag == 0 {
			return nil, false
		}
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		if wt != 2 {
			if !skipField(proto, &off, wt) {
				return nil, false
			}
			continue
		}
		length, ln, ok := readVarint(proto, off)
		if !ok {
			return nil, false
		}
		off += ln
		payloadEnd := off + int(length)
		if payloadEnd > len(proto) {
			payloadEnd = len(proto)
		}
		payload := proto[off:payloadEnd]
		off = payloadEnd
		if fn == fieldNum {
			return payload, true
		}
	}
	return nil, false
}

// readFieldString parses a length-delimited string field.
func readFieldString(proto []byte, fieldNum int) (string, bool) {
	payload, ok := readField(proto, fieldNum)
	if !ok {
		return "", false
	}
	return string(payload), true
}

// skipField advances off past non-length-delimited fields.
func skipField(proto []byte, off *int, wt int) bool {
	switch wt {
	case 0:
		_, n, ok := readVarint(proto, *off)
		if !ok {
			return false
		}
		*off += n
		return true
	case 1:
		if *off+8 > len(proto) {
			return false
		}
		*off += 8
		return true
	case 5:
		if *off+4 > len(proto) {
			return false
		}
		*off += 4
		return true
	default:
		return false
	}
}

// extractEmbeddedJSON scans raw frame bytes for a JSON object starting with
// `{"` and ending with the matching `}`. Returns the JSON substring.
func extractEmbeddedJSON(data []byte) (string, bool) {
	// Find the first `{"` occurrence (the start of the embedded JSON).
	start := -1
	for i := 0; i <= len(data)-2; i++ {
		if data[i] == '{' && data[i+1] == '"' {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false
	}
	// Find the matching closing brace (track depth, respect string literals).
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(data); i++ {
		c := data[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return string(data[start : i+1]), true
			}
		}
	}
	return "", false
}

// firstDelta scans a JSON-Patch operations array for the first {"op":"delta"} value.
func firstDelta(ops any) string {
	arr, ok := ops.([]any)
	if !ok {
		return ""
	}
	for _, op := range arr {
		m, ok := op.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := m["op"].(string); t == "delta" {
			if v, ok := m["value"].(string); ok {
				return v
			}
		}
	}
	return ""
}

// joinSectionText walks response.sections[].view_model.primitive and joins
// text values. SKIP thought_text — those are "thinking" status messages
// like "Confirming user's name" that should NOT be shown to the user.
func joinSectionText(sections any) string {
	arr, ok := sections.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, sec := range arr {
		sm, ok := sec.(map[string]any)
		if !ok {
			continue
		}
		vm, _ := sm["view_model"].(map[string]any)
		if vm == nil {
			continue
		}
		prim, _ := vm["primitive"].(map[string]any)
		if prim == nil {
			continue
		}
		// The main response text is in "text".
		// SKIP "thought_text" — those are thinking step labels, not the actual answer.
		if t, ok := prim["text"].(string); ok && t != "" {
			parts = append(parts, t)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += "\n" + p
	}
	return out
}
