package metaai

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the resolved configuration for a Client. Built from functional
// options with environment-variable fallbacks.
type Config struct {
	// Auth
	Cookies     map[string]string // explicit cookies (highest priority)
	AccessToken string            // ecto1:… token (META_AI_ACCESS_TOKEN or scraped)
	FBEmail     string
	FBPassword  string
	ProxyURL    string // HTTP(S) proxy
	LoadDotEnv  bool   // load ./.env on construction (default true)
	DotEnvPath  string // override .env path (default "./.env")

	// Default chat config (constructor > environment > zero).
	DefaultMode     string
	DefaultThinking bool
	DefaultInstant  bool

	// SystemInstruction, when non-empty, is prepended to every chat/analyze
	// message as a "[System]\n...\n\n" prefix (mirrors the proxy's prompt
	// assembly). Per-call ChatOptions.SystemInstruction overrides this.
	// Set via WithSystemInstruction or META_AI_SYSTEM_INSTRUCTION.
	SystemInstruction string

	// HTTP
	HTTPTimeout time.Duration
	UserAgent   string

	// Logger receives debug output for best-effort operations (e.g.
	// markConversationSeen). nil resolves to a no-op logger.
	Logger Logger
}

// Option configures a Client.
type Option func(*Config)

// WithCookies sets the cookie map (highest auth priority).
func WithCookies(c map[string]string) Option {
	return func(cfg *Config) {
		cfg.Cookies = c
	}
}

// WithAccessToken sets the ecto1: access token explicitly.
func WithAccessToken(token string) Option {
	return func(cfg *Config) { cfg.AccessToken = token }
}

// WithProxy sets an HTTP(S) proxy URL.
func WithProxy(proxyURL string) Option {
	return func(cfg *Config) { cfg.ProxyURL = proxyURL }
}

// WithFBCredentials sets Facebook login credentials (cookie bootstrap path).
func WithFBCredentials(email, password string) Option {
	return func(cfg *Config) {
		cfg.FBEmail = email
		cfg.FBPassword = password
	}
}

// WithDefaultMode sets the default chat mode.
func WithDefaultMode(mode string) Option {
	return func(cfg *Config) { cfg.DefaultMode = mode }
}

// WithDefaultThinking enables thinking mode by default.
func WithDefaultThinking(b bool) Option {
	return func(cfg *Config) { cfg.DefaultThinking = b }
}

// WithDefaultInstant enables instant mode by default.
func WithDefaultInstant(b bool) Option {
	return func(cfg *Config) { cfg.DefaultInstant = b }
}


// WithSystemInstruction sets a global system instruction prepended to every
// chat/analyze message. Per-call ChatOptions.SystemInstruction takes priority.
func WithSystemInstruction(s string) Option {
	return func(cfg *Config) { cfg.SystemInstruction = s }
}

// WithHTTPTimeout sets the HTTP client timeout.
func WithHTTPTimeout(d time.Duration) Option {
	return func(cfg *Config) { cfg.HTTPTimeout = d }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(cfg *Config) { cfg.UserAgent = ua }
}

// WithLogger supplies a debug logger for best-effort operations (e.g.
// markConversationSeen failures). Adapts any structured logger via a one-method
// shim implementing Logger. Without this, the SDK uses a silent no-op logger.
func WithLogger(l Logger) Option {
	return func(cfg *Config) { cfg.Logger = l }
}

// WithDotEnv disables/enables .env loading and optionally sets its path.
func WithDotEnv(enable bool, path string) Option {
	return func(cfg *Config) {
		cfg.LoadDotEnv = enable
		if path != "" {
			cfg.DotEnvPath = path
		}
	}
}

// newConfig applies options and then resolves environment-variable fallbacks:
//   - cookies: explicit > env (META_AI_DATR required) > none
//   - access token: explicit > META_AI_ACCESS_TOKEN > (scraped lazily by caller)
//   - defaults: option > env > zero; thinking+instant mutual exclusion forces thinking off
func newConfig(opts []Option) *Config {
	cfg := &Config{
		LoadDotEnv:  true,
		DotEnvPath:  "./.env",
		HTTPTimeout: 60 * time.Second,
		UserAgent:   DefaultUserAgent,
	}
	for _, o := range opts {
		o(cfg)
	}

	// .env loading (errors are non-fatal — the file is optional).
	if cfg.LoadDotEnv {
		path := cfg.DotEnvPath
		if path == "" {
			path = "./.env"
		}
		_ = godotenv.Load(path)
	}

	// Cookies: explicit wins; otherwise try env.
	if len(cfg.Cookies) == 0 {
		if env := loadCookiesFromEnv(); env != nil {
			cfg.Cookies = env
		}
	}

	// Access token: explicit wins; otherwise META_AI_ACCESS_TOKEN.
	if cfg.AccessToken == "" {
		cfg.AccessToken = strings.TrimSpace(os.Getenv("META_AI_ACCESS_TOKEN"))
	}

	// Default chat config fallbacks (env), then mutual exclusion.
	if cfg.DefaultMode == "" {
		cfg.DefaultMode = strings.TrimSpace(os.Getenv("META_AI_DEFAULT_MODE"))
	}
	if !cfg.DefaultThinking {
		cfg.DefaultThinking = strings.EqualFold(strings.TrimSpace(os.Getenv("META_AI_DEFAULT_THINKING")), "true")
	}
	if !cfg.DefaultInstant {
		cfg.DefaultInstant = strings.EqualFold(strings.TrimSpace(os.Getenv("META_AI_DEFAULT_INSTANT")), "true")
	}
	if cfg.DefaultThinking && cfg.DefaultInstant {
		cfg.DefaultThinking = false // instant mode takes precedence
	}

	// System instruction: explicit wins; otherwise META_AI_SYSTEM_INSTRUCTION.
	if cfg.SystemInstruction == "" {
		cfg.SystemInstruction = strings.TrimSpace(os.Getenv("META_AI_SYSTEM_INSTRUCTION"))
	}

	return cfg
}

// httpClient builds the *http.Client honoring proxy + timeout options.
func (cfg *Config) httpClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.ProxyURL != "" {
		if u, err := url.Parse(cfg.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.HTTPTimeout,
	}
}
