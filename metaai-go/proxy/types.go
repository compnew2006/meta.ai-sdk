package proxy

// OpenAI-compatible Chat Completions types (subset of the API surface; only
// what the proxy needs to read and emit).

type oaiMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content,omitempty"`
	Name    string      `json:"name,omitempty"`
}

// oaiContentPart documents the typed content-part shape (text or image_url).
// It is not parsed directly — incoming content is decoded generically via
// normalizeContent so unknown part types are tolerated.
type oaiContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *oaiImageURL `json:"image_url,omitempty"`
}

type oaiImageURL struct {
	URL string `json:"url"`
}

type oaiChatRequest struct {
	Model       string       `json:"model"`
	Messages    []oaiMessage `json:"messages"`
	Stream      bool         `json:"stream,omitempty"`
	MaxTokens   *int         `json:"max_tokens,omitempty"`
	Temperature *float64     `json:"temperature,omitempty"`
	TopP        *float64     `json:"top_p,omitempty"`
	User        string       `json:"user,omitempty"`
}

type oaiChatResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []oaiChoice `json:"choices"`
	Usage   oaiUsage    `json:"usage"`
}

type oaiChoice struct {
	Index        int        `json:"index"`
	Message      oaiMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type oaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type oaiChatChunk struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []oaiChunkChoice `json:"choices"`
}

type oaiChunkChoice struct {
	Index        int      `json:"index"`
	Delta        oaiDelta `json:"delta"`
	FinishReason *string  `json:"finish_reason"`
}

type oaiDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type oaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type oaiModelsResponse struct {
	Object string     `json:"object"`
	Data   []oaiModel `json:"data"`
}

// Anthropic-compatible Messages types (subset).

type anthMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content,omitempty"`
}

type anthContentPart struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	Source *anthSource `json:"source,omitempty"`
}

type anthSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type anthMessagesRequest struct {
	Model       string        `json:"model"`
	Messages    []anthMessage `json:"messages"`
	System      interface{}   `json:"system,omitempty"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

type anthMessagesResponse struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Role         string            `json:"role"`
	Model        string            `json:"model"`
	Content      []anthContentPart `json:"content"`
	StopReason   *string           `json:"stop_reason"`
	StopSequence *string           `json:"stop_sequence"`
	Usage        anthUsage         `json:"usage"`
}

type anthUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
