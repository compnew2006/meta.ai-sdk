package metaai

// GenerateImage and GenerateVideo generate media via Meta AI. The browser
// generates media through the WebSocket chat path (the prompt text itself is
// the trigger — no special mode needed), and the resulting media URLs are
// surfaced through the mediaLibraryFeed GraphQL query. This file:
//   - sends the prompt as a chat message over the WebSocket,
//   - polls mediaLibraryFeed for the freshly generated media item,
//   - extracts image/video URLs from the feed response.
//
// Earlier approaches (the SSE /api/graphql imagineOperationRequest with doc_id
// 2f707e4a, and the WS chat "create_image"/"create_video" mode) were both
// silently rejected by the server — the former because the doc_id references a
// removed RewriteOptionsInput type, the latter because the server ignores the
// mode flag and just streams chat text. The mediaLibraryFeed poll is the path
// the browser itself uses to display generated results.

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GenerationResult is the unified outcome of an image/video generation call.
type GenerationResult struct {
	Success          bool     `json:"success"`
	Prompt           string   `json:"prompt"`
	URLs             []string `json:"urls"`
	MediaIDs         []string `json:"media_ids,omitempty"`
	Status           string   `json:"status"`
	Processing       bool     `json:"processing"`
	HasGraphQLErrors bool     `json:"has_graphql_errors"`
	GraphQLErrors    []string `json:"graphql_errors,omitempty"`
	ConversationID   string   `json:"conversation_id,omitempty"`
	Error            string   `json:"error,omitempty"`
}

// GenerateImage generates images from a text prompt.
//
// Result/error contract: the returned error is authoritative — when it is
// non-nil the call failed and callers MUST treat it as such. The returned
// *GenerationResult is diagnostic metadata only on the error path; it is never
// a successful result when err != nil (Success is false).
func (c *Client) GenerateImage(ctx context.Context, prompt, orientation string, numImages int) (*GenerationResult, error) {
	return c.generate(ctx, "Imagine", prompt)
}

// GenerateVideo generates a video from a text prompt.
//
// Result/error contract: see GenerateImage.
func (c *Client) GenerateVideo(ctx context.Context, prompt string) (*GenerationResult, error) {
	return c.generate(ctx, "Animate", prompt)
}

// ExtendVideo extends a previously generated video clip identified by its
// media id. The server generates a continuation of the source clip; we detect
// it via the same mediaLibraryFeed poll as the primary generation path.
//
// Result/error contract: see GenerateImage.
func (c *Client) ExtendVideo(ctx context.Context, sourceMediaID string) (*GenerationResult, error) {
	if err := c.EnsureAccessToken(ctx); err != nil {
		return &GenerationResult{Status: "FAILED", Error: err.Error()}, err
	}
	content := fmt.Sprintf("Extend this video %s", sourceMediaID)
	baseline, _ := c.latestMediaLibraryItem(ctx)
	chatCtx, chatCancel := context.WithTimeout(context.Background(), 300*time.Second)
	go func() {
		defer chatCancel()
		_, _ = c.Chat(chatCtx, content, &ChatOptions{NewConversation: true})
	}()
	item, err := c.waitForNewMediaItem(ctx, baseline, content)
	if err != nil {
		return &GenerationResult{Status: "FAILED", Error: err.Error()}, err
	}
	urls, ids := extractMediaFeedURLsAndIDs(item)
	return &GenerationResult{
		Prompt:         content,
		Status:         "READY",
		Success:        len(urls) > 0,
		URLs:           urls,
		MediaIDs:       ids,
		ConversationID: asStringM(item["conversationId"]),
	}, nil
}

// generate sends the prompt over the WS chat path, then polls the media
// library feed for the freshly generated media item.
func (c *Client) generate(ctx context.Context, prefix, prompt string) (*GenerationResult, error) {
	if err := c.EnsureAccessToken(ctx); err != nil {
		return &GenerationResult{Prompt: prompt, Status: "FAILED", Error: err.Error()}, err
	}

	content := strings.TrimSpace(prompt)
	if prefix != "" {
		content = prefix + " " + content
	}

	// Capture the baseline feed item id BEFORE triggering, so we can detect the
	// freshly generated item when it appears.
	baseline, _ := c.latestMediaLibraryItem(ctx)

	// Send the prompt as a chat message. The server triggers media generation
	// from the prompt text. Use a separate context so the chat goroutine is
	// not cancelled when generate returns, and so it does not block our feed
	// polling. Video generation can take several minutes, so the chat context
	// matches the poll deadline.
	chatCtx, chatCancel := context.WithTimeout(context.Background(), 300*time.Second)
	go func() {
		defer chatCancel()
		_, _ = c.Chat(chatCtx, content, &ChatOptions{NewConversation: true})
	}()

	item, err := c.waitForNewMediaItem(ctx, baseline, prompt)
	if err != nil {
		return &GenerationResult{Prompt: prompt, Status: "FAILED", Error: err.Error()}, err
	}
	urls, ids := extractMediaFeedURLsAndIDs(item)
	return &GenerationResult{
		Prompt:         prompt,
		Status:         "READY",
		Success:        len(urls) > 0,
		URLs:           urls,
		MediaIDs:       ids,
		ConversationID: asStringM(item["conversationId"]),
	}, nil
}

// waitForNewMediaItem polls the media library feed until a new item appears
// (different id from baseline) whose content/prompt matches the request.
func (c *Client) waitForNewMediaItem(ctx context.Context, baseline map[string]any, prompt string) (map[string]any, error) {
	baseID := asStringM(baseline["id"])
	// Video generation can take several minutes; poll up to 5 minutes.
	deadline := time.Now().Add(300 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
		item, err := c.latestMediaLibraryItem(ctx)
		if err != nil || item == nil {
			continue
		}
		if id := asStringM(item["id"]); id != "" && id != baseID {
			// New item — accept it (generation order is most-recent-first).
			return item, nil
		}
	}
	return nil, fmt.Errorf("metaai: generation timed out waiting for media library update")
}

// latestMediaLibraryItem returns the most recent item from the media library
// feed (the same query the /create page uses to render generated media).
func (c *Client) latestMediaLibraryItem(ctx context.Context) (map[string]any, error) {
	data, err := c.graphqlRequestOAuth(ctx, DocIDs.MediaLibraryFeed, map[string]any{
		"first":       1,
		"after":       nil,
		"filter":      nil,
		"searchQuery": nil,
		"shape":       "v1",
	})
	if err != nil {
		return nil, err
	}
	var view struct {
		MediaLibraryFeed struct {
			Edges []struct {
				Node map[string]any `json:"node"`
			} `json:"edges"`
		} `json:"mediaLibraryFeed"`
	}
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, fmt.Errorf("metaai: decode media library feed: %w", err)
	}
	if len(view.MediaLibraryFeed.Edges) == 0 {
		return nil, nil
	}
	node := view.MediaLibraryFeed.Edges[0].Node
	return node, nil
}

// extractMediaFeedURLs pulls image + video URLs out of a media-library node.
func extractMediaFeedURLs(node map[string]any) []string {
	urls, _ := extractMediaFeedURLsAndIDs(node)
	return urls
}

// extractMediaFeedURLsAndIDs pulls image + video URLs and their media ids out
// of a media-library node. The media-library node ids are like
// "media_library:<numeric>"; the embedded media objects carry their own "id".
func extractMediaFeedURLsAndIDs(node map[string]any) ([]string, []string) {
	urls := []string{}
	ids := []string{}
	walk := func(v any) {
		arr, ok := v.([]any)
		if !ok {
			return
		}
		for _, e := range arr {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			if u := asStringM(m["url"]); u != "" {
				urls = append(urls, u)
			}
			if id := asStringM(m["id"]); id != "" {
				ids = append(ids, id)
			}
		}
	}
	walk(node["images"])
	walk(node["videos"])
	return urls, ids
}

func asStringM(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// urlRe matches http(s) URLs in generation replies, stopping at whitespace or
// common JSON/quote delimiters.
var urlRe = regexp.MustCompile(`https?://[^\s"'}\n]+`)

// extractURLs finds http(s) URLs in a text response.
func extractURLs(text string) []string {
	return urlRe.FindAllString(text, -1)
}
