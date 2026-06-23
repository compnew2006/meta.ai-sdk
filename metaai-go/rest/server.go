package rest

// server.go implements the REST API server: routing, optional bearer auth,
// the shared client interface, and response helpers. Handlers live in
// handlers.go.

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/smart-studio/metaai-go"
	"github.com/smart-studio/metaai-go/internal/upload"
)

// restClient is the subset of *metaai.Client the REST server uses. Declared as
// an interface so unit tests can inject a fake client.
type restClient interface {
	Chat(ctx context.Context, message string, opts *metaai.ChatOptions) (string, error)
	StreamChat(ctx context.Context, message string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk
	LastConversationID() string
	UploadImage(ctx context.Context, path string) (*upload.Result, error)
	AnalyzeImage(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) (string, error)
	AnalyzeImageStream(ctx context.Context, imagePath, mediaID, question string, opts *metaai.ChatOptions) <-chan metaai.ChatChunk
	GenerateImage(ctx context.Context, prompt, orientation string, numImages int) (*metaai.GenerationResult, error)
	GenerateVideo(ctx context.Context, prompt string) (*metaai.GenerationResult, error)
	ExtendVideo(ctx context.Context, sourceMediaID string) (*metaai.GenerationResult, error)
}

// Config configures a REST Server.
type Config struct {
	Client *metaai.Client // required
	Token  string         // optional bearer clients must present
	Addr   string         // listen address (default ":8000")
	// CORSOrigin is the value of Access-Control-Allow-Origin for browser
	// callers. Defaults to the META_AI_CORS_ORIGIN env var, then
	// "http://localhost:3000". Empty string disables CORS (e.g. when serving
	// the SPA same-origin in production).
	CORSOrigin string
}

// Server is a REST API server over the Meta AI SDK, mirroring the canonical
// SDK REST surface (/healthz, /upload, /image, /image/fetch, /video,
// /video/extend, /video/async, /video/jobs/{job_id}, /chat).
type Server struct {
	client     restClient
	token      string
	addr       string
	cors       string
	sameOrigin bool // true when Jenta SPA is embedded → CORS becomes a no-op
	jobs       *jobStore
	imageHTTP  *http.Client // used by /image/fetch to pull fbcdn URLs server-side

	// chatMu serializes chat turns (the SDK shares one WS per conversation).
	chatMu sync.Mutex
}

// New constructs a REST Server from Config.
func New(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8000"
	}
	cors := strings.TrimSpace(cfg.CORSOrigin)
	if cors == "" {
		cors = strings.TrimSpace(os.Getenv("META_AI_CORS_ORIGIN"))
	}
	if cors == "" {
		cors = "http://localhost:3000"
	}
	return &Server{
		client:    cfg.Client,
		token:     strings.TrimSpace(cfg.Token),
		addr:      cfg.Addr,
		cors:      cors,
		jobs:      newJobStore(),
		imageHTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.addr }

// Handler returns the HTTP mux (mountable on any ServeMux).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/chat", s.auth(s.handleChat))
	mux.HandleFunc("/upload", s.auth(s.handleUpload))
	mux.HandleFunc("/analyze", s.auth(s.handleAnalyze))
	mux.HandleFunc("/image", s.auth(s.handleImage))
	mux.HandleFunc("/image/fetch", s.auth(s.handleImageFetch))
	mux.HandleFunc("/video", s.auth(s.handleVideo))
	mux.HandleFunc("/video/extend", s.auth(s.handleVideoExtend))
	mux.HandleFunc("/video/async", s.auth(s.handleVideoAsync))
	mux.HandleFunc("/video/jobs/", s.auth(s.handleVideoJob))

	// SMART Studio SPA: in production (`make build-prod`) the React build is
	// embedded under jenta/dist and served at "/" same-origin (no CORS). In dev
	// only a dev.txt marker is present, so we fall back to the JSON service index.
	if jentaHasIndex() {
		if subFS, err := fs.Sub(metaai.JentaDistFS, "jenta/dist"); err == nil {
			mux.Handle("/", http.FileServer(spaFileSystem{fs: http.FS(subFS)}))
		}
		// SPA is served same-origin in prod → CORS is unnecessary and noisy.
		// Flag it so withCORS becomes a no-op, without mutating s.cors (which
		// callers may set explicitly for testing).
		s.sameOrigin = true
	} else {
		mux.HandleFunc("/", s.handleIndex)
	}

	return s.withCORS(s.logRequest(mux))
}

// jentaHasIndex reports whether the embedded Jenta build actually contains an
// index.html (i.e. `make build-jenta` ran). When false we serve the API JSON
// index instead, so the dev workflow (Vite at :3000 + this API at :8000) keeps
// working without a Jenta build.
func jentaHasIndex() bool {
	data, err := metaai.JentaDistFS.ReadFile("jenta/dist/index.html")
	return err == nil && len(data) > 0
}

// ListenAndServe starts the HTTP server (blocking).
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.token != "" && bearerToken(r) != s.token {
			writeError(w, http.StatusUnauthorized, "missing or invalid API key")
			return
		}
		next(w, r)
	}
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	for _, k := range []string{"X-Api-Key", "X-API-Key", "x-api-key"} {
		if v := r.Header.Get(k); v != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error":   message,
	})
}

type spaFileSystem struct {
	fs http.FileSystem
}

func (sfs spaFileSystem) Open(name string) (http.File, error) {
	f, err := sfs.fs.Open(name)
	if err == nil {
		return f, nil
	}
	fIndex, errIndex := sfs.fs.Open("/index.html")
	if errIndex == nil {
		return fIndex, nil
	}
	return nil, err
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(srw, r)
		duration := time.Since(start)

		log.Printf("[rest] %s %s - %s %s - %d %s (%v, %d bytes) UA=%q",
			r.RemoteAddr,
			r.Proto,
			r.Method,
			r.RequestURI,
			srw.statusCode,
			http.StatusText(srw.statusCode),
			duration,
			srw.bytesWritten,
			r.UserAgent(),
		)
	})
}

// withCORS wraps next with permissive CORS for the configured origin. CORS is
// skipped entirely when s.cors is empty OR s.sameOrigin is set (i.e. the Jenta
// SPA is embedded and served from the same origin, so cross-origin requests
// never happen). Preflight OPTIONS are answered with 204 and short-circuited.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cors == "" || s.sameOrigin {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", s.cors)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Api-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
