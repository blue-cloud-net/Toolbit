<script setup lang="ts">
/**
 * 侧栏导航。
 *
 * 结构：
 *   1. 过滤框：按工具名即时过滤（窄屏下尤为重要）；
 *   2. 三个特殊入口：全部工具 / 最近使用 / 我的收藏；
 *   3. 分类分组：分类可折叠，展开后列出该分类下的工具；
 *   4. 折叠态：收成图标轨道，只保留图标与分类首字；
 *   5. 窄屏（≤900px）：转为抽屉，由顶栏按钮开合，分类树**不被隐藏**。
 */
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'
import { isImplemented } from '@/tools'
import { categoryIcon } from '@/tools/categories'

const props = defineProps<{
  /** 是否折叠。 */
  collapsed: boolean
}>()

const { t, locale } = useI18n()
const route = useRoute()
const catalog = useCatalogStore()
const prefs = usePrefsStore()

/** 是否折叠成图标轨道。 */
const isCollapsed = computed(() => props.collapsed)

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

/** catalog 的 `icon` 字段 → 内联图标名。 */
function toolName(tool: { name: string; nameEn: string }): string {
  return locale.value === 'en-US' ? tool.nameEn : tool.name
}

/** 分类首字（折叠态用）。中文取两个字，英文取两个字母。 */
function categoryInitial(slug: string, fallback: string): string {
  return categoryName(slug, fallback).slice(0, 2)
}

function isExpanded(slug: string): boolean {
  return expanded.value.includes(slug)
}

function toggleCategory(slug: string): void {
  expanded.value = isExpanded(slug)
    ? expanded.value.filter((item) => item !== slug)
    : [...expanded.value, slug]
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
    if (meta && !isExpanded(meta.category)) {
      expanded.value = [...expanded.value, meta.category]
    }
  },
  { immediate: true },
)

/** 路由变化时关闭移动端抽屉。 */
watch(
  () => route.fullPath,
  () => {
    prefs.closeMobileNav()
  },
)
</script>

<template>
  <nav
    class="sidebar"
    :class="{
      'sidebar--collapsed': isCollapsed,
      'sidebar--mobile-open': prefs.mobileNavOpen,
    }"
    :aria-label="t('common.nav.sidebar')"
  >
    <div class="sidebar__scroll">
      <label
        v-if="!isCollapsed"
        class="sidebar__filter"
      >
        <span class="tb-sr-only">{{ t('common.nav.filterTools') }}</span>
        <input
          v-model="filterText"
          type="search"
          class="sidebar__filter-input"
          :placeholder="t('common.nav.filterPlaceholder')"
        >
      </label>

      <p
        v-if="!isCollapsed"
        class="sidebar__section-label"
      >
        {{ t('common.nav.shortcuts') }}
      </p>

      <ul class="sidebar__shortcuts">
        <li
          v-for="item in shortcuts"
          :key="item.name"
        >
          <RouterLink
            class="sidebar__shortcut"
            :to="{ name: item.name }"
            :title="isCollapsed ? item.label : undefined"
          >
            <span class="sidebar__chip">
              <AppIcon :name="item.icon" />
            </span>
            <span
              v-if="!isCollapsed"
              class="sidebar__label"
            >{{ item.label }}</span>
            <span
              v-if="!isCollapsed && item.count > 0"
              class="sidebar__count"
            >{{ item.count }}</span>
          </RouterLink>
        </li>
      </ul>

      <p
        v-if="!isCollapsed"
        class="sidebar__section-label"
      >
        {{ t('common.nav.categories') }}
      </p>

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
            <span class="sidebar__chip sidebar__chip--category">
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
          >
            {{ categoryInitial(group.cat.slug, group.cat.name) }}
          </RouterLink>

          <ul
            v-if="!isCollapsed && isExpanded(group.cat.slug)"
            class="sidebar__tools"
          >
            <li
              v-for="tool in group.tools"
              :key="tool.slug"
            >
              <RouterLink
                class="sidebar__tool"
                :class="{ 'sidebar__tool--soon': !isImplemented(tool.slug) }"
                :to="{ name: 'tool', params: { slug: tool.slug } }"
              >
                <span class="sidebar__label">{{ toolName(tool) }}</span>
                <span
                  v-if="!isImplemented(tool.slug)"
                  class="sidebar__badge"
                >
                  {{ t('home.soon') }}
                </span>
              </RouterLink>
            </li>
          </ul>
        </li>
      </ul>
    </div>
  </nav>

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

.sidebar--collapsed {
  flex: 0 0 var(--tb-sidebar-rail-w);
  width: var(--tb-sidebar-rail-w);
  padding-left: var(--tb-space-2);
  overflow-x: hidden;
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
  font-size: var(--tb-font-xs);
  font-weight: 600;
  letter-spacing: 0;
  color: var(--tb-text-tertiary);
}

.sidebar__scroll > .sidebar__section-label:first-child {
  margin-top: var(--tb-space-1);
}

.sidebar__shortcuts {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.sidebar__shortcut {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px var(--tb-space-2);
  font-size: var(--tb-font-sm);
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

.sidebar__shortcut.router-link-exact-active,
.sidebar__shortcut.router-link-active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.sidebar--collapsed .sidebar__shortcut {
  justify-content: center;
  padding: 6px 0;
}

/* 图标胶囊：统一软色背景，不引入图标字体 */
.sidebar__chip {
  display: grid;
  place-items: center;
  flex: none;
  width: 26px;
  height: 26px;
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border-radius: 7px;
}

.sidebar__chip--category {
  width: 26px;
  height: 26px;
}

.sidebar__shortcut:hover .sidebar__chip,
.sidebar__category:hover .sidebar__chip {
  color: var(--tb-text);
}

.sidebar__shortcut.router-link-active .sidebar__chip {
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
  align-items: center;
}

.sidebar__category {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 5px var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  font-weight: 500;
  color: var(--tb-text-secondary);
  text-align: left;
  background: transparent;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
}

.sidebar__category:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
  text-decoration: none;
}

.sidebar__category--rail {
  justify-content: center;
  width: 34px;
  height: 30px;
  padding: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
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
  /* 与分类名对齐：左侧内边距(12) + 图标胶囊(26) + 间距(10) - 侧栏边框(2) */
  margin: 2px 0 var(--tb-space-2)
    calc(var(--tb-space-3) + var(--tb-sidebar-chip) + 10px - 2px);
  padding-left: 10px;
  border-left: 1px solid var(--tb-border);
}

.sidebar__tool {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  padding: 4px var(--tb-space-2);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
  border-radius: 7px;
}

.sidebar__tool:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
  text-decoration: none;
}

.sidebar__tool.router-link-active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
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
  padding: 5px var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  transition:
    border-color var(--tb-transition),
    box-shadow var(--tb-transition);
}

.sidebar__filter-input:focus {
  border-color: var(--tb-primary);
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--tb-primary) 22%, transparent);
}

.sidebar__filter-input::placeholder {
  color: var(--tb-text-tertiary);
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

  .sidebar--collapsed {
    width: min(280px, 85vw);
    padding-left: var(--tb-space-3);
  }

  .sidebar__scroll {
    overflow-y: auto;
  }

  .sidebar__shortcuts {
    flex-direction: column;
    gap: 3px;
    overflow-x: visible;
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
