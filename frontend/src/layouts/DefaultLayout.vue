<script setup lang="ts">
/**
 * 默认布局。
 *
 * 提供 header / nav / main / footer 四个地标与「跳到主内容」链接。
 */
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import AppFooter from '@/components/layout/AppFooter.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import ShortcutHelp from '@/components/common/ShortcutHelp.vue'
import CommandPalette from '@/components/search/CommandPalette.vue'
import { useGlobalHotkeys } from '@/composables/useGlobalHotkeys'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const prefs = usePrefsStore()
const catalog = useCatalogStore()
const route = useRoute()

// 全局快捷键（`/` 聚焦搜索、`?` 快捷键说明、Alt+↑/↓ 切换最近使用）。
useGlobalHotkeys()

const sidebarCollapsed = computed(() => prefs.sidebarCollapsed)

/**
 * 专注模式。
 *
 * 由 URL 查询参数 `?focus=1` 驱动 —— 参数只承载界面状态，
 * **不承载任何输入内容**（AGENTS.md §4.11）。
 */
const focusMode = computed(() => route.query.focus === '1')

onMounted(() => {
  prefs.init()
  void catalog.load()
})
</script>

<template>
  <div
    class="layout"
    :class="{ 'layout--focus': focusMode }"
  >
    <a
      class="tb-skip-link"
      href="#main-content"
    >{{ t('common.nav.skipToContent') }}</a>

    <AppHeader />

    <CommandPalette />

    <ShortcutHelp />

    <div class="layout__body">
      <AppSidebar :collapsed="sidebarCollapsed" />

      <div class="layout__content">
        <main
          id="main-content"
          class="layout__main"
          tabindex="-1"
        >
          <RouterView />
        </main>

        <AppFooter />
      </div>
    </div>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

/* 侧栏与内容同高；页脚只属于内容列，侧栏因此直通到底部 */
.layout__body {
  display: flex;
  flex: 1;
  align-items: stretch;
  min-height: 0;
}

.layout__content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.layout__main {
  flex: 1;
  width: 100%;
  max-width: var(--tb-container-w);
  margin: 0 auto;
  padding: var(--tb-space-6) var(--tb-space-6) var(--tb-space-8);
}

.layout__main:focus {
  outline: none;
}

/* 专注模式：隐藏侧栏与页脚，只留输入 / 输出。 */
.layout--focus :deep(.sidebar),
.layout--focus :deep(.footer) {
  display: none;
}

.layout--focus .layout__main {
  max-width: none;
}

@media (max-width: 900px) {
  .layout__body {
    flex-direction: column;
  }

  .layout__main {
    padding: var(--tb-space-4);
  }
}
</style>
