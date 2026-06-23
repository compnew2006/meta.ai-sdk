package proxy

// prompt.go assembles a single Meta AI prompt from OpenAI/Anthropic message
// arrays and normalizes multimodal content (text + images) carried by either
// API shape.

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// imageRef is a normalized image attached to a user message.
type imageRef struct {
	URL       string // http(s) or data: URL (set for URL-style sources)
	MediaType string
	Data      []byte // populated for base64 / data: URLs
}

// normalizeContent flattens an OpenAI/Anthropic content field (string or array
// of typed parts) into plain text plus any image references.
func normalizeContent(content interface{}) (text string, images []imageRef) {
	switch v := content.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case []interface{}:
		var b strings.Builder
		for _, part := range v {
			m, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			switch m["type"] {
			case "text", "input_text", "output_text":
				if t, ok := m["text"].(string); ok {
					if b.Len() > 0 {
						b.WriteString("\n")
					}
					b.WriteString(t)
				}
			case "image_url":
				if iu, ok := m["image_url"].(map[string]interface{}); ok {
					if u, ok := iu["url"].(string); ok {
						images = append(images, parseImageURL(u))
					}
				}
			case "image":
				if src, ok := m["source"].(map[string]interface{}); ok {
					images = append(images, parseAnthropicSource(src))
				}
			}
		}
		return b.String(), images
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func parseImageURL(u string) imageRef {
	r := imageRef{URL: u}
	if strings.HasPrefix(u, "data:") {
		if comma := strings.Index(u, ","); comma > 0 {
			r.MediaType = dataMIME(u[:comma])
			if b, err := base64.StdEncoding.DecodeString(u[comma+1:]); err == nil {
				r.Data = b
			}
		}
	}
	return r
}

func parseAnthropicSource(src map[string]interface{}) imageRef {
	r := imageRef{}
	switch src["type"] {
	case "base64":
		r.MediaType, _ = src["media_type"].(string)
		if d, ok := src["data"].(string); ok {
			if b, err := base64.StdEncoding.DecodeString(d); err == nil {
				r.Data = b
			}
		}
	case "url":
		r.URL, _ = src["url"].(string)
		if mt, ok := src["media_type"].(string); ok {
			r.MediaType = mt
		}
	}
	return r
}

func dataMIME(meta string) string {
	i := strings.Index(meta, ":")
	if i < 0 {
		return "image/png"
	}
	rest := meta[i+1:]
	if j := strings.Index(rest, ";"); j >= 0 {
		return rest[:j]
	}
	return rest
}

// materialize writes the image to a temp file and returns its path (used by
// AnalyzeImage, which needs a file path to upload).
func (r imageRef) materialize(client *http.Client) (string, error) {
	if len(r.Data) > 0 {
		f, err := os.CreateTemp("", "metaai-img-*"+mimeExt(r.MediaType))
		if err != nil {
			return "", err
		}
		if _, err := f.Write(r.Data); err != nil {
			f.Close()
			os.Remove(f.Name())
			return "", err
		}
		f.Close()
		return f.Name(), nil
	}
	if r.URL == "" {
		return "", fmt.Errorf("image has no data or url")
	}
	c := client
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Get(r.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("image download failed: %s", resp.Status)
	}
	mt := r.MediaType
	if mt == "" {
		mt = resp.Header.Get("Content-Type")
	}
	f, err := os.CreateTemp("", "metaai-img-*"+mimeExt(mt))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func mimeExt(mt string) string {
	switch strings.ToLower(mt) {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	}
	switch {
	case strings.Contains(mt, "png"):
		return ".png"
	case strings.Contains(mt, "jpeg"), strings.Contains(mt, "jpg"):
		return ".jpg"
	case strings.Contains(mt, "gif"):
		return ".gif"
	case strings.Contains(mt, "webp"):
		return ".webp"
	}
	return ".png"
}

func systemText(system interface{}) string {
	if system == nil {
		return ""
	}
	t, _ := normalizeContent(system)
	return strings.TrimSpace(t)
}

// assembleOaiTranscript folds OpenAI messages into a single labeled prompt and
// returns any images attached to the last user message (the only ones that can
// be attached to the outgoing Meta AI turn). OpenAI folds system into messages,
// so system here is nil.
func assembleOaiTranscript(system interface{}, msgs []oaiMessage) (string, []imageRef) {
	var b strings.Builder
	if sys := systemText(system); sys != "" {
		fmt.Fprintf(&b, "[System]\n%s\n\n", sys)
	}
	lastUser := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if strings.EqualFold(msgs[i].Role, "user") {
			lastUser = i
			break
		}
	}
	var lastImages []imageRef
	for i, m := range msgs {
		text, imgs := normalizeContent(m.Content)
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role == "" {
			role = "user"
		}
		switch role {
		case "system":
			fmt.Fprintf(&b, "[System]\n%s\n\n", text)
		case "user":
			fmt.Fprintf(&b, "[User]\n%s\n\n", text)
			if i == lastUser {
				lastImages = imgs
			}
		case "assistant":
			fmt.Fprintf(&b, "[Assistant]\n%s\n\n", text)
		case "tool", "function":
			fmt.Fprintf(&b, "[Tool]\n%s\n\n", text)
		default:
			fmt.Fprintf(&b, "[%s]\n%s\n\n", title(role), text)
		}
	}
	return strings.TrimRight(b.String(), "\n"), lastImages
}

// assembleAnthTranscript is the Anthropic analogue of assembleOaiTranscript.
func assembleAnthTranscript(system interface{}, msgs []anthMessage) (string, []imageRef) {
	var b strings.Builder
	if sys := systemText(system); sys != "" {
		fmt.Fprintf(&b, "[System]\n%s\n\n", sys)
	}
	lastUser := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if strings.EqualFold(msgs[i].Role, "user") {
			lastUser = i
			break
		}
	}
	var lastImages []imageRef
	for i, m := range msgs {
		text, imgs := normalizeContent(m.Content)
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role == "" {
			role = "user"
		}
		switch role {
		case "system":
			fmt.Fprintf(&b, "[System]\n%s\n\n", text)
		case "user":
			fmt.Fprintf(&b, "[User]\n%s\n\n", text)
			if i == lastUser {
				lastImages = imgs
			}
		case "assistant":
			fmt.Fprintf(&b, "[Assistant]\n%s\n\n", text)
		default:
			fmt.Fprintf(&b, "[%s]\n%s\n\n", title(role), text)
		}
	}
	return strings.TrimRight(b.String(), "\n"), lastImages
}

func title(s string) string {
	if s == "" {
		return "User"
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
