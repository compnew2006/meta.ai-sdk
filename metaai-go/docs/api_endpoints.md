# Meta AI Go REST API Documentation

This document describes the REST API endpoints exposed by the `metaai-go` REST server. 

By default, the server listens on `http://localhost:8000`.

---

## Authentication

If the server is started with a required token, all endpoints (except `/`, `/healthz`, and `/ui/`) require authentication. Authentication can be provided in one of the following ways:

1. **Authorization Header**:
   ```http
   Authorization: Bearer <your-api-key>
   ```
2. **Custom API Key Headers** (e.g. `X-Api-Key`, `X-API-Key`, or `x-api-key`):
   ```http
   X-Api-Key: <your-api-key>
   ```

---

## Endpoints Summary

| Method | Path | Auth Required | Description |
|:---|:---|:---:|:---|
| **GET** | `/` | No | Service information and list of endpoints |
| **GET** | `/healthz` | No | Liveness/health probe |
| **POST** | `/chat` | Yes | Send chat message (supports streaming) |
| **POST** | `/upload` | Yes | Upload an image for analysis |
| **POST** | `/analyze` | Yes | Send an image (by `media_id`) and prompt to Meta AI (supports streaming) |
| **POST** | `/image` | Yes | Generate images from a text prompt |
| **GET** | `/image/fetch` | Yes | Fetch a generated image URL server-side, return base64 (CORS-safe for browsers) |
| **POST** | `/video` | Yes | Generate a video from a text prompt (blocks) |
| **POST** | `/video/extend` | Yes | Extend an existing video |
| **POST** | `/video/async` | Yes | Generate a video asynchronously (background job) |
| **GET** | `/video/jobs/{job_id}` | Yes | Check the status of a background video job |
| **GET** | `/ui/` | No | Embedded Vue-based Web Dashboard |

---

## Endpoint Details

### 1. Service Info
* **Method**: `GET`
* **Path**: `/`
* **Response Payload (`application/json`)**:
  ```json
  {
    "name": "metaai-rest",
    "endpoints": [
      "/healthz",
      "/chat",
      "/upload",
      "/analyze",
      "/image",
      "/video",
      "/video/extend",
      "/video/async",
      "/video/jobs/{job_id}"
    ]
  }
  ```

---

### 2. Health Check
* **Method**: `GET`
* **Path**: `/healthz`
* **Response Payload (`application/json`)**:
  ```json
  {
    "status": "ok"
  }
  ```

---

### 3. Chat
* **Method**: `POST`
* **Path**: `/chat`
* **Request Payload (`application/json`)**:
  ```json
  {
    "message": "Hello Meta AI!",
    "new_conversation": true,
    "stream": false,
    "thinking": false,
    "instant": true,
    "mode": "chat"
  }
  ```
  * `message` (string, required): The prompt text.
  * `new_conversation` (boolean, optional): Whether to start a new chat thread.
  * `stream` (boolean, optional): Set to `true` to receive chunked EventStream updates.
  * `thinking` (boolean, optional): Request thinking/reasoning outputs.
  * `instant` (boolean, optional): Request quick/instant answers.
  * `mode` (string, optional): Specific operation mode.
* **Response Payload - Non-Streaming (`application/json`)**:
  ```json
  {
    "success": true,
    "message": "Hello! How can I help you today?",
    "conversation_id": "optional-uuid"
  }
  ```
* **Response Payload - Streaming (`text/event-stream`)**:
  Chunks are sent in SSE format: `data: <json_payload>\n\n`.
  ```json
  {"success":true,"message":"Hello"}
  {"success":true,"message":"!"}
  ```

---

### 4. Upload Image
* **Method**: `POST`
* **Path**: `/upload`
* **Request Payload (`multipart/form-data`)**:
  * `file` (binary, required): The raw image file (limited to 32 MB).
* **Response Payload (`application/json`)**:
  ```json
  {
    "success": true,
    "media_id": "867051314767696",
    "file_name": "photo.png",
    "file_size": 24536,
    "mime_type": "image/png"
  }
  ```

---

### 5. Analyze Image
* **Method**: `POST`
* **Path**: `/analyze`
* **Request Payload (`application/json`)**:
  ```json
  {
    "media_id": "867051314767696",
    "question": "What is in this picture?",
    "stream": false
  }
  ```
  * `media_id` (string, required): The ID of the image returned from `/upload`.
  * `question` (string, required): The prompt or instructions for the image analysis.
  * `stream` (boolean, optional): Set to `true` to receive chunked EventStream updates as the analysis is generated (mirrors `/chat` streaming). Default `false` returns the full analysis in a single JSON body.
* **Response Payload - Non-Streaming (`application/json`)**:
  ```json
  {
    "success": true,
    "message": "This is a picture containing..."
  }
  ```
* **Response Payload - Streaming (`text/event-stream`)**:
  Chunks are sent in SSE format: `data: <json_payload>\n\n`, one frame per
  generated token, flushed immediately (mirrors `/chat`). The stream closes
  when generation completes; on failure a final `{"error":"..."}` frame is
  emitted before close.
  ```json
  {"success":true,"message":"This"}
  {"success":true,"message":" is"}
  {"success":true,"message":" a picture"}
  ```

---

### 6. Generate Image
* **Method**: `POST`
* **Path**: `/image`
* **Request Payload (`application/json`)**:
  ```json
  {
    "prompt": "A futuristic city under a neon sky",
    "orientation": "SQUARE"
  }
  ```
  * `prompt` (string, required): Text description of the image to generate.
  * `orientation` (string, optional): `SQUARE` (default), `LANDSCAPE`, or `VERTICAL`.
* **Response Payload (`application/json`)**:
  ```json
  {
    "success": true,
    "prompt": "A futuristic city under a neon sky",
    "image_urls": ["https://scontent.xx.fbcdn.net/..."],
    "media_ids": ["12345678"],
    "status": "READY",
    "conversation_id": "conv-uuid"
  }
  ```

---

### 7. Generate Video (Synchronous)
* **Method**: `POST`
* **Path**: `/video`
* **Request Payload (`application/json`)**:
  ```json
  {
    "prompt": "A cat playing with a ball of yarn"
  }
  ```
* **Response Payload (`application/json`)**:
  Blocks until the video generation is complete.
  ```json
  {
    "success": true,
    "prompt": "A cat playing with a ball of yarn",
    "video_urls": ["https://..."],
    "media_ids": ["87654321"],
    "status": "READY",
    "conversation_id": "conv-uuid"
  }
  ```

---

### 8. Extend Video
* **Method**: `POST`
* **Path**: `/video/extend`
* **Request Payload (`application/json`)**:
  ```json
  {
    "media_id": "87654321"
  }
  ```
  * `media_id` (string, required): The media ID of the original video to extend.
* **Response Payload (`application/json`)**:
  ```json
  {
    "success": true,
    "video_urls": ["https://..."],
    "media_ids": ["87654322"],
    "status": "READY"
  }
  ```

---

### 9. Generate Video (Asynchronous)
* **Method**: `POST`
* **Path**: `/video/async`
* **Request Payload (`application/json`)**:
  ```json
  {
    "prompt": "A spaceship warping through hyper-space"
  }
  ```
* **Response Payload (`application/json`)**:
  Returns immediately with a `job_id`.
  ```json
  {
    "success": true,
    "job_id": "job_1718873612",
    "status": "queued"
  }
  ```

---

### 10. Check Video Job Status
* **Method**: `GET`
* **Path**: `/video/jobs/{job_id}`
* **Response Payload (`application/json`)**:
  ```json
  {
    "job_id": "job_1718873612",
    "status": "completed",
    "result": {
      "success": true,
      "prompt": "A spaceship warping through hyper-space",
      "video_urls": ["https://..."],
      "media_ids": ["99988877"],
      "status": "READY",
      "conversation_id": "conv-uuid"
    }
  }
  ```
  * `status` can be one of: `queued`, `running`, `completed`, `failed`.
  * `result` is only present when the status is `completed`.

---

### 11. Embedded Web Dashboard
* **Method**: `GET`
* **Path**: `/ui/`
* **Response**: Serves the Vue 3 + TypeScript single-page app dashboard compiled files. SPA routing falls back to `/ui/index.html` dynamically.

---

### 12. Fetch a generated image as base64 (CORS-safe)
* **Method**: `GET`
* **Path**: `/image/fetch?url=<fbcdn-url>`
* **Auth**: required

Meta AI surfaces generated images on `*.fbcdn.net`, whose `Access-Control-Allow-Origin`
header is inconsistent from the browser. This endpoint fetches the URL
**server-side** and returns the bytes as base64, so a browser SPA can convert a
generated image into its own base64 image representation without a flaky
cross-origin fetch.

**Strict allow-list (SSRF guard):** only hosts ending in `.fbcdn.net` or
`.cdninstagram.com` are accepted. Any other host returns `400`. The response
body is capped at 32 MB.

* **Query**: `url` — the absolute `https://*.fbcdn.net/...` URL returned in an
  `image_urls` / `video_urls` field.
* **Response** (`200`, `ImageFetchResponse`):
  ```json
  {
    "success": true,
    "base64": "<base64-encoded image bytes>",
    "mime_type": "image/png"
  }
  ```
* **Errors**:
  - `400` — missing `url`, unparseable URL, or host not on the fbcdn allow-list.
  - `502` — upstream fetch failed or returned non-200.

---

## CORS (browser callers)

The server wraps every route in a CORS middleware so a browser SPA on a
different origin (e.g. a Vite dev server on `http://localhost:3000`) can call
the API directly.

* `Access-Control-Allow-Origin` is set from, in priority order:
  1. `Config.CORSOrigin`,
  2. the `META_AI_CORS_ORIGIN` env var,
  3. the default `http://localhost:3000`.
* Preflight `OPTIONS` requests are answered with `204 No Content` and
  `Access-Control-Allow-Headers: Authorization, Content-Type, X-Api-Key` plus
  `Access-Control-Allow-Methods: GET, POST, OPTIONS`.
* Set `META_AI_CORS_ORIGIN=` (empty) to **disable** CORS — recommended for
  production, where the SPA is served same-origin from this server (see the
  `make build-jenta` target, which bundles the Jenta React app into
  `jenta/dist`).

