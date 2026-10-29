<script setup lang="ts">
/**
 * 顶栏：品牌、搜索、GitHub / 全屏 / 主题 / 设置。
 *
 * 布局分三段：左段与侧栏等宽（品牌贴左边缘、折叠按钮落在侧栏右边缘），
 * 中段搜索在视口内精确居中，右段图标贴视口右边缘。
 * 语言与「跟随系统」都收进设置菜单；GitHub 图标是外部链接，带可自部署的提示。
 */
import { ElTooltip } from 'element-plus/es/components/tooltip/index'
import 'element-plus/es/components/tooltip/style/css'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import BrandMark from '@/components/common/BrandMark.vue'
import SettingsMenu from '@/components/layout/SettingsMenu.vue'
import SearchTrigger from '@/components/search/SearchTrigger.vue'
import { REPO_URL } from '@/config/site'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const router = useRouter()
const prefs = usePrefsStore()

/**
 * 全屏。
 *
 * 原来的「专注模式」靠 URL 参数 `?focus=1` 隐藏侧栏与页脚，与「折叠侧栏」
 * 按钮职责重叠；现改为浏览器 Fullscreen API —— 隐藏的是浏览器外壳，
 * 侧栏折叠仍由自己的按钮负责。全屏状态由浏览器持有，不写任何存储。
 */
const isFullscreen = ref(false)

function syncFullscreen(): void {
  isFullscreen.value = typeof document !== 'undefined' && document.fullscreenElement !== null
}

/** 切换全屏；浏览器拒绝（非用户手势 / 策略禁止）时保持原状。 */
async function toggleFullscreen(): Promise<void> {
  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen()
    } else {
      await document.documentElement.requestFullscreen()
    }
  } catch {
    // 策略或权限拒绝时静默忽略，不影响其余操作。
  }
}

const fullscreenLabel = computed(() =>
  isFullscreen.value ? t('common.actions.exitFullscreen') : t('common.actions.enterFullscreen'),
)

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

/** 侧栏收起成图标轨道时，左段（品牌 + 折叠按钮）跟着收窄。 */
const railMode = computed(() => prefs.sidebarCollapsed && !isNarrow.value)

/**
 * 伸缩按钮的方向：态指向「点下去会去哪里」。
 * 展开中 → 指向左侧（收起）；已收起 → 指向右侧（展开）。
 * 窄屏下按钮开合的是抽屉，朝向含义相同。
 */
const toggleIcon = computed(() => {
  const collapsed = isNarrow.value ? !prefs.mobileNavOpen : prefs.sidebarCollapsed
  return collapsed ? 'panel-expand' : 'panel-collapse'
})

function goHome(): void {
  void router.push({ name: 'home' })
}

onMounted(() => {
  syncFullscreen()
  document.addEventListener('fullscreenchange', syncFullscreen)

  if (typeof window === 'undefined' || !window.matchMedia) {
    return
  }
  viewportQuery = window.matchMedia('(max-width: 900px)')
  viewportQuery.addEventListener('change', syncViewport)
  syncViewport()
})

onBeforeUnmount(() => {
  document.removeEventListener('fullscreenchange', syncFullscreen)
  viewportQuery?.removeEventListener('change', syncViewport)
  viewportQuery = null
})
</script>

<template>
  <header class="header">
    <div class="header__inner">
      <!-- 左段：品牌贴左（对齐侧栏菜单），折叠按钮落在侧栏右边缘 -->
      <div
        class="header__lead"
        :class="{ 'header__lead--rail': railMode }"
      >
        <a
          class="header__brand"
          href="/"
          @click.prevent="goHome"
        >
          <BrandMark :size="30" />
          <span class="header__brand-text">
            <strong>{{ t('common.appName') }}</strong>
          </span>
        </a>

        <button
          type="button"
          class="header__icon-button"
          :aria-label="t('common.nav.toggleSidebar')"
          :title="t('common.nav.toggleSidebar')"
          :aria-pressed="prefs.sidebarCollapsed"
          @click="toggleSidebar"
        >
          <AppIcon :name="toggleIcon" />
        </button>
      </div>

      <!-- 中段：搜索，在视口内居中 -->
      <div class="header__search">
        <SearchTrigger />
      </div>

      <!-- 右段：GitHub / 全屏 / 主题 / 设置，贴右边缘 -->
      <div class="header__actions">
        <ElTooltip
          placement="bottom"
          :show-after="200"
          popper-class="tb-hint"
        >
          <template #content>
            <p class="tb-hint__title">
              {{ t('common.github.label') }}
            </p>
            <p class="tb-hint__body">
              {{ t('common.github.deployHint') }}
            </p>
          </template>
          <a
            class="header__icon-button header__github"
            :href="REPO_URL"
            target="_blank"
            rel="noopener noreferrer"
            :aria-label="t('common.github.label')"
          >
            <AppIcon
              name="github"
              :size="18"
            />
          </a>
        </ElTooltip>

        <button
          type="button"
          class="header__icon-button"
          :class="{ 'header__icon-button--on': isFullscreen }"
          :aria-label="fullscreenLabel"
          :aria-pressed="isFullscreen"
          :title="fullscreenLabel"
          @click="toggleFullscreen"
        >
          <AppIcon :name="isFullscreen ? 'fullscreen-exit' : 'fullscreen'" />
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
  display: grid;
  /* `1fr auto 1fr` 让中段搜索在视口内精确居中，右段才能贴到右边缘 */
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: var(--tb-space-4);
  min-height: var(--tb-header-h);
  padding-right: var(--tb-space-4);
}

/* 左段：与侧栏等宽 —— 品牌对齐侧栏菜单，折叠按钮落在侧栏右边缘 */
.header__lead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  width: var(--tb-sidebar-w);
  max-width: 100%;
  padding-left: var(--tb-space-3);
  overflow: hidden;
  /* 与侧栏宽度过渡同步，收起时折叠按钮不会瞬间跳到最左 */
  transition: width var(--tb-transition);
}

/* 侧栏收起时左段跟着收成图标轨道宽度，品牌让位给折叠按钮 */
.header__lead--rail {
  justify-content: center;
  width: var(--tb-sidebar-rail-w);
  padding-left: 0;
}

.header__lead--rail .header__brand {
  display: none;
}

.header__icon-button {
  display: grid;
  place-items: center;
  flex: none;
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

/* GitHub 是链接而不是按钮：去掉下划线，其余外观与图标按钮一致 */
.header__github:hover {
  text-decoration: none;
}

/* 全屏开启时按钮保持激活态。 */
.header__icon-button--on {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.header__brand {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  min-width: 0;
  color: var(--tb-text);
  text-decoration: none;
}

.header__brand:hover {
  text-decoration: none;
}

.header__brand-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  line-height: 1.2;
}

/* 品牌全称较长（如 en-US 的 Blue Cloud Toolbit），收窄时省略而不是换行挤压搜索框。 */
.header__brand-text strong {
  font-size: var(--tb-font-2xl);
  font-weight: 700;
  letter-spacing: -0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header__search {
  display: flex;
  justify-self: center;
  width: min(560px, 34vw);
}

.header__actions {
  display: flex;
  align-items: center;
  justify-self: end;
  gap: var(--tb-space-2);
}

@media (max-width: 900px) {
  .header__inner {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: var(--tb-space-2);
    padding: 0 var(--tb-space-3);
  }

  .header__lead {
    justify-content: flex-start;
    width: auto;
    padding-left: 0;
  }

  .header__search {
    width: auto;
  }
}
</style>
