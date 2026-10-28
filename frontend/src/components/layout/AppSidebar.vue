<script setup lang="ts">
/**
 * 侧栏导航。
 *
 * 结构：
 *   1. 三个特殊入口：全部工具 / 最近使用 / 我的收藏；
 *   2. 分类分组：分类可折叠，展开后列出该分类下的工具；
 *   3. 折叠态：收成图标轨道，只保留图标与分类首字。
 */
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import AppIcon, { type IconName } from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'
import { isImplemented } from '@/tools'

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
const CATEGORY_ICONS: Record<string, IconName> = {
  lock: 'lock',
  key: 'key',
  certificate: 'certificate',
  code: 'code',
  indent: 'indent',
  text: 'text',
  data: 'data',
  monitor: 'monitor',
  connection: 'connection',
  document: 'document',
}

function categoryIcon(icon: string): IconName {
  return CATEGORY_ICONS[icon] ?? 'document'
}

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
</script>

<template>
  <nav
    class="sidebar"
    :class="{ 'sidebar--collapsed': isCollapsed }"
    :aria-label="t('common.nav.sidebar')"
  >
    <div class="sidebar__scroll">
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
          v-for="cat in categories"
          :key="cat.slug"
          class="sidebar__group"
        >
          <button
            v-if="!isCollapsed"
            type="button"
            class="sidebar__category"
            :aria-expanded="isExpanded(cat.slug)"
            @click="toggleCategory(cat.slug)"
          >
            <span class="sidebar__chip sidebar__chip--category">
              <AppIcon
                :name="categoryIcon(cat.icon)"
                :size="16"
              />
            </span>
            <span class="sidebar__label">{{ categoryName(cat.slug, cat.name) }}</span>
            <span class="sidebar__count">{{ catalog.getToolsByCategory(cat.slug).length }}</span>
            <AppIcon
              class="sidebar__chevron"
              :class="{ 'sidebar__chevron--open': isExpanded(cat.slug) }"
              name="chevron"
              :size="13"
            />
          </button>

          <RouterLink
            v-else
            class="sidebar__category sidebar__category--rail"
            :to="{ name: 'category', params: { category: cat.slug } }"
            :title="categoryName(cat.slug, cat.name)"
          >
            {{ categoryInitial(cat.slug, cat.name) }}
          </RouterLink>

          <ul
            v-if="!isCollapsed && isExpanded(cat.slug)"
            class="sidebar__tools"
          >
            <li
              v-for="tool in catalog.getToolsByCategory(cat.slug)"
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
</template>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  position: sticky;
  top: var(--tb-header-h);
  flex: 0 0 236px;
  width: 236px;
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
  flex: 0 0 56px;
  width: 56px;
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
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--tb-text-tertiary);
  text-transform: uppercase;
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
  background: var(--tb-bg-sunken);
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
  background: var(--tb-bg-sunken);
  text-decoration: none;
}

.sidebar__category--rail {
  justify-content: center;
  width: 34px;
  height: 30px;
  padding: 0;
  font-size: 11px;
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
  font-size: 11px;
  font-weight: 400;
  color: var(--tb-text-tertiary);
  text-align: right;
}

.sidebar__tools {
  display: flex;
  flex-direction: column;
  gap: 1px;
  /* 与分类名对齐：8(行内边距) + 26(图标胶囊) + 10(间距) ≈ 44 */
  margin: 2px 0 var(--tb-space-2) 33px;
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
  background: var(--tb-bg-sunken);
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
  padding: 0 var(--tb-space-1);
  font-size: 10px;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

@media (max-width: 900px) {
  .sidebar {
    position: static;
    flex: none;
    width: 100%;
    height: auto;
    padding: var(--tb-space-2) var(--tb-space-3);
    border-right: 0;
    border-bottom: 1px solid var(--tb-border);
  }

  .sidebar__scroll {
    overflow: visible;
  }

  .sidebar__shortcuts {
    flex-direction: row;
    gap: var(--tb-space-1);
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .sidebar__shortcuts::-webkit-scrollbar {
    width: 0;
    height: 0;
  }

  .sidebar__shortcut {
    white-space: nowrap;
  }

  .sidebar__section-label,
  .sidebar__categories {
    display: none;
  }
}
</style>
