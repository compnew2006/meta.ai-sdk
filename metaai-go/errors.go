package metaai

import "errors"

// Sentinel errors shared by the public client and its transport layers.
var (
	// ErrInvalidCredentials is raised when Facebook login fails.
	ErrInvalidCredentials = errors.New("metaai: invalid facebook credentials")

	// ErrRegionBlocked is raised when Meta returns a region/challenge block.
	ErrRegionBlocked = errors.New("metaai: region blocked or challenge required")

	// ErrNotAuthenticated is returned when an operation requires auth but the
	// client has no cookies / access token.
	ErrNotAuthenticated = errors.New("metaai: not authenticated (no cookies or access token)")

	// ErrGraphQLValidation is raised when Meta returns a persisted-query
	// GRAPHQL_VALIDATION_FAILED error (doc_id drift).
	ErrGraphQLValidation = errors.New("metaai: graphql validation failed (doc_id may be stale)")

	// ErrAccessTokenMissing is returned when chat/upload needs an ecto1: token
	// but none is configured.
	ErrAccessTokenMissing = errors.New("metaai: missing ecto1 access token")

	// ErrInvalidMode is returned when prompt/chat receives a mode outside the
	// allowed set.
	ErrInvalidMode = errors.New("metaai: invalid mode (want learn|analyze|create_image|create_video)")

	// ErrConflictingModes is returned when both thinking and instant are true.
	ErrConflictingModes = errors.New("metaai: thinking and instant are mutually exclusive")
)
