// Package clippy implements the Meta AI "clippy" WebSocket binary protocol.
//
// Wire-format ground truth is documented in docs/protocol.md and derived from a
// live capture made on 2026-06-19.
//
// Frame layers:
//  1. FRAME header: [type:1][seq:2 LE][len:2 LE]([flags...])[payload…]
//     - type 0x0f (CONNECT): 6-byte header [type][seq][len][flags=0x00] + JSON
//     - type 0x0d (PROTO_INSIDE_JSON): 8-byte header [type][seq][len][0x00][sub][0x80] + JSON
//  2. OUTER JSON (type 0x0d): {"req-id":"<uuid>","payload":"<base64 inner protobuf>"}
//  3. INNER protobuf: standard varint/LEN wire encoding (see buildChatProto).
package clippy

import "encoding/binary"

// ── Protobuf wire-encoding primitives ──────────────────────────────────────
// These primitives produce the protobuf wire values used by captured frames.

// encodeVarint appends an unsigned varint to dst.
func encodeVarint(dst []byte, v uint64) []byte {
	for v >= 0x80 {
		dst = append(dst, byte(v)|0x80)
		v >>= 7
	}
	return append(dst, byte(v))
}

// encodeTag appends a (field<<3 | wireType) varint tag.
func encodeTag(dst []byte, field, wireType int) []byte {
	return encodeVarint(dst, uint64(field<<3|wireType))
}

// encodeString appends a field (LEN wire type 2) holding a UTF-8 string.
func encodeString(dst []byte, field int, s string) []byte {
	dst = encodeTag(dst, field, 2)
	dst = encodeVarint(dst, uint64(len(s)))
	return append(dst, s...)
}

// encodeBytes appends a field (LEN wire type 2) holding raw bytes.
func encodeBytes(dst []byte, field int, b []byte) []byte {
	dst = encodeTag(dst, field, 2)
	dst = encodeVarint(dst, uint64(len(b)))
	return append(dst, b...)
}

// encodeMessage appends a nested LEN field whose content is the given sub-message bytes.
func encodeMessage(dst []byte, field int, msg []byte) []byte {
	dst = encodeTag(dst, field, 2)
	dst = encodeVarint(dst, uint64(len(msg)))
	return append(dst, msg...)
}

// encodeVarintField appends a field (VARINT wire type 0).
func encodeVarintField(dst []byte, field int, v uint64) []byte {
	dst = encodeTag(dst, field, 0)
	return encodeVarint(dst, v)
}

// ── Frame header packing (little-endian) ───────────────────────────────────

// packConnectHeader builds the 6-byte type-0x0f header.
// Layout: [0x0f][seq LE16][len LE16][0x00]
func packConnectHeader(payloadLen int) []byte {
	h := make([]byte, 6)
	h[0] = 0x0f
	binary.LittleEndian.PutUint16(h[1:3], 0) // seq
	binary.LittleEndian.PutUint16(h[3:5], uint16(payloadLen))
	h[5] = 0x00
	return h
}

// packChatHeader builds the 8-byte type-0x0d header.
// Layout: [0x0d][seq LE16][len LE16][0x00][subType][0x80]
//
// subType is 0x00 for a SEND frame per the live capture.
func packChatHeader(payloadLen int, subType byte) []byte {
	h := make([]byte, 8)
	h[0] = 0x0d
	binary.LittleEndian.PutUint16(h[1:3], 0) // seq
	binary.LittleEndian.PutUint16(h[3:5], uint16(payloadLen))
	h[5] = 0x00
	h[6] = subType
	h[7] = 0x80
	return h
}
