<script setup lang="ts">
/**
 * 工具页骨架。
 *
 * 标题行 + 面板 + 结果区；视口 ≥1024 时切换为左右分栏，
 * 中间的分隔条可拖拽调整比例（比例记忆在 `localStorage`）。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { usePrefsStore } from '@/stores/prefs'
import { isImplemented } from '@/tools'

/** 分栏比例的持久化键（属界面偏好，非工具输入）。 */
const SPLIT_STORAGE_KEY = 'toolbit.split'
/** 分栏比例的合法区间。 */
const SPLIT_MIN = 0.25
const SPLIT_MAX = 0.75

const props = withDefaults(
  defineProps<{
    /** 工具 slug。 */
    slug: string
    /** 中文名。 */
    name: string
    /** 英文名。 */
    nameEn: string
    /** 是否双栏布局。 */
    wide?: boolean
  }>(),
  { wide: false },
)

const { t, locale } = useI18n()
const prefs = usePrefsStore()

const displayName = computed(() => (locale.value === 'en-US' ? props.nameEn : props.name))
const implemented = computed(() => isImplemented(props.slug))

const favorite = computed(() => prefs.isFavorite(props.slug))
const favoriteLabel = computed(() =>
  favorite.value ? t('tool.favorite.remove') : t('tool.favorite.add'),
)

function toggleFavorite(): void {
  prefs.toggleFavorite(props.slug)
}

/** 词条可能尚未登记（工具 chunk 未加载），缺失时返回空串。 */
function i18nOrEmpty(key: string): string {
  const full = `tools.${props.slug}.${key}`
  const translated = t(full)
  return translated === full ? '' : translated
}

const description = computed(() => i18nOrEmpty('description'))

/** 使用说明与 FAQ 默认折叠，避免长页面拖慢首屏扫读。 */
const usageOpen = ref(false)
const faqOpen = ref(false)

// ── 可拖拽分隔条 ──────────────────────────────────────────────
const grid = ref<HTMLElement | null>(null)
/** 左栏占比（0.25 ~ 0.75）。 */
const split = ref(0.5)
/** 拖拽中标记：拖拽期间禁用文本选中。 */
const dragging = ref(false)

/** 读取已保存的比例。 */
function loadSplit(): void {
  if (typeof localStorage === 'undefined') {
    return
  }
  try {
    const raw = localStorage.getItem(SPLIT_STORAGE_KEY)
    if (raw === null) {
      return
    }
    const value = Number.parseFloat(raw)
    if (Number.isFinite(value) && value >= SPLIT_MIN && value <= SPLIT_MAX) {
      split.value = value
    }
  } catch {
    // localStorage 不可用时用默认值。
  }
}

/** 写入比例到 localStorage。 */
function saveSplit(): void {
  if (typeof localStorage === 'undefined') {
    return
  }
  try {
    localStorage.setItem(SPLIT_STORAGE_KEY, String(split.value))
  } catch {
    // 存储不可用时忽略，不影响布局。
  }
}

/** 把比例应用到网格列模板。 */
function applySplit(): void {
  const el = grid.value
  if (!el) {
    return
  }
  el.style.setProperty('--tb-split', `${split.value}fr`)
  el.style.setProperty('--tb-split-rest', `${1 - split.value}fr`)
}

/** 拖拽开始。 */
function onHandlePointerDown(event: PointerEvent): void {
  dragging.value = true
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
}

/** 拖拽中：根据指针位置计算新比例。 */
function onHandlePointerMove(event: PointerEvent): void {
  if (!dragging.value) {
    return
  }
  const el = grid.value
  if (!el) {
    return
  }
  const rect = el.getBoundingClientRect()
  if (rect.width === 0) {
    return
  }
  const ratio = (event.clientX - rect.left) / rect.width
  split.value = Math.min(SPLIT_MAX, Math.max(SPLIT_MIN, ratio))
  applySplit()
}

/** 拖拽结束：持久化。 */
function onHandlePointerUp(event: PointerEvent): void {
  if (!dragging.value) {
    return
  }
  dragging.value = false
  ;(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId)
  saveSplit()
}

/** 键盘调整：左右方向键微调比例。 */
function onHandleKeydown(event: KeyboardEvent): void {
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') {
    return
  }
  event.preventDefault()
  const delta = event.key === 'ArrowLeft' ? -0.05 : 0.05
  split.value = Math.min(SPLIT_MAX, Math.max(SPLIT_MIN, split.value + delta))
  applySplit()
  saveSplit()
}

onMounted(() => {
  loadSplit()
  applySplit()
})

onBeforeUnmount(() => {
  dragging.value = false
})
</script>

<template>
  <article
    class="tool-page"
    :class="{ 'tool-page--wide': wide }"
  >
    <header class="tool-page__head">
      <nav
        class="tool-page__breadcrumb"
        :aria-label="t('common.nav.home')"
      >
        <RouterLink :to="{ name: 'home' }">
          {{ t('common.nav.home') }}
        </RouterLink>
        <span aria-hidden="true">/</span>
        <span>{{ displayName }}</span>
      </nav>

      <div class="tool-page__title-row">
        <h1 class="tool-page__title">
          {{ displayName }}
        </h1>
        <span
          v-if="!implemented"
          class="tool-page__badge"
        >{{ t('home.soon') }}</span>

        <button
          v-if="implemented"
          type="button"
          class="tool-page__favorite"
          :class="{ 'tool-page__favorite--on': favorite }"
          :aria-pressed="favorite"
          :aria-label="favoriteLabel"
          :title="favoriteLabel"
          @click="toggleFavorite"
        >
          <AppIcon
            :name="favorite ? 'star-filled' : 'star'"
            :size="17"
          />
        </button>
      </div>

      <p
        v-if="description"
        class="tool-page__desc"
      >
        {{ description }}
      </p>
    </header>

    <div
      v-if="!implemented"
      class="tool-page__soon"
    >
      <p>{{ t('home.soonHint') }}</p>
    </div>

    <template v-else>
      <div
        ref="grid"
        class="tool-page__grid"
      >
        <div class="tool-page__primary">
          <slot name="panel" />
        </div>
        <div
          class="tool-page__handle"
          role="separator"
          :aria-orientation="'vertical'"
          :aria-valuenow="Math.round(split * 100)"
          :aria-valuemin="Math.round(SPLIT_MIN * 100)"
          :aria-valuemax="Math.round(SPLIT_MAX * 100)"
          tabindex="0"
          @pointerdown="onHandlePointerDown"
          @pointermove="onHandlePointerMove"
          @pointerup="onHandlePointerUp"
          @pointercancel="onHandlePointerUp"
          @keydown="onHandleKeydown"
        />
        <div class="tool-page__secondary">
          <slot name="result" />
        </div>
      </div>

      <section
        v-if="$slots.usage"
        class="tool-page__content"
      >
        <h2
          id="tool-usage"
          class="tool-page__section-title"
        >
          {{ t('tool.sections.usage') }}
        </h2>
        <div class="tool-page__collapse">
          <button
            type="button"
            class="tool-page__collapse-toggle"
            :aria-expanded="usageOpen"
            aria-controls="tool-usage-body"
            @click="usageOpen = !usageOpen"
          >
            <AppIcon
              name="chevron"
              :size="14"
              class="tool-page__collapse-chevron"
              :class="{ 'tool-page__collapse-chevron--open': usageOpen }"
            />
            {{ usageOpen ? t('tool.sections.collapse') : t('tool.sections.expand') }}
          </button>
          <div
            v-show="usageOpen"
            id="tool-usage-body"
            class="tool-page__collapse-body"
          >
            <slot name="usage" />
          </div>
        </div>
      </section>

      <section
        v-if="$slots.faq"
        class="tool-page__content"
      >
        <h2
          id="tool-faq"
          class="tool-page__section-title"
        >
          {{ t('tool.sections.faq') }}
        </h2>
        <div class="tool-page__collapse">
          <button
            type="button"
            class="tool-page__collapse-toggle"
            :aria-expanded="faqOpen"
            aria-controls="tool-faq-body"
            @click="faqOpen = !faqOpen"
          >
            <AppIcon
              name="chevron"
              :size="14"
              class="tool-page__collapse-chevron"
              :class="{ 'tool-page__collapse-chevron--open': faqOpen }"
            />
            {{ faqOpen ? t('tool.sections.collapse') : t('tool.sections.expand') }}
          </button>
          <div
            v-show="faqOpen"
            id="tool-faq-body"
            class="tool-page__collapse-body"
          >
            <slot name="faq" />
          </div>
        </div>
      </section>
    </template>
  </article>
</template>

<style scoped>
.tool-page {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-5);
}

.tool-page__head {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.tool-page__breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--tb-space-1);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

/* 上级页与当前页拉开层次。 */
.tool-page__breadcrumb a {
  color: var(--tb-text-tertiary);
  text-decoration: none;
}

.tool-page__breadcrumb a:hover {
  color: var(--tb-text-secondary);
}

.tool-page__breadcrumb span[aria-hidden='true'] {
  color: var(--tb-border-strong);
}

.tool-page__breadcrumb span:last-child {
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.tool-page__title-row {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.tool-page__favorite {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  color: var(--tb-text-tertiary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: color var(--tb-transition), background var(--tb-transition);
}

.tool-page__favorite:hover {
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
}

.tool-page__favorite--on {
  color: var(--tb-warning);
}

.tool-page__title {
  margin: 0;
  font-size: var(--tb-font-2xl);
  font-weight: 700;
}

.tool-page__badge {
  padding: 0 var(--tb-space-2);
  font-size: 11px;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.tool-page__desc {
  margin: 0;
  color: var(--tb-text-secondary);
}

.tool-page__soon {
  padding: var(--tb-space-6) var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
  background: var(--tb-bg-elevated);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-panel);
}

.tool-page__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--tb-space-4);
}

.tool-page__primary,
.tool-page__secondary {
  min-width: 0;
}

/* 分隔条：仅在双栏时出现。 */
.tool-page__handle {
  display: none;
  align-self: stretch;
  width: 10px;
  margin: 0 calc(var(--tb-space-2) * -1);
  cursor: col-resize;
  touch-action: none;
  background: transparent;
  border-radius: var(--tb-radius-pill);
  transition: background var(--tb-transition);
}

.tool-page__handle:hover,
.tool-page__handle:focus-visible {
  background: var(--tb-primary-soft);
}

.tool-page__content {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.tool-page__section-title {
  margin: 0;
  font-size: var(--tb-font-md);
  font-weight: 600;
}

/* 折叠区块：使用说明与 FAQ。 */
.tool-page__collapse {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.tool-page__collapse-toggle {
  display: inline-flex;
  align-items: center;
  gap: var(--tb-space-1);
  align-self: flex-start;
  padding: var(--tb-space-1) var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: color var(--tb-transition), border-color var(--tb-transition);
}

.tool-page__collapse-toggle:hover {
  color: var(--tb-primary);
  border-color: var(--tb-primary);
}

.tool-page__collapse-chevron {
  transition: transform var(--tb-transition);
}

.tool-page__collapse-chevron--open {
  transform: rotate(90deg);
}

.tool-page__collapse-body {
  font-size: var(--tb-font-sm);
  line-height: 1.65;
  color: var(--tb-text-secondary);
}

@media (min-width: 1024px) {
  .tool-page__grid {
    grid-template-columns: var(--tb-split, 1fr) var(--tb-split-rest, 1fr);
    gap: var(--tb-space-3);
    align-items: start;
  }

  .tool-page__handle {
    display: block;
  }

  .tool-page__primary,
  .tool-page__secondary {
    max-height: calc(100vh - var(--tb-header-h) - 96px);
    overflow-y: auto;
    overscroll-behavior: contain;
  }
}
</style>
