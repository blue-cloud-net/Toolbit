import { fileURLToPath, URL } from 'node:url'

import { defineConfig, mergeConfig } from 'vitest/config'

import viteConfig from './vite.config'

export default mergeConfig(
  viteConfig,
  defineConfig({
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      include: ['src/**/*.test.ts'],
      setupFiles: ['./src/test/setup.ts'],
      restoreMocks: true,
      // Element Plus 的组件样式是裸 CSS 导入，外部化后会由 Node 直接加载而报
      // `Unknown file extension ".css"`；交给 Vite 转换为空模块即可。
      server: { deps: { inline: [/element-plus/] } },
    },
  }),
)
