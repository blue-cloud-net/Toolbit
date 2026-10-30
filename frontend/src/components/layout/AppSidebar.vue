<script setup lang="ts">
/**
 * 侧栏导航。
 *
 * 结构：
 *   1. 过滤框：按工具名即时过滤（窄屏下尤为重要）；
 *   2. 三个特殊入口：全部工具 / 最近使用 / 我的收藏；
 *   3. 分类分组：分类可折叠，展开后列出该分类下的工具
 *      （带工具图标、整行选中底色含收藏按钮、左侧引导竖线）；
 *   4. 折叠态：收成图标轨道，只保留分类图标；
 *      悬浮或聚焦分类图标时在右侧弹出子菜单，展示该分类下的全部工具与一句话描述；
 *   5. 窄屏（≤900px）：转为抽屉，由顶栏按钮开合，**始终**显示完整分类树。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'
import { isImplemented } from '@/tools'
import { categoryIcon } from '@/tools/categories'
import { ensureToolDescriptions, toolDescription } from '@/tools/toolDescriptions'
import { toolIcon } from '@/tools/toolIcons'
import { categoryTone, toolTone, toneStyle } from '@/tools/tones'
import type { Tone } from '@/tools/tones'

const props = defineProps<{
  /** 是否折叠。 */
  collapsed: boolean
}>()

const { t, locale } = useI18n()
const route = useRoute()
const catalog = useCatalogStore()
const prefs = usePrefsStore()

/** 窄屏断点：与样式表及顶栏的 `900px` 保持一致。 */
const NARROW_QUERY = '(max-width: 900px)'

/** 悬浮子菜单与轨道、视口边缘的间距。 */
const FLYOUT_GAP = 8
const FLYOUT_MARGIN = 12
/** 移出后延迟关闭，给指针留出跨过轨道与菜单之间间隙的时间。 */
const FLYOUT_CLOSE_DELAY = 120

const isNarrow = ref(false)
let viewportQuery: MediaQueryList | null = null

function syncViewport(): void {
  if (typeof window === 'undefined') {
    return
  }
  isNarrow.value = window.matchMedia
    ? window.matchMedia(NARROW_QUERY).matches
    : window.innerWidth <= 900
}

/** 是否折叠成图标轨道；窄屏下侧栏是抽屉，始终按展开态渲染。 */
const isCollapsed = computed(() => props.collapsed && !isNarrow.value)

const categories = computed(() => [...catalog.categories].sort((a, b) => a.order - b.order))

/** 已展开的分类 slug。 */
const expanded = ref<string[]>([])

/** 过滤关键字。 */
const filterText = ref('')

/** 过滤后仍有工具的分类。 */
const visibleCategories = computed(() => {
  return categories.value
    .map((cat) => ({
      cat,
      tools: catalog.getToolsByCategory(cat.slug).filter((tool) => matchFilter(tool)),
    }))
    .filter((group) => group.tools.length > 0)
})

/** 过滤匹配：工具的中英文名任一命中即算匹配。 */
function matchFilter(tool: { name: string; nameEn: string }): boolean {
  const q = filterText.value.trim().toLowerCase()
  if (q === '') {
    return true
  }
  return tool.name.toLowerCase().includes(q) || tool.nameEn.toLowerCase().includes(q)
}

/** 特殊入口。 */
const shortcuts = computed(() => [
  {
    name: 'home',
    icon: 'all-tools' as const,
    label: t('common.nav.allTools'),
    count: catalog.tools.length,
  },
  {
    name: 'recent',
    icon: 'clock' as const,
    label: t('common.nav.recent'),
    count: prefs.recent.length,
  },
  {
    name: 'favorites',
    icon: 'star' as const,
    label: t('common.nav.favorites'),
    count: prefs.favorites.length,
  },
])

function categoryName(slug: string, fallback: string): string {
  const key = `home.categories.${slug}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

/** 工具显示名：按当前语言取中英文名。 */
function toolName(tool: { name: string; nameEn: string }): string {
  return locale.value === 'en-US' ? tool.nameEn : tool.name
}

/** 工具图标：未登记时回落到所属分类的图标。 */
function itemIcon(slug: string, icon: string | undefined): ReturnType<typeof toolIcon> {
  return toolIcon(slug, categoryIcon(icon))
}

/** 工具图标色调：未登记时回落到所属分类的色调。 */
function itemTone(slug: string, category: string): Tone {
  return toolTone(slug, category)
}

/** 色调对应的 CSS 变量（前景色 + 底色）。 */
function toneVars(tone: Tone): Record<string, string> {
  return toneStyle(tone)
}

/** 工具的一句话描述（未加载时为空串）。 */
function itemDescription(slug: string): string {
  return toolDescription(locale.value, slug)
}

function isExpanded(slug: string): boolean {
  return expanded.value.includes(slug)
}

function toggleCategory(slug: string): void {
  closeFlyout()
  expanded.value = isExpanded(slug)
    ? expanded.value.filter((item) => item !== slug)
    : [...expanded.value, slug]
}

/** 当前路由是否停在该工具页。 */
function isActiveTool(slug: string): boolean {
  return route.name === 'tool' && route.params.slug === slug
}

/** 侧栏可滚区；用于把选中项定位到可视区中央。 */
const scrollBox = ref<HTMLElement | null>(null)

/** 是否已做过首次定位（只在加载 / 刷新时定位一次）。 */
let initialRevealDone = false

/** 把当前选中项滚到可滚区中央。
 *
 * **只在首次加载（含刷新）时执行一次**：之后的路由切换不动滚动位置，
 * 免得用户手动滑过菜单后被强行拉回。
 */
async function revealActiveOnce(): Promise<void> {
  if (initialRevealDone) {
    return
  }
  await nextTick()
  const box = scrollBox.value
  const row = box?.querySelector<HTMLElement>('.sidebar__tool-row--active')
  if (!box || !row) {
    return
  }
  initialRevealDone = true
  const boxRect = box.getBoundingClientRect()
  const rowRect = row.getBoundingClientRect()
  // 直接算 scrollTop 并赋值：不用 scrollIntoView（它会把整页也一起滚），也不带动画。
  const centered = box.scrollTop + (rowRect.top - boxRect.top) - box.clientHeight / 2 + rowRect.height / 2
  const max = Math.max(box.scrollHeight - box.clientHeight, 0)
  box.scrollTop = Math.min(Math.max(centered, 0), max)
}

/** 是否有分类处于展开状态。 */
const hasExpanded = computed(() => expanded.value.length > 0)

/** 一键折叠按钮的文案：有展开项时提示收起，否则提示展开。 */
const collapseAllLabel = computed(() =>
  hasExpanded.value ? t('common.nav.collapseAll') : t('common.nav.expandAll'),
)

/** 一键收起 / 展开全部分类。 */
function toggleAllCategories(): void {
  expanded.value = hasExpanded.value ? [] : categories.value.map((cat) => cat.slug)
}

/** 收藏状态与切换（收藏是界面偏好，不写任何工具输入）。 */
function isFavorite(slug: string): boolean {
  return prefs.isFavorite(slug)
}

function toggleFavorite(slug: string): void {
  prefs.toggleFavorite(slug)
}

function favoriteLabel(slug: string): string {
  return isFavorite(slug) ? t('tool.favorite.remove') : t('tool.favorite.add')
}

// ── 折叠态悬浮子菜单 ────────────────────────────────────────────

/** 悬浮子菜单的位置与内容。 */
interface FlyoutState {
  /** 分类 slug。 */
  slug: string
  /** 面板顶边（视口坐标）。 */
  top: number
  /** 面板左边（视口坐标）。 */
  left: number
}

const flyout = ref<FlyoutState | null>(null)
const flyoutEl = ref<HTMLElement | null>(null)
let closeTimer: number | undefined

/** 正在弹出子菜单的分类（含过滤后的工具）。 */
const flyoutGroup = computed(
  () => visibleCategories.value.find((group) => group.cat.slug === flyout.value?.slug) ?? null,
)

/** 取消待执行的关闭。 */
function cancelClose(): void {
  if (closeTimer !== undefined) {
    window.clearTimeout(closeTimer)
    closeTimer = undefined
  }
}

/** 立即关闭子菜单。 */
function closeFlyout(): void {
  cancelClose()
  flyout.value = null
}

/** 延迟关闭：指针从轨道移入面板需要一点时间。 */
function scheduleClose(): void {
  cancelClose()
  closeTimer = window.setTimeout(() => {
    flyout.value = null
    closeTimer = undefined
  }, FLYOUT_CLOSE_DELAY)
}

/** 打开某个分类的子菜单，并预取该分类工具的描述。 */
async function openFlyout(slug: string, event: Event): Promise<void> {
  if (!isCollapsed.value) {
    return
  }
  cancelClose()

  const trigger = event.currentTarget as HTMLElement | null
  if (!trigger) {
    return
  }
  const rect = trigger.getBoundingClientRect()
  flyout.value = { slug, top: rect.top - FLYOUT_GAP, left: rect.right + FLYOUT_GAP }

  // 面板高度取决于条目数与描述是否已加载，渲染后再校正一次位置。
  await nextTick()
  const panel = flyoutEl.value
  if (!panel || flyout.value?.slug !== slug) {
    return
  }
  const maxTop = window.innerHeight - panel.offsetHeight - FLYOUT_MARGIN
  const top = Math.max(FLYOUT_MARGIN, Math.min(rect.top - FLYOUT_GAP, maxTop))
  flyout.value = { ...flyout.value, top }

  const group = visibleCategories.value.find((item) => item.cat.slug === slug)
  if (group) {
    void ensureToolDescriptions(
      locale.value,
      group.tools.map((tool) => tool.slug),
    )
  }
}

/** 点击子菜单里的条目或分类图标后关闭面板。 */
function onFlyoutNavigate(): void {
  closeFlyout()
}

/** 进入工具页时自动展开所属分类。
 *
 * 同时监听目录长度：首次进入时目录可能尚未加载完成。
 */
watch(
  [() => route.params.slug, () => catalog.tools.length],
  ([slug]) => {
    if (typeof slug !== 'string') {
      return
    }
    const meta = catalog.getTool(slug)
    if (!meta) {
      return
    }
    if (!isExpanded(meta.category)) {
      expanded.value = [...expanded.value, meta.category]
    }
    void revealActiveOnce()
  },
  { immediate: true },
)

/** 路由变化时关闭移动端抽屉与悬浮子菜单。 */
watch(
  () => route.fullPath,
  () => {
    closeFlyout()
    prefs.closeMobileNav()
  },
)

/** 回到展开态时关掉悬浮子菜单。 */
watch(isCollapsed, (collapsed) => {
  if (!collapsed) {
    closeFlyout()
  }
})

/** Escape 关闭悬浮子菜单。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    closeFlyout()
  }
}

onMounted(() => {
  if (typeof window === 'undefined') {
    return
  }
  // 两个来源都听：`change` 在部分浏览器/内嵌环境不派发，`resize` 兜底。
  window.addEventListener('resize', syncViewport)
  if (window.matchMedia) {
    viewportQuery = window.matchMedia(NARROW_QUERY)
    viewportQuery.addEventListener('change', syncViewport)
  }
  syncViewport()
})

onBeforeUnmount(() => {
  cancelClose()
  window.removeEventListener('resize', syncViewport)
  viewportQuery?.removeEventListener('change', syncViewport)
  viewportQuery = null
})
</script>

<template>
  <nav
    class="sidebar"
    :class="{
      'sidebar--collapsed': isCollapsed,
      'sidebar--mobile-open': prefs.mobileNavOpen,
    }"
    :aria-label="t('common.nav.sidebar')"
    @keydown="onKeydown"
  >
    <!-- 固定区：快捷入口、分类标题与过滤不随列表滚动 -->
    <div class="sidebar__fixed">
      <div class="sidebar__collapsible">
        <div>
          <p class="sidebar__section-label">
            {{ t('common.nav.shortcuts') }}
          </p>
        </div>
      </div>

      <ul class="sidebar__shortcuts">
        <li
          v-for="item in shortcuts"
          :key="item.name"
        >
          <RouterLink
            class="sidebar__shortcut"
            :to="{ name: item.name }"
            :title="item.label"
          >
            <span class="sidebar__chip">
              <AppIcon :name="item.icon" />
            </span>
            <span class="sidebar__label">{{ item.label }}</span>
            <span
              v-if="item.count > 0"
              class="sidebar__count"
            >{{ item.count }}</span>
          </RouterLink>
        </li>
      </ul>

      <!-- 「快捷」与「分类」之间的组分隔：拉出间距并画线 -->
      <hr class="sidebar__divider">

      <div class="sidebar__collapsible">
        <div>
          <div class="sidebar__section-head">
            <p class="sidebar__section-label">
              {{ t('common.nav.categories') }}
            </p>
            <button
              type="button"
              class="sidebar__collapse-all"
              :aria-label="collapseAllLabel"
              :title="collapseAllLabel"
              @click="toggleAllCategories"
            >
              <AppIcon
                name="chevron"
                :size="14"
                class="sidebar__chevron"
                :class="{ 'sidebar__chevron--open': hasExpanded }"
              />
            </button>
          </div>
        </div>
      </div>

      <!-- 过滤只作用于分类下的工具，因此放在分类标题下方 -->
      <div class="sidebar__collapsible">
        <div>
          <label class="sidebar__filter">
            <span class="tb-sr-only">{{ t('common.nav.filterTools') }}</span>
            <ElInput
              v-model="filterText"
              class="sidebar__filter-input"
              :placeholder="t('common.nav.filterPlaceholder')"
              :aria-label="t('common.nav.filterTools')"
              clearable
            />
          </label>
        </div>
      </div>
    </div>

    <!-- 滚动区：只让分类列表滚动 -->
    <div
      ref="scrollBox"
      class="sidebar__scroll"
      @scroll="closeFlyout"
    >
      <ul class="sidebar__categories">
        <li
          v-for="group in visibleCategories"
          :key="group.cat.slug"
          class="sidebar__group"
        >
          <button
            v-if="!isCollapsed"
            type="button"
            class="sidebar__category"
            :aria-expanded="isExpanded(group.cat.slug)"
            @click="toggleCategory(group.cat.slug)"
          >
            <span
              class="sidebar__chip sidebar__chip--category"
              :style="toneVars(categoryTone(group.cat.slug))"
            >
              <AppIcon
                :name="categoryIcon(group.cat.icon)"
                :size="16"
              />
            </span>
            <span class="sidebar__label">{{ categoryName(group.cat.slug, group.cat.name) }}</span>
            <span class="sidebar__count">{{ group.tools.length }}</span>
            <AppIcon
              class="sidebar__chevron"
              :class="{ 'sidebar__chevron--open': isExpanded(group.cat.slug) }"
              name="chevron"
              :size="13"
            />
          </button>

          <RouterLink
            v-else
            class="sidebar__category sidebar__category--rail"
            :to="{ name: 'category', params: { category: group.cat.slug } }"
            :title="categoryName(group.cat.slug, group.cat.name)"
            :aria-label="categoryName(group.cat.slug, group.cat.name)"
            @mouseenter="openFlyout(group.cat.slug, $event)"
            @mouseleave="scheduleClose"
            @focus="openFlyout(group.cat.slug, $event)"
            @blur="scheduleClose"
            @click="onFlyoutNavigate"
          >
            <span
              class="sidebar__chip sidebar__chip--category"
              :style="toneVars(categoryTone(group.cat.slug))"
            >
              <AppIcon
                :name="categoryIcon(group.cat.icon)"
                :size="16"
              />
            </span>
          </RouterLink>

          <ul
            v-if="!isCollapsed && isExpanded(group.cat.slug)"
            class="sidebar__tools"
          >
            <li
              v-for="tool in group.tools"
              :key="tool.slug"
              class="sidebar__tool-row"
              :class="{ 'sidebar__tool-row--active': isActiveTool(tool.slug) }"
            >
              <RouterLink
                class="sidebar__tool"
                :class="{ 'sidebar__tool--soon': !isImplemented(tool.slug) }"
                :to="{ name: 'tool', params: { slug: tool.slug } }"
              >
                <span
                  class="sidebar__tool-icon"
                  :style="toneVars(itemTone(tool.slug, group.cat.slug))"
                >
                  <AppIcon
                    :name="itemIcon(tool.slug, group.cat.icon)"
                    :size="16"
                  />
                </span>
                <span class="sidebar__label">{{ toolName(tool) }}</span>
                <span
                  v-if="!isImplemented(tool.slug)"
                  class="sidebar__badge"
                >
                  {{ t('home.soon') }}
                </span>
              </RouterLink>
              <button
                type="button"
                class="sidebar__favorite"
                :class="{ 'sidebar__favorite--on': isFavorite(tool.slug) }"
                :aria-pressed="isFavorite(tool.slug)"
                :aria-label="favoriteLabel(tool.slug)"
                :title="favoriteLabel(tool.slug)"
                @click="toggleFavorite(tool.slug)"
              >
                <AppIcon
                  :name="isFavorite(tool.slug) ? 'star-filled' : 'star'"
                  :size="16"
                />
              </button>
            </li>
          </ul>
        </li>
      </ul>
    </div>
  </nav>

  <!-- 折叠态悬浮子菜单：teleport 到 body，避免被侧栏的 overflow 裁掉 -->
  <Teleport to="body">
    <div
      v-if="flyout && flyoutGroup"
      ref="flyoutEl"
      class="nav-flyout"
      :style="{ top: `${flyout.top}px`, left: `${flyout.left}px` }"
      @mouseenter="cancelClose"
      @mouseleave="scheduleClose"
    >
      <p class="nav-flyout__title">
        <span
          class="nav-flyout__title-icon"
          :style="toneVars(categoryTone(flyoutGroup.cat.slug))"
        >
          <AppIcon
            :name="categoryIcon(flyoutGroup.cat.icon)"
            :size="15"
          />
        </span>
        <span class="nav-flyout__title-text">
          {{ categoryName(flyoutGroup.cat.slug, flyoutGroup.cat.name) }}
        </span>
        <span class="nav-flyout__count">{{ flyoutGroup.tools.length }}</span>
      </p>

      <ul class="nav-flyout__list">
        <li
          v-for="tool in flyoutGroup.tools"
          :key="tool.slug"
        >
          <RouterLink
            class="nav-flyout__item"
            :class="{ 'nav-flyout__item--active': isActiveTool(tool.slug) }"
            :to="{ name: 'tool', params: { slug: tool.slug } }"
            @click="onFlyoutNavigate"
          >
            <span
              class="nav-flyout__icon"
              :style="toneVars(itemTone(tool.slug, flyoutGroup.cat.slug))"
            >
              <AppIcon
                :name="itemIcon(tool.slug, flyoutGroup.cat.icon)"
                :size="16"
              />
            </span>
            <span class="nav-flyout__text">
              <span class="nav-flyout__name">{{ toolName(tool) }}</span>
              <span
                v-if="itemDescription(tool.slug)"
                class="nav-flyout__desc"
              >{{ itemDescription(tool.slug) }}</span>
            </span>
            <span
              v-if="!isImplemented(tool.slug)"
              class="nav-flyout__badge"
            >{{ t('home.soon') }}</span>
          </RouterLink>
        </li>
      </ul>
    </div>
  </Teleport>

  <button
    v-if="prefs.mobileNavOpen"
    type="button"
    class="sidebar__scrim"
    :aria-label="t('common.actions.close')"
    @click="prefs.closeMobileNav()"
  />
</template>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  position: sticky;
  top: var(--tb-header-h);
  flex: 0 0 var(--tb-sidebar-w);
  width: var(--tb-sidebar-w);
  height: calc(100vh - var(--tb-header-h));
  padding: var(--tb-space-3) var(--tb-space-2) var(--tb-space-3) var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border-right: 1px solid var(--tb-border);
  overflow: hidden;
  /* 展开 / 收起时宽度平滑过渡；内部文字靠 overflow 裁切，不参与宽度协商 */
  transition:
    flex-basis var(--tb-transition),
    width var(--tb-transition),
    padding var(--tb-transition);
}

/* 固定区：快捷入口、分类标题与过滤，不参与滚动 */
.sidebar__fixed {
  flex: none;
}

.sidebar__scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  /* 隐藏滚动条但保留滚动：分类展开后菜单会变长，滚动条会破坏侧栏观感 */
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.sidebar__scroll::-webkit-scrollbar {
  width: 0;
  height: 0;
}

/*
 * 折叠态：收成图标轨道。
 * 轨道宽 52 = 左边距 12 + 图标 28 + 右边距 11 + 右边框 1，图标因此视觉居中。
 */
.sidebar--collapsed {
  flex: 0 0 var(--tb-sidebar-rail-w);
  width: var(--tb-sidebar-rail-w);
  padding-left: var(--tb-space-3);
  padding-right: 11px;
  overflow-x: hidden;
}

/*
 * 折叠时整块淡出并收高（grid-template-rows 0fr→1fr 可过渡），
 * 避免「收起瞬间顶部内容消失、下面的图标突然上跳」。
 */
.sidebar__collapsible {
  display: grid;
  grid-template-rows: 1fr;
  opacity: 1;
  transition:
    grid-template-rows var(--tb-transition),
    opacity var(--tb-transition);
}

.sidebar__collapsible > * {
  min-height: 0;
  overflow: hidden;
}

.sidebar--collapsed .sidebar__collapsible {
  grid-template-rows: 0fr;
  opacity: 0;
}

.sidebar--collapsed .sidebar__collapsible > * {
  pointer-events: none;
}

.sidebar__shortcuts,
.sidebar__categories,
.sidebar__tools {
  margin: 0;
  padding: 0;
  list-style: none;
}

/* 分组标题：快捷 / 分类 */
.sidebar__section-label {
  margin: var(--tb-space-3) 0 var(--tb-space-1);
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  letter-spacing: 0;
  color: var(--tb-text-tertiary);
}

.sidebar__fixed > .sidebar__collapsible:first-child .sidebar__section-label {
  margin-top: var(--tb-space-1);
}

/* 「快捷」与「分类」之间的分隔线：上下留白让两组拉开距离 */
.sidebar__divider {
  margin: var(--tb-space-3) var(--tb-space-2) var(--tb-space-1);
  border: 0;
  border-top: 1px solid var(--tb-border);
}

/* 轨道态没有文字，分隔线改成整宽、两端留白一致 */
.sidebar--collapsed .sidebar__divider {
  margin: var(--tb-space-2) 0;
}

/* 分类标题行：标题 + 一键收起 / 展开全部 */
.sidebar__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin: var(--tb-space-3) 0 var(--tb-space-1);
  padding: 0 var(--tb-space-2);
}

.sidebar__section-head .sidebar__section-label {
  margin: 0;
  padding: 0;
}

.sidebar__collapse-all {
  display: grid;
  place-items: center;
  width: 22px;
  height: 22px;
  color: var(--tb-text-tertiary);
  background: transparent;
  border: 0;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: color var(--tb-transition), background var(--tb-transition);
}

.sidebar__collapse-all:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
}

.sidebar__shortcuts {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.sidebar__shortcut {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px var(--tb-space-2);
  font-size: var(--tb-font-base);
  font-weight: 500;
  color: var(--tb-text-secondary);
  border-radius: 8px;
  transition: background var(--tb-transition), color var(--tb-transition);
}

.sidebar__shortcut:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
  text-decoration: none;
}

/* 快捷入口只在对应路由本身高亮：`/` 是所有路径的前缀，
   用 `router-link-active` 会让「全部工具」在任意工具页都亮着。 */
.sidebar__shortcut.router-link-exact-active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

/* 轨道态：图标贴左（与分类图标同一条竖线），其余内容由 overflow 裁掉 */
.sidebar--collapsed .sidebar__shortcut {
  padding: 8px 0;
}

/* 图标胶囊：统一软色背景，不引入图标字体 */
.sidebar__chip {
  display: grid;
  place-items: center;
  flex: none;
  width: var(--tb-sidebar-chip);
  height: var(--tb-sidebar-chip);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border-radius: 7px;
}

.sidebar__shortcut:hover .sidebar__chip {
  color: var(--tb-text);
}

/* 分类图标胶囊：每个分类一个颜色（不随 hover 变灰） */
.sidebar__chip--category {
  color: var(--tb-tone, var(--tb-text-secondary));
  background: var(--tb-tone-soft, var(--tb-bg-sunken));
}

.sidebar__shortcut.router-link-exact-active .sidebar__chip {
  color: var(--tb-primary);
  background: var(--tb-bg-elevated);
}

.sidebar__categories {
  display: flex;
  flex-direction: column;
  gap: 1px;
  margin-top: 0;
}

.sidebar--collapsed .sidebar__categories {
  margin-top: var(--tb-space-2);
}

.sidebar__category {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 7px var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-base);
  font-weight: 500;
  color: var(--tb-text-secondary);
  text-align: left;
  background: transparent;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
  transition: background var(--tb-transition), color var(--tb-transition);
}

.sidebar__category:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
  text-decoration: none;
}

/* 轨道态分类项：宽度撑满 28px 内容区，图标位置与展开态完全一致，切换时不跳动 */
.sidebar__category--rail {
  justify-content: flex-start;
  padding: 7px 0;
  border-radius: 8px;
}

.sidebar__category--rail:focus-visible {
  outline: 2px solid var(--tb-primary);
  outline-offset: 2px;
}

.sidebar__chevron {
  color: var(--tb-text-tertiary);
  transition: transform var(--tb-transition);
}

.sidebar__chevron--open {
  transform: rotate(90deg);
}

.sidebar__label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar__count {
  flex: none;
  min-width: 2ch;
  font-size: var(--tb-font-2xs);
  font-weight: 400;
  color: var(--tb-text-tertiary);
  text-align: right;
}

.sidebar__tools {
  display: flex;
  flex-direction: column;
  gap: 1px;
  /* 比分类图标略缩进一级：左侧内边距(12) + 图标胶囊 + 间距(12) - 侧栏边框(2) - 再左移 8 */
  margin: 2px 0 var(--tb-space-2)
    calc(var(--tb-space-3) + var(--tb-sidebar-chip) + 12px - 2px - var(--tb-space-2));
}

/* 工具行：链接 + 收藏按钮并排，hover / 选中底色覆盖整行（含收藏按钮） */
.sidebar__tool-row {
  display: flex;
  align-items: center;
  gap: 2px;
  border-radius: 7px;
  transition: background var(--tb-transition);
}

.sidebar__tool-row:hover:not(.sidebar__tool-row--active) {
  background: var(--tb-bg-hover);
}

.sidebar__tool-row--active {
  background: var(--tb-primary-soft);
}

.sidebar__tool {
  display: flex;
  flex: 1;
  align-items: center;
  gap: var(--tb-space-2);
  min-width: 0;
  padding: 6px var(--tb-space-2);
  font-size: var(--tb-font-base);
  color: var(--tb-text-secondary);
  border-radius: 7px;
  transition: color var(--tb-transition);
}

.sidebar__tool:hover {
  color: var(--tb-text);
  text-decoration: none;
}

/* 工具图标：每个工具一个颜色，hover / 选中时保留自身色调 */
.sidebar__tool-icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 18px;
  color: var(--tb-tone, var(--tb-text-tertiary));
  transition: color var(--tb-transition);
}

.sidebar__tool-row:hover .sidebar__tool-icon {
  color: var(--tb-tone, var(--tb-text-secondary));
}

.sidebar__tool-row--active .sidebar__tool {
  color: var(--tb-primary);
}

.sidebar__tool-row--active .sidebar__tool-icon {
  color: var(--tb-tone, var(--tb-primary));
}

/* 收藏：默认淡出，hover / 已收藏 / 选中 / 键盘聚焦时可见 */
.sidebar__favorite {
  display: grid;
  place-items: center;
  flex: none;
  width: 26px;
  height: 26px;
  margin-right: 2px;
  color: var(--tb-text-tertiary);
  background: transparent;
  border: 0;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  opacity: 0;
  transition: opacity var(--tb-transition), color var(--tb-transition);
}

.sidebar__tool-row:hover .sidebar__favorite,
.sidebar__tool-row--active .sidebar__favorite,
.sidebar__favorite:focus-visible {
  opacity: 1;
}

.sidebar__favorite:hover {
  color: var(--tb-warning);
}

.sidebar__favorite--on {
  opacity: 1;
  color: var(--tb-warning);
}

/* 触屏没有 hover，收藏按钮常显 */
@media (hover: none) {
  .sidebar__favorite {
    opacity: 1;
  }
}

.sidebar__tool--soon {
  color: var(--tb-text-tertiary);
}

.sidebar__badge {
  flex: none;
  padding: 1px var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  line-height: 1.6;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

/* 过滤框 */
.sidebar__filter {
  display: block;
  margin: var(--tb-space-1) 0 var(--tb-space-2);
}

.sidebar__filter-input {
  width: 100%;
}

.sidebar__filter-input:focus {
  border-color: var(--tb-primary);
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--tb-primary) 22%, transparent);
}

.sidebar__filter-input::placeholder {
  color: var(--tb-text-tertiary);
}

/* ── 折叠态悬浮子菜单 ─────────────────────────────────────────── */
.nav-flyout {
  position: fixed;
  z-index: 60;
  width: 288px;
  max-height: min(480px, calc(100vh - var(--tb-space-6)));
  overflow-y: auto;
  padding: var(--tb-space-2) var(--tb-space-1);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  box-shadow: var(--tb-shadow-lg);
  animation: nav-flyout-in var(--tb-transition) both;
}

@keyframes nav-flyout-in {
  from {
    opacity: 0;
    transform: translateX(-6px);
  }

  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.nav-flyout__title {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  margin: 0 0 var(--tb-space-1);
  padding: 0 var(--tb-space-2) var(--tb-space-1);
  font-size: var(--tb-font-xs);
  font-weight: 600;
  color: var(--tb-text-tertiary);
  border-bottom: 1px solid var(--tb-border);
}

/* 悬浮子菜单标题的分类图标：与侧栏分类图标同色 */
.nav-flyout__title-icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 22px;
  height: 22px;
  color: var(--tb-tone, var(--tb-text-secondary));
  background: var(--tb-tone-soft, transparent);
  border-radius: var(--tb-radius-control);
}

.nav-flyout__title-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-flyout__count {
  font-size: var(--tb-font-2xs);
  font-weight: 400;
}

.nav-flyout__list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.nav-flyout__item {
  display: flex;
  align-items: flex-start;
  gap: var(--tb-space-2);
  padding: 6px var(--tb-space-2);
  color: var(--tb-text-secondary);
  border-radius: 7px;
  transition: background var(--tb-transition), color var(--tb-transition);
}

.nav-flyout__item:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
  text-decoration: none;
}

.nav-flyout__item--active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.nav-flyout__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 18px;
  height: 22px;
  color: var(--tb-tone, var(--tb-text-tertiary));
}

.nav-flyout__item:hover .nav-flyout__icon,
.nav-flyout__item--active .nav-flyout__icon {
  color: var(--tb-tone, inherit);
}

.nav-flyout__text {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.nav-flyout__name {
  overflow: hidden;
  font-size: var(--tb-font-sm);
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-flyout__desc {
  display: -webkit-box;
  overflow: hidden;
  font-size: var(--tb-font-2xs);
  line-height: 1.5;
  color: var(--tb-text-tertiary);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.nav-flyout__item--active .nav-flyout__desc {
  color: inherit;
  opacity: 0.8;
}

.nav-flyout__badge {
  flex: none;
  align-self: center;
  padding: 1px var(--tb-space-1);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

@media (max-width: 900px) {
  /*
   * 窄屏：侧栏转为抽屉。
   * 分类树**不再隐藏**（此前 `display: none` 导致移动端完全无法导航）。
   */
  .sidebar {
    position: fixed;
    top: var(--tb-header-h);
    bottom: 0;
    left: 0;
    z-index: 40;
    flex: none;
    width: min(280px, 85vw);
    height: auto;
    padding: var(--tb-space-3) var(--tb-space-2) var(--tb-space-3) var(--tb-space-3);
    border-right: 1px solid var(--tb-border);
    border-bottom: 0;
    box-shadow: var(--tb-shadow-lg);
    transform: translateX(-100%);
    transition: transform var(--tb-transition);
  }

  .sidebar--mobile-open {
    transform: translateX(0);
  }

  .sidebar__scroll {
    overflow-y: auto;
  }

  .sidebar__shortcut {
    white-space: nowrap;
  }
}

/* 抽屉遮罩 */
.sidebar__scrim {
  position: fixed;
  inset: var(--tb-header-h) 0 0 0;
  z-index: 35;
  background: rgb(11 15 23 / 45%);
  border: 0;
  cursor: pointer;
}

@media (min-width: 901px) {
  .sidebar__scrim {
    display: none;
  }
}
</style>
