// Package uuid provides the small set of UUID/random-id helpers shared across
// the metaai-go internal packages (transport, upload) and the root package.
//
// All generators panic on crypto/rand failure: a failing CSPRNG means the
// system cannot produce the unpredictable identifiers the Meta AI protocol
// relies on (WebSocket offsets, rupload session ids, conversation ids). A
// silent zero/deterministic fallback would corrupt server-side tracking
// invisibly, so we surface the failure loudly instead.
package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// V4 returns a fresh RFC 4122 version-4 UUID string.
func V4() string {
	var b [16]byte
	mustRead(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// NumericString returns a uniform random n-digit decimal string (zero-padded).
// Used for clippy message ids.
func NumericString(digits int) string {
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		panic(fmt.Sprintf("uuid: crypto/rand failed: %v", err))
	}
	s := n.String()
	for len(s) < digits {
		s = "0" + s
	}
	if len(s) > digits {
		s = s[:digits]
	}
	return s
}

// OfflineThreadingID produces the threading identifier used by the web client:
//
//	threading_id = ((timestamp_ms << 22) | (random_64bit & (2^22-1))) & (2^64-1)
//
// Returned as a decimal string.
func OfflineThreadingID() string {
	ts := uint64(time.Now().UnixMilli())
	const mask22 = (1 << 22) - 1
	const max64 = (1 << 64) - 1
	var rnd [8]byte
	mustRead(rnd[:])
	r := binary.LittleEndian.Uint64(rnd[:])
	id := ((ts << 22) | (r & mask22)) & max64
	return strconv.FormatUint(id, 10)
}

// mustRead fills b from the CSPRNG, panicking on failure.
func mustRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("uuid: crypto/rand failed: %v", err))
	}
}
