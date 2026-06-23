package metaai

// session.go implements topic-based conversation session management.
// The methods in this file manage named conversation topics.
//
// Each topic maps to a conversation id (auto-remembered after the first chat in
// that topic). Topics may carry per-topic default overrides for thinking/instant/
// mode; resolution order is per-call > topic > client defaults (see resolveChatConfig).

// SetTopic switches the current topic. The next Chat without an explicit Topic
// option (and without NewConversation) will continue in this topic's conversation.
func (c *Client) SetTopic(topic string) {
	c.currentTopic = topic
}

// NewTopic registers a fresh topic with optional default overrides and resets its
// conversation binding so the next chat starts a new conversation.
func (c *Client) NewTopic(topic string, thinking, instant *bool, mode *string) {
	if topic == "" {
		return
	}
	c.topicConfigMu.Lock()
	c.topicConfig[topic] = topicOverride{Thinking: thinking, Instant: instant, Mode: mode}
	c.topicConfigMu.Unlock()
	delete(c.topics, topic) // force a new conversation id on next chat
	c.currentTopic = topic
}

// GetTopic returns the conversation id bound to a topic (or the current topic if
// topic is ""). Returns "" when no conversation has happened in that topic yet.
func (c *Client) GetTopic(topic string) string {
	if topic == "" {
		topic = c.currentTopic
	}
	return c.topics[topic]
}

// ListTopics returns all tracked topic→conversation-id bindings.
func (c *Client) ListTopics() map[string]string {
	out := make(map[string]string, len(c.topics))
	for k, v := range c.topics {
		out[k] = v
	}
	return out
}

// CurrentTopic returns the active topic name.
func (c *Client) CurrentTopic() string { return c.currentTopic }

// LastConversationID returns the conversation id used for the most recent chat
// turn (set in streamChat). Callers (e.g. the REST UI's multi-conversation
// feature) use it to learn the id Meta AI assigned to a brand-new conversation
// so they can store and reuse it for subsequent turns.
func (c *Client) LastConversationID() string { return c.externalConversationID }

// SetTopicConfig overrides the thinking/instant/mode defaults for an existing topic.
func (c *Client) SetTopicConfig(topic string, thinking, instant *bool, mode *string) {
	c.topicConfigMu.Lock()
	defer c.topicConfigMu.Unlock()
	ov := c.topicConfig[topic]
	if thinking != nil {
		ov.Thinking = thinking
	}
	if instant != nil {
		ov.Instant = instant
	}
	if mode != nil {
		ov.Mode = mode
	}
	c.topicConfig[topic] = ov
}
