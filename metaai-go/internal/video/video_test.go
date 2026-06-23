package video

import "testing"

// These tests lock the SSE parsing, media extraction, and document-ID contracts.

func TestDocIDs(t *testing.T) {
	if VideoCardDocID != "666834feb70769370072c294c87ebd23" {
		t.Errorf("VideoCardDocID = %q", VideoCardDocID)
	}
	if VideoGenerateDocID != "a3d873304cb1411ba7f056e47060ad1d" {
		t.Errorf("VideoGenerateDocID = %q", VideoGenerateDocID)
	}
	if VideoFetchDocID != "10b7bd5aa8b7537e573e49d701a5b21b" {
		t.Errorf("VideoFetchDocID = %q", VideoFetchDocID)
	}
}

func TestParseSSEResponseValid(t *testing.T) {
	body := "data: {\"type\":\"subscription_start\"}\n" +
		"data: {\"type\":\"message\",\"data\":{\"xfb_kadabra_send_message\":{}}}\n" +
		"data: {\"type\":\"done\"}\n"
	evs := ParseSSEResponse(body)
	if len(evs) != 3 {
		t.Fatalf("len = %d, want 3", len(evs))
	}
	if evs[0]["type"] != "subscription_start" {
		t.Errorf("ev[0] type = %v", evs[0]["type"])
	}
	if evs[1]["type"] != "message" {
		t.Errorf("ev[1] type = %v", evs[1]["type"])
	}
}

func TestParseSSEResponseEmpty(t *testing.T) {
	if evs := ParseSSEResponse(""); len(evs) != 0 {
		t.Errorf("empty → %d events, want 0", len(evs))
	}
}

func TestParseSSEResponseInvalidJSON(t *testing.T) {
	body := "data: {invalid json}\ndata: {\"valid\": \"json\"}"
	evs := ParseSSEResponse(body)
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1 (invalid skipped)", len(evs))
	}
	if evs[0]["valid"] != "json" {
		t.Errorf("ev[0] = %v", evs[0])
	}
}

func TestExtractMediaIDs(t *testing.T) {
	resp := map[string]any{
		"data": map[string]any{
			"xfb_kadabra_send_message": map[string]any{
				"messages": map[string]any{
					"edges": []any{
						map[string]any{
							"node": map[string]any{
								"content": map[string]any{
									"imagine_video": map[string]any{
										"videos": map[string]any{
											"nodes": []any{
												map[string]any{"mediaId": "917535734784048"},
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
	ids := ExtractMediaIDs(resp)
	if len(ids) != 1 || ids[0] != "917535734784048" {
		t.Errorf("ids = %v", ids)
	}
}

func TestExtractVideoURLs(t *testing.T) {
	resp := map[string]any{
		"data": map[string]any{
			"xfb_kadabra_send_message": map[string]any{
				"messages": map[string]any{
					"edges": []any{
						map[string]any{
							"node": map[string]any{
								"content": map[string]any{
									"imagine_video": map[string]any{
										"videos": map[string]any{
											"nodes": []any{
												map[string]any{
													"video_url": "https://example.com/video1.mp4",
													"videoDeliveryResponseResult": map[string]any{
														"progressive_urls": []any{
															map[string]any{"progressive_url": "https://example.com/video1_360p.mp4"},
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
				},
			},
		},
	}
	urls := ExtractVideoURLs(resp)
	if len(urls) < 1 {
		t.Fatalf("urls = %v", urls)
	}
	found := false
	for _, u := range urls {
		if u == "https://example.com/video1.mp4" {
			found = true
		}
	}
	if !found {
		t.Errorf("video1.mp4 missing from %v", urls)
	}
}
