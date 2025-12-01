import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
  // plugins: [vue()],
  plugins: [vue(), wails("./bindings")],
  build: {
    rollupOptions: {
      external: ['@wailsio/runtime'],
      output: {
        entryFileNames: `assets/[name].js`,
        chunkFileNames: `assets/[name].js`,
        assetFileNames: `assets/[name].[ext]`
      }
    }
  }
})