package proxy

// proxy.go bridges an HTTP request to the metaai Client. complete/stream run a
// single chat turn (optionally with an image attachment); doGeneration routes
// image/video model requests to the SDK generators.

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/smart-studio/metaai-go"
)

// chatClient is the subset of *metaai.Client the proxy depends on. Declaring it
// as an interface lets tests inject a fake client instead of hitting Meta AI.
// *metaai.Client satisfies it implicitly.
type chatClient interface {
	Chat(ctx context.Context, message string, opts *metaai.ChatOptions) (string, error)
	StreamChat(ctx context.Context, message string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk
	AnalyzeImage(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) (string, error)
	GenerateImage(ctx context.Context, prompt, orientation string, numImages int) (*metaai.GenerationResult, error)
	GenerateVideo(ctx context.Context, prompt string) (*metaai.GenerationResult, error)
}

// complete performs a single (non-streaming) chat turn against Meta AI. When
// images are present, the turn runs through AnalyzeImage so Meta AI can see
// the image content.
func (s *Server) complete(ctx context.Context, prompt string, images []imageRef, opts *metaai.ChatOptions) (string, error) {
	if len(images) > 0 {
		return s.completeWithImages(ctx, prompt, images, opts)
	}
	return s.client.Chat(ctx, prompt, opts)
}

func (s *Server) completeWithImages(ctx context.Context, prompt string, images []imageRef, opts *metaai.ChatOptions) (string, error) {
	path, err := images[0].materialize(s.http)
	if err != nil {
		return "", fmt.Errorf("download image: %w", err)
	}
	defer os.Remove(path)
	return s.client.AnalyzeImage(ctx, path, "", prompt, opts)
}

// stream returns a channel of chat chunks plus a cleanup func (no-op for the
// text path; the image path removes its temp file internally).
func (s *Server) stream(ctx context.Context, prompt string, images []imageRef, opts *metaai.ChatOptions) (<-chan metaai.ChatChunk, func()) {
	out := make(chan metaai.ChatChunk, 8)
	if len(images) == 0 {
		src := s.client.StreamChat(ctx, prompt, opts)
		go func() {
			defer close(out)
			for c := range src {
				out <- c
			}
		}()
		return out, func() {}
	}
	path, err := images[0].materialize(s.http)
	if err != nil {
		go func() {
			defer close(out)
			out <- metaai.ChatChunk{Err: err, Done: true}
		}()
		return out, func() {}
	}
	go func() {
		defer close(out)
		defer os.Remove(path)
		text, err := s.client.AnalyzeImage(ctx, path, "", prompt, opts)
		if err != nil {
			out <- metaai.ChatChunk{Err: err, Done: true}
			return
		}
		out <- metaai.ChatChunk{Text: text, Done: true}
	}()
	return out, func() {}
}

// doGeneration routes an image/video model request to the SDK generators and
// returns the result text (typically media URLs).
func (s *Server) doGeneration(ctx context.Context, model modelInfo, prompt string) (string, error) {
	switch model.Generation {
	case "image":
		res, err := s.client.GenerateImage(ctx, prompt, "SQUARE", 1)
		if err != nil {
			return "", err
		}
		return formatGeneration(res), nil
	case "video":
		res, err := s.client.GenerateVideo(ctx, prompt)
		if err != nil {
			return "", err
		}
		return formatGeneration(res), nil
	}
	return s.client.Chat(ctx, prompt, &metaai.ChatOptions{NewConversation: true})
}

func formatGeneration(res *metaai.GenerationResult) string {
	if len(res.URLs) == 0 {
		if res.Error != "" {
			return res.Error
		}
		return res.Status
	}
	return strings.Join(res.URLs, "\n")
}
