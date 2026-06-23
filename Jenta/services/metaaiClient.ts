// services/metaaiClient.ts
//
// The ONE network layer between Jenta and the metaai-go REST API. All other
// service files go through here so HTTP details (base URL, bearer token,
// multipart vs JSON, base64 conversion) live in exactly one place (DRY).
//
// Endpoint contract: see metaai-go/docs/api_endpoints.md and
// metaai-go/rest/types.go.

/// <reference types="vite/client" />

const BASE = (import.meta.env.VITE_METAAI_URL as string | undefined) ?? 'http://localhost:8000';
const TOKEN = (import.meta.env.VITE_METAAI_TOKEN as string | undefined) ?? '';

function authHeaders(extra: Record<string, string> = {}): Record<string, string> {
  return {
    ...(TOKEN ? { Authorization: `Bearer ${TOKEN}` } : {}),
    ...extra,
  };
}

/**
 * JSON request helper. Sets Content-Type: application/json + bearer auth and
 * parses the JSON response. Throws on non-2xx.
 */
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(init?.headers as Record<string, string> | undefined),
    },
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`${path} → ${res.status}: ${text}`);
  }
  return JSON.parse(text) as T;
}

/**
 * Upload a browser Blob/File to /upload (multipart/form-data) and return the
 * assigned media_id. Do NOT set Content-Type — the browser must set the
 * multipart boundary itself.
 */
export async function uploadFile(file: Blob, fileName = 'upload.png'): Promise<string> {
  const form = new FormData();
  form.append('file', file, fileName);
  const res = await fetch(`${BASE}/upload`, {
    method: 'POST',
    headers: authHeaders(), // no Content-Type on purpose
    body: form,
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`/upload → ${res.status}: ${text}`);
  }
  const json = JSON.parse(text) as { success?: boolean; media_id?: string; error?: string };
  if (!json.success || !json.media_id) {
    throw new Error(json.error || 'upload failed');
  }
  return json.media_id;
}

/**
 * Fetch a generated image URL via the server-side /image/fetch endpoint and
 * return it in Jenta's base64 ImageFile shape. This avoids the fbcdn CDN's
 * inconsistent browser CORS — the SPA never calls fbcdn directly.
 */
export async function fetchImageToBase64(
  url: string,
  name = 'generated.png',
): Promise<{ base64: string; mimeType: string; name: string }> {
  const r = await request<{ success: boolean; base64: string; mime_type?: string; error?: string }>(
    `/image/fetch?url=${encodeURIComponent(url)}`,
  );
  if (!r.success || !r.base64) {
    throw new Error(r.error || 'image fetch returned no data');
  }
  return { base64: r.base64, mimeType: r.mime_type || 'image/png', name };
}

/** Convert a base64 string + mime into a Blob (for multipart upload). */
export function base64ToBlob(b64: string, mime: string): Blob {
  const bin = atob(b64);
  const len = bin.length;
  const bytes = new Uint8Array(len);
  for (let i = 0; i < len; i++) bytes[i] = bin.charCodeAt(i);
  return new Blob([bytes], { type: mime });
}

/** Map Jenta's aspect ratio to metaai-go's orientation enum. */
export function orientationFor(aspect: string): 'LANDSCAPE' | 'VERTICAL' | 'SQUARE' {
  if (aspect === '16:9') return 'LANDSCAPE';
  if (aspect === '1:1') return 'SQUARE';
  return 'VERTICAL'; // 9:16, 4:3, 3:4, default
}

// ---------------------------------------------------------------------------
// Concurrency-limited map (avoids hammering Meta AI with N parallel uploads)
// ---------------------------------------------------------------------------

/**
 * Run an async mapper over `items` with at most `limit` calls in flight at
 * once. Preserves input order in the output. A tiny pLimit — no dependency.
 */
export async function mapLimit<T, R>(
  items: T[],
  limit: number,
  fn: (item: T, index: number) => Promise<R>,
): Promise<R[]> {
  const results = new Array<R>(items.length);
  let cursor = 0;
  async function worker() {
    while (cursor < items.length) {
      const i = cursor++;
      results[i] = await fn(items[i], i);
    }
  }
  const workerCount = Math.max(1, Math.min(limit, items.length));
  await Promise.all(Array.from({ length: workerCount }, () => worker()));
  return results;
}

// ---------------------------------------------------------------------------
// File hashing (cache key for uploads + descriptions)
// ---------------------------------------------------------------------------

/** SHA-256 of a Blob, hex-encoded. Used to dedupe uploads + cache descriptions. */
export async function hashFile(file: Blob): Promise<string> {
  const buf = await file.arrayBuffer();
  const digest = await crypto.subtle.digest('SHA-256', buf);
  return Array.from(new Uint8Array(digest))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

// ---------------------------------------------------------------------------
// Per-studio conversation cache (multi-turn /analyze without re-upload)
// ---------------------------------------------------------------------------

const CONV_PREFIX = 'metaai:conv:';
const DESC_PREFIX = 'metaai:desc:'; // hash → reference-image description (#2)

function lsGet(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch (e) {
    // localStorage disabled (private mode, sandboxed iframe): cache miss degrades
    // gracefully to a fresh /analyze round-trip, so warn-and-return is correct.
    console.warn('metaai cache: localStorage unavailable, treating as miss', e);
    return null;
  }
}
function lsSet(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch (e) {
    // Quota exceeded or storage disabled: the cache is best-effort; a write
    // miss only costs a future /analyze round-trip, never data loss.
    console.warn('metaai cache: localStorage write failed', key, e);
  }
}

/**
 * Get a stored Meta AI conversation_id for a (studio, imageHash) pair, so a
 * follow-up /analyze reuses prior context instead of re-uploading.
 */
export function getConversation(studio: string, imageHash: string): string {
  return lsGet(CONV_PREFIX + studio + ':' + imageHash) ?? '';
}

export function setConversation(studio: string, imageHash: string, conversationId: string): void {
  if (conversationId) lsSet(CONV_PREFIX + studio + ':' + imageHash, conversationId);
}

/**
 * Cache a text description derived from a reference image (#2). Lets us fold a
 * product/style description into the prompt on every /image call without
 * re-running /analyze each time.
 */
export function getDescription(imageHash: string): string | null {
  return lsGet(DESC_PREFIX + imageHash);
}

export function setDescription(imageHash: string, description: string): void {
  if (description) lsSet(DESC_PREFIX + imageHash, description);
}

export { BASE, TOKEN };
