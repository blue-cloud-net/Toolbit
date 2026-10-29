<script setup lang="ts">
/**
 * 顶栏：侧栏折叠、品牌、主题图标、设置入口。
 *
 * 语言与「跟随系统」都收进设置菜单，顶栏只留两个图标按钮。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import BrandMark from '@/components/common/BrandMark.vue'
import SettingsMenu from '@/components/layout/SettingsMenu.vue'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const router = useRouter()
const prefs = usePrefsStore()

/** 当前实际生效的主题。 */
const isDark = computed(() => prefs.resolvedTheme === 'dark')

const themeToggleLabel = computed(() =>
  isDark.value ? t('common.theme.toLight') : t('common.theme.toDark'),
)

function toggleTheme(): void {
  prefs.setTheme(isDark.value ? 'light' : 'dark')
}

function toggleSidebar(): void {
  prefs.toggleSidebar()
}

function goHome(): void {
  void router.push({ name: 'home' })
}
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
  max-width: 1440px;
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
