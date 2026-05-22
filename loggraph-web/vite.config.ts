import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    host: true,
    proxy: {
      '/api/v1/ai': 'http://localhost:8081',
      '/api': 'http://localhost:8080',
    },
  },
})
