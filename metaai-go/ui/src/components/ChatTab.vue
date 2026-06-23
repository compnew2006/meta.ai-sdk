<script setup lang="ts">
import { ref, nextTick, computed } from 'vue';
import { useLocale, useConversations } from '../composables';
import { streamChat } from '../api';
import ErrorAlert from './ErrorAlert.vue';
import { Send, Plus, MessageSquare, Pencil, Trash2, Activity, Settings } from '@lucide/vue';

const { t, locale } = useLocale();
const {
  conversations,
  activeConversation,
  createNew,
  select,
  remove,
  rename,
  appendMessage,
  appendToLastAssistant,
  setMetaConversationId,
  setConvSystemInstruction
} = useConversations();

const prompt = ref('');
const loading = ref(false);
const error = ref('');
const thinking = ref(false);
const instant = ref(false);

// Sidebar inline-edit state
const editingId = ref<string | null>(null);
const editingTitle = ref('');
const menuOpenId = ref<string | null>(null);

// Per-conversation system-instruction editor state.
const editingSysId = ref<string | null>(null);
const editingSysValue = ref('');

const chatAreaRef = ref<HTMLElement | null>(null);
const messages = computed(() => activeConversation.value?.messages ?? []);

function scrollToBottom() {
  nextTick(() => {
    if (chatAreaRef.value) {
      chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight;
    }
  });
}

function startNewChat() {
  if (loading.value) return;
  createNew();
  error.value = '';
  prompt.value = '';
}

function startRename(id: string, currentTitle: string) {
  editingId.value = id;
  editingTitle.value = currentTitle;
  menuOpenId.value = null;
}

function commitRename() {
  if (editingId.value) {
    rename(editingId.value, editingTitle.value);
    editingId.value = null;
  }
}

function startEditSystem(id: string) {
  const conv = conversations.value.find((c) => c.id === id);
  editingSysId.value = id;
  editingSysValue.value = conv?.systemInstruction ?? '';
  menuOpenId.value = null;
}
function commitEditSystem() {
  if (editingSysId.value) {
    setConvSystemInstruction(editingSysId.value, editingSysValue.value.trim());
    editingSysId.value = null;
  }
}

function deleteConversation(id: string) {
  if (loading.value && activeConversation.value?.id === id) return;
  if (window.confirm(t.value.confirmDelete)) {
    remove(id);
  }
  menuOpenId.value = null;
}

function toggleMenu(id: string) {
  menuOpenId.value = menuOpenId.value === id ? null : id;
}

async function handleSend() {
  const msg = prompt.value.trim();
  const conv = activeConversation.value;
  if (!msg || loading.value || !conv) return;

  loading.value = true;
  error.value = '';
  prompt.value = '';

  // Add user message + empty assistant placeholder, then stream into it.
  appendMessage(conv.id, { role: 'user', text: msg });
  appendMessage(conv.id, { role: 'assistant', text: '' });
  scrollToBottom();

  try {
    await streamChat(
      msg,
      {
        thinking: thinking.value,
        instant: instant.value,
        // Resume this conversation's Meta AI context if it has an id yet;
        // otherwise let the server assign one (captured below).
        conversationId: conv.metaConversationId || undefined,
        // Per-conversation system instruction (seeded from the global default
        // at creation, editable via the sidebar menu).
        systemInstruction: conv.systemInstruction || undefined
      },
      (chunk) => {
        appendToLastAssistant(conv.id, chunk);
        scrollToBottom();
      },
      (err) => {
        error.value = err.message || t.value.apiError;
        loading.value = false;
      },
      (metaConversationId) => {
        // First chunk carries the conversation_id Meta AI assigned/reused.
        // Store it so the next turn resumes this conversation's context.
        setMetaConversationId(conv.id, metaConversationId);
      }
    );
    loading.value = false;
  } catch (err: any) {
    error.value = err.message || t.value.apiError;
    loading.value = false;
  }
}
</script>

<template>
  <div class="tab-content animate-fade-in">
    <div class="chat-layout">
      <!-- Conversation Sidebar -->
      <aside class="conv-sidebar glass">
        <button class="new-chat-btn glow-primary" @click="startNewChat" :disabled="loading">
          <Plus class="btn-icon" />
          <span>{{ t.newChat }}</span>
        </button>

        <div class="conv-list">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            class="conv-item"
            :class="{ active: conv.id === activeConversation?.id }"
            @click="select(conv.id)"
          >
            <template v-if="editingId === conv.id">
              <input
                v-model="editingTitle"
                class="rename-input"
                @keyup.enter="commitRename"
                @blur="commitRename"
                @click.stop
                autofocus
              />
            </template>
            <template v-else>
              <MessageSquare class="conv-icon" />
              <span class="conv-title">{{ conv.title }}</span>
              <div class="conv-menu-wrapper">
                <button class="conv-menu-btn" @click.stop="toggleMenu(conv.id)">
                  ⋯
                </button>
                <div v-if="menuOpenId === conv.id" class="conv-menu" @click.stop>
                  <button @click="startRename(conv.id, conv.title)">
                    <Pencil class="menu-icon" /> {{ t.rename }}
                  </button>
                  <button @click="startEditSystem(conv.id)">
                    <Settings class="menu-icon" /> {{ t.systemInstructions }}
                  </button>
                  <button class="danger" @click="deleteConversation(conv.id)">
                    <Trash2 class="menu-icon" /> {{ t.delete }}
                  </button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </aside>

      <!-- Per-conversation system-instruction editor -->
      <div v-if="editingSysId" class="sys-overlay" @click.self="commitEditSystem">
        <div class="sys-modal glass">
          <h3 class="settings-title">{{ t.systemInstructions }}</h3>
          <textarea
            v-model="editingSysValue"
            class="settings-textarea"
            :placeholder="t.systemInstructionsPlaceholder"
            rows="5"
          ></textarea>
          <div class="sys-actions">
            <button class="sys-btn cancel" @click="commitEditSystem">{{ t.cancel }}</button>
            <button class="sys-btn glow-primary" @click="commitEditSystem">{{ t.saveSettings }}</button>
          </div>
        </div>
      </div>

      <!-- Chat Area -->
      <div class="chat-main">
        <div ref="chatAreaRef" class="chat-area glass">
          <!-- Empty State -->
          <div v-if="messages.length === 0" class="empty-chat">
            <Activity class="empty-icon pulse-glow" />
            <h3>{{ t.chat }}</h3>
            <p class="subtitle">{{ t.emptyChat }}</p>
          </div>

          <!-- Message History -->
          <div v-else class="message-list">
            <div
              v-for="(msg, index) in messages"
              :key="index"
              class="message-wrapper"
              :class="msg.role"
            >
              <div class="message-bubble">
                <div v-if="msg.role === 'assistant' && !msg.text && loading" class="streaming-indicator">
                  <span class="pulse-dot"></span>
                </div>
                <p class="message-text">{{ msg.text }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Error Banner -->
        <ErrorAlert :message="error" @close="error = ''" />

        <!-- Chat Controls & Input -->
        <div class="input-section glass">
          <div class="toolbar">
            <label class="option-check">
              <input v-model="thinking" type="checkbox" />
              <span class="label-text">{{ t.thinking }}</span>
            </label>
            <label class="option-check">
              <input v-model="instant" type="checkbox" />
              <span class="label-text">{{ t.instant }}</span>
            </label>
          </div>

          <form class="input-form" @submit.prevent="handleSend">
            <input
              v-model="prompt"
              type="text"
              :placeholder="t.promptPlaceholder"
              class="chat-input"
              :disabled="loading"
            />
            <button type="submit" class="send-btn glow-primary" :disabled="loading || !prompt.trim()">
              <Send class="send-icon" />
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tab-content {
  height: 100%;
}

.chat-layout {
  display: flex;
  gap: 16px;
  height: 100%;
}

/* ── Sidebar ── */
.conv-sidebar {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 14px;
  gap: 12px;
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.new-chat-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 12px;
  background: linear-gradient(135deg, var(--color-secondary), var(--color-primary));
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.new-chat-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.new-chat-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  width: 16px;
  height: 16px;
}

.conv-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-inline: -6px;
  padding-inline: 6px;
}

.conv-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
  min-width: 0;
}

.conv-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.conv-item.active {
  background: rgba(59, 130, 246, 0.12);
}

.conv-icon {
  width: 15px;
  height: 15px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.conv-item.active .conv-icon {
  color: var(--color-primary);
}

.conv-title {
  font-size: 13px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.conv-item.active .conv-title {
  color: var(--text-primary);
  font-weight: 600;
}

.rename-input {
  flex: 1;
  background: var(--bg-input);
  border: 1px solid var(--color-secondary);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  outline: none;
  min-width: 0;
}

.conv-menu-wrapper {
  position: relative;
  flex-shrink: 0;
}

.conv-menu-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 16px;
  padding: 0 4px;
  line-height: 1;
  opacity: 0;
  transition: opacity 0.15s;
}

.conv-item:hover .conv-menu-btn {
  opacity: 1;
}

.conv-menu {
  position: absolute;
  inset-inline-end: 0;
  top: 100%;
  z-index: 10;
  background: var(--bg-card, #1a1a2e);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  overflow: hidden;
  min-width: 140px;
}

.conv-menu button {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 9px 12px;
  background: none;
  border: none;
  color: var(--text-secondary);
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  text-align: start;
}

.conv-menu button:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}

.conv-menu button.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #fca5a5;
}

.menu-icon {
  width: 14px;
  height: 14px;
}

/* ── Chat main ── */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.chat-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-lg);
  min-height: 250px;
}

.empty-chat {
  margin: auto;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-icon {
  width: 48px;
  height: 48px;
  color: var(--color-secondary);
  margin-bottom: 8px;
}

.empty-chat h3 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.subtitle {
  font-size: 14px;
  color: var(--text-secondary);
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-wrapper {
  display: flex;
  width: 100%;
}

.message-wrapper.user {
  justify-content: flex-end;
}

.message-wrapper.assistant {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 75%;
  padding: 12px 18px;
  border-radius: var(--radius-md);
  line-height: 1.5;
  font-size: 15px;
}

.user .message-bubble {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: white;
  border-bottom-inline-end-radius: 4px;
}

.assistant .message-bubble {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  border-bottom-inline-start-radius: 4px;
}

.message-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.streaming-indicator {
  display: flex;
  align-items: center;
  height: 20px;
  padding: 0 4px;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background: var(--color-secondary);
  border-radius: 50%;
  animation: pulseGlow 1.2s infinite ease-in-out;
}

@keyframes pulseGlow {
  0%, 100% { transform: scale(0.6); opacity: 0.5; }
  50% { transform: scale(1.1); opacity: 1; box-shadow: 0 0 8px var(--color-secondary); }
}

/* Input Section */
.input-section {
  padding: 16px;
  border-radius: var(--radius-lg);
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
}

.option-check {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.option-check input {
  accent-color: var(--color-secondary);
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.label-text {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.input-form {
  display: flex;
  gap: 12px;
}

.chat-input {
  flex: 1;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 15px;
  padding: 14px 18px;
  border-radius: var(--radius-md);
  outline: none;
  transition: var(--transition-smooth);
}

.chat-input:focus {
  border-color: var(--color-secondary);
  box-shadow: 0 0 10px var(--color-secondary-glow);
}

.send-btn {
  background: linear-gradient(135deg, var(--color-secondary), var(--color-primary));
  color: white;
  border: none;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.send-btn:hover:not(:disabled) {
  transform: scale(1.05);
  filter: brightness(1.1);
}

.send-btn:disabled {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  cursor: not-allowed;
  box-shadow: none;
}

.send-icon {
  width: 18px;
  height: 18px;
  transform: v-bind('locale === "ar" ? "rotate(180deg)" : "none"');
}

@media (max-width: 1024px) {
  .chat-layout {
    flex-direction: column;
  }
  .conv-sidebar {
    width: 100%;
    max-height: 180px;
  }
}

/* Per-conversation system-instruction modal */
.sys-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.55);
  display: flex; align-items: center; justify-content: center; z-index: 100; padding: 20px;
}
.sys-modal {
  width: 100%; max-width: 520px; padding: 24px;
  border-radius: var(--radius-lg); display: flex; flex-direction: column; gap: 16px;
}
.settings-title { font-size: 18px; font-weight: 600; color: var(--text-primary); margin: 0; }
.settings-textarea {
  background: var(--bg-input); border: 1px solid var(--border-color); color: var(--text-primary);
  font-family: inherit; font-size: 14px; line-height: 1.5; padding: 12px 14px;
  border-radius: var(--radius-md); outline: none; resize: vertical; min-height: 110px;
}
.settings-textarea:focus { border-color: var(--color-secondary); box-shadow: 0 0 10px var(--color-secondary-glow); }
.sys-actions { display: flex; justify-content: flex-end; gap: 12px; }
.sys-btn {
  font-family: inherit; font-size: 13px; font-weight: 600; padding: 10px 16px;
  border-radius: var(--radius-md); border: 1px solid var(--border-color); cursor: pointer;
  background: rgba(255,255,255,0.03); color: var(--text-secondary); transition: var(--transition-smooth);
}
.sys-btn.cancel:hover { background: rgba(255,255,255,0.06); color: var(--text-primary); }
.sys-btn.glow-primary { background: linear-gradient(135deg, var(--color-secondary), var(--color-primary)); color: white; border-color: transparent; }
</style>
