package clippy

// substituteMediaIDInAttachment walks the message-block payload, finds the f3
// attachment field, and replaces the media_ent_id varint + mime + filename inside it.
func substituteMediaIDInAttachment(msgBlock []byte, mediaID, mime, filename string) []byte {
	mediaIDNum := uint64(0)
	for _, c := range []byte(mediaID) {
		if c < '0' || c > '9' {
			return msgBlock // not a number; skip
		}
		mediaIDNum = mediaIDNum*10 + uint64(c-'0')
	}

	// Build the new attachment proto
	attInner := encodeVarintField(nil, 1, mediaIDNum)
	newAtt := encodeMessage(nil, 1, attInner)
	newAtt = encodeVarintField(newAtt, 2, 1)
	newAtt = encodeMessage(newAtt, 3, nil)
	newAtt = encodeVarintField(newAtt, 5, 0)
	newAtt = encodeString(newAtt, 6, mime)
	newAtt = encodeString(newAtt, 7, filename)

	// Walk msgBlock, replace the f3 field
	var out []byte
	off := 0
	replaced := false
	for off < len(msgBlock) {
		tag, n, ok := readVarint(msgBlock, off)
		if !ok {
			out = append(out, msgBlock[off:]...)
			break
		}
		tagBytes := append([]byte(nil), msgBlock[off:off+n]...)
		off += n
		fn, wt := int(tag>>3), int(tag&7)
		if wt == 2 {
			length, ln, _ := readVarint(msgBlock, off)
			payloadStart := off + ln
			payloadEnd := payloadStart + int(length)
			if payloadEnd > len(msgBlock) {
				out = append(out, tagBytes...)
				out = append(out, msgBlock[off:]...)
				break
			}
			payload := msgBlock[payloadStart:payloadEnd]
			off = payloadEnd

			if fn == 3 && !replaced {
				// Replace this f3 with the new attachment
				out = encodeMessage(out, 3, newAtt)
				replaced = true
				continue
			}
			out = append(out, tagBytes...)
			out = encodeVarint(out, length)
			out = append(out, payload...)
		} else if wt == 0 {
			_, vn, _ := readVarint(msgBlock, off)
			out = append(out, tagBytes...)
			out = append(out, msgBlock[off:off+vn]...)
			off += vn
		} else if wt == 5 {
			out = append(out, tagBytes...)
			out = append(out, msgBlock[off:off+4]...)
			off += 4
		} else if wt == 1 {
			out = append(out, tagBytes...)
			out = append(out, msgBlock[off:off+8]...)
			off += 8
		} else {
			out = append(out, tagBytes...)
			out = append(out, msgBlock[off:]...)
			break
		}
	}
	return out
}
