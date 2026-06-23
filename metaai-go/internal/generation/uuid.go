package generation

import "github.com/smart-studio/metaai-go/internal/uuid"

// newUUID returns a UUIDv4 string. Delegates to the shared internal/uuid
// package, which panics on CSPRNG failure rather than silently returning a
// collision-prone zero UUID.
func newUUID() string {
	return uuid.V4()
}
