<script setup lang="ts">
import { ref } from 'vue';
import { useLocale, usePersistentRef } from '../composables';
import { generateImage } from '../api';
import ErrorAlert from './ErrorAlert.vue';
import MediaPreview from './MediaPreview.vue';
import { Sparkles, RefreshCw, LayoutGrid, Layout, Compass } from '@lucide/vue';

const { t, locale } = useLocale();

const prompt = usePersistentRef<string>('metaai_image_prompt', '');
const orientation = usePersistentRef<'SQUARE' | 'LANDSCAPE' | 'VERTICAL'>('metaai_image_orientation', 'SQUARE');

const loading = ref(false);
const resultImages = usePersistentRef<string[]>('metaai_image_results', []);
const error = ref('');

async function handleGenerate() {
  if (!prompt.value.trim() || loading.value) return;
  loading.value = true;
  error.value = '';
  resultImages.value = [];
  
  try {
    const res = await generateImage({
      prompt: prompt.value,
      orientation: orientation.value
    });
    if (res.success && res.image_urls && res.image_urls.length > 0) {
      resultImages.value = res.image_urls;
    } else {
      throw new Error(res.error || 'Failed to generate image');
    }
  } catch (err: any) {
    error.value = err.message || t.value.apiError;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="tab-content animate-fade-in">
    <div class="image-layout">
      <!-- Input Panel -->
      <div class="panel-input glass">
        <h3 class="panel-title">{{ t.imageGen }}</h3>

        <form @submit.prevent="handleGenerate" class="gen-form">
          <!-- Prompt field -->
          <div class="field-group">
            <label class="field-label">{{ t.prompt }}</label>
            <textarea
              v-model="prompt"
              :placeholder="locale === 'ar' ? 'اكتب وصفاً مفصلاً للصورة التي تريد توليدها...' : 'Describe what you want to generate in detail...'"
              class="prompt-input"
              :disabled="loading"
            ></textarea>
          </div>

          <!-- Orientation options -->
          <div class="field-group">
            <label class="field-label">{{ t.orientation }}</label>
            <div class="orientation-selector">
              <!-- Square -->
              <label class="orientation-option" :class="{ active: orientation === 'SQUARE' }">
                <input
                  v-model="orientation"
                  type="radio"
                  value="SQUARE"
                  name="orientation"
                  class="sr-only"
                />
                <LayoutGrid class="orientation-icon" />
                <span>{{ t.square }} (1:1)</span>
              </label>

              <!-- Landscape -->
              <label class="orientation-option" :class="{ active: orientation === 'LANDSCAPE' }">
                <input
                  v-model="orientation"
                  type="radio"
                  value="LANDSCAPE"
                  name="orientation"
                  class="sr-only"
                />
                <Layout class="orientation-icon landscape-rotate" />
                <span>{{ t.landscape }} (16:9)</span>
              </label>

              <!-- Vertical -->
              <label class="orientation-option" :class="{ active: orientation === 'VERTICAL' }">
                <input
                  v-model="orientation"
                  type="radio"
                  value="VERTICAL"
                  name="orientation"
                  class="sr-only"
                />
                <Layout class="orientation-icon" />
                <span>{{ t.vertical }} (9:16)</span>
              </label>
            </div>
          </div>

          <button type="submit" class="submit-btn glow-primary" :disabled="loading || !prompt.trim()">
            <Sparkles v-if="!loading" class="btn-icon" />
            <RefreshCw v-else class="btn-icon spin-icon" />
            <span>{{ loading ? t.generating : t.generate }}</span>
          </button>
        </form>
      </div>

      <!-- Result Panel -->
      <div class="panel-result glass">
        <h3 class="panel-title">{{ t.results }}</h3>

        <!-- Error Banner -->
        <ErrorAlert :message="error" @close="error = ''" />

        <div class="result-display">
          <!-- Loading State -->
          <div v-if="loading" class="result-placeholder">
            <RefreshCw class="placeholder-icon spin-icon" />
            <p>{{ t.generating }}</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="resultImages.length === 0" class="result-placeholder">
            <Compass class="placeholder-icon pulse-glow" />
            <p>{{ locale === 'ar' ? 'ستظهر الصور المولدة هنا' : 'Your generated images will appear here' }}</p>
          </div>

          <!-- Display Images -->
          <div v-else class="results-grid">
            <MediaPreview
              v-for="(url, idx) in resultImages"
              :key="idx"
              :src="url"
              type="image"
              :title="prompt"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tab-content {
  height: 100%;
}

.image-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  height: 100%;
}

@media (max-width: 768px) {
  .image-layout {
    grid-template-columns: 1fr;
    overflow-y: auto;
  }
}

.panel-input, .panel-result {
  padding: 24px;
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
}

.panel-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 12px;
}

.gen-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex: 1;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.prompt-input {
  width: 100%;
  height: 120px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 15px;
  padding: 14px;
  border-radius: var(--radius-md);
  outline: none;
  resize: none;
  transition: var(--transition-smooth);
}

.prompt-input:focus:not(:disabled) {
  border-color: var(--color-primary);
  box-shadow: 0 0 10px var(--color-primary-glow);
}

/* Orientation Selector */
.orientation-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.orientation-option {
  border: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.01);
  padding: 14px 10px;
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: var(--transition-smooth);
}

.orientation-option:hover {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.15);
}

.orientation-option.active {
  background: rgba(59, 130, 246, 0.08);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.orientation-icon {
  width: 20px;
  height: 20px;
}

.landscape-rotate {
  transform: rotate(90deg);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

.submit-btn {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: white;
  border: none;
  font-family: inherit;
  font-size: 15px;
  font-weight: 600;
  padding: 14px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  transition: var(--transition-smooth);
  margin-top: auto;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.1);
}

.submit-btn:disabled {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  cursor: not-allowed;
  box-shadow: none;
}

.btn-icon {
  width: 16px;
  height: 16px;
}

.spin-icon {
  width: 16px;
  height: 16px;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Result Display */
.result-display {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.result-placeholder {
  margin: auto;
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.placeholder-icon {
  width: 48px;
  height: 48px;
  color: rgba(255, 255, 255, 0.1);
  margin-bottom: 8px;
}

.results-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}
</style>
