package generation

import "testing"

// === BUILD VARIABLES TESTS ===
// Tests that BuildBaseVariables produces correct shapes for each operation.

func TestBuildVariablesImageToImageProducesSourceMediaEntId(t *testing.T) {
	vars, err := BuildBaseVariables(BuildVariablesOptions{
		Prompt:    "make it blue",
		Operation: OpTextToImage,
		MediaIDs:  []string{"123456"},
		NumImages: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	ir := vars["imagineOperationRequest"].(map[string]any)
	if ir["operation"] != "IMAGE_TO_IMAGE" {
		t.Errorf("operation = %v, want IMAGE_TO_IMAGE", ir["operation"])
	}
	params := ir["imageToImageParams"].(map[string]any)
	if params["sourceMediaEntId"] != "123456" {
		t.Errorf("sourceMediaEntId = %v", params["sourceMediaEntId"])
	}
	if params["numMedia"] != 2 {
		t.Errorf("numMedia = %v, want 2", params["numMedia"])
	}
}

func TestBuildVariablesImageToVideoProducesImageToVideoParams(t *testing.T) {
	vars, err := BuildBaseVariables(BuildVariablesOptions{
		Prompt:    "animate this",
		Operation: OpTextToVideo,
		MediaIDs:  []string{"789"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ir := vars["imagineOperationRequest"].(map[string]any)
	if ir["operation"] != "IMAGE_TO_VIDEO" {
		t.Errorf("operation = %v", ir["operation"])
	}
	params := ir["imageToVideoParams"].(map[string]any)
	if params["sourceMediaEntId"] != "789" {
		t.Errorf("sourceMediaEntId = %v", params["sourceMediaEntId"])
	}
}

func TestBuildVariablesExtendVideoRequiresSourceID(t *testing.T) {
	_, err := BuildBaseVariables(BuildVariablesOptions{
		Prompt:    "extend",
		Operation: OpExtendVideo,
	})
	if err == nil {
		t.Error("expected error for extend video without source ID")
	}
}

func TestBuildVariablesExtendVideoProducesCorrectShape(t *testing.T) {
	vars, err := BuildBaseVariables(BuildVariablesOptions{
		Prompt:          "extend it",
		Operation:       OpExtendVideo,
		ExtendSourceID:  "src-1",
		ExtendSourceURL: "https://example.com/v.mp4",
	})
	if err != nil {
		t.Fatal(err)
	}
	ir := vars["imagineOperationRequest"].(map[string]any)
	if ir["operation"] != "EXTEND_VIDEO" {
		t.Errorf("operation = %v", ir["operation"])
	}
	params := ir["extendVideoParams"].(map[string]any)
	if params["sourceMediaEntId"] != "src-1" {
		t.Errorf("sourceMediaEntId = %v", params["sourceMediaEntId"])
	}
}

func TestBuildVariablesContentPrefixApplied(t *testing.T) {
	vars, _ := BuildBaseVariables(BuildVariablesOptions{
		Prompt:        "astronaut",
		Operation:     OpTextToImage,
		ContentPrefix: "Imagine",
	})
	if vars["content"] != "Imagine astronaut" {
		t.Errorf("content = %v, want 'Imagine astronaut'", vars["content"])
	}
}

func TestBuildVariablesThinkingSetsForceThinkingEnabled(t *testing.T) {
	vars, _ := BuildBaseVariables(BuildVariablesOptions{
		Prompt:    "test",
		Operation: OpTextToImage,
		Thinking:  true,
	})
	dev := vars["developerOverridesForMessage"]
	if dev == nil {
		t.Fatal("developerOverridesForMessage is nil")
	}
	m := dev.(map[string]any)
	if m["forceThinkingEnabled"] != true {
		t.Error("forceThinkingEnabled not set")
	}
}

func TestBuildVariablesInstantSetsSkipThinking(t *testing.T) {
	vars, _ := BuildBaseVariables(BuildVariablesOptions{
		Prompt:    "test",
		Operation: OpTextToImage,
		Instant:   true,
	})
	m := vars["developerOverridesForMessage"].(map[string]any)
	if m["skipThinking"] != true {
		t.Error("skipThinking not set")
	}
}

func TestBuildVariablesEntryPointIsUnifiedCanvasForExtend(t *testing.T) {
	vars, _ := BuildBaseVariables(BuildVariablesOptions{
		Prompt:         "ext",
		Operation:      OpExtendVideo,
		ExtendSourceID: "s",
	})
	if vars["entryPoint"] != "KADABRA__IMAGINE_UNIFIED_CANVAS" {
		t.Errorf("entryPoint = %v", vars["entryPoint"])
	}
	if vars["currentBranchPath"] != "2" {
		t.Errorf("currentBranchPath = %v", vars["currentBranchPath"])
	}
}

func TestBuildVariablesRejectsEmptyPrompt(t *testing.T) {
	_, err := BuildBaseVariables(BuildVariablesOptions{Operation: OpTextToImage})
	if err == nil {
		t.Error("expected error for empty prompt")
	}
}

// === DOC ID TESTS ===

func TestDocIDEnvOverrideForTextToVideo(t *testing.T) {
	t.Setenv("META_AI_DOC_ID_TEXT_TO_VIDEO", "customvideodoc123")
	if got := DocID(OpTextToVideo); got != "customvideodoc123" {
		t.Errorf("DocID = %q, want 'customvideodoc123'", got)
	}
}

func TestDocIDEnvOverrideFallsBackToSharedDocID(t *testing.T) {
	t.Setenv("META_AI_DOC_ID_TEXT_TO_VIDEO", "")
	t.Setenv("META_AI_DOC_ID", "sharedoverride")
	if got := DocID(OpTextToImage); got != "sharedoverride" {
		t.Errorf("DocID = %q, want 'sharedoverride'", got)
	}
}

func TestDocIDIgnoresNonAlphanumericOverride(t *testing.T) {
	t.Setenv("META_AI_DOC_ID_TEXT_TO_IMAGE", "notvalid123!")
	if got := DocID(OpTextToImage); got == "notvalid123!" {
		t.Error("non-alphanumeric override should be ignored")
	}
}

// === SSE PARSE TESTS ===

func TestParseSSEStreamingDoneState(t *testing.T) {
	body := `data: {"data":{"sendMessageStream":{"streamingState":"STREAMING"}}}`
	r := ParseSSE(body)
	if r.StreamingState != "STREAMING" {
		t.Errorf("state = %q", r.StreamingState)
	}
}

func TestParseSSEVideosExtracted(t *testing.T) {
	body := `data: {"data":{"sendMessageStream":{"videos":[{"url":"https://example.com/v.mp4"}]}}}`
	r := ParseSSE(body)
	if len(r.Videos) != 1 || r.Videos[0] != "https://example.com/v.mp4" {
		t.Errorf("videos = %v", r.Videos)
	}
}

func TestParseSSEConversationIDExtracted(t *testing.T) {
	body := `data: {"data":{"sendMessageStream":{"conversationId":"conv-abc"}}}`
	r := ParseSSE(body)
	if r.ConversationID != "conv-abc" {
		t.Errorf("conversationId = %q", r.ConversationID)
	}
}

func TestParseSSESkipsDoneSentinels(t *testing.T) {
	body := "data: [DONE]\ndata: null\ndata: "
	r := ParseSSE(body)
	if r.StatusCode != 200 {
		t.Errorf("status = %d", r.StatusCode)
	}
}

// === MEDIA EXTRACTION TESTS ===

func TestExtractMediaURLsFindsProgressiveURLs(t *testing.T) {
	data := map[string]any{
		"progressive_url": "https://example.com/video_720p.mp4",
	}
	urls := ExtractMediaURLs(data)
	if len(urls) != 1 {
		t.Fatalf("len = %d, want 1", len(urls))
	}
}

func TestIsMediaReadyEmptyReturnsFalse(t *testing.T) {
	if IsMediaReady(nil) {
		t.Error("nil should be false")
	}
}

func TestIsMediaReadyErrorReturnsFalse(t *testing.T) {
	if IsMediaReady(map[string]any{"error": "bad"}) {
		t.Error("error key should be false")
	}
}
