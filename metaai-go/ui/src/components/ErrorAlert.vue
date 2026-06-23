<script setup lang="ts">
import { AlertCircle, X } from '@lucide/vue';

defineProps<{
  message: string;
}>();

defineEmits<{
  (e: 'close'): void;
}>();
</script>

<template>
  <div v-if="message" class="error-alert animate-fade-in">
    <div class="alert-content">
      <AlertCircle class="alert-icon" />
      <span class="alert-text">{{ message }}</span>
    </div>
    <button type="button" class="close-btn" @click="$emit('close')">
      <X class="close-icon" />
    </button>
  </div>
</template>

<style scoped>
.error-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  padding: 12px 16px;
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  width: 100%;
}

.alert-content {
  display: flex;
  align-items: center;
  gap: 10px;
}

.alert-icon {
  width: 18px;
  height: 18px;
  color: var(--color-error);
  flex-shrink: 0;
}

.alert-text {
  font-size: 14px;
  font-weight: 500;
  color: #fca5a5;
  line-height: 1.4;
}

.close-btn {
  background: none;
  border: none;
  color: #fca5a5;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: var(--transition-smooth);
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: white;
}

.close-icon {
  width: 16px;
  height: 16px;
}
</style>
