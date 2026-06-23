<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue';
import { useLocale, usePersistentRef } from '../composables';
import { streamChat } from '../api';
import ErrorAlert from './ErrorAlert.vue';
import { Send, MessageSquarePlus, Activity } from '@lucide/vue';

const { t, locale } = useLocale();

const messages = usePersistentRef<{ role: 'user' | 'assistant'; text: string }[]>(
  'metaai_stream_history',
  []
);

onMounted(() => {
  scrollToBottom();
});
const prompt = ref('');
const loading = ref(false);
const error = ref('');
const chatAreaRef = ref<HTMLElement | null>(null);

// Options
const thinking = ref(false);
const instant = ref(false);
const newConversation = ref(false);

function scrollToBottom() {
  nextTick(() => {
    if (chatAreaRef.value) {
      chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight;
    }
  });
}

async function handleSend() {
  const msg = prompt.value.trim();
  if (!msg || loading.value) return;

  // Add user message
  messages.value.push({ role: 'user', text: msg });
  prompt.value = '';
  loading.value = true;
  error.value = '';

  // Add empty placeholder assistant message
  messages.value.push({ role: 'assistant', text: '' });
  const assistantMsgIndex = messages.value.length - 1;
  scrollToBottom();

  try {
    await streamChat(
      msg,
      {
        thinking: thinking.value,
        instant: instant.value,
        newConversation: newConversation.value
      },
      (chunk) => {
        // Append chunk to the last assistant message
        messages.value[assistantMsgIndex].text += chunk;
        scrollToBottom();
      },
      (err) => {
        // If error happens during stream
        error.value = err.message || t.value.apiError;
        // Clean up last empty message if stream failed immediately
        if (!messages.value[assistantMsgIndex].text) {
          messages.value.pop();
        }
        loading.value = false;
      }
    );

    // Stream finished successfully
    loading.value = false;
    if (newConversation.value) {
      // Clear history keep only the last turn
      messages.value = [
        { role: 'user', text: msg },
        { role: 'assistant', text: messages.value[assistantMsgIndex].text }
      ];
      newConversation.value = false;
    }
  } catch (err: any) {
    error.value = err.message || t.value.apiError;
    if (!messages.value[assistantMsgIndex].text) {
      messages.value.pop();
    }
    loading.value = false;
  }
}
</script>

<template>
  <div class="tab-content animate-fade-in">
    <div ref="chatAreaRef" class="chat-area glass">
      <!-- Empty State -->
      <div v-if="messages.length === 0" class="empty-chat">
        <Activity class="empty-icon pulse-glow" />
        <h3>{{ t.stream }}</h3>
        <p class="subtitle">{{ t.promptPlaceholder }}</p>
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
            <!-- Indicator during thinking / stream loader -->
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
      <!-- Options -->
      <div class="toolbar">
        <label class="option-check">
          <input v-model="thinking" type="checkbox" />
          <span class="label-text">{{ t.thinking }}</span>
        </label>
        <label class="option-check">
          <input v-model="instant" type="checkbox" />
          <span class="label-text">{{ t.instant }}</span>
        </label>
        <label class="option-check">
          <input v-model="newConversation" type="checkbox" />
          <span class="label-text">
            <MessageSquarePlus class="icon-inline" />
            {{ t.newConv }}
          </span>
        </label>
      </div>

      <!-- Form -->
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
</template>

<style scoped>
.tab-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 16px;
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
  display: flex;
  align-items: center;
  gap: 4px;
}

.icon-inline {
  width: 14px;
  height: 14px;
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
</style>
