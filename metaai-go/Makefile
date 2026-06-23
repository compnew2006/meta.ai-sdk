.PHONY: all build clean fmt test verify vet run run-rest run-hot run-dev build-jenta build-prod run-migrate

all: verify

fmt:
	test -z "$$(gofmt -l .)"

test: build-ui
	go test ./... -race -coverprofile=coverage.out

vet: build-ui
	go vet ./...

build-ui:
	cd ui && npm install && npm run build

# build-jenta builds the Jenta React app (../Jenta) and copies its dist into
# ./jenta/dist so it can be served same-origin from this server in production,
# eliminating CORS. Mirrors build-ui. Optional: only needed for prod deploys;
# the dev workflow uses Vite at :3000 + the CORS middleware.
build-jenta:
	cd ../Jenta && npm install && npm run build
	mkdir -p jenta/dist
	cp -r ../Jenta/dist/. jenta/dist/

build: build-ui
	go build ./...

# build-prod produces ONE self-contained binary that serves the Jenta React
# frontend AND the metaai-go REST API on the same origin (no CORS in prod).
# Pipeline: build Jenta → copy into embed dir → embed into the Go binary.
# Output: ./bin/metaai-rest (static assets baked in).
build-prod: build-jenta
	mkdir -p bin
	go build -o bin/metaai-rest ./cmd/metaai-rest
	@echo ""
	@echo "✓ Built bin/metaai-rest (SMART Studio SPA + REST API, single binary)"
	@echo "  Run with: META_AI_REST_TOKEN=... ./bin/metaai-rest"

# run-migrate rebuilds Jenta's dist into the embed dir without recompiling Go,
# useful after a frontend-only change while iterating on a running prod binary.
# (Re-run build-prod to bake the new assets into a fresh binary.)
run-migrate: build-jenta
	@echo "✓ Jenta dist refreshed at jenta/dist (run 'make build-prod' to embed into binary)"

# run boots the OpenAI/Anthropic-compatible API proxy server.
run:
	go run ./cmd/metaai-server

# run-rest boots the REST API server (/chat, /image, /video, ...).
run-rest:
	go run ./cmd/metaai-rest

# run-hot starts the REST API with air hot-reload: every .go or .env change
# rebuilds + restarts the server automatically (like Vite HMR but for Go).
# Requires `go install github.com/air-verse/air@latest`.
run-hot:
	$$(go env GOPATH)/bin/air

# run-dev starts BOTH the hot-reloaded Go API AND the Jenta Vite dev server,
# so a single command gives you the full dev loop with auto-reload on both
# sides. Ctrl-C kills both.
run-dev:
	@echo "Starting Go API (air hot-reload on :8000) + Jenta Vite (HMR on :3000)..."
	@trap 'kill 0' INT; \
	( $$(go env GOPATH)/bin/air ) & api_pid=$$!; \
	( cd ../Jenta && npm run dev ) & vite_pid=$$!; \
	wait

# run-token boots the proxy with a bearer token clients must present.
run-token:
	go run ./cmd/metaai-server -token "$$META_AI_PROXY_TOKEN"

verify: fmt test vet build

clean:
	go clean ./...
	rm -f coverage.out coverage.html
	rm -rf ui/dist bin
	rm -f jenta/dist/index.html jenta/dist/assets -r 2>/dev/null || true

