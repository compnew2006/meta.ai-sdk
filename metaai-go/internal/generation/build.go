package generation

import (
	"fmt"
)

// BuildVariablesOptions are the inputs to BuildBaseVariables.
type BuildVariablesOptions struct {
	Prompt          string
	Operation       Operation
	ContentPrefix   string // "" → use Prompt as-is; else "{prefix} {prompt}"
	Thinking        bool
	Instant         bool
	Mode            *string // nil → "create" (or nil for extend)
	ConversationID  string  // "" → generated
	PromptSessionID string  // "" → generated
	MediaIDs        []string
	NumImages       int // image-to-image numMedia (default 4)
	ExtendSourceID  string
	ExtendSourceURL string
	RequestID       *string
}

// BuildBaseVariables produces the
// "variables" JSON for a generation GraphQL request. The imagineOperationRequest
// shape depends on Operation + MediaIDs (image-to-image / image-to-video /
// extend-video / text-to-*).
func BuildBaseVariables(opts BuildVariablesOptions) (map[string]any, error) {
	if opts.Prompt == "" {
		return nil, fmt.Errorf("generation: Prompt is required")
	}
	op := opts.Operation
	isExtend := op == OpExtendVideo
	isImg2Img := op == OpTextToImage && len(opts.MediaIDs) > 0
	isImg2Vid := op == OpTextToVideo && len(opts.MediaIDs) > 0

	content := opts.Prompt
	if opts.ContentPrefix != "" {
		content = opts.ContentPrefix + " " + opts.Prompt
	}

	conversationID := opts.ConversationID
	if conversationID == "" {
		conversationID = newUUID()
	}
	promptSessionID := opts.PromptSessionID
	if promptSessionID == "" {
		promptSessionID = newUUID()
	}

	var attachmentsV2 []string
	for _, m := range opts.MediaIDs {
		attachmentsV2 = append(attachmentsV2, m)
	}

	// imagineOperationRequest
	var imagine map[string]any
	switch {
	case isImg2Img:
		num := opts.NumImages
		if num <= 0 {
			num = 4
		}
		imagine = map[string]any{
			"operation": string(OpImageToImage),
			"imageToImageParams": map[string]any{
				"sourceMediaEntId": opts.MediaIDs[0],
				"instruction":      opts.Prompt,
				"imageSource":      "USER_UPLOADED",
				"imageUploadType":  "GENAI_UPLOADED_FILE",
				"mediaType":        "UPLOADED_IMAGE",
				"numMedia":         num,
			},
		}
	case isImg2Vid:
		imagine = map[string]any{
			"operation": string(OpImageToVideo),
			"imageToVideoParams": map[string]any{
				"sourceMediaEntId": opts.MediaIDs[0],
				"prompt":           opts.Prompt,
				"numMedia":         1,
			},
		}
	case isExtend:
		if opts.ExtendSourceID == "" {
			return nil, fmt.Errorf("generation: ExtendSourceID is required for EXTEND_VIDEO")
		}
		imagine = map[string]any{
			"operation": string(OpExtendVideo),
			"extendVideoParams": map[string]any{
				"sourceMediaEntId": opts.ExtendSourceID,
				"sourceMediaUrl":   opts.ExtendSourceURL,
				"numMedia":         1,
			},
		}
	default:
		imagine = map[string]any{
			"operation":         string(op),
			"textToImageParams": map[string]any{"prompt": opts.Prompt},
			"requestId":         nil, // overridden below if set
		}
	}
	if _, ok := imagine["requestId"]; !ok {
		if opts.RequestID != nil {
			imagine["requestId"] = *opts.RequestID
		} else {
			imagine["requestId"] = nil
		}
	}

	// effective mode
	mode := opts.Mode
	if mode == nil {
		if isExtend {
			mode = nil
		} else {
			m := "create"
			mode = &m
		}
	}

	// developer overrides
	devOverrides := map[string]any{}
	if opts.Thinking {
		devOverrides["forceThinkingEnabled"] = true
	}
	if opts.Instant {
		devOverrides["skipThinking"] = true
	}
	var devAny any
	if len(devOverrides) > 0 {
		devAny = devOverrides
	}

	entryPoint := "KADABRA__UNKNOWN"
	branchPath := "0"
	if isExtend {
		entryPoint = "KADABRA__IMAGINE_UNIFIED_CANVAS"
		branchPath = "2"
	}

	return map[string]any{
		"conversationId":               conversationID,
		"content":                      content,
		"userMessageId":                newUUID(),
		"assistantMessageId":           newUUID(),
		"userUniqueMessageId":          uniqueID(),
		"turnId":                       newUUID(),
		"mode":                         derefString(mode),
		"attachments":                  nil,
		"attachmentsV2":                attachmentsV2,
		"mentions":                     nil,
		"clippyIp":                     nil,
		"isNewConversation":            true,
		"imagineOperationRequest":      imagine,
		"qplJoinId":                    nil,
		"clientTimezone":               "UTC",
		"developerOverridesForMessage": devAny,
		"clientLatitude":               nil,
		"clientLongitude":              nil,
		"devicePixelRatio":             1.25,
		"entryPoint":                   entryPoint,
		"promptSessionId":              promptSessionID,
		"promptType":                   nil,
		"conversationStarterId":        nil,
		"userAgent":                    DefaultUserAgent,
		"currentBranchPath":            branchPath,
		"promptEditType":               "new_message",
		"userLocale":                   "en-US",
		"userEventId":                  nil,
		"requestedToolCall":            nil,
	}, nil
}

func derefString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}
