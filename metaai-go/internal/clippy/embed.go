package clippy

import _ "embed"

//go:embed testdata/template_frame.b64
var embeddedTemplateB64 string

//go:embed testdata/attachment_frame.b64
var embeddedAttachmentTemplateB64 string

// DefaultTemplateB64 returns the embedded captured working frame (base64).
// The embedded frame was captured from a live meta.ai browser session on
// 2026-06-19 with text "TEMPLATEMARKER reply ok" on conversation
// 43427357-1e17-4493-9e22-4130be3c53da.
func DefaultTemplateB64() string { return embeddedTemplateB64 }

// DefaultAttachmentTemplateB64 returns the embedded captured working attachment frame (base64).
func DefaultAttachmentTemplateB64() string { return embeddedAttachmentTemplateB64 }

// DefaultTemplateText is the original text in the embedded template frame.
const DefaultTemplateText = "TEMPLATEMARKER reply ok"
