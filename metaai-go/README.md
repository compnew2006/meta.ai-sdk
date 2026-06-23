# 🤖 MetaAI Go SDK

> **Chat with Meta AI, analyze images, and generate content from Go.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)]()
[![Status](https://img.shields.io/badge/Status-All%20Features%20Working-238636)]()
[![Tests](https://img.shields.io/badge/Tests-179%20passing-blue)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()

---

## 🔥 ما الذي يقدمه المشروع؟

مكتبة Go مستقلة للتعامل مع Meta AI عبر HTTP وGraphQL وبروتوكول Clippy
الثنائي على WebSocket. يعتمد تنفيذ المحادثة على إطارات ملتقطة من المتصفح
ومثبتة باختبارات golden لضمان صحة البنية والـ subtype.

---

## ✅ الميزات

| الميزة | الحالة | الوصف |
|---------|------|--------|
| 💬 **Chat** | ✅ يعمل | محادثة نصية عبر WebSocket (template mode) |
| 📡 **Streaming** | ✅ يعمل | بث الردود عبر Go channels |
| 📤 **رفع صور** | ✅ يعمل | رفع لـ rupload.meta.ai (مع lowercase headers) |
| 🔍 **تحليل صور** | ✅ يعمل | رفع + سؤال → الذكاء الاصطناعي **يرى الصورة** |
| 🎨 **توليد صور** | ✅ يعمل | نص → صورة عبر بث الشات + استطلاع mediaLibraryFeed |
| 🎬 **توليد فيديو** | ✅ يعمل | نص → فيديو عبر بث الشات + استطلاع mediaLibraryFeed |
| 📜 **السجل** | ✅ يعمل | استرجاع المحادثات السابقة |
| 🗂️ **المواضيع** | ✅ يعمل | إدارة محادثات متعددة |

---

## 📦 التثبيت

```bash
go get github.com/smart-studio/metaai-go
```

## 🔑 الإعداد

```bash
# متغيرات البيئة
export META_AI_DATR=your_datr_cookie
export META_AI_ECTO_1_SESS=your_session_cookie
export META_AI_ACCESS_TOKEN=ecto1:your_oauth_token
```

<details>
<summary>📖 كيفية الحصول على الـ cookies</summary>

1. سجل الدخول إلى [meta.ai](https://meta.ai)
2. افتح DevTools (F12) → Application → Cookies → `https://meta.ai`
3. انسخ قيم `datr` و `ecto_1_sess`
4. للحصول على التوكن، افتح Console واكتب:
```js
document.documentElement.outerHTML.match(/ecto1:[A-Za-z0-9_-]{20,}/)[0]
```

</details>

---

## 🚀 البداية السريعة

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/smart-studio/metaai-go"
)

func main() {
    client, err := metaai.NewClient()
    if err != nil { log.Fatal(err) }
    ctx := context.Background()

    // ═══ 💬 Chat ═══
    reply, _ := client.Chat(ctx, "What is 2+2?", nil)
    fmt.Println(reply)  // "2 + 2 = 4"

    // ═══ 📤 رفع + 🔍 تحليل صورة ═══
    analysis, _ := client.AnalyzeImage(ctx, "photo.png", "", "What's in this image?", nil)
    fmt.Println(analysis)  // AI تصف الصورة بالتفصيل

    // ═══ 🎨 توليد صورة ═══
    img, _ := client.GenerateImage(ctx, "a sunset over mountains", "LANDSCAPE", 1)
    fmt.Println(img.Status)  // "READY"

    // ═══ 📡 بث متدفق ═══
    for chunk := range client.StreamChat(ctx, "Write a poem about Go", nil) {
        if chunk.Err != nil { break }
        fmt.Print(chunk.Text)  // طباعة لحظية
    }
}
```

<details>
<summary>🔧 خيارات متقدمة</summary>

```go
// خيارات العميل
client, _ := metaai.NewClient(
    metaai.WithCookies(map[string]string{"datr": "...", "ecto_1_sess": "..."}),
    metaai.WithAccessToken("ecto1:..."),
    metaai.WithProxy("http://proxy:8080"),
    metaai.WithDefaultThinking(true),
    metaai.WithDefaultMode("analyze"),
)

// خيارات المحادثة
opts := &metaai.ChatOptions{
    Topic:           "coding",      // محادثة في موضوع محدد
    NewConversation: true,           // محادثة جديدة
    Thinking:        boolPtr(true),  // وضع التفكير
    Mode:            strPtr("analyze"),
}

// المواضيع
client.NewTopic("math", boolPtr(true), nil, nil)
client.SetTopic("math")
topics := client.ListTopics()

// السجل
history, _ := client.GetConversationHistory(ctx, 20, 0)
```

</details>

---

## 🌐 API Proxy — OpenAI & Anthropic compatibility

The `cmd/metaai-server` binary exposes the SDK as an HTTP server that speaks
**both** the OpenAI Chat Completions API and the Anthropic Messages API, so
agents like **Claude Code** (and any OpenAI-compatible tool) can use Meta AI as
a backend.

<table>
<tr>
<td width="50%" align="center">

### 🔌 **OpenAI Compatible**

`/v1/chat/completions` • `/v1/models`

Drop-in for any OpenAI SDK / LangChain / Cursor

</td>
<td width="50%" align="center">

### 🧠 **Anthropic Compatible**

`/v1/messages`

Point Claude Code at it via `ANTHROPIC_BASE_URL`

</td>
</tr>
</table>

### API Endpoints

| Endpoint                  | Method | Shape    | Description                                          | Status     |
| ------------------------- | ------ | -------- | ---------------------------------------------------- | ---------- |
| `/`                       | GET    | —        | Service info + available models                      | ✅ Working |
| `/health`                 | GET    | —        | Liveness probe (`{"status":"ok"}`)                   | ✅ Working |
| `/v1/models`              | GET    | OpenAI   | List virtual models (`meta-ai`, `meta-ai-think`, …) | ✅ Working |
| `/v1/chat/completions`    | POST   | OpenAI   | Chat completion — streaming (`stream:true`) supported | ✅ Working |
| `/chat/completions`       | POST   | OpenAI   | Alias (no `/v1` prefix) for older clients            | ✅ Working |
| `/v1/messages`            | POST   | Anthropic | Messages API — what Claude Code uses                | ✅ Working |
| `/messages`               | POST   | Anthropic | Alias (no `/v1` prefix)                             | ✅ Working |

**Authentication:** optional bearer token via `META_AI_PROXY_TOKEN`. Clients
send it in `Authorization: Bearer …` (OpenAI) or `x-api-key: …` (Anthropic).
If unset, the proxy accepts any caller.

### Virtual Models

| Model name                | Maps to                              | Example use                         |
| ------------------------- | ------------------------------------ | ----------------------------------- |
| `meta-ai` _(default)_     | normal chat                          | general Q&A, code, explanations     |
| `meta-ai-think`           | extended-thinking mode               | complex reasoning, math, planning   |
| `meta-ai-fast`            | instant mode                         | quick replies                       |
| `meta-ai-analyze`         | chat mode override                   | focused analysis                    |
| `meta-ai-learn`           | chat mode override                   | learning/explainer mode             |
| `meta-ai-image`           | image generation                     | returns image URL(s) in the content |
| `meta-ai-video`           | video generation                     | returns video URL(s) in the content |

> Unknown model names map heuristically: `gpt-4o`, `gpt-3.5-turbo`,
> `claude-3-5-sonnet`, `claude-3-opus`, etc. all work (falling back to default
> chat). Vision input (OpenAI `image_url` and Anthropic `image` content blocks)
> is uploaded and analyzed. Tool/function-calling fields are accepted but
> ignored (Meta AI has no native tool API).

### Run it

```bash
# Set the SDK credentials (same ones the library uses)
export META_AI_DATR=...
export META_AI_ECTO_1_SESS=...
export META_AI_ACCESS_TOKEN=ecto1:...

# Optional: require a bearer token from clients
export META_AI_PROXY_TOKEN=change-me

go run ./cmd/metaai-server            # default :8787
# or: make run
```

### Use it with Claude Code

```bash
export ANTHROPIC_BASE_URL=http://localhost:8787
export ANTHROPIC_API_KEY=$META_AI_PROXY_TOKEN   # must match the proxy token
# then launch Claude Code — its /v1/messages calls are proxied to Meta AI
```

### Use it as an OpenAI backend

```bash
curl http://localhost:8787/v1/chat/completions \
  -H "Authorization: Bearer $META_AI_PROXY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-ai-think",
    "messages": [{"role":"user","content":"Explain goroutines"}],
    "stream": true
  }'
```

**Response (OpenAI shape):**

```json
{
  "id": "chatcmpl-7a3f9c1b2e4d",
  "object": "chat.completion",
  "created": 1781964000,
  "model": "meta-ai-think",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "Goroutines are lightweight threads..."},
    "finish_reason": "stop"
  }],
  "usage": {"prompt_tokens": 4, "completion_tokens": 96, "total_tokens": 100}
}
```

### Use it as an Anthropic (Messages) backend

```bash
curl http://localhost:8787/v1/messages \
  -H "x-api-key: $META_AI_PROXY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-ai",
    "max_tokens": 1024,
    "stream": true,
    "messages": [{"role":"user","content":"What is 2+2?"}]
  }'
```

**Response (Anthropic shape, streamed):**

```
event: message_start
data: {"type":"message_start","message":{"id":"msg_...","role":"assistant",...}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Two plus two equals four."}}

event: message_stop
data: {"type":"message_stop"}
```

### Vision (image input)

Both APIs accept images in the latest user message. OpenAI uses `image_url`,
Anthropic uses `image` content blocks. They are downloaded (or decoded from
`data:` URLs) and run through `AnalyzeImage`:

```bash
curl http://localhost:8787/v1/chat/completions \
  -H "Authorization: Bearer $META_AI_PROXY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-ai",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "What colors are in this image?"},
        {"type": "image_url", "image_url": {"url": "https://example.com/photo.jpg"}}
      ]
    }]
  }'
```

> **Concurrency:** the proxy serializes chat turns (the SDK reuses one
> WebSocket per conversation). Fine for agent-style sequential use; add a
> client pool if you need parallel turns. Each request is stateless and sends
> the full transcript to Meta AI, matching how OpenAI/Anthropic clients behave.

---

## 🌍 REST API Server

The `cmd/metaai-rest` binary exposes the SDK as a REST API server with the same
surface as the canonical SDK REST API — a drop-in for any HTTP client.

### API Endpoints

| Endpoint               | Method | Description                              | Status     |
| ---------------------- | ------ | ---------------------------------------- | ---------- |
| `/`                    | GET    | Service info + available endpoints       | ✅ Working |
| `/healthz`             | GET    | Health check (`{"status":"ok"}`)         | ✅ Working |
| `/chat`                | POST   | Send chat messages (streaming supported) | ✅ Working |
| `/upload`              | POST   | Upload images for generation/analysis    | ✅ Working |
| `/image`               | POST   | Generate images from text                | ✅ Working |
| `/video`               | POST   | Generate video (blocks until complete)   | ✅ Working |
| `/video/extend`        | POST   | Extend a video from a media ID           | ✅ Working |
| `/video/async`         | POST   | Start async video generation             | ✅ Working |
| `/video/jobs/{job_id}` | GET    | Poll an async video job's status         | ✅ Working |

### Run it

```bash
# Credentials (same ones the SDK uses)
export META_AI_DATR=...
export META_AI_ECTO_1_SESS=...

# Optional bearer token clients must send
export META_AI_REST_TOKEN=change-me

go run ./cmd/metaai-rest             # default :8000
# or: make run-rest
```

### Examples

**Chat:**

```bash
curl http://localhost:8000/chat \
  -H "Authorization: Bearer $META_AI_REST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message":"What is 2+2?"}'
# → {"success":true,"message":"Two plus two equals four."}
```

**Generate image:**

```bash
curl http://localhost:8000/image \
  -H "Authorization: Bearer $META_AI_REST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"a red apple","orientation":"SQUARE"}'
# → {"success":true,"prompt":"a red apple","image_urls":["https://scontent..."],"media_ids":["1102..."],"status":"READY"}
```

**Async video generation + polling:**

```bash
# 1) Start the job
JOB=$(curl -s http://localhost:8000/video/async \
  -H "Authorization: Bearer $META_AI_REST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"sunset over ocean"}' | jq -r .job_id)

# 2) Poll until "completed"
curl -s http://localhost:8000/video/jobs/$JOB \
  -H "Authorization: Bearer $META_AI_REST_TOKEN"
# → {"job_id":"...","status":"completed","result":{"success":true,"video_urls":[...]}}
```

**Upload an image (multipart):**

```bash
curl http://localhost:8000/upload \
  -H "Authorization: Bearer $META_AI_REST_TOKEN" \
  -F "file=@photo.jpg"
# → {"success":true,"media_id":"1575...","file_name":"photo.jpg","file_size":12345,"mime_type":"image/jpeg"}
```

> **Authentication:** optional bearer token via `META_AI_REST_TOKEN`. Clients
> send it in `Authorization: Bearer …` or `x-api-key: …`. If unset, the server
> accepts any caller.

---

## 📖 مرجع API

| الدالة | الوصف | المثال |
|--------|--------|---------|
| `Chat(ctx, msg, opts)` | إرسال رسالة واستلام الرد | `client.Chat(ctx, "Hello", nil)` |
| `StreamChat(ctx, msg, opts)` | بث الرد عبر channel | `range client.StreamChat(ctx, "Hi", nil)` |
| `UploadImage(ctx, path)` | رفع صورة → media_id | `client.UploadImage(ctx, "photo.jpg")` |
| `AnalyzeImage(ctx, path, mediaID, q, opts)` | رفع + تحليل | `client.AnalyzeImage(ctx, "p.png", "", "Describe", nil)` |
| `GenerateImage(ctx, prompt, orient, num)` | توليد صورة | `client.GenerateImage(ctx, "a cat", "SQUARE", 1)` |
| `GenerateVideo(ctx, prompt)` | توليد فيديو | `client.GenerateVideo(ctx, "waves")` |
| `GetConversationHistory(ctx, limit, offset)` | السجل | `client.GetConversationHistory(ctx, 10, 0)` |
| `NewTopic / SetTopic / ListTopics` | إدارة المواضيع | `client.NewTopic("work", nil, nil, nil)` |

---

## 🏗️ البنية المعمارية

```
go/
├── metaai.go              # العميل (نقطة الدخول)
├── chat.go                # المحادثة (WebSocket template mode)
├── analyze.go             # تحليل الصور (حقن attachment)
├── generation_client.go   # توليد الصور/الفيديو
├── session.go             # إدارة المواضيع
├── graphql.go             # GraphQL HTTP
├── internal/
│   ├── clippy/            # البروتوكول الثنائي (encoder + parser + template)
│   │   ├── proto.go       # protobuf primitives
│   │   ├── frame.go       # بناء الإطار
│   │   ├── parse.go       # تحليل الردود
│   │   ├── template.go    # Template mode (استبدال الإطار المسجّل)
│   │   └── testdata/      # إطارات مسجّلة من المتصفح
│   ├── transport/         # gorilla/websocket dialer
│   ├── upload/            # rupload.meta.ai (lowercase headers)
│   └── generation/        # SSE parsing
├── cmd/livecheck/         # اختبارات حية
├── cmd/metaai-server/     # خادم API proxy (OpenAI + Anthropic)
├── cmd/metaai-rest/       # خادم REST API
├── proxy/                 # طبقة الـ HTTP proxy (types, handlers, SSE)
├── rest/                  # طبقة الـ REST API (handlers, async jobs)
└── examples/              # أمثلة
```

### 🔬 كيف يعمل Template Mode؟

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  إطار مسجّل  │ ──► │  استبدال     │ ──► │  إطار جديد   │
│  من المتصفح  │     │  النص + ID   │     │  جاهز للإرسال│
│              │     │  + attachment│     │              │
│  ✓ شغال     │     │  فقط        │     │  ✓ شغال     │
└──────────────┘     └──────────────┘     └──────────────┘
      │                                           │
      │  //go:embed                               │  WebSocket
      │  (مضمّن في الكود)                         │  إرسال
      ▼                                           ▼
  testdata/                               gateway.meta.ai
  template_frame.b64                      → رد فوري ✅
```

---

## 🧪 الاختبار

```bash
# 179 اختبار وحدة
go test ./...

# اختبارات حية (تتطلب cookies صالحة)
META_AI_DATR=... META_AI_ECTO_1_SESS=... META_AI_ACCESS_TOKEN=ecto1:... \
    go run ./cmd/livecheck

# تقرير التغطية
go test ./... -cover
```

### نتائج الاختبار الحي

```
✅ Chat:       "HELLO123" (رد مطابق)
✅ Upload:     mediaID=2019608138695944
✅ Analyze:    "Green" (صورة خضراء → AI شافها ووصفها)
✅ Generate:   تم الإرسال والمعالجة
✅ Stream:     39 chunks / 932 chars
✅ History:    success=true
✅ Topics:     NewTopic/SetTopic/Chat/ListTopics
```

---

## 📄 الترخيص

MIT

## 🙏 شكر وتقدير

- الهندسة العكسية للبروتوكول: التقاط حي من المتصفح (2026-06-19)
