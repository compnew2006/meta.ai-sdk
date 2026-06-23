package proxy

// server.go implements the OpenAI/Anthropic-compatible HTTP proxy: the Server,
// routing, optional bearer auth, and shared response helpers.

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/smart-studio/metaai-go"
)

// Config configures a proxy Server.
type Config struct {
	Client     *metaai.Client // required: the Meta AI client to proxy through
	Token      string         // optional bearer token clients must present
	Addr       string         // listen address (default ":8787")
	HTTPClient *http.Client   // optional, used to download remote image URLs
}

// Server is an OpenAI- and Anthropic-compatible HTTP proxy over Meta AI.
type Server struct {
	client chatClient
	token  string
	addr   string
	http   *http.Client

	// chatMu serializes chat turns. The underlying SDK reuses a single
	// WebSocket connection per conversation, so concurrent turns on one Client
	// would interleave clippy frames; serialize to keep the protocol sane.
	chatMu sync.Mutex
}

// New constructs a Server from Config.
func New(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8787"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Server{
		client: cfg.Client,
		token:  strings.TrimSpace(cfg.Token),
		addr:   cfg.Addr,
		http:   cfg.HTTPClient,
	}
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.addr }

// Handler returns the mux used by the server (mountable on any ServeMux).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/models", s.auth(s.handleModels))
	mux.HandleFunc("/v1/chat/completions", s.auth(s.handleChatCompletions))
	mux.HandleFunc("/chat/completions", s.auth(s.handleChatCompletions))
	mux.HandleFunc("/v1/messages", s.auth(s.handleMessages))
	mux.HandleFunc("/messages", s.auth(s.handleMessages))
	return mux
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
			writeError(w, http.StatusUnauthorized, "authentication_error", "missing or invalid API key")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not_found", "unknown endpoint")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":      "metaai-proxy",
		"endpoints": []string{"/v1/chat/completions", "/v1/messages", "/v1/models", "/health"},
		"models":    modelNames(),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, modelsResponse())
}

// bearerToken extracts a bearer/api key from the request, checking both the
// OpenAI-style Authorization header and the Anthropic-style x-api-key header.
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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiErrorBody is shaped to satisfy both OpenAI ({error:{type,message}}) and
// Anthropic ({type:"error",error:{type,message}}) error readers.
type apiErrorBody struct {
	Error apiErrorDetail `json:"error"`
	Type  string         `json:"type,omitempty"`
}

type apiErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, errType, message string) {
	writeJSON(w, status, apiErrorBody{
		Error: apiErrorDetail{Type: errType, Message: message},
		Type:  "error",
	})
}
