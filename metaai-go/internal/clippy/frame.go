package clippy

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// UUIDFunc returns a fresh UUIDv4 string. Exported so tests can pin it.
var UUIDFunc = randUUID4

// ChatMessageOptions controls the chat SEND frame contents. Zero-value fields
// are replaced with sensible defaults at build time (see DefaultChatMessageOptions).
type ChatMessageOptions struct {
	Text           string   // the user's message (required)
	ConversationID string   // conversation UUID (required)
	Thinking       bool     // thinking/reasoning mode
	EntryPoint     string   // Kadabra entry point
	AgentType      string   // agent type string (inner.f7)
	Platform       string   // platform string (envelope.f16)
	Timezone       string   // IANA tz (top.f15.f4)
	UserAgent      string   // UA (envelope.f15)
	Locale         string   // locale (top.f9.f2)
	AppID          string   // x-dgw-appid (inner.f2)
	Capabilities   []string // capability names (top.f18 repeated)
	CapabilityHash string   // envelope.f19.f1 hash

	// Pre-generated IDs. If any are empty, BuildChatMessage generates them via
	// UUIDFunc / numericID. Tests pin these to assert byte-for-byte output.
	RequestID          string
	UserMessageID      string
	AssistantMessageID string
	MessageIDSuffix    string // inner.f4 (e.g. "5a5b-8d4e-…")
	PromptSessionID    string // top.f10.f1
	SessionID          string // top.f10.f3 (separate uuid in capture)
	TurnID             string // message-block.f4.f1 depth as string ("0","1",…)

	// TimestampMs pins the frame timestamp (tests). Zero → time.Now().UnixMilli().
	TimestampMs int64
	// TurnCounter pins the envelope.f2 counter (tests). Zero → derived from RequestID.
	TurnCounter uint64
	// MessageNonce pins the message-id-wrapper.f3 nonce (tests). Zero → derived.
	MessageNonce uint64
}

// DefaultChatMessageOptions returns o with zero-value fields filled in to match
// the live-captured browser defaults.
func DefaultChatMessageOptions(o ChatMessageOptions) ChatMessageOptions {
	if o.EntryPoint == "" {
		o.EntryPoint = "KADABRA__HOME__UNIFIED_INPUT_BAR"
	}
	if o.AgentType == "" {
		o.AgentType = "HUMAN_AGENT"
	}
	if o.Platform == "" {
		o.Platform = "desktop_web"
	}
	if o.Timezone == "" {
		o.Timezone = "Africa/Cairo"
	}
	if o.UserAgent == "" {
		o.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"
	}
	if o.Locale == "" {
		o.Locale = "en-US"
	}
	if o.AppID == "" {
		o.AppID = "1522763855472543"
	}
	if o.Capabilities == nil {
		o.Capabilities = []string{
			"stocks", "weather", "meta_knowledge_search_carousel",
			"meta_catalog_search_carousel", "media_gallery",
		}
	}
	if o.CapabilityHash == "" {
		o.CapabilityHash = "e2b88f9846379cbc26960fa3ae1d22201dfb19df7890ae6a3ac8a28870bac682"
	}
	if o.TurnID == "" {
		o.TurnID = "0"
	}
	return o
}

// BuildConnectMessage builds the captured type-0x0f CONNECT/handshake frame.
func BuildConnectMessage(conversationID string) ([]byte, error) {
	return BuildConnectMessageWithRequestID(conversationID, "")
}

// BuildConnectMessageWithRequestID is like BuildConnectMessage but allows pinning
// the request id (tests).
func BuildConnectMessageWithRequestID(conversationID, requestID string) ([]byte, error) {
	payload := map[string]string{
		"x-dgw-app-x-ecto-conversation-id": conversationID,
		"x-dgw-app-client-payload-type":    "PROTO_INSIDE_JSON",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("clippy: marshal connect json: %w", err)
	}
	header := packConnectHeader(len(body))
	return append(header, body...), nil
}

// BuildChatMessage builds the type-0x0d chat SEND frame per the captured schema.
func BuildChatMessage(opts ChatMessageOptions) ([]byte, error) {
	o := DefaultChatMessageOptions(opts)
	if o.Text == "" {
		return nil, fmt.Errorf("clippy: Text is required")
	}
	if o.ConversationID == "" {
		return nil, fmt.Errorf("clippy: ConversationID is required")
	}
	if o.RequestID == "" {
		o.RequestID = UUIDFunc()
	}
	if o.UserMessageID == "" {
		o.UserMessageID = numericID(15)
	}
	if o.AssistantMessageID == "" {
		o.AssistantMessageID = numericID(15)
	}
	if o.MessageIDSuffix == "" {
		o.MessageIDSuffix = randIDSuffix()
	}
	if o.PromptSessionID == "" {
		o.PromptSessionID = UUIDFunc()
	}
	if o.SessionID == "" {
		o.SessionID = UUIDFunc()
	}

	proto := buildChatProto(o)
	innerB64 := base64.StdEncoding.EncodeToString(proto)

	// Outer JSON field ORDER matters: the live capture is
	// {"req-id":"...","payload":"..."}. Use a struct (not a map, which Go sorts
	// alphabetically → {"payload":...,"req-id":...}) so the bytes match the browser.
	type outerJSON struct {
		ReqID   string `json:"req-id"`
		Payload string `json:"payload"`
	}
	body, err := json.Marshal(outerJSON{ReqID: o.RequestID, Payload: innerB64})
	if err != nil {
		return nil, fmt.Errorf("clippy: marshal outer json: %w", err)
	}
	// sub_type 0x00 per the live capture.
	header := packChatHeader(len(body), 0x00)
	return append(header, body...), nil
}

// buildChatProto assembles the inner protobuf per docs/protocol.md, reconstructed
// byte-accurately from the live-captured chat SEND frame.
//
// The inner proto has exactly TWO top-level fields:
//
//	f1 ENVELOPE (841-ish bytes) — see below
//	f2 MESSAGE-BLOCK — {f1: msg-id-wrapper, f2: "<text>", f4: {f1: "<turn>"}}
//
// ENVELOPE (top.f1) direct fields:
//
//	f1  INNER (467-ish bytes) — entry point + app + ids + ECTO1 + Abra + mode + UA + …
//	f2  {f1,f2: turn counters, f3: 2}
//	f3  {f4: VARINT=1}
//	f4  empty
//	f5  {f1: submitMs, f3: subscribeMs}
//	f6  "<requestID>"
//	f7  {f12:1, f13:1}
//	f9  {f2: "<locale>"}
//	f10 SESSION-REF {f1: promptSessionID, f3: sessionID, f4: conversationID}
//	f15 {f4: "<timezone>"}
//	f16 {f22: VARINT=1}
//	f18 REPEATED {f1: "<cap>", f2: {f1:1}}   // one per capability
//	f20 {0x03}                                // opaque marker (captured verbatim)
//	f26 VARINT=0
//
// INNER (envelope.f1) direct fields:
//
//	f1  "<entry_point>"
//	f2  "<appid>"
//	f4  "<message-id-suffix>"
//	f5  {f1: "<conversation_id>"}   (double-nested; live capture)
//	f6  VARINT=5
//	f7  "<agent_type>"
//	f8  {f1: userMessageID, f2: assistantMessageID}
//	f10 "ECTO1"
//	f11 "Abra Web Main Key"
//	f12 mode config (thinking-dependent)
//	f13 "Mac OS X"
//	f14 "user_input"
//	f15 "<user agent>"
//	f16 "<platform>"
//	f19 {f1: "<capability hash>"}
func buildChatProto(o ChatMessageOptions) []byte {
	now := o.now()

	// INNER (envelope.f1)
	var inner []byte
	inner = encodeString(inner, 1, o.EntryPoint)
	inner = encodeString(inner, 2, o.AppID)
	inner = encodeString(inner, 4, o.MessageIDSuffix)
	// inner.f5 = {f5: {f1: "<conversationID>"}} — DOUBLE-nested per live capture
	// (the inner f5 wraps a nested message whose f5 wraps the conv-id string).
	inner = encodeMessage(inner, 5, encodeMessage(nil, 5, encodeString(nil, 1, o.ConversationID)))
	inner = encodeVarintField(inner, 6, 5)
	inner = encodeString(inner, 7, o.AgentType)
	var ids []byte
	ids = encodeString(ids, 1, o.UserMessageID)
	ids = encodeString(ids, 2, o.AssistantMessageID)
	inner = encodeMessage(inner, 8, ids)
	inner = encodeString(inner, 10, "ECTO1")
	inner = encodeString(inner, 11, "Abra Web Main Key")
	inner = encodeMessage(inner, 12, modeConfig(o.Thinking))
	inner = encodeString(inner, 13, "Mac OS X")
	inner = encodeString(inner, 14, "user_input")
	inner = encodeString(inner, 15, o.UserAgent)
	inner = encodeString(inner, 16, o.Platform)
	inner = encodeMessage(inner, 19, capabilityHashMsg(o.CapabilityHash))

	// ENVELOPE (top.f1)
	var envelope []byte
	envelope = encodeMessage(envelope, 1, inner)
	// envelope.f2 = {f1,f2: turn counters, f3: 2}
	var envF2 []byte
	envF2 = encodeVarintField(envF2, 1, o.turnCounter())
	envF2 = encodeVarintField(envF2, 2, o.turnCounter())
	envF2 = encodeVarintField(envF2, 3, 2)
	envelope = encodeMessage(envelope, 2, envF2)
	envelope = encodeMessage(envelope, 3, encodeVarintField(nil, 4, 1)) // f3 = {f4:1}
	envelope = encodeMessage(envelope, 4, nil)                          // f4 = empty
	envelope = encodeMessage(envelope, 5, timestampPair(now))           // f5 = {f1,f3: ms pair}
	envelope = encodeString(envelope, 6, o.RequestID)                   // f6 = requestID
	var envF7 []byte
	envF7 = encodeVarintField(envF7, 12, 1)
	envF7 = encodeVarintField(envF7, 13, 1)
	envelope = encodeMessage(envelope, 7, envF7)                          // f7
	envelope = encodeMessage(envelope, 9, encodeString(nil, 2, o.Locale)) // f9 = {f2: locale}
	var sessRef []byte
	sessRef = encodeString(sessRef, 1, o.PromptSessionID)
	sessRef = encodeString(sessRef, 3, o.SessionID)
	sessRef = encodeString(sessRef, 4, o.ConversationID)
	envelope = encodeMessage(envelope, 10, sessRef)                          // f10 = session-ref
	envelope = encodeMessage(envelope, 15, encodeString(nil, 4, o.Timezone)) // f15 = {f4: tz}
	envelope = encodeMessage(envelope, 16, encodeVarintField(nil, 22, 1))    // f16 = {f22:1}
	for _, cap := range o.Capabilities {                                     // f18 repeated
		var entry []byte
		entry = encodeString(entry, 1, cap)
		entry = encodeMessage(entry, 2, encodeVarintField(nil, 1, 1))
		envelope = encodeMessage(envelope, 18, entry)
	}
	envelope = encodeMessage(envelope, 20, []byte{0x03}) // f20 = opaque marker
	envelope = encodeVarintField(envelope, 26, 0)        // f26 = 0

	// MESSAGE-BLOCK (top.f2)
	var msgBlock []byte
	msgBlock = encodeMessage(msgBlock, 1, messageIDWrapper(o, now))
	msgBlock = encodeString(msgBlock, 2, o.Text)
	msgBlock = encodeMessage(msgBlock, 4, encodeString(nil, 1, o.TurnID))

	// TOP LEVEL: exactly two fields.
	var top []byte
	top = encodeMessage(top, 1, envelope)
	top = encodeMessage(top, 2, msgBlock)
	return top
}

// messageIDWrapper builds message-block.f1:
//
//	{f1: "<messageUUID>", f2: {f1: conversationID, f2: submitMs, f3: streamStartMs, f5: text, f6: …}, f5: VARINT=1}
//
// Matches the captured layout. The wrapper.f2 nested message carries the
// conversation id plus timing varints; only f1 (conversation id) is semantically
// required, the rest mirror the capture.
func messageIDWrapper(o ChatMessageOptions, now int64) []byte {
	var w []byte
	w = encodeString(w, 1, o.RequestID) // message uuid reuses request id
	var f2 []byte
	f2 = encodeString(f2, 1, o.ConversationID)
	f2 = encodeVarintField(f2, 2, uint64(now)) // submit timestamp ms
	// f3 = a large ~19-digit numeric id (same varint width as the capture).
	// Distinct from the envelope turn-counter; opaque to the server.
	f2 = encodeVarintField(f2, 3, o.messageNonce())
	w = encodeMessage(w, 2, f2)
	w = encodeVarintField(w, 5, 1)
	return w
}

// messageNonce returns a stable ~19-digit numeric id for the message-id-wrapper
// f3 field. Deterministic per request id (for tests); opaque to the server.
func (o ChatMessageOptions) messageNonce() uint64 {
	if o.MessageNonce != 0 {
		return o.MessageNonce
	}
	// Full-width FNV-1a → ~19-digit value, same varint width as the capture.
	var h uint64 = 14695981039346656037
	for i := 0; i < len(o.RequestID); i++ {
		h ^= uint64(o.RequestID[i])
		h *= 1099511628211
	}
	return h
}

// modeConfig builds envelope.f12. Matches capture:
//
//	thinking=true  → {f3: {f1:1001, f2:"mode_thinking"}, f4: {f1:1}}
//	thinking=false → {f1:1001, f2:"MODE_FAST"}
func modeConfig(thinking bool) []byte {
	if thinking {
		var inner []byte
		inner = encodeVarintField(inner, 1, 1001)
		inner = encodeString(inner, 2, "mode_thinking")
		var m []byte
		m = encodeMessage(m, 3, inner)
		m = encodeMessage(m, 4, encodeVarintField(nil, 1, 1))
		return m
	}
	var m []byte
	m = encodeVarintField(m, 1, 1001)
	m = encodeString(m, 2, "MODE_FAST")
	return m
}

// capabilityHashMsg builds inner.f19 = {f1: "<sha256 hash>", f2: float32 1.0}.
// Matches capture: f2 is a wire-type-5 (I32) holding the little-endian bytes of
// the float32 value 1.0 (0x3f800000 → bytes 00 00 80 3f).
func capabilityHashMsg(hash string) []byte {
	var m []byte
	m = encodeString(m, 1, hash)
	m = appendFixed32(m, 2, 0x3f800000) // f2 = float32 1.0
	return m
}

// appendFixed32 appends a wire-type-5 (32-bit fixed) field. v is the raw
// little-endian uint32 value.
func appendFixed32(dst []byte, field int, v uint32) []byte {
	dst = encodeTag(dst, field, 5)
	return append(dst, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// timestampPair builds top.f5 = {f1: submitMs, f3: subscribeMs}.
// now is the submit timestamp (ms); subscribe is ~30ms earlier in captures.
func timestampPair(now int64) []byte {
	var t []byte
	t = encodeVarintField(t, 1, uint64(now))
	t = encodeVarintField(t, 3, uint64(now-30))
	return t
}

// now returns the timestamp (ms) the frame is built with, or a fixed value when
// the caller pinned TimestampMs (tests).
func (o ChatMessageOptions) now() int64 {
	if o.TimestampMs != 0 {
		return o.TimestampMs
	}
	return time.Now().UnixMilli()
}

// turnCounter returns a stable per-turn counter value for envelope.f2. The
// capture uses a 15-digit numeric id (e.g. 588298564377564) — same shape as the
// user/assistant message ids. We force the result into the [10^14, 10^15) range
// so it always encodes to the same varint width (8 bytes) as the capture.
// Deterministic per request id (for tests); opaque to the server.
func (o ChatMessageOptions) turnCounter() uint64 {
	if o.TurnCounter != 0 {
		return o.TurnCounter
	}
	var h uint64 = 14695981039346656037
	for i := 0; i < len(o.RequestID); i++ {
		h ^= uint64(o.RequestID[i])
		h *= 1099511628211
	}
	// Force into [1e14, 1e15) → always 15 digits → always 8-byte varint.
	return 100000000000000 + (h % 900000000000000)
}

// ── ID generation ──────────────────────────────────────────────────────────

// randUUID4 returns a random UUIDv4 string (crypto/rand).
func randUUID4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read on Linux/macOS only errors if /dev/urandom is unavailable;
		// fall back to a deterministic-ish value rather than panic.
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// numericID returns a random n-digit decimal string (for user/assistant message ids).
func numericID(digits int) string {
	const max = 15
	if digits > max {
		digits = max
	}
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(max)), nil)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "000000000000000"
	}
	s := n.String()
	if len(s) > digits {
		s = s[:digits]
	}
	for len(s) < digits {
		s = "0" + s
	}
	return s
}

// randIDSuffix returns a random dash-separated suffix for inner.f4
// (shape like "5a5b-8d4e-f054-99ef-b2de-db02-0d05-52c7"). The exact content is
// opaque to the server; this matches the captured shape.
func randIDSuffix() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%x-%x-%x-%x-%x-%x-%x-%x",
		b[0:2], b[2:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:14], b[14:16])
}
