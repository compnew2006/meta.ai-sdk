package metaai

// history.go implements conversation history listing + search.
// Grounded in the captured doc_id e6214ae9… (META_AI_HISTORY_DOC_ID) and the
// observed conversation-history response shape.

import (
	"context"
	"encoding/json"
	"fmt"
)

// Conversation is one entry in the user's conversation history.
type Conversation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	LastMessage string `json:"last_message"`
	UpdatedAt   string `json:"updated_at"`
}

// HistoryResult is the result of GetConversationHistory.
type HistoryResult struct {
	Success       bool           `json:"success"`
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
	Error         string         `json:"error,omitempty"`
}

// GetConversationHistory lists the user's past conversations.
// Uses the persisted-query doc_id META_AI_HISTORY_DOC_ID (e6214ae9…) with
// "Authorization: OAuth <token>".
//
// Result/error contract: the returned error is authoritative. On the error
// path the *HistoryResult is diagnostic metadata (Error describes the failure)
// and Success is false; it is never a success when err != nil.
func (c *Client) GetConversationHistory(ctx context.Context, limit, offset int) (*HistoryResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if err := c.EnsureAccessToken(ctx); err != nil {
		return &HistoryResult{Error: err.Error()}, err
	}
	data, err := c.graphqlRequestOAuth(ctx, DocIDs.History, map[string]any{"first": limit, "offset": offset})
	if err != nil {
		return &HistoryResult{Error: err.Error()}, err
	}

	var view struct {
		Data struct {
			Viewer struct {
				Conversations struct {
					Edges []struct {
						Node struct {
							ID          string `json:"id"`
							Title       string `json:"title"`
							Name        string `json:"name"`
							UpdatedTime string `json:"updated_time"`
							LastUpdated string `json:"last_updated"`
							LastMessage struct {
								Text string `json:"text"`
							} `json:"last_message"`
						} `json:"node"`
					} `json:"edges"`
					Count int `json:"count"`
				} `json:"conversations"`
			} `json:"viewer"`
		} `json:"data"`
	}
	if jerr := json.Unmarshal(data, &view); jerr != nil {
		// Schema drift: the response did not match the expected shape. Surface
		// it rather than silently returning an empty "success".
		return &HistoryResult{
			Success: false,
			Limit:   limit,
			Offset:  offset,
			Error:   fmt.Sprintf("decode history response: %v (body: %s)", jerr, truncate(string(data), 200)),
		}, fmt.Errorf("metaai: decode history response: %w", jerr)
	}

	res := &HistoryResult{Success: true, Limit: limit, Offset: offset}
	for _, e := range view.Data.Viewer.Conversations.Edges {
		n := e.Node
		title := n.Title
		if title == "" {
			title = n.Name
		}
		if title == "" {
			title = "Untitled"
		}
		updated := n.UpdatedTime
		if updated == "" {
			updated = n.LastUpdated
		}
		res.Conversations = append(res.Conversations, Conversation{
			ID: n.ID, Title: title, LastMessage: n.LastMessage.Text, UpdatedAt: updated,
		})
	}
	res.Total = view.Data.Viewer.Conversations.Count
	if res.Total == 0 {
		res.Total = len(res.Conversations)
	}
	return res, nil
}

// SearchResult is one hit from SearchConversations.
type SearchResult struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	Snippet        string `json:"snippet"`
	Timestamp      string `json:"timestamp"`
}

// SearchConversations searches the user's conversation history.
// The search GraphQL response shape is not yet captured from a live session;
// this returns an explicit unimplemented error until that work is done.
func (c *Client) SearchConversations(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if err := c.EnsureAccessToken(ctx); err != nil {
		return nil, err
	}
	_ = query
	_ = limit
	return nil, fmt.Errorf("metaai: SearchConversations not yet implemented (response shape not captured)")
}
