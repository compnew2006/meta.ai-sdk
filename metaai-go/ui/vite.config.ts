import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  base: '/ui/',
  server: {
    proxy: {
      '/chat': 'http://localhost:8000',
      '/upload': 'http://localhost:8000',
      '/analyze': 'http://localhost:8000',
      '/image': 'http://localhost:8000',
      '/video': 'http://localhost:8000',
    }
  }
})
