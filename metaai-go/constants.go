// Package metaai provides programmatic
// access to Meta AI (https://meta.ai) — chat (via the clippy WebSocket binary
// protocol), image/video generation, image upload/analysis, conversation history,
// and topic-based session management.
//
// The primary entry point is the Client type (see NewClient).
//
// Wire-format ground truth for the clippy protocol is documented in
// docs/protocol.md and derived from a live browser capture (2026-06-19).
package metaai

// Endpoint URLs observed in the live web client.
const (
	WebsocketURL    = "wss://gateway.meta.ai/ws/clippy"
	GraphqlURL      = "https://meta.ai/api/graphql"
	ChatGraphqlURL  = "https://www.meta.ai/api/graphql"
	MetaAIOrigin    = "https://meta.ai"
	MetaAIHomeURL   = "https://meta.ai"
	UploadURLFormat = "https://rupload.meta.ai/gen_ai_document_gen_ai_tenant/%s"
)

// DGW WebSocket query-string parameter values for the clippy connection.
// Captured from the live browser; must match exactly.
const (
	DGWAppID      = "1522763855472543"
	DGWAppVersion = "1.0.0"
	DGWAuthType   = "15:0"
	DGWVersion    = "5"
	DGWUUID       = "0"
	DGWTier       = "prod"
)

// DefaultUserAgent mirrors the browser UA used by the live site (Chrome 149).
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36"

// Chat modes accepted by prompt and chat requests.
const (
	ModeLearn       = "learn"
	ModeAnalyze     = "analyze"
	ModeCreateImage = "create_image"
	ModeCreateVideo = "create_video"
)

// Kadabra entry points observed in live captures.
const (
	EntryPointHome = "KADABRA__HOME__UNIFIED_INPUT_BAR"
	EntryPointChat = "KADABRA__CHAT__UNIFIED_INPUT_BAR"
)

// Default capability strings sent in the chat SEND frame (top.f18 repeated).
// Captured from the live browser.
var DefaultCapabilities = []string{
	"stocks",
	"weather",
	"meta_knowledge_search_carousel",
	"meta_catalog_search_carousel",
	"media_gallery",
}

// Capability hash sent in envelope.f19.f1. Captured from the live browser;
// opaque to this client — copied verbatim.
const CapabilityHash = "e2b88f9846379cbc26960fa3ae1d22201dfb19df7890ae6a3ac8a28870bac682"

// MaxRetries is the default retry limit for transient requests.
const MaxRetries = 3

// DocID holds persisted-query document IDs for the GraphQL HTTP endpoint.
// Supporting operations only (config, mark-seen, latency, history, media).
// Chat sendMessage itself goes over the WebSocket.
//
// Values are overridable via the META_AI_* env vars in EnvKey. Defaults match
// the live capture.
type DocID struct {
	Default string   // hardcoded default
	EnvKey  []string // env var names, tried in order
}

// DocIDs groups all persisted-query doc IDs used by the SDK.
var DocIDs = struct {
	ConversationRegistration string // e7f80258… (pre-send: register conversation id)
	ConversationConfig       string // 2b7fcdc8… (logInvitationImpression / conversationDepth / agentType)
	ModeSetter               string // c32bbe99… (pre-send: set conversation mode think_hard/fast)
	BootstrapA               string // 954e9b19… (variables:{})
	BootstrapB               string // 1a336150… (variables:{})
	BootstrapC               string // 659ce505… (variables:{})
	FetchConversationStatus  string // 9ec38c71… ({id, includeMessageList:false})
	MarkSeen                 string // 0b2cb3a4… (= META_AI_LAST_SEEN_DOC_ID)
	Latency                  string // 26999e5d… (= META_AI_LATENCY_DOC_ID)
	History                  string // e6214ae9… (META_AI_HISTORY_DOC_ID)
	FetchMedia               string // ecc43cc5… (META_AI_DOC_ID_FETCH_MEDIA)
	TextToImage              string // 2f707e4a… (image gen SSE)
	TextToVideo              string // 2f707e4a… (video gen SSE)
	ExtendVideo              string // 865d6fe8…
	FetchConversation        string // 9f7f4e20…
	PollMedia                string // 335a1ff1…
	MediaLibraryFeed         string // 3ee648d9… (media library feed query)
}{
	ConversationRegistration: "e7f802582dbfed8e181b012e010993eb",
	ConversationConfig:       "2b7fcdc841885a7263eccebaab5cfa3e",
	ModeSetter:               "c32bbe999c48e64e855dc63177d5153f",
	BootstrapA:               "954e9b193487fa4af750af87906e4313",
	BootstrapB:               "1a336150dbfcb081bfca357bac91b601",
	BootstrapC:               "659ce5056fb5c37f7311ab04d8e928f3",
	FetchConversationStatus:  "9ec38c71319524c705534312900bf6db",
	MarkSeen:                 "0b2cb3a499606dfe326807a2b8b7ee20",
	Latency:                  "26999e5d1366c257595b7fafa7822c31",
	History:                  "e6214ae98a2da65944658b5b49cc5f00",
	FetchMedia:               "ecc43cc5adc3443611ed22bd8608a371",
	TextToImage:              "2f707e4a86f4b01adba97e1376cbdc14",
	MediaLibraryFeed:         "3ee648d9976908559a9013a0b520322b",
	TextToVideo:              "2f707e4a86f4b01adba97e1376cbdc14",
	ExtendVideo:              "865d6fe804a7ea98fbce7e562b1d61ce",
	FetchConversation:        "9f7f4e20336400df0ea882b6131d2dd6",
	PollMedia:                "335a1ff137a82e22e0a9724d4bf70b6f",
}
