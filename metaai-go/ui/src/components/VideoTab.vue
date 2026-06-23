<script setup lang="ts">
import { ref, onBeforeUnmount, computed } from 'vue';
import { useLocale, usePersistentRef } from '../composables';
import { generateVideoAsync, getJobStatus, extendVideo } from '../api';
import ErrorAlert from './ErrorAlert.vue';
import MediaPreview from './MediaPreview.vue';
import { Video, RefreshCw, Film, Plus } from '@lucide/vue';

const { t, locale } = useLocale();

const prompt = usePersistentRef<string>('metaai_video_prompt', '');
const loading = ref(false);
const error = ref('');

// Video outputs
const videoUrl = usePersistentRef<string>('metaai_video_url', '');
const mediaId = usePersistentRef<string>('metaai_video_media_id', '');

// Extend states
const extending = ref(false);

// Polling states
const jobStatus = ref<'queued' | 'running' | 'completed' | 'failed' | ''>('');
const jobId = ref('');
let pollInterval: any = null;

const statusText = computed(() => {
  if (!jobStatus.value) return '';
  return (t.value as any)[jobStatus.value] || jobStatus.value;
});

function startPolling(id: string) {
  jobId.value = id;
  jobStatus.value = 'queued';
  
  pollInterval = setInterval(async () => {
    try {
      const statusRes = await getJobStatus(id);
      jobStatus.value = statusRes.status;
      
      if (statusRes.status === 'completed' && statusRes.result) {
        clearInterval(pollInterval);
        loading.value = false;
        if (statusRes.result.video_urls && statusRes.result.video_urls.length > 0) {
          videoUrl.value = statusRes.result.video_urls[0];
        }
        if (statusRes.result.media_ids && statusRes.result.media_ids.length > 0) {
          mediaId.value = statusRes.result.media_ids[0];
        }
      } else if (statusRes.status === 'failed') {
        clearInterval(pollInterval);
        loading.value = false;
        error.value = statusRes.error || 'Video generation job failed';
      }
    } catch (err: any) {
      clearInterval(pollInterval);
      loading.value = false;
      error.value = err.message || t.value.apiError;
    }
  }, 2000);
}

async function handleGenerate() {
  if (!prompt.value.trim() || loading.value) return;
  loading.value = true;
  error.value = '';
  videoUrl.value = '';
  mediaId.value = '';
  jobId.value = '';
  jobStatus.value = '';
  
  if (pollInterval) {
    clearInterval(pollInterval);
  }

  try {
    const res = await generateVideoAsync({
      prompt: prompt.value
    });
    if (res.success && res.job_id) {
      startPolling(res.job_id);
    } else {
      throw new Error('Async job registration failed');
    }
  } catch (err: any) {
    error.value = err.message || t.value.apiError;
    loading.value = false;
  }
}

async function handleExtend() {
  if (!mediaId.value || extending.value) return;
  extending.value = true;
  error.value = '';

  try {
    const res = await extendVideo(mediaId.value);
    if (res.success && res.video_urls && res.video_urls.length > 0) {
      videoUrl.value = res.video_urls[0];
      if (res.media_ids && res.media_ids.length > 0) {
        mediaId.value = res.media_ids[0];
      }
    } else {
      throw new Error(res.error || 'Failed to extend video');
    }
  } catch (err: any) {
    error.value = err.message || t.value.apiError;
  } finally {
    extending.value = false;
  }
}

onBeforeUnmount(() => {
  if (pollInterval) {
    clearInterval(pollInterval);
  }
});
</script>

<template>
  <div class="tab-content animate-fade-in">
    <div class="video-layout">
      <!-- Input Panel -->
      <div class="panel-input glass">
        <h3 class="panel-title">{{ t.videoGen }}</h3>

        <form @submit.prevent="handleGenerate" class="gen-form">
          <div class="field-group">
            <label class="field-label">{{ t.prompt }}</label>
            <textarea
              v-model="prompt"
              :placeholder="locale === 'ar' ? 'اكتب وصفاً مفصلاً للحركة في الفيديو...' : 'Describe what you want to animate in detail...'"
              class="prompt-input"
              :disabled="loading"
            ></textarea>
          </div>

          <button type="submit" class="submit-btn glow-secondary" :disabled="loading || !prompt.trim()">
            <Video v-if="!loading" class="btn-icon" />
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
          <!-- Polling Loading States -->
          <div v-if="loading && jobStatus" class="result-placeholder">
            <div class="spinner-container">
              <RefreshCw class="placeholder-icon spin-icon" />
              <div class="status-badge" :class="jobStatus">
                {{ statusText }}
              </div>
            </div>
            
            <p v-if="jobId" class="job-id-text">
              {{ t.jobId }}: <code>{{ jobId }}</code>
            </p>
          </div>

          <!-- Empty State -->
          <div v-else-if="!videoUrl" class="result-placeholder">
            <Film class="placeholder-icon pulse-glow" />
            <p>{{ locale === 'ar' ? 'سيظهر الفيديو المولد هنا' : 'Your generated video will appear here' }}</p>
          </div>

          <!-- Display Video -->
          <div v-else class="video-preview-container animate-fade-in">
            <MediaPreview
              :src="videoUrl"
              type="video"
              :title="prompt"
            />
            
            <div class="actions-footer" v-if="mediaId">
              <button
                class="extend-btn glow-primary"
                :disabled="extending"
                @click="handleExtend"
              >
                <Plus v-if="!extending" class="btn-icon" />
                <RefreshCw v-else class="btn-icon spin-icon" />
                <span>{{ extending ? t.extending : t.extendVideo }}</span>
              </button>
            </div>
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

.video-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  height: 100%;
}

@media (max-width: 768px) {
  .video-layout {
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
  border-color: var(--color-secondary);
  box-shadow: 0 0 10px var(--color-secondary-glow);
}

.submit-btn {
  background: linear-gradient(135deg, var(--color-secondary), var(--color-primary));
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

.spinner-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.status-badge {
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid rgba(255, 255, 255, 0.1);
  text-transform: capitalize;
}

.status-badge.queued {
  background: rgba(245, 158, 11, 0.1);
  color: #fcd34d;
  border-color: rgba(245, 158, 11, 0.2);
}

.status-badge.running {
  background: rgba(59, 130, 246, 0.1);
  color: #93c5fd;
  border-color: rgba(59, 130, 246, 0.2);
}

.job-id-text {
  font-size: 11px;
}

.job-id-text code {
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
  color: #a5b4fc;
}

.placeholder-icon {
  width: 48px;
  height: 48px;
  color: rgba(255, 255, 255, 0.1);
  margin-bottom: 8px;
}

.video-preview-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.actions-footer {
  display: flex;
  justify-content: center;
}

.extend-btn {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: white;
  border: none;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  padding: 12px 24px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.extend-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.1);
}

.extend-btn:disabled {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  cursor: not-allowed;
  box-shadow: none;
}
</style>
