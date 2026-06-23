// Command metaai-rest runs the REST API server over the Meta AI Go SDK. It
// mirrors the canonical SDK REST surface (/healthz, /upload, /image, /video,
// /video/extend, /video/async, /video/jobs/{job_id}, /chat).
//
// Credentials come from the environment (META_AI_*), the same ones used by the
// SDK and the livecheck command. Run:
//
//	go run ./cmd/metaai-rest -addr :8000
//
// Optional bearer token clients must send in Authorization or x-api-key:
//
//	export META_AI_REST_TOKEN=change-me
package main

import (
	"flag"
	"log"
	"os"

	"github.com/smart-studio/metaai-go"
	"github.com/smart-studio/metaai-go/rest"
)

func main() {
	addr := flag.String("addr", "", "listen address (default :8000, env META_AI_REST_ADDR)")
	token := flag.String("token", "", "bearer token clients must send (env META_AI_REST_TOKEN)")
	flag.Parse()

	client, err := metaai.NewClient()
	if err != nil {
		log.Fatalf("construct metaai client: %v", err)
	}

	a := *addr
	if a == "" {
		a = getenv("META_AI_REST_ADDR", ":8000")
	}
	t := *token
	if t == "" {
		t = os.Getenv("META_AI_REST_TOKEN")
	}

	srv := rest.New(rest.Config{Client: client, Token: t, Addr: a})
	log.Printf("metaai-rest listening on %s (auth=%v)", a, t != "")
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
