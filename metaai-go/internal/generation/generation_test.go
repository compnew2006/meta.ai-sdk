package generation

import (
	"testing"
)

// TestBuildVariablesTextToVideo locks down the TEXT_TO_VIDEO variable shape, the
// contract observed in the latest text-to-video capture.
func TestBuildVariablesTextToVideo(t *testing.T) {
	vars, err := BuildBaseVariables(BuildVariablesOptions{
		Prompt:        "astronaut in space",
		Operation:     OpTextToVideo,
		ContentPrefix: "Animate",
	})
	if err != nil {
		t.Fatal(err)
	}
	ir := vars["imagineOperationRequest"].(map[string]any)
	if ir["operation"] != "TEXT_TO_VIDEO" {
		t.Errorf("operation = %v, want TEXT_TO_VIDEO", ir["operation"])
	}
	t2i := ir["textToImageParams"].(map[string]any)
	if t2i["prompt"] != "astronaut in space" {
		t.Errorf("prompt = %v", t2i["prompt"])
	}
	if ir["requestId"] != nil {
		t.Errorf("requestId = %v, want nil", ir["requestId"])
	}
	if vars["entryPoint"] != "KADABRA__UNKNOWN" {
		t.Errorf("entryPoint = %v, want KADABRA__UNKNOWN", vars["entryPoint"])
	}
	if vars["currentBranchPath"] != "0" {
		t.Errorf("currentBranchPath = %v, want 0", vars["currentBranchPath"])
	}
}

// TestParseSSEGraphQLErrors locks "HTTP 200 with GraphQL errors → FAILED".
func TestParseSSEGraphQLErrors(t *testing.T) {
	body := `data: {"errors":[{"message":"Cannot query field \"name\" on type \"User\".","extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}]}`
	r := ParseSSE(body)
	if !r.HasGraphQLErrors {
		t.Fatal("HasGraphQLErrors = false, want true")
	}
	if r.StreamingState != "FAILED" {
		t.Errorf("StreamingState = %q, want FAILED", r.StreamingState)
	}
	if len(r.GraphQLErrors) != 1 {
		t.Fatalf("len(GraphQLErrors) = %d, want 1", len(r.GraphQLErrors))
	}
	if r.GraphQLErrors[0].Code != "GRAPHQL_VALIDATION_FAILED" {
		t.Errorf("Code = %q", r.GraphQLErrors[0].Code)
	}
}

// TestParseSSEImages locks the sendMessageStream images extraction.
func TestParseSSEImages(t *testing.T) {
	body := `data: {"data":{"sendMessageStream":{"streamingState":"OVERALL_DONE","conversationId":"conv_1","images":[{"id":"img_1","url":"https://example.com/image.jpg"}]}}}`
	r := ParseSSE(body)
	if len(r.Images) != 1 || r.Images[0] != "https://example.com/image.jpg" {
		t.Errorf("Images = %v", r.Images)
	}
	if r.ConversationID != "conv_1" {
		t.Errorf("ConversationID = %q", r.ConversationID)
	}
}

func TestParseSSEInvalidJSONSkipped(t *testing.T) {
	body := "data: {invalid json}\ndata: {\"valid\": \"json\"}"
	r := ParseSSE(body)
	// Invalid line skipped, no panic, no crash.
	if r == nil {
		t.Fatal("ParseSSE returned nil")
	}
}

// TestExtractMediaURLs locks the image/video URL extraction contract.
func TestExtractMediaURLs(t *testing.T) {
	data := map[string]any{
		"data": map[string]any{
			"xfb_imagine_send_message": map[string]any{
				"messages": map[string]any{
					"edges": []any{
						map[string]any{
							"node": map[string]any{
								"content": map[string]any{
									"imagine_media": map[string]any{
										"images": map[string]any{
											"nodes": []any{
												map[string]any{"uri": "https://example.com/image1.jpg"},
												map[string]any{"uri": "https://example.com/image2.jpg"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	urls := ExtractMediaURLs(data)
	if len(urls) != 2 {
		t.Fatalf("len(urls) = %d, want 2", len(urls))
	}
}

// TestIsMediaReady locks the media-readiness contract.
func TestIsMediaReady(t *testing.T) {
	if IsMediaReady(nil) {
		t.Error("empty → should be false")
	}
	ready := map[string]any{
		"data": map[string]any{
			"xfb_genai_fetch_post": map[string]any{
				"messages": map[string]any{
					"edges": []any{
						map[string]any{
							"node": map[string]any{
								"content": map[string]any{
									"imagine_video": map[string]any{
										"videos": map[string]any{
											"nodes": []any{
												map[string]any{"uri": "https://example.com/video_ready.mp4"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if !IsMediaReady(ready) {
		t.Error("urls present → should be ready")
	}
}

// TestDocIDEnvOverride locks the META_AI_DOC_ID_TEXT_TO_IMAGE override path.
func TestDocIDEnvOverride(t *testing.T) {
	t.Setenv("META_AI_DOC_ID_TEXT_TO_IMAGE", "abc123override")
	if got := DocID(OpTextToImage); got != "abc123override" {
		t.Errorf("DocID = %q, want abc123override", got)
	}
}

// TestDocIDDefault locks the default doc_id when no env override is present.
func TestDocIDDefault(t *testing.T) {
	t.Setenv("META_AI_DOC_ID_TEXT_TO_IMAGE", "")
	if got := DocID(OpTextToImage); got != "2f707e4a86f4b01adba97e1376cbdc14" {
		t.Errorf("DocID default = %q", got)
	}
}
