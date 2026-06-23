<script setup lang="ts">
import { ref } from 'vue';
import { useAuth, useLocale } from '../composables';
import { KeyRound, ShieldAlert } from '@lucide/vue';

const { login } = useAuth();
const { t } = useLocale();

const tokenInput = ref('');
const error = ref('');

function handleSubmit() {
  if (!tokenInput.value.trim()) {
    error.value = t.value.tokenPlaceholder;
    return;
  }
  login(tokenInput.value);
}
</script>

<template>
  <div class="auth-overlay">
    <div class="auth-card glass animate-fade-in">
      <div class="header-icon-container">
        <KeyRound class="header-icon" />
      </div>
      
      <h2>{{ t.title }}</h2>
      <p class="subtitle">{{ t.enterToken }}</p>
      
      <form @submit.prevent="handleSubmit">
        <div class="input-group">
          <KeyRound class="field-icon" />
          <input
            v-model="tokenInput"
            type="password"
            :placeholder="t.tokenPlaceholder"
            class="auth-input"
            @input="error = ''"
          />
        </div>
        
        <div v-if="error" class="error-msg animate-fade-in">
          <ShieldAlert class="error-icon" />
          <span>{{ error }}</span>
        </div>
        
        <button type="submit" class="submit-btn glow-primary">
          {{ t.save }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.auth-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(4, 6, 10, 0.85);
  backdrop-filter: blur(20px);
  z-index: 1000;
  padding: 20px;
}

.auth-card {
  width: 100%;
  max-width: 440px;
  padding: 40px;
  text-align: center;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
}

.header-icon-container {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 24px;
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.4);
}

.header-icon {
  width: 32px;
  height: 32px;
  color: white;
}

h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #fff 40%, #a5b4fc 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 32px;
}

.input-group {
  position: relative;
  margin-bottom: 20px;
}

.field-icon {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  inset-inline-start: 16px;
  width: 18px;
  height: 18px;
  color: var(--text-muted);
}

.auth-input {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 15px;
  padding: 14px 16px;
  padding-inline-start: 46px;
  border-radius: var(--radius-md);
  outline: none;
  transition: var(--transition-smooth);
}

.auth-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 10px var(--color-primary-glow);
}

.error-msg {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fca5a5;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 24px;
}

.error-icon {
  width: 16px;
  height: 16px;
  color: var(--color-error);
}

.submit-btn {
  width: 100%;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: white;
  border: none;
  font-family: inherit;
  font-size: 15px;
  font-weight: 600;
  padding: 14px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: var(--transition-smooth);
}

.submit-btn:hover {
  transform: translateY(-2px);
  filter: brightness(1.1);
}

.submit-btn:active {
  transform: translateY(0);
}
</style>
