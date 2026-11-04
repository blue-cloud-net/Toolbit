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

/**
 * 开发模式代理的后端地址。
 *
 * 默认指向本机 30234（与后端默认监听端口一致）；
 * 指向别处时用 `TOOLBIT_API_TARGET` 覆盖。
 */
const API_TARGET = process.env.TOOLBIT_API_TARGET ?? 'http://127.0.0.1:30234'

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
        target: API_TARGET,
        changeOrigin: true,
        /*
         * 流式工具（`ssl-check`）走 WebSocket：`/api/v1/tools/<slug>/stream`。
         * 不开 `ws` 时升级请求不会转发，页面会一直停在「连接中」。
         */
        ws: true,
        /*
         * 浏览器发起的 POST 会带上 `Origin: http://127.0.0.1:5173`，
         * 而后端的 Origin 校验要求「Origin 的 host 与请求 Host 一致」。
         * 代理转发时 Host 已是目标地址，Origin 却是开发服务器地址，
         * 因此 POST 一律被判 `CROSS_ORIGIN_BLOCKED`。
         *
         * 这里把 Origin 一并改写成目标地址，与 `changeOrigin` 对齐。
         */
        headers: {
          Origin: API_TARGET,
        },
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
