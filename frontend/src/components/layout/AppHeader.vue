<script setup lang="ts">
/**
 * 顶栏：侧栏折叠、品牌、主题图标、设置入口。
 *
 * 语言与「跟随系统」都收进设置菜单，顶栏只留两个图标按钮。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import BrandMark from '@/components/common/BrandMark.vue'
import SettingsMenu from '@/components/layout/SettingsMenu.vue'
import SearchTrigger from '@/components/search/SearchTrigger.vue'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const prefs = usePrefsStore()

/** 专注模式：由 URL 查询参数 `focus=1` 驱动。 */
const focusMode = computed(() => route.query.focus === '1')

/** 进入 / 退出专注模式。 */
function toggleFocus(): void {
  const query = { ...route.query }
  if (focusMode.value) {
    delete query.focus
  } else {
    query.focus = '1'
  }
  void router.push({ query })
}

/** 当前实际生效的主题。 */
const isDark = computed(() => prefs.resolvedTheme === 'dark')

const themeToggleLabel = computed(() =>
  isDark.value ? t('common.theme.toLight') : t('common.theme.toDark'),
)

function toggleTheme(): void {
  prefs.setTheme(isDark.value ? 'light' : 'dark')
}

/**
 * 汉堡按钮的双行为：
 *   - 宽屏：折叠 / 展开侧栏；
 *   - 窄屏：开合导航抽屉（侧栏在窄屏下是抽屉，见 AppSidebar）。
 */
const isNarrow = ref(false)
let viewportQuery: MediaQueryList | null = null

function syncViewport(): void {
  isNarrow.value = viewportQuery?.matches ?? false
}

function toggleSidebar(): void {
  if (isNarrow.value) {
    prefs.toggleMobileNav()
    return
  }
  prefs.toggleSidebar()
}

function goHome(): void {
  void router.push({ name: 'home' })
}

onMounted(() => {
  if (typeof window === 'undefined' || !window.matchMedia) {
    return
  }
  viewportQuery = window.matchMedia('(max-width: 900px)')
  viewportQuery.addEventListener('change', syncViewport)
  syncViewport()
})

onBeforeUnmount(() => {
  viewportQuery?.removeEventListener('change', syncViewport)
  viewportQuery = null
})
</script>

<template>
  <header class="header">
    <div class="header__inner">
      <button
        type="button"
        class="header__icon-button"
        :aria-label="t('common.nav.toggleSidebar')"
        :aria-pressed="prefs.sidebarCollapsed"
        @click="toggleSidebar"
      >
        <AppIcon name="menu" />
      </button>

      <a
        class="header__brand"
        href="/"
        @click.prevent="goHome"
      >
        <BrandMark :size="28" />
        <span class="header__brand-text">
          <strong>{{ t('common.appName') }}</strong>
          <small>{{ t('common.appTagline') }}</small>
        </span>
      </a>

      <div class="header__spacer" />

      <div class="header__search">
        <SearchTrigger />
      </div>

      <button
        type="button"
        class="header__icon-button"
        :class="{ 'header__icon-button--on': focusMode }"
        :aria-label="focusMode ? t('common.actions.exitFocus') : t('common.actions.enterFocus')"
        :aria-pressed="focusMode"
        :title="focusMode ? t('common.actions.exitFocus') : t('common.actions.enterFocus')"
        @click="toggleFocus"
      >
        <AppIcon name="focus" />
      </button>

      <button
        type="button"
        class="header__icon-button"
        :aria-label="themeToggleLabel"
        :title="themeToggleLabel"
        @click="toggleTheme"
      >
        <AppIcon :name="isDark ? 'sun' : 'moon'" />
      </button>

      <SettingsMenu />
    </div>
  </header>
</template>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 20;
  min-height: var(--tb-header-h);
  background: var(--tb-bg-elevated);
  border-bottom: 1px solid var(--tb-border);
}

.header__inner {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  min-height: var(--tb-header-h);
  max-width: var(--tb-container-w);
  margin: 0 auto;
  padding: var(--tb-space-2) var(--tb-space-6);
}

.header__icon-button {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  color: var(--tb-text-secondary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: background var(--tb-transition), color var(--tb-transition);
}

.header__icon-button:hover {
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
}

/* 专注模式开启时按钮保持激活态。 */
.header__icon-button--on {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.header__brand {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  margin-left: var(--tb-space-1);
  color: var(--tb-text);
  text-decoration: none;
}

.header__brand:hover {
  text-decoration: none;
}

.header__brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.header__brand-text small {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.header__spacer {
  flex: 1;
}

@media (max-width: 900px) {
  .header__inner {
    padding: var(--tb-space-2) var(--tb-space-4);
  }

  .header__brand-text small {
    display: none;
  }
}
</style>
