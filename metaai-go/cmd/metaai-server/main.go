// Command metaai-server runs an OpenAI/Anthropic-compatible HTTP proxy over
// the Meta AI Go SDK. It lets tools like Claude Code (Anthropic /v1/messages)
// and any OpenAI-compatible client (/v1/chat/completions) use Meta AI as a
// backend.
//
// Credentials come from the environment (META_AI_*), the same ones used by the
// SDK and the livecheck command. Run:
//
//	go run ./cmd/metaai-server -addr :8787
//
// Then point a client at it, e.g. for Claude Code:
//
//	export ANTHROPIC_BASE_URL=http://localhost:8787
//	export ANTHROPIC_API_KEY=$META_AI_PROXY_TOKEN
package main

import (
	"flag"
	"log"
	"os"

	"github.com/smart-studio/metaai-go"
	"github.com/smart-studio/metaai-go/proxy"
)

func main() {
	addr := flag.String("addr", "", "listen address (default :8787, env META_AI_PROXY_ADDR)")
	token := flag.String("token", "", "bearer token clients must send (env META_AI_PROXY_TOKEN)")
	flag.Parse()

	client, err := metaai.NewClient()
	if err != nil {
		log.Fatalf("construct metaai client: %v", err)
	}

	a := *addr
	if a == "" {
		a = getenv("META_AI_PROXY_ADDR", ":8787")
	}
	t := *token
	if t == "" {
		t = os.Getenv("META_AI_PROXY_TOKEN")
	}

	srv := proxy.New(proxy.Config{Client: client, Token: t, Addr: a})
	log.Printf("metaai-proxy listening on %s (auth=%v)", a, t != "")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getenv(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}
