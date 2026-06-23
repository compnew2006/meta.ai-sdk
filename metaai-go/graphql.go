package metaai

// graphql.go implements the persisted-query GraphQL HTTP transport used by
// meta.ai for SUPPORTING operations (conversation config, mark-seen, latency,
// history, media fetch, generation). Chat sendMessage itself goes over the
// clippy WebSocket.
//
// Request shape (captured): POST https://meta.ai/api/graphql with JSON
// {"doc_id": "...", "variables": {...}}, headers Authorization (ecto1: token),
// Origin, Referer, browser UA, plus auth cookies.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// graphqlResponse is the raw JSON envelope returned by /api/graphql.
type graphqlResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

// graphQLError mirrors a Meta GraphQL error entry.
type graphQLError struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

// code returns the GraphQL error extension code (e.g. GRAPHQL_VALIDATION_FAILED).
func (e graphQLError) code() string {
	if e.Extensions == nil {
		return ""
	}
	if c, ok := e.Extensions["code"].(string); ok {
		return c
	}
	return ""
}

// graphqlRequest sends a persisted-query GraphQL POST and decodes the JSON.
// The body is {"doc_id": docID, "variables": variables}. The Authorization
// header is set to the raw access token (matches the capture; the history/search
// paths use "OAuth <token>" via graphqlRequestOAuth).
//
// Raises ErrGraphQLValidation when Meta returns a GRAPHQL_VALIDATION_FAILED
// error (doc_id drift).
func (c *Client) graphqlRequest(ctx context.Context, docID string, variables any) (json.RawMessage, error) {
	return c.graphqlRequestAuth(ctx, docID, variables, false)
}

// graphqlRequestOAuth is like graphqlRequest but sends "Authorization: OAuth <token>"
// (used by viewer/history/search queries).
func (c *Client) graphqlRequestOAuth(ctx context.Context, docID string, variables any) (json.RawMessage, error) {
	return c.graphqlRequestAuth(ctx, docID, variables, true)
}

func (c *Client) graphqlRequestAuth(ctx context.Context, docID string, variables any, oauth bool) (json.RawMessage, error) {
	body, err := json.Marshal(map[string]any{"doc_id": docID, "variables": variables})
	if err != nil {
		return nil, fmt.Errorf("metaai: marshal graphql body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GraphqlURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.meta.ai")
	req.Header.Set("Referer", "https://www.meta.ai/")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if c.accessToken != "" {
		if oauth {
			req.Header.Set("Authorization", "OAuth "+c.accessToken)
		} else {
			// Supporting ops send the bare ecto1: token.
			req.Header.Set("Authorization", c.accessToken)
		}
	}
	attachCookies(req, c.cookies)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("metaai: graphql POST: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("metaai: read graphql response: %w", err)
	}
	var gr graphqlResponse
	if jerr := json.Unmarshal(raw, &gr); jerr != nil {
		return nil, fmt.Errorf("metaai: decode graphql response: %w (body: %s)", jerr, truncate(string(raw), 200))
	}
	if len(gr.Errors) > 0 {
		if gr.Errors[0].code() == "GRAPHQL_VALIDATION_FAILED" {
			return nil, fmt.Errorf("%w: %s", ErrGraphQLValidation, gr.Errors[0].Message)
		}
		return nil, fmt.Errorf("metaai: graphql error: %s", gr.Errors[0].Message)
	}
	return gr.Data, nil
}

// RegisterConversation POSTs the conversation-registration mutation
// (doc_id e7f80258…) captured in the new-conversation lifecycle. Optional —
// the default chat flow does not require it, but callers wanting to mirror the
// full browser lifecycle may call it before StreamChat.
func (c *Client) RegisterConversation(ctx context.Context, conversationID string) error {
	_, err := c.graphqlRequest(ctx, DocIDs.ConversationRegistration, map[string]any{
		"conversationId": conversationID,
	})
	return err
}

// SetConversationMode POSTs the mode-setter mutation (doc_id c32bbe99…).
// mode is "think_hard" when thinking is enabled, else "fast". Optional.
func (c *Client) SetConversationMode(ctx context.Context, conversationID string, thinking bool) error {
	mode := "fast"
	if thinking {
		mode = "think_hard"
	}
	_, err := c.graphqlRequest(ctx, DocIDs.ModeSetter, map[string]any{
		"input": map[string]any{
			"conversationId": conversationID,
			"mode":           mode,
		},
	})
	return err
}

// markConversationSeen POSTs the mark-seen mutation (doc_id 0b2cb3a4...,
// = META_AI_LAST_SEEN_DOC_ID) captured in the live flow. Errors are logged
// because this is a best-effort notification; the chat stream continues
// regardless of whether the server acknowledges the seen marker.
func (c *Client) markConversationSeen(ctx context.Context, conversationID, lastSeenMessageID string) {
	if conversationID == "" || lastSeenMessageID == "" {
		return
	}
	vars := map[string]any{
		"input": map[string]any{
			"conversationId":    conversationID,
			"lastSeenMessageId": lastSeenMessageID,
		},
	}
	if _, err := c.graphqlRequest(ctx, DocIDs.MarkSeen, vars); err != nil {
		// Non-fatal: mark-seen is a best-effort server notification and the
		// chat stream continues regardless. Surface the failure at debug level
		// so operators can detect persistent breakage (e.g. doc_id drift).
		c.logger.Debugf("metaai: mark-seen failed for conversation %s (lastSeen=%s): %v",
			conversationID, lastSeenMessageID, err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
