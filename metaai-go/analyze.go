package metaai

// AnalyzeImage uploads an image and sends a chat message with the image
// attached via the clippy WebSocket protocol.

import (
	"context"
	"fmt"
	"path/filepath"
)

// AnalyzeImage uploads an image (if imagePath is given) and sends a chat message
// with the image attached via the WebSocket clippy protocol. The attachment is
// injected into the message-block f3 field, so Meta AI can actually SEE the
// image content and analyze it. It blocks until the full response is received;
// for token-by-token streaming use AnalyzeImageStream.
func (c *Client) AnalyzeImage(ctx context.Context, imagePath, mediaID, question string, opts *ChatOptions) (string, error) {
	question = defaultQuestion(question)

	o := opts
	if o == nil {
		o = &ChatOptions{}
	}

	// Multi-turn: a ConversationID means follow-up question (text-only, resume);
	// otherwise this is the first turn (upload + attach image, new conversation).
	if o.ConversationID == "" {
		mid, filename, mime, err := c.resolveAttachInfo(ctx, imagePath, mediaID)
		if err != nil {
			return "", err
		}
		// Store attachment info so streamChat can pass it to BuildFromTemplate.
		// Cleared by the deferred nil after Chat() has fully drained the stream
		// (Chat blocks until the StreamChat goroutine finishes, so the read of
		// attachMedia in streamChat always happens before this clear).
		c.attachMedia = &attachInfo{
			MediaID:  mid,
			Mime:     mime,
			Filename: filename,
		}
		defer func() { c.attachMedia = nil }()
		o.NewConversation = true
	}
	return c.Chat(ctx, question, o)
}

// AnalyzeImageStream is the streaming variant of AnalyzeImage. It uploads the
// image (if imagePath is given), attaches it, and returns a channel of
// ChatChunks as Meta AI streams the analysis token-by-token — so callers can
// render the answer incrementally instead of waiting for the whole response.
//
// attachMedia lifecycle (concurrency): StreamChat spawns a goroutine that reads
// c.attachMedia asynchronously, AFTER AnalyzeImageStream returns. To keep that
// read non-nil for the whole stream, cleanup is deferred INSIDE the forwarding
// goroutine here (tied to stream completion, not to this function's return).
// The chatMu held by the REST/proxy server serializes turns, so there is never
// more than one live stream racing on attachMedia.
func (c *Client) AnalyzeImageStream(ctx context.Context, imagePath, mediaID, question string, opts *ChatOptions) <-chan ChatChunk {
	out := make(chan ChatChunk, 16)
	question = defaultQuestion(question)

	o := opts
	if o == nil {
		o = &ChatOptions{}
	}

	// Multi-turn: if a ConversationID is provided, this is a follow-up question
	// about the SAME image — send a text-only frame and resume the conversation
	// (Meta AI remembers the image from turn 1 via the conversation id). No
	// attachment, no forced new conversation.
	// First turn (no ConversationID): upload/attach the image and start a new
	// conversation.
	isFollowUp := o.ConversationID != ""
	if !isFollowUp {
		mid, filename, mime, err := c.resolveAttachInfo(ctx, imagePath, mediaID)
		if err != nil {
			go func() {
				defer close(out)
				out <- ChatChunk{Err: err, Done: true}
			}()
			return out
		}
		c.attachMedia = &attachInfo{
			MediaID:  mid,
			Mime:     mime,
			Filename: filename,
		}
		o.NewConversation = true
	}

	go func() {
		defer close(out)
		// Clear attachMedia only after the stream has finished consuming it
		// (only relevant on the first turn; a no-op clear on follow-ups).
		defer func() { c.attachMedia = nil }()
		for chunk := range c.StreamChat(ctx, question, o) {
			out <- chunk
		}
	}()
	return out
}

// resolveAttachInfo resolves the media id, filename, and MIME type for an
// attachment. If mediaID is already provided it is used as-is; otherwise the
// image at imagePath is uploaded. Shared by AnalyzeImage and AnalyzeImageStream
// to keep their media handling identical (DRY).
func (c *Client) resolveAttachInfo(ctx context.Context, imagePath, mediaID string) (mid, filename, mime string, err error) {
	mid = mediaID
	if mid == "" && imagePath != "" {
		res, err := c.UploadImage(ctx, imagePath)
		if err != nil {
			return "", "", "", err
		}
		mid = res.MediaID
		filename = filepath.Base(imagePath)
		if m := res.MimeType; m != "" {
			mime = m
		}
	}
	if mid == "" {
		return "", "", "", fmt.Errorf("metaai: AnalyzeImage requires either imagePath or mediaID")
	}
	if mime == "" {
		mime = "image/png"
	}
	return mid, filename, mime, nil
}

// defaultQuestion returns q, or the default analyze prompt when q is empty.
func defaultQuestion(q string) string {
	if q == "" {
		return "Describe what's happening in this image"
	}
	return q
}

// attachInfo holds image attachment data for the current chat message.
type attachInfo struct {
	MediaID  string
	Mime     string
	Filename string
}
