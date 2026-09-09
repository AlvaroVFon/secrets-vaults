import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/secrets': 'http://localhost:8080',
      '/management': 'http://localhost:8080',
    },
  },
})
