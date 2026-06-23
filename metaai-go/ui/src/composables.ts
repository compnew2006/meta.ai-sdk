import { ref, computed, watch, type Ref } from 'vue';
import { currentLocale, toggleLocale, t } from './i18n';

// ── Multi-conversation model (ChatGPT-style) ──────────────────────────────

export interface ChatMessage {
  role: 'user' | 'assistant';
  text: string;
}

export interface Conversation {
  id: string;          // client-side id (uuid); distinct from metaConversationId
  metaConversationId: string; // Meta AI conversation_id (assigned by server; "" until first turn)
  title: string;
  messages: ChatMessage[];
  createdAt: number;
  updatedAt: number;
  // Per-conversation system instruction. Empty → no instruction. Initialized
  // from the global default (useSystemInstruction) when a conversation is
  // created, then editable independently per conversation.
  systemInstruction?: string;
}

const safeSessionStorage = typeof sessionStorage !== 'undefined' ? sessionStorage : {
  getItem: () => '',
  setItem: () => {},
  removeItem: () => {}
};

const safeLocalStorage = typeof localStorage !== 'undefined' ? localStorage : {
  getItem: () => '',
  setItem: () => {},
  removeItem: () => {}
};

const tokenRef = ref(safeSessionStorage.getItem('metaai_token') || '');

export function useAuth() {
  const isAuthenticated = computed(() => !!tokenRef.value.trim());

  function login(newToken: string) {
    const cleanToken = newToken.trim();
    safeSessionStorage.setItem('metaai_token', cleanToken);
    tokenRef.value = cleanToken;
  }

  function logout() {
    safeSessionStorage.removeItem('metaai_token');
    tokenRef.value = '';
  }

  return {
    token: computed(() => tokenRef.value),
    isAuthenticated,
    login,
    logout
  };
}

export function useLocale() {
  return {
    locale: currentLocale,
    t,
    toggleLocale
  };
}

/**
 * usePersistentRef wraps a Vue ref and mirrors its value to localStorage so the
 * value survives page refresh. Shared by all tabs (Chat/Stream/Analyze/Image/
 * Video) to keep persistence behavior identical and DRY.
 *
 * - Reads the initial value from localStorage (falling back to `initialValue`
 *   on miss or JSON parse error, so corrupted storage never wipes state).
 * - Deep-watches the ref and writes JSON to localStorage on change. Writes are
 *   debounced (default 200ms) so high-frequency streaming appends don't thrash
 *   localStorage on every token.
 * - SSR/no-storage environments fall through to the safe no-op storage shim.
 */
export function usePersistentRef<T>(key: string, initialValue: T, debounceMs = 200): Ref<T> {
  let stored: T = initialValue;
  const raw = safeLocalStorage.getItem(key);
  if (raw !== null && raw !== '') {
    try {
      stored = JSON.parse(raw) as T;
    } catch {
      // Corrupted entry — discard and start fresh rather than crash.
      safeLocalStorage.removeItem(key);
    }
  }
  const r = ref(stored) as Ref<T>;
  let writeTimer: ReturnType<typeof setTimeout> | null = null;
  watch(
    r,
    (val) => {
      if (writeTimer) clearTimeout(writeTimer);
      writeTimer = setTimeout(() => {
        try {
          safeLocalStorage.setItem(key, JSON.stringify(val));
        } catch {
          // QuotaExceeded or serialization failure — best-effort, ignore.
        }
      }, debounceMs);
    },
    { deep: true }
  );
  return r;
}

function genId(): string {
  return 'c-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 8);
}

function titleFromMessage(msg: string): string {
  const clean = msg.replace(/\s+/g, ' ').trim();
  return clean.length > 40 ? clean.slice(0, 40) + '…' : (clean || 'New chat');
}

// Global system-instruction default (the ⚙️ settings value). New conversations
// seed their per-conversation systemInstruction from this; the global value is
// NOT read at send time anymore (each conversation carries its own copy).
const systemInstructionRef = usePersistentRef<string>('metaai_system_instruction', '');

/**
 * createConversationStore is the shared multi-conversation store engine
 * (ChatGPT-style). Both the chat tab (useConversations) and the analyze tab
 * (useAnalyzeConversations) build on it so the sidebar logic stays DRY.
 *
 * Holds a list of conversations + the active one, persisted to localStorage.
 * One store per (listKey, activeKey) pair — module-level so the sidebar and the
 * main area in a tab share state.
 */
function createConversationStore<T extends Conversation>(options: {
  listKey: string;
  activeKey: string;
  makeNew: () => T;
  migrate?: (list: Ref<T[]>) => void;
}) {
  const { listKey, activeKey, makeNew, migrate } = options;
  const listRef = usePersistentRef<T[]>(listKey, []);
  const activeIdRef = usePersistentRef<string>(activeKey, '');

  let migratedDone = false;
  function runMigration(): void {
    if (migratedDone || listRef.value.length > 0) { migratedDone = true; return; }
    migratedDone = true;
    if (migrate) migrate(listRef);
  }

  const conversations = computed(() => listRef.value);
  const activeId = computed(() => activeIdRef.value);
  const activeConversation = computed(() =>
    listRef.value.find((c) => c.id === activeIdRef.value) || null
  );

  function createNew(): T {
    const conv = makeNew();
    listRef.value = [conv, ...listRef.value];
    activeIdRef.value = conv.id;
    return conv;
  }

  function ensureAtLeastOne(): T | null {
    if (listRef.value.length === 0) {
      return createNew();
    }
    if (!activeIdRef.value || !listRef.value.find((c) => c.id === activeIdRef.value)) {
      activeIdRef.value = listRef.value[0].id;
    }
    return activeConversation.value;
  }

  function select(id: string): void {
    if (listRef.value.find((c) => c.id === id)) {
      activeIdRef.value = id;
    }
  }

  function remove(id: string): void {
    const idx = listRef.value.findIndex((c) => c.id === id);
    if (idx < 0) return;
    listRef.value = listRef.value.filter((c) => c.id !== id);
    if (activeIdRef.value === id) {
      activeIdRef.value = listRef.value[0]?.id || '';
      if (listRef.value.length === 0) {
        ensureAtLeastOne();
      }
    }
  }

  function rename(id: string, title: string): void {
    const conv = listRef.value.find((c) => c.id === id);
    if (conv) {
      conv.title = title.trim() || conv.title;
      conv.updatedAt = Date.now();
      listRef.value = [...listRef.value];
    }
  }

  /** Append a message to a conversation and update its timestamp. */
  function appendMessage(id: string, msg: ChatMessage): T | null {
    const conv = listRef.value.find((c) => c.id === id);
    if (!conv) return null;
    conv.messages = [...conv.messages, msg];
    conv.updatedAt = Date.now();
    // Auto-title from the first user message if still the default.
    if (conv.title === 'New chat' && msg.role === 'user') {
      conv.title = titleFromMessage(msg.text);
    }
    listRef.value = [...listRef.value];
    return conv;
  }

  /** Patch the last assistant message of a conversation (streaming append). */
  function appendToLastAssistant(id: string, chunk: string): T | null {
    const conv = listRef.value.find((c) => c.id === id);
    if (!conv) return null;
    const msgs = conv.messages;
    if (msgs.length > 0 && msgs[msgs.length - 1].role === 'assistant') {
      msgs[msgs.length - 1] = { ...msgs[msgs.length - 1], text: msgs[msgs.length - 1].text + chunk };
    }
    conv.updatedAt = Date.now();
    listRef.value = [...listRef.value];
    return conv;
  }

  /** Record the Meta AI conversation_id assigned by the server for a conversation. */
  function setMetaConversationId(id: string, metaId: string): void {
    const conv = listRef.value.find((c) => c.id === id);
    if (conv && !conv.metaConversationId) {
      conv.metaConversationId = metaId;
      listRef.value = [...listRef.value];
    }
  }

  /** Update the per-conversation system instruction. */
  function setConvSystemInstruction(id: string, instr: string): void {
    const conv = listRef.value.find((c) => c.id === id);
    if (conv) {
      conv.systemInstruction = instr;
      conv.updatedAt = Date.now();
      listRef.value = [...listRef.value];
    }
  }

  return {
    conversations,
    activeId,
    activeConversation,
    createNew,
    ensureAtLeastOne,
    select,
    remove,
    rename,
    appendMessage,
    appendToLastAssistant,
    setMetaConversationId,
    setConvSystemInstruction,
    runMigration
  };
}

const CHAT_LIST_KEY = 'metaai_conversations';
const CHAT_ACTIVE_KEY = 'metaai_active_conversation';

/** Legacy migration: fold the old single-history keys into conversations. */
function migrateLegacyChatHistory(list: Ref<Conversation[]>): void {
  const now = Date.now();
  const imported: Conversation[] = [];
  for (const [legacyKey, label] of [
    ['metaai_stream_history', 'Imported stream chat'],
    ['metaai_chat_history', 'Imported chat']
  ] as const) {
    const raw = safeLocalStorage.getItem(legacyKey);
    if (!raw) continue;
    try {
      const msgs = JSON.parse(raw) as ChatMessage[];
      if (Array.isArray(msgs) && msgs.length > 0) {
        const firstUser = msgs.find((m) => m.role === 'user');
        imported.push({
          id: genId(),
          metaConversationId: '',
          title: firstUser ? titleFromMessage(firstUser.text) : label,
          messages: msgs,
          createdAt: now,
          updatedAt: now
        });
      }
    } catch {
      // corrupt legacy data — skip
    }
  }
  if (imported.length > 0) {
    list.value = imported;
  }
}

// Module-level singletons (one shared store per tab type).
const chatStore = createConversationStore<Conversation>({
  listKey: CHAT_LIST_KEY,
  activeKey: CHAT_ACTIVE_KEY,
  makeNew: () => ({
    id: genId(),
    metaConversationId: '',
    title: 'New chat',
    messages: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
    systemInstruction: systemInstructionRef.value
  }),
  migrate: migrateLegacyChatHistory
});

/**
 * useConversations is the multi-conversation store for the Chat tab
 * (ChatGPT-style sidebar). Module-level state shared across the sidebar and
 * chat area. Migrates legacy single-history keys on first load.
 */
export function useConversations() {
  chatStore.runMigration();
  chatStore.ensureAtLeastOne();
  return chatStore;
}

// ── Analyze conversations (image + multi-turn Q&A) ────────────────────────

export interface AnalyzeConversation extends Conversation {
  mediaId: string;  // uploaded image media_id (turn 1 only; "" until first turn)
  imageUrl: string; // local object URL preview (turn 1 only; not persisted meaningfully)
}

const ANALYZE_LIST_KEY = 'metaai_analyze_conversations';
const ANALYZE_ACTIVE_KEY = 'metaai_active_analyze_conversation';

const analyzeStore = createConversationStore<AnalyzeConversation>({
  listKey: ANALYZE_LIST_KEY,
  activeKey: ANALYZE_ACTIVE_KEY,
  makeNew: () => ({
    id: genId(),
    metaConversationId: '',
    title: 'New analysis',
    messages: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
    mediaId: '',
    imageUrl: '',
    systemInstruction: systemInstructionRef.value
  })
});

/** Patch the image fields of an analyze conversation (set on turn 1). */
function useAnalyzeImagePatch() {
  function setAnalyzeImage(id: string, mediaId: string, imageUrl: string): void {
    const conv = analyzeStore.activeConversation.value;
    const target = (conv && conv.id === id) ? conv : null;
    if (target) {
      target.mediaId = mediaId;
      target.imageUrl = imageUrl;
    }
  }
  return { setAnalyzeImage };
}

/**
 * useAnalyzeConversations is the multi-conversation store for the Analyze tab:
 * each conversation is a chat-about-an-image (upload once, ask follow-ups).
 * Mirrors useConversations but carries the turn-1 image fields.
 */
export function useAnalyzeConversations() {
  analyzeStore.ensureAtLeastOne();
  return { ...analyzeStore, ...useAnalyzeImagePatch() };
}

/**
 * useSystemInstruction exposes the global system-instruction DEFAULT (the ⚙️
 * settings value). It seeds new conversations; each conversation then carries
 * its own copy (editable per-session). Persisted to localStorage.
 */
export function useSystemInstruction() {
  return {
    systemInstruction: computed(() => systemInstructionRef.value),
    setSystemInstruction: (v: string) => { systemInstructionRef.value = v; }
  };
}
