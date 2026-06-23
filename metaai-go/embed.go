package metaai

import "embed"

// JentaDistFS holds the SMART Studio React production build (jenta/dist). In
// dev this only contains a dev.txt marker; `make build-prod` (which runs
// `make build-jenta`) populates it with the real Vite build so a single Go
// binary serves the SPA + the API on the same origin (no CORS in prod).
//
// The directory is named `jenta/dist` for historical reasons (the project was
// renamed Jenta → SMART Studio); the user-facing brand is SMART Studio.
//
//go:embed jenta/dist
var JentaDistFS embed.FS
