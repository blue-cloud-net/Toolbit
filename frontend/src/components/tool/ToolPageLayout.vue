<script setup lang="ts">
/**
 * 工具页骨架。
 *
 * 标题行 + 面板 + 结果区；默认单列上下排列，`wide` 工具在视口 ≥1280 时
 * 切换为左右分栏，中间的分隔条可拖拽调整比例（比例记忆在 `localStorage`）。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
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
const catalog = useCatalogStore()

const displayName = computed(() => (locale.value === 'en-US' ? props.nameEn : props.name))
const implemented = computed(() => isImplemented(props.slug))

/** 面包屑中的分类层级：从目录元信息推导，缺失时只显示首页与工具名。 */
const category = computed(() => {
  const meta = catalog.getTool(props.slug)
  if (!meta) {
    return null
  }
  return catalog.categoryMap.get(meta.category) ?? null
})

/** 分类名：优先用词条，未登记时用目录里的中文名。 */
const categoryName = computed(() => {
  const cat = category.value
  if (!cat) {
    return ''
  }
  const key = `home.categories.${cat.slug}`
  const translated = t(key)
  if (translated !== key) {
    return translated
  }
  return locale.value === 'en-US' ? cat.nameEn : cat.name
})

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

/** 使用说明与 FAQ 默认展开：折叠按钮改成标题本身，不再占一行独立小胶囊。 */
const usageOpen = ref(true)
const faqOpen = ref(true)

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
        :aria-label="t('common.nav.breadcrumb')"
      >
        <RouterLink :to="{ name: 'home' }">
          {{ t('common.nav.home') }}
        </RouterLink>
        <template v-if="category">
          <span aria-hidden="true">/</span>
          <RouterLink :to="{ name: 'category', params: { category: category.slug } }">
            {{ categoryName }}
          </RouterLink>
        </template>
        <span aria-hidden="true">/</span>
        <span aria-current="page">{{ displayName }}</span>
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
            :size="20"
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
        v-if="$slots.reference"
        class="tool-page__content"
      >
        <h2 class="tool-page__section-title">
          {{ t('tool.sections.reference') }}
        </h2>
        <slot name="reference" />
      </section>

      <section
        v-if="$slots.usage"
        class="tool-page__content"
      >
        <h2 class="tool-page__section-title">
          <button
            type="button"
            class="tool-page__section-toggle"
            :aria-expanded="usageOpen"
            aria-controls="tool-usage-body"
            @click="usageOpen = !usageOpen"
          >
            {{ t('tool.sections.usage') }}
            <AppIcon
              name="chevron"
              :size="16"
              class="tool-page__section-chevron"
              :class="{ 'tool-page__section-chevron--open': usageOpen }"
            />
          </button>
        </h2>
        <div
          v-show="usageOpen"
          id="tool-usage-body"
          class="tool-page__section-body"
        >
          <slot name="usage" />
        </div>
      </section>

      <section
        v-if="$slots.faq"
        class="tool-page__content"
      >
        <h2 class="tool-page__section-title">
          <button
            type="button"
            class="tool-page__section-toggle"
            :aria-expanded="faqOpen"
            aria-controls="tool-faq-body"
            @click="faqOpen = !faqOpen"
          >
            {{ t('tool.sections.faq') }}
            <AppIcon
              name="chevron"
              :size="16"
              class="tool-page__section-chevron"
              :class="{ 'tool-page__section-chevron--open': faqOpen }"
            />
          </button>
        </h2>
        <div
          v-show="faqOpen"
          id="tool-faq-body"
          class="tool-page__section-body"
        >
          <slot name="faq" />
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

/* 当前页加粗，与上级页拉开层次。 */
.tool-page__breadcrumb span[aria-current='page'] {
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
  width: 32px;
  height: 32px;
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
  display: flex;
  flex-direction: column;
  min-width: 0;
}

/* 面板撑满整列，左右两栏的框因此视觉上等长。
   `flex-basis: auto` 必须保留 —— 写 0 会让列的固有高度归零，面板内容被裁掉。 */
.tool-page__primary > :deep(*),
.tool-page__secondary > :deep(*) {
  flex: 1 0 auto;
  min-height: 0;
}

/* 分隔条：仅在双栏时出现，宽度须与列模板里的固定列一致。 */
.tool-page__handle {
  display: none;
  align-self: stretch;
  width: 10px;
  margin: 0;
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

/* 分区标题：标题整行可点，点击折叠该分区正文（默认展开）。 */
.tool-page__section-toggle {
  display: inline-flex;
  align-items: center;
  gap: var(--tb-space-1);
  padding: 0;
  font: inherit;
  color: inherit;
  background: none;
  border: 0;
  cursor: pointer;
}

.tool-page__section-toggle:hover {
  color: var(--tb-primary);
}

.tool-page__section-toggle:focus-visible {
  outline: 2px solid var(--tb-primary);
  outline-offset: 3px;
  border-radius: var(--tb-radius-control);
}

.tool-page__section-chevron {
  color: var(--tb-text-tertiary);
  transition: transform var(--tb-transition);
}

.tool-page__section-chevron--open {
  transform: rotate(90deg);
}

.tool-page__section-body {
  font-size: var(--tb-font-sm);
  line-height: 1.65;
  color: var(--tb-text-secondary);
}

/* 双栏仅对 `wide` 工具生效：列模板必须是「面板 · 分隔条 · 面板」三列，
   否则分隔条会占掉第二列、结果面板被挤到下一行。
   两栏等高（`stretch`），且**不**设内部滚动条 —— 内容超出时由页面整体滚动。 */
@media (min-width: 1280px) {
  .tool-page--wide .tool-page__grid {
    grid-template-columns:
      minmax(0, var(--tb-split, 1fr))
      10px
      minmax(0, var(--tb-split-rest, 1fr));
    gap: var(--tb-space-3);
    align-items: stretch;
  }

  .tool-page--wide .tool-page__handle {
    display: block;
  }
}
</style>
