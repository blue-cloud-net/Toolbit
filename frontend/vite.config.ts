import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import unocss from 'unocss/vite'
import { defineConfig } from 'vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

/**
 * 工具组件按 `tools/<category>/<slug>/` 拆成独立 chunk，
 * 同时对齐「单工具 chunk < 40 KB（gzip）」的体积预算。
 */
const TOOL_CHUNK_RE = /\/src\/tools\/[a-z0-9-]+\/([a-z0-9-]+)\//

export default defineConfig({
  plugins: [
    vue(),
    unocss(),
    Components({
      dts: 'components.d.ts',
      resolvers: [ElementPlusResolver({ importStyle: 'css' })],
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:30234',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2022',
    sourcemap: false,
    chunkSizeWarningLimit: 512,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          const tool = TOOL_CHUNK_RE.exec(id)
          if (tool) {
            return `tool-${tool[1]}`
          }
          if (id.includes('node_modules/element-plus')) {
            return 'element-plus'
          }
          if (id.includes('node_modules/vue-i18n')) {
            return 'vue-i18n'
          }
          if (id.includes('node_modules')) {
            return 'vendor'
          }
          return undefined
        },
      },
    },
  },
})
