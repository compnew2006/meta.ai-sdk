# 🤖 SMART Studio — Meta AI Go API & React 19 Frontend

> **🌐 Languages:** [🇬🇧 English](README.md) · [🇪🇬 العربية](README.ar.md) (Looking for the Arabic version? Check [README.ar.md](README.ar.md))

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React Version](https://img.shields.io/badge/React-19.0.0-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/github/license/compnew2006/meta.ai-sdk)](https://github.com/compnew2006/meta.ai-sdk/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/compnew2006/meta.ai-sdk)](https://github.com/compnew2006/meta.ai-sdk)

<p align="center">
  <img src="docs/images/demo-multiplatform.png" alt="SMART Studio — Meta AI Web UI ↔ Telegram Hermes Bot" width="800">
  <br><em>Meta AI running locally, bridged to Telegram &amp; WhatsApp via <a href="https://github.com/nousresearch/hermes-agent">Hermes Agent</a></em>
</p>

**Unleash the Power of Meta AI with Go & React** 🚀

A modern, feature-rich web application combining a React 19 design tool (featuring 11 specialized studios) with a high-performance Go API server acting as a wrapper around Meta AI. All without API keys!

> [!WARNING]
> **Educational Purposes Only:** This project is created strictly for educational, research, and development purposes. It is an unofficial integration and should not be used in commercial production environments without respecting Meta's official guidelines and terms of service.

🎯 [Project Overview](#-project-overview) • 🏗️ [Architecture](#-architecture) • 📦 [Prerequisites](#-prerequisites) • 🔐 [Cookie Setup](#-getting-cookies-from-meta-ai-step-by-step) • 🚀 [Running Modes](#-running--3-modes) • 🔧 [REST API](#-rest-api-reference)

---

## 🔍 Project Overview

* **[smart-studio](file:///Users/noiemany/Downloads/meta.ai/smart-studio)**: A modern marketing and design interface (React 19) offering 11 specialized studios for branding, photography, video, voiceovers, campaigns, and market analysis.
* **[metaai-go](file:///Users/noiemany/Downloads/meta.ai/metaai-go)**: A high-performance Go API wrapper for Meta AI. It intercepts requests from SMART Studio and interacts directly with Meta AI using cookie authentication.

> [!NOTE]
> **Integration Details:**
> Every AI call in SMART Studio routes through a central service (`smart-studio/services/geminiService.ts` which re-exports from `aiService.ts`), communicating directly with the `metaai-go` REST API.

---

## 🌟 Core Capabilities

| Studio / Feature | Description | Engine | Status |
| :--- | :--- | :--- | :--- |
| 💬 **Intelligent Chat** | Powered by Muse Spark with real-time Bing search | Meta AI | ✅ Working |
| 🎨 **Creator Studio** | Generate product images in custom orientations | Meta AI | ✅ Working |
| 🎬 **Video Studio** | Generate cinematic videos from text/image references | Meta AI | ✅ Working |
| 🔍 **Image Analysis** | Describe, extract prompts, and audit brand aesthetics | Meta AI | ✅ Working |
| 📢 **Plan Studio** | Generate 9-post campaigns localized in Egyptian colloquial Arabic | Meta AI | ✅ Working |
| 🎨 **Branding Studio** | Generate logo variations and extract brand colors | Meta AI | ✅ Working |
| 🗣️ **Voice Over Studio** | Text-to-Speech (TTS) conversion | Gemini | 🔶 Requires API Key |

---

## 🏗️ Architecture & Flow

```
┌────────────────────────────────────────────────────────────────┐
│  Browser (SMART Studio React SPA)                              │
│                                                                │
│  components/*.tsx ──imports──▶ services/geminiService.ts        │
│    (11 studios)                   │                            │
│                                   ▼ (re-export shim)            │
│                          services/aiService.ts                  │
│                                   │                            │
│                                   ▼                            │
│                          services/metaaiClient.ts               │
│                       (single network layer: fetch + Bearer)    │
└────────────────────────────────────────┬───────────────────────┘
                                         │ HTTPS / same-origin
                                         ▼
┌────────────────────────────────────────────────────────────────┐
│  metaai-go REST Server  (Go binary)                            │
│                                                                │
│  /chat  /analyze  /upload  /image  /image/fetch  /video*       │
│      rest/handlers.go  rest/analyze_handler.go                 │
│                                                                │
│  + embedded SMART Studio SPA at /  (in prod build)             │
└────────────────────────────────────────┬───────────────────────┘
                                         │ WebSocket + GraphQL
                                         │ (cookies + access token)
                                         ▼
┌────────────────────────────────────────────────────────────────┐
│  Meta AI  (meta.ai)                                            │
│  - generated images on scontent-arn2-1.xx.fbcdn.net            │
│  - generated videos on video-*.xx.fbcdn.net                    │
└────────────────────────────────────────────────────────────────┘
```

---

## 📦 Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| **Go** | 1.21+ | Run and compile the `metaai-go` backend |
| **Node.js** | 18+ | Package installation and building for `smart-studio` |
| **npm** | 9+ | Package manager for `smart-studio` and Go UI dependencies |
| **Meta AI Account** | - | Free logged-in account at `meta.ai` |

---

## 🔐 Getting Cookies from Meta AI (Step-by-Step)

To authenticate with Meta AI without API keys, you must extract cookies from your logged-in browser session.

1. Go to [meta.ai](https://www.meta.ai) in your browser and sign in.
2. Open DevTools (**F12** or right-click → Inspect).
3. Navigate to the **Application** tab → **Storage** → **Cookies** → `https://www.meta.ai`.
4. Copy values for the following cookies:
   * `datr` (Required - long-lived device token)
   * `ecto_1_sess` (Required - session token, expires periodically)
   * `rd_challenge` (Recommended - bypasses regional verification blocks)
   * `abra_sess` (Optional - improves compatibility in some regions)

---

## 🛠️ Configuration (`.env`)

### Backend Configuration (`metaai-go/.env`)
Create a `.env` file inside the `metaai-go` directory:
```env
META_AI_DATR=your_datr_cookie_here
META_AI_ECTO_1_SESS=your_ecto_1_sess_cookie_here

# Recommended Config
META_AI_RD_CHALLENGE=your_rd_challenge_cookie
META_AI_DPR=1
META_AI_WD=1837x1240
META_AI_PS_L=1
META_AI_PS_N=1

# Optional Server Config
META_AI_REST_ADDR=:8000
META_AI_REST_TOKEN=smart-studio-dev-token
META_AI_CORS_ORIGIN=http://localhost:3000
```

### Frontend Configuration (`smart-studio/.env`)
Create a `.env` file inside the `smart-studio` directory:
```env
VITE_METAAI_URL=http://localhost:8000
VITE_METAAI_TOKEN=smart-studio-dev-token
GEMINI_API_KEY=your_gemini_api_key_here  # Only required for Voice Over Studio (TTS)
```

---

## 🚀 Running — 3 Modes

### 1. Dev Mode (Recommended for development)
Starts Vite HMR dev server for the frontend, and standard API server for the backend.
```bash
# Start backend on :8000
cd metaai-go
make run-rest

# In a new terminal: Start frontend on :3000
cd smart-studio
npm install
npm run dev
```

Alternatively, if you have `air` installed, run both with a single command:
```bash
cd metaai-go
make run-dev
```

### 2. Single Binary Production Build
Combines SMART Studio frontend and Go backend into a single executable binary.
```bash
cd metaai-go
make build-prod

# Run the unified server (accessible at http://localhost:8000)
META_AI_REST_TOKEN=smart-studio-dev-token ./bin/metaai-rest
```

### 3. Background Run (For remote VMs/Servers)
```bash
cd metaai-go
go build -o /tmp/metaai-rest ./cmd/metaai-rest
nohup /tmp/metaai-rest > /tmp/metaai-rest.log 2>&1 &

# Follow logs
tail -f /tmp/metaai-rest.log
```

---

## 🔧 REST API Reference

All requests must include the REST Token as a Bearer Auth header:
```text
Authorization: Bearer <META_AI_REST_TOKEN>
```

| Endpoint | Method | Description | Status |
| :--- | :--- | :--- | :--- |
| `/healthz` | GET | Health check (No Auth) | ✅ Working |
| `/chat` | POST | Send messages to Muse Spark | ✅ Working |
| `/upload` | POST | Upload reference images (multipart form-data) | ✅ Working |
| `/analyze` | POST | Analyze uploaded images | ✅ Working |
| `/image` | POST | Generate images from prompts | ✅ Working |
| `/image/fetch` | GET | Fetch fbcdn image URL and convert to Base64 | ✅ Working |
| `/video/async` | POST | Start asynchronous video generation | ✅ Working |
| `/video/jobs/{id}` | GET | Poll status of asynchronous video job | ✅ Working |

### Example: Generate Image (cURL)
```bash
curl -X POST http://localhost:8000/image \
  -H "Authorization: Bearer smart-studio-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Cyberpunk cityscape at night", "orientation": "LANDSCAPE"}'
```

---

## 🌟 Project Structure

```text
meta.ai-sdk/
│
├── 📁 smart-studio/           # React 19 Frontend Web App
│   ├── 📁 components/         # 11 Marketing Studios UI
│   ├── 📁 services/           # AI Client and routing services
│   └── 📄 package.json        # Dependencies & Scripts
│
├── 📁 metaai-go/              # Go Backend API Wrapper
│   ├── 📁 cmd/                # Entrypoints (metaai-rest, server)
│   ├── 📁 rest/               # REST API Router & Handler logic
│   ├── 📁 ui/                 # Embedded administration console
│   ├── 📄 Makefile            # Build pipeline commands
│   └── 📄 go.mod              # Go modules dependency declaration
│
├── 📄 README.md               # English documentation
└── 📄 README.ar.md            # Arabic documentation
```

---

## 📜 License & Disclaimer

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

### ⚖️ Disclaimer
This project is an independent implementation and is not officially affiliated with Meta Platforms, Inc. or any of its affiliates.
* ✅ Educational and development purposes
* ✅ Use responsibly and ethically
* ✅ Comply with Meta's Terms of Service
* ✅ Respect usage limits and policies

**Muse Spark License:** Visit [llama.com/muse-spark](https://llama.com/muse-spark/license) for Muse Spark usage terms.
