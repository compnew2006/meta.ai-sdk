export interface ChatRequest {
  message: string;
  new_conversation?: boolean;
  stream?: boolean;
  thinking?: boolean;
  instant?: boolean;
  mode?: string;
  conversation_id?: string;
  system_instruction?: string;
}

export interface ChatResponse {
  success: boolean;
  message: string;
  conversation_id?: string;
  error?: string;
}

export interface UploadResponse {
  success: boolean;
  media_id: string;
  file_name?: string;
  file_size?: number;
  mime_type?: string;
  error?: string;
}

export interface AnalyzeRequest {
  media_id?: string;
  question: string;
  stream?: boolean;
  conversation_id?: string;
  system_instruction?: string;
}

export interface AnalyzeResponse {
  success: boolean;
  message: string;
  conversation_id?: string;
  error?: string;
}

export interface ImageRequest {
  prompt: string;
  orientation?: 'LANDSCAPE' | 'VERTICAL' | 'SQUARE';
}

export interface ImageResponse {
  success: boolean;
  prompt?: string;
  image_urls?: string[];
  media_ids?: string[];
  status?: string;
  error?: string;
}

export interface VideoRequest {
  prompt: string;
  auto_poll?: boolean;
}

export interface VideoResponse {
  success: boolean;
  prompt?: string;
  video_urls?: string[];
  media_ids?: string[];
  status?: string;
  error?: string;
}

export interface AsyncJobResponse {
  success: boolean;
  job_id: string;
  status: string;
}

export interface JobStatusResponse {
  job_id: string;
  status: 'queued' | 'running' | 'completed' | 'failed';
  result?: VideoResponse;
  error?: string;
}

function getHeaders(): HeadersInit {
  const token = sessionStorage.getItem('metaai_token') || '';
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    let parsedErr = '';
    try {
      const data = JSON.parse(text);
      parsedErr = data.error || data.message;
    } catch {
      parsedErr = text;
    }
    throw new Error(parsedErr || `HTTP Error ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function chat(req: ChatRequest): Promise<ChatResponse> {
  const res = await fetch('/chat', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(req)
  });
  return handleResponse<ChatResponse>(res);
}

export async function uploadFile(file: File): Promise<UploadResponse> {
  const token = sessionStorage.getItem('metaai_token') || '';
  const fd = new FormData();
  fd.append('file', file);
  
  const res = await fetch('/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: fd
  });
  return handleResponse<UploadResponse>(res);
}

export async function analyze(req: AnalyzeRequest): Promise<AnalyzeResponse> {
  const res = await fetch('/analyze', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(req)
  });
  return handleResponse<AnalyzeResponse>(res);
}

export async function generateImage(req: ImageRequest): Promise<ImageResponse> {
  const res = await fetch('/image', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify(req)
  });
  return handleResponse<ImageResponse>(res);
}

export async function generateVideo(req: VideoRequest): Promise<VideoResponse> {
  const res = await fetch('/video', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ ...req, auto_poll: true })
  });
  return handleResponse<VideoResponse>(res);
}

export async function generateVideoAsync(req: VideoRequest): Promise<AsyncJobResponse> {
  const res = await fetch('/video/async', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ ...req, auto_poll: false })
  });
  return handleResponse<AsyncJobResponse>(res);
}

export async function getJobStatus(jobId: string): Promise<JobStatusResponse> {
  const res = await fetch(`/video/jobs/${jobId}`, {
    method: 'GET',
    headers: getHeaders()
  });
  return handleResponse<JobStatusResponse>(res);
}

export async function extendVideo(mediaId: string): Promise<VideoResponse> {
  const res = await fetch('/video/extend', {
    method: 'POST',
    headers: getHeaders(),
    body: JSON.stringify({ media_id: mediaId })
  });
  return handleResponse<VideoResponse>(res);
}

export async function streamChat(
  message: string,
  options: { thinking?: boolean; instant?: boolean; newConversation?: boolean; mode?: string; conversationId?: string; systemInstruction?: string },
  onChunk: (text: string) => void,
  onError: (err: any) => void,
  onConversationId?: (id: string) => void
) {
  const token = sessionStorage.getItem('metaai_token') || '';
  try {
    const response = await fetch('/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        message,
        stream: true,
        thinking: options.thinking,
        instant: options.instant,
        new_conversation: options.newConversation,
        mode: options.mode,
        conversation_id: options.conversationId || undefined,
        system_instruction: options.systemInstruction || undefined
      })
    });
    if (!response.ok) {
      const errText = await response.text();
      let parsedErr = '';
      try {
        parsedErr = JSON.parse(errText).error;
      } catch {
        parsedErr = errText;
      }
      throw new Error(parsedErr || `HTTP ${response.status}`);
    }
    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('ReadableStream not supported by browser');
    }
    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';
      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith('data:')) {
          const jsonStr = trimmed.substring(5).trim();
          if (jsonStr) {
            try {
              const data = JSON.parse(jsonStr);
              if (data.error) {
                onError(new Error(data.error));
                return;
              }
              // Surface the Meta AI conversation_id as soon as it arrives
              // (every chunk carries it). The UI stores it to resume context.
              if (data.conversation_id && onConversationId) {
                onConversationId(data.conversation_id);
              }
              if (data.success && data.message) {
                onChunk(data.message);
              }
            } catch (e) {
              // ignore invalid line JSON parse
            }
          }
        }
      }
    }
  } catch (err) {
    onError(err);
  }
}

/**
 * streamAnalyze is the streaming variant of `analyze`. It POSTs to /analyze
 * with stream:true and reads the Server-Sent Events stream, invoking onChunk
 * for each token and onError on failure. The SSE parse loop is identical to
 * streamChat (no done-marker — completion is implicit when the stream closes).
 */
export async function streamAnalyze(
  req: AnalyzeRequest,
  onChunk: (text: string) => void,
  onError: (err: any) => void,
  onConversationId?: (id: string) => void
) {
  const token = sessionStorage.getItem('metaai_token') || '';
  try {
    const response = await fetch('/analyze', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ ...req, stream: true })
    });
    if (!response.ok) {
      const errText = await response.text();
      let parsedErr = '';
      try {
        parsedErr = JSON.parse(errText).error;
      } catch {
        parsedErr = errText;
      }
      throw new Error(parsedErr || `HTTP ${response.status}`);
    }
    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error('ReadableStream not supported by browser');
    }
    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';
      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith('data:')) {
          const jsonStr = trimmed.substring(5).trim();
          if (jsonStr) {
            try {
              const data = JSON.parse(jsonStr);
              if (data.error) {
                onError(new Error(data.error));
                return;
              }
              if (data.conversation_id && onConversationId) {
                onConversationId(data.conversation_id);
              }
              if (data.success && data.message) {
                onChunk(data.message);
              }
            } catch {
              // ignore invalid line JSON parse
            }
          }
        }
      }
    }
  } catch (err) {
    onError(err);
  }
}
