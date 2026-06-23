<script setup lang="ts">
import { ref } from 'vue';
import { Eye, Download, Film, Image as ImageIcon } from '@lucide/vue';

defineProps<{
  src: string;
  type: 'image' | 'video';
  title?: string;
}>();

const loading = ref(true);

function handleLoad() {
  loading.value = false;
}

function downloadMedia(url: string, title?: string) {
  const a = document.createElement('a');
  a.href = url;
  a.download = title || 'download';
  a.target = '_blank';
  a.click();
}
</script>

<template>
  <div class="media-container glass animate-fade-in">
    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>
    
    <div class="media-wrapper">
      <img
        v-if="type === 'image'"
        :src="src"
        alt="Generated Media"
        class="media-element"
        @load="handleLoad"
      />
      <video
        v-else-if="type === 'video'"
        :src="src"
        controls
        autoplay
        loop
        class="media-element"
        @loadeddata="handleLoad"
      ></video>
    </div>
    
    <div class="hover-actions" :class="{ 'has-loaded': !loading }">
      <button class="action-btn" @click="downloadMedia(src, title)">
        <Download class="btn-icon" />
      </button>
      <a :href="src" target="_blank" class="action-btn">
        <Eye class="btn-icon" />
      </a>
      <span class="media-badge">
        <ImageIcon v-if="type === 'image'" class="badge-icon" />
        <Film v-else class="badge-icon" />
        {{ type.toUpperCase() }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.media-container {
  position: relative;
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
  border-radius: var(--radius-md);
  overflow: hidden;
  aspect-ratio: 16 / 9;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(8, 11, 17, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 5;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.media-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-element {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.hover-actions {
  position: absolute;
  bottom: 0;
  inset-inline: 0;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.85));
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  opacity: 0;
  transform: translateY(10px);
  transition: var(--transition-smooth);
  z-index: 10;
}

.media-container:hover .hover-actions.has-loaded {
  opacity: 1;
  transform: translateY(0);
}

.action-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: white;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  text-decoration: none;
  transition: var(--transition-smooth);
}

.action-btn:hover {
  background: var(--color-primary);
  border-color: var(--color-primary);
  transform: scale(1.1);
}

.btn-icon {
  width: 18px;
  height: 18px;
}

.media-badge {
  margin-inline-start: auto;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: #fff;
}

.badge-icon {
  width: 13px;
  height: 13px;
}
</style>
