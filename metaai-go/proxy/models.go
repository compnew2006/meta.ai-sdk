package proxy

import "strings"

// modelInfo maps a virtual model name (exposed via /v1/models) to Meta AI
// chat-config behavior.
type modelInfo struct {
	Name        string
	Description string
	Mode        string // learn | analyze | create_image | create_video ("" = default)
	Thinking    bool
	Instant     bool
	Generation  string // "" | "image" | "video"
}

// modelRegistry is the list of virtual models the proxy advertises. The names
// are stable identifiers clients put in the "model" field.
var modelRegistry = []modelInfo{
	{Name: "meta-ai", Description: "Meta AI (default chat)"},
	{Name: "meta-ai-fast", Description: "Meta AI fast / instant mode", Instant: true},
	{Name: "meta-ai-think", Description: "Meta AI with extended thinking", Thinking: true},
	{Name: "meta-ai-analyze", Description: "Meta AI analyze mode", Mode: "analyze"},
	{Name: "meta-ai-learn", Description: "Meta AI learn mode", Mode: "learn"},
	{Name: "meta-ai-image", Description: "Meta AI image generation", Mode: "create_image", Generation: "image"},
	{Name: "meta-ai-video", Description: "Meta AI video generation", Mode: "create_video", Generation: "video"},
}

// resolveModel maps a client-supplied model name to a modelInfo. Unknown names
// are matched heuristically (so "gpt-4", "claude-3-5-sonnet", etc. work), then
// fall back to the default chat model.
func resolveModel(name string) modelInfo {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, m := range modelRegistry {
		if m.Name == n {
			return m
		}
	}
	switch {
	case strings.Contains(n, "image"):
		return resolveModel("meta-ai-image")
	case strings.Contains(n, "video"):
		return resolveModel("meta-ai-video")
	case strings.Contains(n, "think") || strings.Contains(n, "reason") || strings.Contains(n, "opus"):
		return resolveModel("meta-ai-think")
	case strings.Contains(n, "fast") || strings.Contains(n, "instant") || strings.Contains(n, "haiku"):
		return resolveModel("meta-ai-fast")
	}
	return resolveModel("meta-ai")
}

func modelNames() []string {
	out := make([]string, 0, len(modelRegistry))
	for _, m := range modelRegistry {
		out = append(out, m.Name)
	}
	return out
}

func modelsResponse() oaiModelsResponse {
	resp := oaiModelsResponse{Object: "list"}
	for _, m := range modelRegistry {
		resp.Data = append(resp.Data, oaiModel{
			ID: m.Name, Object: "model", Created: 1718841600, OwnedBy: "meta",
		})
	}
	return resp
}
