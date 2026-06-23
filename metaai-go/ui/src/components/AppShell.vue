<script setup lang="ts">
import { ref } from 'vue';
import { useAuth, useLocale, useSystemInstruction } from '../composables';
import AuthDialog from './AuthDialog.vue';
import ChatTab from './ChatTab.vue';
import AnalyzeTab from './AnalyzeTab.vue';
import ImageTab from './ImageTab.vue';
import VideoTab from './VideoTab.vue';

// Icons
import {
  MessageCircle,
  Maximize2,
  Image as ImageIcon,
  Film,
  Languages,
  LogOut,
  Settings
} from '@lucide/vue';

const { isAuthenticated, logout } = useAuth();
const { t, locale, toggleLocale } = useLocale();
const { systemInstruction, setSystemInstruction } = useSystemInstruction();

type Tab = 'chat' | 'analyze' | 'image' | 'video';
const activeTab = ref<Tab>('chat');

// System-instructions modal state.
const showSettings = ref(false);
const settingsDraft = ref(systemInstruction.value);
function openSettings() {
  settingsDraft.value = systemInstruction.value;
  showSettings.value = true;
}
function saveSettings() {
  setSystemInstruction(settingsDraft.value.trim());
  showSettings.value = false;
}
</script>

<template>
  <!-- Force Login if not authenticated -->
  <AuthDialog v-if="!isAuthenticated" />

  <div v-else class="app-layout">
    <!-- Header -->
    <header class="app-header glass">
      <div class="logo-area">
        <div class="logo-glow"></div>
        <h1>{{ t.title }}</h1>
      </div>
      
      <div class="actions-area">
        <!-- Settings (System Instructions) -->
        <button class="header-btn" @click="openSettings" :class="{ 'has-value': !!systemInstruction }">
          <Settings class="btn-icon" />
          <span>{{ t.settings }}</span>
        </button>

        <!-- Language Switcher -->
        <button class="header-btn" @click="toggleLocale">
          <Languages class="btn-icon" />
          <span>{{ locale === 'ar' ? 'English' : 'العربية' }}</span>
        </button>

        <!-- Logout Button -->
        <button class="header-btn logout-btn" @click="logout">
          <LogOut class="btn-icon" />
          <span>{{ t.logout }}</span>
        </button>
      </div>
    </header>

    <!-- System Instructions Modal -->
    <div v-if="showSettings" class="settings-overlay" @click.self="showSettings = false">
      <div class="settings-modal glass">
        <h3 class="settings-title">{{ t.systemInstructions }}</h3>
        <p class="settings-hint">{{ t.systemInstructionsGlobalHint }}</p>
        <textarea
          v-model="settingsDraft"
          class="settings-textarea"
          :placeholder="t.systemInstructionsPlaceholder"
          rows="5"
        ></textarea>
        <div class="settings-actions">
          <button class="header-btn" @click="showSettings = false">{{ t.cancel }}</button>
          <button class="header-btn glow-primary save-btn" @click="saveSettings">{{ t.saveSettings }}</button>
        </div>
      </div>
    </div>

    <!-- Nav & Main Tabs -->
    <main class="app-main">
      <!-- Tabs Navigation -->
      <nav class="tabs-nav glass">
        <!-- Chat Tab -->
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'chat' }"
          @click="activeTab = 'chat'"
        >
          <MessageCircle class="tab-icon" />
          <span class="tab-label">{{ t.chat }}</span>
        </button>

        <!-- Analyze Tab -->
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'analyze' }"
          @click="activeTab = 'analyze'"
        >
          <Maximize2 class="tab-icon" />
          <span class="tab-label">{{ t.analyze }}</span>
        </button>

        <!-- Image Gen Tab -->
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'image' }"
          @click="activeTab = 'image'"
        >
          <ImageIcon class="tab-icon" />
          <span class="tab-label">{{ t.imageGen }}</span>
        </button>

        <!-- Video Gen Tab -->
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'video' }"
          @click="activeTab = 'video'"
        >
          <Film class="tab-icon" />
          <span class="tab-label">{{ t.videoGen }}</span>
        </button>
      </nav>

      <!-- Active Tab Content Panel -->
      <div class="tabs-content-wrapper">
        <KeepAlive>
          <component :is="
            activeTab === 'chat' ? ChatTab :
            activeTab === 'analyze' ? AnalyzeTab :
            activeTab === 'image' ? ImageTab : VideoTab
          " />
        </KeepAlive>
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 20px;
  gap: 20px;
}

.app-header {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.logo-area {
  position: relative;
  display: flex;
  align-items: center;
}

.logo-glow {
  position: absolute;
  width: 32px;
  height: 32px;
  background: var(--color-primary);
  filter: blur(15px);
  border-radius: 50%;
  inset-inline-start: -8px;
  opacity: 0.6;
}

h1 {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.5px;
  background: linear-gradient(135deg, #ffffff, #a5b4fc);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.actions-area {
  display: flex;
  gap: 12px;
}

.header-btn {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.logout-btn {
  color: #fca5a5;
  border-color: rgba(239, 68, 68, 0.15);
}

.logout-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: white;
}

.btn-icon {
  width: 16px;
  height: 16px;
}

/* Main Content Area */
.app-main {
  flex: 1;
  display: flex;
  gap: 20px;
  height: calc(100vh - 132px);
  min-height: 0;
}

@media (max-width: 1024px) {
  .app-main {
    flex-direction: column;
    height: auto;
  }
}

.tabs-nav {
  width: 260px;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 8px;
}

@media (max-width: 1024px) {
  .tabs-nav {
    width: 100%;
    flex-direction: row;
    overflow-x: auto;
    padding: 8px;
  }
}

.tab-btn {
  background: none;
  border: 1px solid transparent;
  color: var(--text-secondary);
  font-family: inherit;
  font-size: 14px;
  font-weight: 500;
  padding: 14px 18px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  text-align: start;
  width: 100%;
  transition: var(--transition-smooth);
}

@media (max-width: 1024px) {
  .tab-btn {
    width: auto;
    flex-shrink: 0;
    padding: 10px 14px;
  }
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-primary);
}

.tab-btn.active {
  background: rgba(59, 130, 246, 0.08);
  border-color: rgba(59, 130, 246, 0.25);
  color: var(--color-primary);
  font-weight: 600;
}

.tab-btn.active .tab-icon {
  color: var(--color-primary);
  filter: drop-shadow(0 0 5px var(--color-primary-glow));
}

.tab-icon {
  width: 18px;
  height: 18px;
  color: var(--text-muted);
  transition: var(--transition-smooth);
}

.tab-label {
  white-space: nowrap;
}

.tabs-content-wrapper {
  flex: 1;
  min-height: 0;
  height: 100%;
}

@media (max-width: 1024px) {
  .tabs-content-wrapper {
    height: auto;
  }
}

/* Settings button highlight when a system instruction is set */
.header-btn.has-value {
  color: var(--color-primary);
  border-color: rgba(59, 130, 246, 0.35);
}

/* System Instructions modal */
.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 20px;
}

.settings-modal {
  width: 100%;
  max-width: 520px;
  padding: 24px;
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.settings-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.settings-textarea {
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 14px;
  line-height: 1.5;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  outline: none;
  resize: vertical;
  min-height: 110px;
}

.settings-textarea:focus {
  border-color: var(--color-secondary);
  box-shadow: 0 0 10px var(--color-secondary-glow);
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.save-btn {
  background: linear-gradient(135deg, var(--color-secondary), var(--color-primary));
  color: white;
  border-color: transparent;
}
</style>
