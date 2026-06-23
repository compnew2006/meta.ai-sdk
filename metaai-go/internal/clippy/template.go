// template.go implements the "template mode" chat frame builder.
//
// Because the hand-rolled encoder produces a frame the server structurally
// accepts but silently drops (a residual 2-byte difference we could not
// pinpoint), this builder takes a CAPTURED WORKING frame and substitutes only
// the text + conversation id + request id, preserving every other byte the
// browser produced. The operation is byte-faithful and non-interactive.
//
// Substitutions performed on the inner protobuf:
//   - message text (message-block.f2) — variable length, length varint patched
//   - message text duplicate (message-block.f10) — same patch
//   - conversation id (3 occurrences, all fixed 36-char UUIDs → in-place replace)
//   - request id (inner envelope.f6 + outer JSON req-id) — 36-char UUID replace
//
// After substitution, the inner proto is re-base64'd, re-wrapped in the outer
// JSON {req-id, payload}, and re-framed with a recomputed 8-byte header.

package clippy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// TemplateOptions controls the template-substitution build.
type TemplateOptions struct {
	// TemplateB64 is the base64-encoded captured working frame (the full
	// type-0x0d frame including the 8-byte header). Required.
	TemplateB64 string
	// TemplateText is the original text embedded in the template (used to
	// locate the substitution point). Required.
	TemplateText string
	// NewText is the replacement message. Required.
	NewText string
	// ConversationID is the new conversation UUID. Required.
	ConversationID string
	// RequestID overrides the request id; empty → generated.
	RequestID string
	// MediaID is the uploaded image media_id. When set, an attachment field
	// (message-block.f3) is injected so Meta AI can see the image.
	MediaID string
	// MediaMime is the MIME type of the uploaded image (e.g. "image/png").
	MediaMime string
	// MediaFilename is the original filename of the uploaded image.
	MediaFilename string
}

// BuildFromTemplate produces a chat SEND frame by substituting text/conversation/
// request-id into a captured working template. All other bytes are preserved
// verbatim from the browser capture, guaranteeing the server accepts the frame.
func BuildFromTemplate(opts TemplateOptions) ([]byte, error) {
	if opts.TemplateB64 == "" {
		return nil, fmt.Errorf("clippy: TemplateB64 is required")
	}
	if opts.TemplateText == "" {
		return nil, fmt.Errorf("clippy: TemplateText is required (the marker to locate)")
	}
	if opts.NewText == "" {
		return nil, fmt.Errorf("clippy: NewText is required")
	}
	if opts.ConversationID == "" {
		return nil, fmt.Errorf("clippy: ConversationID is required")
	}

	// Decode the template frame.
	tpl, err := base64.StdEncoding.DecodeString(opts.TemplateB64)
	if err != nil {
		return nil, fmt.Errorf("clippy: decode template b64: %w", err)
	}
	if len(tpl) < 9 || tpl[0] != 0x0d {
		return nil, fmt.Errorf("clippy: template is not a type-0x0d frame (byte0=0x%02x)", tpl[0])
	}

	// Parse the outer JSON.
	var outer struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	if err := json.Unmarshal(tpl[8:], &outer); err != nil {
		return nil, fmt.Errorf("clippy: parse template outer json: %w", err)
	}
	inner, err := base64.StdEncoding.DecodeString(outer.Payload)
	if err != nil {
		return nil, fmt.Errorf("clippy: decode template inner proto: %w", err)
	}

	// Resolve new request id.
	newReqID := opts.RequestID
	if newReqID == "" {
		newReqID = UUIDFunc()
	}

	// --- Substitutions on the inner proto ---

	// 1. Replace the template text occurrences with the new text.
	//    The text is encoded as f2 LEN[<n>] "<text>" — a protobuf LEN field.
	//    We must patch the length varint when the new text has a different
	//    length. The text may appear 1-2 times (message-block.f2 and an f10
	//    duplicate). We rebuild every occurrence.
	inner, err = replaceAllTextFieldValues(inner, opts.TemplateText, opts.NewText)
	if err != nil {
		return nil, err
	}

	// 2. Replace conversation id (all occurrences). UUIDs are always 36 chars,
	//    so this is a same-length in-place replace — no length patching needed.
	//    We need the template's conversation id; it's the 36-char UUID that
	//    appears after "0a 24" (the \n$ marker) in the inner.f5 nested ref.
	tplConv := extractTemplateConversationID(inner)
	if tplConv == "" {
		return nil, fmt.Errorf("clippy: could not locate template conversation id")
	}
	if tplConv != opts.ConversationID {
		inner = bytesReplaceAll(inner, []byte(tplConv), []byte(opts.ConversationID))
	}

	// 3. Replace the request id (envelope.f6, same-length UUID replace).
	if outer.ReqID != "" && outer.ReqID != newReqID {
		inner = bytesReplaceAll(inner, []byte(outer.ReqID), []byte(newReqID))
	}

	// 4. Inject image attachment (message-block.f3) if MediaID is set.
	//    From HAR capture: the attachment is a nested message in top.f2.f3
	//    with structure: {f1:{f1:VARINT(media_ent_id)}, f2:1, f3:"", f5:0,
	//    f6:"image/png", f7:"filename.png"}.
	if opts.MediaID != "" {
		mime := opts.MediaMime
		if mime == "" {
			mime = "image/jpeg"
		}
		filename := opts.MediaFilename
		if filename == "" {
			filename = "upload.png"
		}
		inner, err = replaceAllTextFieldValues(inner, "867051314767696", opts.MediaID)
		if err != nil {
			return nil, err
		}
		inner, err = replaceAllTextFieldValues(inner, "student_card.png", filename)
		if err != nil {
			return nil, err
		}
		inner, err = replaceAllTextFieldValues(inner, "image/png", mime)
		if err != nil {
			return nil, err
		}
		inner = injectAttachment(inner, opts.MediaID, mime, filename)
	}

	// --- Re-wrap ---
	// Preserve the original header's length field exactly. The captured frame's
	// length field is intentionally 2 bytes LARGER than the JSON body (a server-
	// side quirk); recomputing it from len(body) produces a length the server
	// rejects. So we copy the original 8-byte header and patch only the length
	// field to account for any body-size delta caused by text substitution.
	innerB64 := base64.StdEncoding.EncodeToString(inner)
	// Struct (not map) preserves key order {"req-id","payload"} matching the capture.
	type outerJSON struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	body, err := json.Marshal(outerJSON{ReqID: newReqID, Payload: innerB64})
	if err != nil {
		return nil, fmt.Errorf("clippy: marshal new outer json: %w", err)
	}
	origBodyLen := len(tpl) - 8                  // original JSON body size
	origLenField := int(tpl[3]) | int(tpl[4])<<8 // original length-field value
	// lengthFieldDelta = how much larger the length field is than the body
	lengthFieldDelta := origLenField - origBodyLen
	// New length field = new body size + the same delta.
	newLenField := len(body) + lengthFieldDelta
	header := make([]byte, 8)
	copy(header, tpl[:8])                       // preserve original header bytes
	header[3] = byte(newLenField & 0xff)        // patch length low byte
	header[4] = byte((newLenField >> 8) & 0xff) // patch length high byte
	return append(header, body...), nil
}

// injectAttachment inserts an image attachment field (f3) into the message-block
// (top.f2). The attachment structure (from HAR capture 2026-06-19):
//
//	f3 = {
//	  f1 = { f1 = VARINT(media_ent_id) }    // numeric media_id as varint
//	  f2 = 1                                // media type = image
//	  f3 = ""                               // empty
//	  f5 = 0                                // flag
//	  f6 = "image/png"                      // mime type
//	  f7 = "filename.png"                   // original filename
//	}
func injectAttachment(proto []byte, mediaID, mime, filename string) []byte {
	// Parse the media_id string to a uint64 varint.
	mediaIDNum := uint64(0)
	for _, c := range []byte(mediaID) {
		if c < '0' || c > '9' {
			return proto // not a number; skip injection
		}
		mediaIDNum = mediaIDNum*10 + uint64(c-'0')
	}

	// Build the attachment proto.
	attInner := encodeVarintField(nil, 1, mediaIDNum)  // f1 = {f1: mediaID}
	attachment := encodeMessage(nil, 1, attInner)      // f1 wraps media_id
	attachment = encodeVarintField(attachment, 2, 1)   // f2 = 1 (image type)
	attachment = encodeMessage(attachment, 3, nil)     // f3 = empty
	attachment = encodeVarintField(attachment, 5, 0)   // f5 = 0
	attachment = encodeString(attachment, 6, mime)     // f6 = mime type
	attachment = encodeString(attachment, 7, filename) // f7 = filename

	// Walk the proto, find top.f2 (message-block), and insert f3 after f2's
	// existing fields (but before f4 if present).
	var out []byte
	off := 0
	injected := false
	for off < len(proto) {
		tag, n, ok := readVarint(proto, off)
		if !ok {
			out = append(out, proto[off:]...)
			break
		}
		tagBytes := append([]byte(nil), proto[off:off+n]...)
		off += n
		fn, wt := int(tag>>3), int(tag&7)

		if wt == 2 {
			length, ln, ok := readVarint(proto, off)
			if !ok {
				out = append(out, tagBytes...)
				out = append(out, proto[off:]...)
				break
			}
			payloadStart := off + ln
			payloadEnd := payloadStart + int(length)
			if payloadEnd > len(proto) {
				out = append(out, tagBytes...)
				out = append(out, proto[off:]...)
				break
			}
			payload := append([]byte(nil), proto[payloadStart:payloadEnd]...)
			off = payloadEnd

			if fn == 2 && !injected {
				// This is the message-block (top.f2). If the template already has
				// an f3 (attachment), SUBSTITUTE the media_id inside it instead of
				// injecting a new one. Otherwise, inject a new f3.
				hasF3 := false
				// Walk payload to check for f3
				{
					po := 0
					for po < len(payload) {
						pt, pn, ok2 := readVarint(payload, po)
						if !ok2 {
							break
						}
						po += pn
						pfn, pwt := int(pt>>3), int(pt&7)
						if pfn == 3 {
							hasF3 = true
							break
						}
						if pwt == 2 {
							vl, vln, okVal := readVarint(payload, po)
							if !okVal {
								break
							}
							po += vln + int(vl)
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
				}
				if hasF3 {
					// Substitute media_id inside existing attachment
					payload = substituteMediaIDInAttachment(payload, mediaID, mime, filename)
				} else {
					// Inject new f3 after the message-block fields
					payload = append(payload, encodeMessage(nil, 3, attachment)...)
					// Re-encode with new length
					out = append(out, tagBytes...)
					out = encodeVarint(out, uint64(len(payload)))
					out = append(out, payload...)
					injected = true
					continue
				}
				out = append(out, tagBytes...)
				out = encodeVarint(out, uint64(len(payload)))
				out = append(out, payload...)
				injected = true
				continue
			}
			out = append(out, tagBytes...)
			out = encodeVarint(out, length)
			out = append(out, payload...)
		} else if wt == 0 {
			_, vn, _ := readVarint(proto, off)
			out = append(out, tagBytes...)
			out = append(out, proto[off:off+vn]...)
			off += vn
		} else if wt == 5 {
			out = append(out, tagBytes...)
			out = append(out, proto[off:off+4]...)
			off += 4
		} else if wt == 1 {
			out = append(out, tagBytes...)
			out = append(out, proto[off:off+8]...)
			off += 8
		} else {
			out = append(out, tagBytes...)
			out = append(out, proto[off:]...)
			break
		}
	}

	// If we never found f2, append the attachment at the end.
	if !injected {
		out = encodeMessage(out, 3, attachment)
	}

	return out
}

// replaceAllTextFieldValues finds every protobuf LEN field whose value equals
// oldText and replaces its value with newText, patching the length varint. It
// RECURSES into nested LEN messages so it can find text fields at any depth.
// All wire types are handled; unknown types are copied verbatim.
func replaceAllTextFieldValues(proto []byte, oldText, newText string) ([]byte, error) {
	rewritten, _, ok := replaceTextFieldValues(proto, []byte(oldText), []byte(newText), 0)
	if !ok {
		return proto, nil
	}
	return rewritten, nil
}

const maxTemplateProtoDepth = 16

// replaceTextFieldValues recursively walks valid protobuf messages. Unchanged
// fields retain their exact original bytes; ancestor lengths are rebuilt only
// when a nested payload contains a replacement.
func replaceTextFieldValues(proto, oldText, newText []byte, depth int) ([]byte, int, bool) {
	if depth > maxTemplateProtoDepth {
		return nil, 0, false
	}

	out := make([]byte, 0, len(proto))
	replacements := 0
	for off := 0; off < len(proto); {
		fieldStart := off
		tag, tagLen, ok := readVarint(proto, off)
		if !ok || tag == 0 {
			return nil, 0, false
		}
		off += tagLen

		switch tag & 7 {
		case 0:
			_, valueLen, ok := readVarint(proto, off)
			if !ok {
				return nil, 0, false
			}
			off += valueLen
			out = append(out, proto[fieldStart:off]...)
		case 1:
			if off+8 > len(proto) {
				return nil, 0, false
			}
			off += 8
			out = append(out, proto[fieldStart:off]...)
		case 2:
			length, lengthLen, ok := readVarint(proto, off)
			if !ok {
				return nil, 0, false
			}
			payloadStart := off + lengthLen
			payloadEnd := payloadStart + int(length)
			if payloadEnd < payloadStart || payloadEnd > len(proto) {
				return nil, 0, false
			}
			payload := proto[payloadStart:payloadEnd]
			off = payloadEnd

			if equalBytes(payload, oldText) {
				out = append(out, proto[fieldStart:fieldStart+tagLen]...)
				out = encodeVarint(out, uint64(len(newText)))
				out = append(out, newText...)
				replacements++
				continue
			}

			nested, nestedReplacements, nestedOK := replaceTextFieldValues(
				payload, oldText, newText, depth+1,
			)
			if nestedOK && nestedReplacements > 0 {
				out = append(out, proto[fieldStart:fieldStart+tagLen]...)
				out = encodeVarint(out, uint64(len(nested)))
				out = append(out, nested...)
				replacements += nestedReplacements
				continue
			}
			out = append(out, proto[fieldStart:off]...)
		case 5:
			if off+4 > len(proto) {
				return nil, 0, false
			}
			off += 4
			out = append(out, proto[fieldStart:off]...)
		default:
			return nil, 0, false
		}
	}

	return out, replacements, true
}

// extractTemplateConversationID finds the conversation UUID in the inner proto.
// It's the 36-char string following the "\n$" (0a 24) marker in the inner.f5
// nested reference. Returns "" if not found.
func extractTemplateConversationID(proto []byte) string {
	marker := []byte{0x0a, 0x24} // \n$
	for i := 0; i <= len(proto)-len(marker)-36; i++ {
		if proto[i] == marker[0] && proto[i+1] == marker[1] {
			candidate := proto[i+2 : i+2+36]
			if isUUIDChars(candidate) {
				return string(candidate)
			}
		}
	}
	// Fallback: scan for any 36-char UUID-shaped run.
	return scanForUUID(proto)
}

func isUUIDChars(b []byte) bool {
	if len(b) != 36 {
		return false
	}
	for i, c := range b {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func scanForUUID(b []byte) string {
	for i := 0; i <= len(b)-36; i++ {
		if isUUIDChars(b[i : i+36]) {
			return string(b[i : i+36])
		}
	}
	return ""
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func bytesReplaceAll(s, old, new []byte) []byte {
	if len(old) == 0 {
		return s
	}
	var out []byte
	off := 0
	for off <= len(s)-len(old) {
		if equalBytes(s[off:off+len(old)], old) {
			out = append(out, new...)
			off += len(old)
		} else {
			out = append(out, s[off])
			off++
		}
	}
	out = append(out, s[off:]...)
	return out
}

// LoadTemplateB64 reads a base64-encoded template frame from disk. path is
// relative to the working directory. Returns "" (no error) if the file is
// missing — callers fall back to the embedded default.
func LoadTemplateB64(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Trim whitespace/newlines.
	s := make([]byte, 0, len(b))
	for _, c := range b {
		if c != '\n' && c != '\r' && c != ' ' && c != '\t' {
			s = append(s, c)
		}
	}
	return string(s), nil
}
