<script setup lang="ts">
/**
 * 默认布局。
 *
 * 提供 header / nav / main / footer 四个地标与「跳到主内容」链接。
 */
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

import AppFooter from '@/components/layout/AppFooter.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const prefs = usePrefsStore()
const catalog = useCatalogStore()

const sidebarCollapsed = computed(() => prefs.sidebarCollapsed)

onMounted(() => {
  prefs.init()
  void catalog.load()
})
</script>

<template>
  <div class="layout">
    <a
      class="tb-skip-link"
      href="#main-content"
    >{{ t('common.nav.skipToContent') }}</a>

    <AppHeader />

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
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--tb-space-6) var(--tb-space-6) var(--tb-space-8);
}

.layout__main:focus {
  outline: none;
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
