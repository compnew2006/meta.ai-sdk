package metaai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/smart-studio/metaai-go/internal/clippy"
	"github.com/smart-studio/metaai-go/internal/transport"
)

// ChatOptions controls a single chat turn. Zero values inherit the client's
// defaults (constructor option > env var > topic override).
type ChatOptions struct {
	Thinking        *bool
	Instant         *bool
	Mode            *string
	NewConversation bool
	Topic           string // route to a tracked topic conversation
	// ConversationID, when set, resumes an existing Meta AI conversation by
	// sending this id in the CONNECT/CHAT frames so Meta AI reuses its prior
	// context. Highest priority in resolveConversationID (overrides Topic and
	// the sticky current). Used by the REST UI's multi-conversation feature.
	ConversationID string
	// SystemInstruction, when set, overrides the client-wide
	// Config.SystemInstruction for this single turn. Prepended to the message
	// as "[System]\n...\n\n".
	SystemInstruction string
}

// ChatChunk is one streamed fragment of a chat response.
type ChatChunk struct {
	Text    string
	Sources []map[string]any
	Media   []map[string]any
	Err     error
	Done    bool
}

// Chat sends a message and returns the full response text (non-streaming).
// For streaming, use StreamChat.
func (c *Client) Chat(ctx context.Context, message string, opts *ChatOptions) (string, error) {
	ch := c.StreamChat(ctx, message, opts)
	var full string
	for chunk := range ch {
		if chunk.Err != nil {
			return full, chunk.Err
		}
		full += chunk.Text
	}
	if full == "" {
		return "", errors.New("metaai: chat produced no response (server may have rejected the message format)")
	}
	return full, nil
}

// StreamChat sends a message and returns a channel of streamed chunks. The
// channel closes when the response completes (or ctx is cancelled). The final
// chunk carries Done=true; an error (if any) is delivered on a chunk with
// Err set, after which the channel closes.
//
// Transport: clippy WebSocket (primary). Connect + chat SEND frames are built by
// internal/clippy from the captured spec; streamed RECV frames are parsed with
// clippy.ResponseText.
func (c *Client) StreamChat(ctx context.Context, message string, opts *ChatOptions) <-chan ChatChunk {
	if opts == nil {
		opts = &ChatOptions{}
	}
	out := make(chan ChatChunk, 16)
	go func() {
		defer close(out)
		if err := c.streamChat(ctx, message, opts, out); err != nil {
			select {
			case <-ctx.Done():
			default:
				out <- ChatChunk{Err: err, Done: true}
			}
		}
	}()
	return out
}

// streamChat is the streaming implementation. It runs on the calling goroutine's
// behalf (called from the StreamChat goroutine).
func (c *Client) streamChat(ctx context.Context, message string, opts *ChatOptions, out chan<- ChatChunk) error {
	thinking, instant, mode := c.resolveChatConfig(opts)
	if thinking && instant {
		return ErrConflictingModes
	}
	if mode != "" && !isValidMode(mode) {
		return fmt.Errorf("%w: %s", ErrInvalidMode, mode)
	}
	if err := c.EnsureAccessToken(ctx); err != nil {
		return err
	}
	conversationID := c.resolveConversationID(opts)
	isNew := conversationID == ""
	if isNew {
		conversationID = newConversationID()
	}
	c.externalConversationID = conversationID
	if opts != nil && opts.Topic != "" {
		c.topics[opts.Topic] = conversationID
	}

	c.logger.Debugf("clippy: streamChat initialized for conversationID %q", conversationID)
	// Keep the WS connection alive within the same conversation. Only
	// reconnect if NewConversation is set and the old connection belongs to
	// a different conversation, or the connection died.
	c.wsMu.Lock()
	if c.ws != nil && opts != nil && opts.NewConversation && c.wsConvID != "" && c.wsConvID != conversationID {
		c.logger.Debugf("clippy: closing old WebSocket for conversation %s to start new conversation %s", c.wsConvID, conversationID)
		_ = c.ws.tc.Close()
		c.ws = nil
	}
	if c.ws == nil {
		conn, err := c.dialClippy(ctx)
		if err != nil {
			c.wsMu.Unlock()
			return fmt.Errorf("metaai: connect clippy: %w", err)
		}
		c.ws = conn
		c.wsConvID = conversationID
	}
	conn := c.ws
	c.wsMu.Unlock()

	// Send CONNECT handshake for this conversation (safe to repeat).
	connectFrame, err := clippy.BuildConnectMessage(conversationID)
	if err != nil {
		return err
	}
	c.logger.Debugf("clippy: sending CONNECT frame...")
	if err := conn.tc.SendBinary(connectFrame); err != nil {
		c.onWSFailure()
		return fmt.Errorf("metaai: send connect frame: %w", err)
	}
	c.logger.Debugf("clippy: CONNECT frame sent successfully")
	// 300ms is the minimum observed server-side debounce between CONNECT and
	// the subsequent CHAT frame in live captures; shorter delays cause the
	// server to silently drop the chat frame (it has not yet registered the
	// conversation from the CONNECT ack).
	time.Sleep(300 * time.Millisecond)

	// Template mode: substitute the user's text, conversation id,
	// and request id into a captured working browser frame. The hand-rolled
	// encoder produces a frame the server structurally accepts but silently
	// drops (a ~2-byte residual difference); template substitution preserves
	// every browser byte except the replaced fields, guaranteeing acceptance.
	//
	// System instruction: prepend a "[System]\n...\n\n" block to the message
	// (mirrors proxy/prompt.go). Per-call ChatOptions.SystemInstruction wins
	// over the client-wide Config.SystemInstruction.
	if sys := c.resolveSystemInstruction(opts); sys != "" {
		message = "[System]\n" + sys + "\n\n" + message
	}
	var chatFrame []byte
	hasAttachment := c.attachMedia != nil
	if tplB64 := clippy.DefaultTemplateB64(); tplB64 != "" {
		tplOpts := clippy.TemplateOptions{
			TemplateB64:    tplB64,
			TemplateText:   clippy.DefaultTemplateText,
			NewText:        message,
			ConversationID: conversationID,
		}
		// When an attachment is present, use the attachment template instead
		// of the text-only template. The attachment template was captured from
		// a live image-analysis message and already contains the f3 attachment
		// field — we just substitute the media_id, mime, and filename.
		if c.attachMedia != nil {
			tplOpts.TemplateB64 = clippy.DefaultAttachmentTemplateB64()
			tplOpts.TemplateText = "What do you see ?"
			tplOpts.MediaID = c.attachMedia.MediaID
			tplOpts.MediaMime = c.attachMedia.Mime
			tplOpts.MediaFilename = c.attachMedia.Filename
		}
		chatFrame, err = clippy.BuildFromTemplate(tplOpts)
	} else {
		chatFrame, err = clippy.BuildChatMessage(clippy.ChatMessageOptions{
			Text:           message,
			ConversationID: conversationID,
			Thinking:       thinking,
		})
	}
	if err != nil {
		return err
	}
	c.logger.Debugf("clippy: sending CHAT frame (len=%d, attachment=%v)...", len(chatFrame), hasAttachment)
	if err := conn.tc.SendBinary(chatFrame); err != nil {
		c.onWSFailure()
		return fmt.Errorf("metaai: send chat frame: %w", err)
	}
	c.logger.Debugf("clippy: CHAT frame sent successfully, entering recvLoop")

	return c.recvLoop(ctx, conn, out, conversationID)
}

// recvLoop reads clippy RECV frames, emits text chunks, and stops on
// end-of-stream markers according to the captured streaming contract.
//
// It is intentionally pure orchestration: read a frame, classify it, and
// dispatch. The two classification rules — duplicate-"full" filtering and
// end-of-stream detection — live in their own helpers below.
//
// readTimeout is the per-read deadline (reset before every RecvBinary), i.e.
// the maximum GAP between frames. Vision/image turns (AnalyzeImage) have high
// first-token latency: the server can spend 60s+ running the model before it
// emits any frame. 30s was too short and killed the connection mid-turn,
// surfacing as the misleading "chat produced no response" error even though
// meta.ai's own UI kept the socket open and rendered the answer. 120s leaves
// ample headroom while still recovering from a truly dead socket.
func (c *Client) recvLoop(ctx context.Context, conn *clippyConn, out chan<- ChatChunk, conversationID string) error {
	const readTimeout = 120 * time.Second
	seenFull := false   // whether we have already emitted a type:"full" response
	receivedAny := false
	var lastFullText string
	var patchDedup patchDeduper
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, _, err := conn.tc.RecvBinary(readTimeout)
		if err != nil {
			c.logger.Debugf("clippy: recvLoop RecvBinary error: %v", err)
			c.onWSFailure()
			// If we already received some data, the stream ended (cleanly or
			// otherwise) after the turn progressed — return nil so the caller
			// keeps whatever text was emitted. If NO data ever arrived, the
			// timeout/EOF is a genuine failure: return the error so Chat()
			// surfaces the real cause instead of the generic "no response".
			if receivedAny || errors.Is(err, transport.ErrClosed) {
				return nil
			}
			return fmt.Errorf("metaai: stream ended before any response: %w", err)
		}
		if len(raw) == 0 {
			continue
		}
		receivedAny = true
		c.logger.Debugf("clippy: recvLoop received frame: len=%d type=%d", len(raw), raw[0])

		txt, ok, isFull := clippy.ResponseTextInfo(raw)
		if ok && txt != "" {
			c.logger.Debugf("clippy: recvLoop extracted text (len=%d, isFull=%v): %q", len(txt), isFull, txt)
			if shouldSkipDuplicateFull(raw, &seenFull) {
				// A duplicate type:"full" response is Meta AI's stream-completion
				// signal: the server re-sends the complete answer once the turn is
				// done, and never sends more content afterward. Terminating here
				// (instead of waiting readTimeout for a connect-ack end marker that
				// never comes for chat/analyze turns) lets the channel close and the
				// HTTP/SSE response flush promptly — otherwise callers wait the full
				// readTimeout (120s) after the answer with the connection idle.
				c.logger.Debugf("clippy: recvLoop stream complete (duplicate full response); closing")
				return nil
			}
			// Patch-delta dedup: Meta AI's clippy transport sends every streaming
			// delta TWICE. See patchDedup for details.
			if !isFull && patchDedup.shouldSkip(raw, txt) {
				c.logger.Debugf("clippy: recvLoop skipped duplicate patch delta: %q", txt)
				continue
			}
			chunkText := txt
			if isFull {
				if lastFullText != "" && strings.HasPrefix(txt, lastFullText) {
					chunkText = strings.TrimPrefix(txt, lastFullText)
				}
				lastFullText = txt
			} else {
				lastFullText += txt
			}
			if chunkText != "" {
				select {
				case out <- ChatChunk{Text: chunkText}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			// Print raw frame content prefix to see what it is
			limit := len(raw)
			if limit > 64 {
				limit = 64
			}
			c.logger.Debugf("clippy: recvLoop frame has no text. raw hex prefix: %s", fmt.Sprintf("%x", raw[:limit]))
		}
		if isEndOfStream(raw) {
			c.logger.Debugf("clippy: recvLoop end of stream detected")
			return nil
		}
	}
}

// patchDeduper tracks the last-emitted patch delta so recvLoop can skip the
// duplicate. Meta AI's clippy transport sends every streaming delta twice (two
// consecutive frames carrying identical text, sometimes sharing a "seq"). The
// first frame wins; the second is dropped. Without this the streamed text
// doubles up into a garbled mess ("مسسمساء النور!").
type patchDeduper struct {
	lastSeq  int
	lastText string
	hasSeq   bool
}

// shouldSkip reports whether a patch-delta frame is a duplicate of the previous
// one and should be dropped. It records the frame as the new "last seen" when it
// is NOT a duplicate.
func (d *patchDeduper) shouldSkip(raw []byte, txt string) bool {
	if seq, hasSeq := clippy.ResponseSeq(raw); hasSeq {
		if d.hasSeq && seq == d.lastSeq {
			return true
		}
		d.lastSeq = seq
		d.hasSeq = true
	}
	if txt == d.lastText {
		return true
	}
	d.lastText = txt
	return false
}

// shouldSkipDuplicateFull reports whether a frame is a duplicate type:"full"
// response that must be dropped. The server sends the complete response
// multiple times; only the first is emitted. seenFull is set true the first
// time a full response is observed.
func shouldSkipDuplicateFull(raw []byte, seenFull *bool) bool {
	if !clippy.IsFullResponse(raw) {
		return false
	}
	if *seenFull {
		return true // already emitted one; skip duplicates
	}
	*seenFull = true
	return false
}

// isEndOfStream reports whether a frame signals the end of the chat stream:
// a type 0x0f (connect) frame carrying a non-200 "code" ack.
func isEndOfStream(raw []byte) bool {
	frame, err := clippy.ParseFrame(raw)
	if err != nil {
		return false
	}
	if frame.Type != clippy.TypeConnect || frame.Payload == nil {
		return false
	}
	code, has := frame.Payload["code"]
	if !has {
		return false
	}
	nc, ok := toInt(code)
	return ok && nc != 200
}

// onWSFailure drops the cached connection so the next chat re-dials.
func (c *Client) onWSFailure() {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	if c.ws != nil {
		_ = c.ws.tc.Close()
		c.ws = nil
	}
}

// resolveChatConfig merges per-call > topic > client defaults for thinking/instant/mode.
func (c *Client) resolveChatConfig(opts *ChatOptions) (thinking, instant bool, mode string) {
	thinking = c.cfg.DefaultThinking
	instant = c.cfg.DefaultInstant
	mode = c.cfg.DefaultMode
	if opts != nil && opts.Topic != "" {
		c.topicConfigMu.RLock()
		if ov, ok := c.topicConfig[opts.Topic]; ok {
			if ov.Thinking != nil {
				thinking = *ov.Thinking
			}
			if ov.Instant != nil {
				instant = *ov.Instant
			}
			if ov.Mode != nil {
				mode = *ov.Mode
			}
		}
		c.topicConfigMu.RUnlock()
	}
	if opts != nil {
		if opts.Thinking != nil {
			thinking = *opts.Thinking
		}
		if opts.Instant != nil {
			instant = *opts.Instant
		}
		if opts.Mode != nil {
			mode = *opts.Mode
		}
	}
	return thinking, instant, mode
}

// resolveConversationID picks the conversation id. Priority order:
// explicit ConversationID > topic binding > NewConversation (force new) >
// sticky current externalConversationID.
func (c *Client) resolveConversationID(opts *ChatOptions) string {
	if opts != nil && opts.ConversationID != "" {
		return opts.ConversationID
	}
	if opts != nil && opts.Topic != "" {
		if id := c.topics[opts.Topic]; id != "" {
			return id
		}
	}
	if opts != nil && opts.NewConversation {
		return ""
	}
	return c.externalConversationID
}


// resolveSystemInstruction returns the per-call SystemInstruction when set,
// otherwise the client-wide Config.SystemInstruction. Empty when neither is set.
func (c *Client) resolveSystemInstruction(opts *ChatOptions) string {
	if opts != nil && opts.SystemInstruction != "" {
		return opts.SystemInstruction
	}
	return c.cfg.SystemInstruction
}

// SetSystemInstruction updates the client-wide system instruction at runtime.
func (c *Client) SetSystemInstruction(s string) {
	c.cfg.SystemInstruction = s
}

func isValidMode(m string) bool {
	switch m {
	case ModeLearn, ModeAnalyze, ModeCreateImage, ModeCreateVideo:
		return true
	}
	return false
}
