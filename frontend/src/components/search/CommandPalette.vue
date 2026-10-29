<script setup lang="ts">
/**
 * 命令面板。
 *
 * 全局搜索浮层：`Ctrl/⌘ + K` 唤起，支持：
 *
 *   - `↑` / `↓` 移动选中项（首尾循环）；
 *   - `Enter` 打开选中项；
 *   - `Esc` 关闭并把焦点还给触发按钮；
 *   - `Tab` 焦点陷阱，焦点始终限制在面板内；
 *   - `role="dialog"` + `aria-activedescendant` 播报当前项。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCommandPalette, type PaletteItem } from '@/composables/useCommandPalette'

const { t, locale } = useI18n()
const router = useRouter()
const palette = useCommandPalette()

const input = ref<HTMLInputElement | null>(null)
const list = ref<HTMLElement | null>(null)
/** 上一次打开面板时的触发元素，用于关闭后归还焦点。 */
const restoreFocus = ref<HTMLElement | null>(null)

const results = computed(() => palette.results.value)
const activeIndex = computed(() => palette.activeIndex.value)

/** 选中项的 DOM id，供 `aria-activedescendant` 引用。 */
function itemId(index: number): string {
  return `command-palette-item-${index}`
}

/** 当前语言下应展示的标签。 */
function displayLabel(item: PaletteItem): string {
  return locale.value.startsWith('en') ? item.labelEn : item.label
}

/** 条目类型的中文标签。 */
function kindLabel(kind: PaletteItem['kind']): string {
  if (kind === 'tool') {
    return t('common.search.tools')
  }
  if (kind === 'category') {
    return t('common.search.categories')
  }
  return t('common.search.pages')
}

/** 打开时聚焦输入框并记录要归还焦点的元素。 */
watch(palette.open, async (next) => {
  if (!next) {
    restoreFocus.value?.focus()
    restoreFocus.value = null
    return
  }
  restoreFocus.value = document.activeElement as HTMLElement | null
  await nextTick()
  input.value?.focus()
})

/** 关键字变化时重置选中项。 */
watch(palette.keyword, () => {
  palette.activeIndex.value = 0
})

/** 打开选中的条目。 */
async function openItem(index: number): Promise<void> {
  const item = results.value[index]
  if (!item) {
    return
  }
  palette.hide()
  await router.push(item.to)
}

/** 面板内的按键处理。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    palette.hide()
    return
  }
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    palette.move(1)
    return
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    palette.move(-1)
    return
  }
  if (event.key === 'Enter') {
    event.preventDefault()
    void openItem(activeIndex.value)
    return
  }
  if (event.key === 'Tab') {
    // 焦点陷阱：面板只有一个可聚焦控件（输入框），Tab 一律拉回输入框。
    event.preventDefault()
    input.value?.focus()
  }
}

/** 点击遮罩关闭。 */
function onBackdropClick(): void {
  palette.hide()
}

// 全局快捷键由本组件独占注册（composable 会被多个组件调用）。
onMounted(() => {
  window.addEventListener('keydown', palette.onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', palette.onKeydown)
})
</script>

<template>
  <div
    v-if="palette.open.value"
    class="palette"
    @click="onBackdropClick"
  >
    <div
      class="palette__panel"
      role="dialog"
      aria-modal="true"
      :aria-label="t('common.search.hint')"
      @click.stop
      @keydown="onKeydown"
    >
      <div class="palette__input-row">
        <AppIcon
          name="search"
          :size="18"
          class="palette__icon"
        />
        <input
          ref="input"
          v-model="palette.keyword.value"
          type="search"
          class="palette__input"
          :placeholder="t('common.search.placeholder')"
          :aria-label="t('common.search.placeholder')"
          role="combobox"
          aria-expanded="true"
          aria-controls="command-palette-list"
          :aria-activedescendant="results.length ? itemId(activeIndex) : undefined"
          autocomplete="off"
          spellcheck="false"
        >
        <kbd class="palette__kbd">Esc</kbd>
      </div>

      <div
        id="command-palette-list"
        ref="list"
        class="palette__list"
        role="listbox"
        :aria-label="t('common.search.placeholder')"
      >
        <p
          v-if="palette.loading.value"
          class="palette__state"
        >
          {{ t('common.states.loading') }}
        </p>
        <p
          v-else-if="results.length === 0"
          class="palette__state"
        >
          {{ t('common.search.empty') }}
        </p>
        <button
          v-for="(item, index) in results"
          v-else
          :id="itemId(index)"
          :key="`${item.kind}-${item.label}-${index}`"
          type="button"
          role="option"
          class="palette__item"
          :class="{ 'palette__item--active': index === activeIndex }"
          :aria-selected="index === activeIndex"
          @mouseenter="palette.activeIndex.value = index"
          @click="openItem(index)"
        >
          <span class="palette__item-kind">{{ kindLabel(item.kind) }}</span>
          <span class="palette__item-label">{{ displayLabel(item) }}</span>
          <span
            v-if="!item.implemented"
            class="palette__item-badge"
          >
            {{ t('common.search.unimplemented') }}
          </span>
        </button>
      </div>

      <div class="palette__footer">
        <span class="palette__hint">
          <kbd class="palette__kbd">↑</kbd>
          <kbd class="palette__kbd">↓</kbd>
          {{ t('common.search.navigate') }}
        </span>
        <span class="palette__hint">
          <kbd class="palette__kbd">Enter</kbd>
          {{ t('common.search.open') }}
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.palette {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: calc(var(--tb-header-h) + var(--tb-space-6)) var(--tb-space-4) var(--tb-space-4);
  background: rgb(0 0 0 / 45%);
  backdrop-filter: blur(2px);
}

.palette__panel {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 620px;
  max-height: 70vh;
  overflow: hidden;
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
  box-shadow: var(--tb-shadow-lg);
}

.palette__input-row {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-3) var(--tb-space-4);
  border-bottom: 1px solid var(--tb-border);
}

.palette__icon {
  flex: none;
  color: var(--tb-text-tertiary);
}

.palette__input {
  flex: 1;
  min-width: 0;
  padding: 0;
  font-size: var(--tb-font-lg);
  color: var(--tb-text);
  background: transparent;
  border: none;
  outline: none;
}

.palette__input::placeholder {
  color: var(--tb-text-tertiary);
}

.palette__kbd {
  flex: none;
  padding: 1px var(--tb-space-2);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-hover);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.palette__list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: var(--tb-space-2);
}

.palette__state {
  margin: 0;
  padding: var(--tb-space-6) var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
}

.palette__item {
  display: flex;
  align-items: center;
  gap: var(--tb-space-3);
  width: 100%;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-size: var(--tb-font-base);
  color: var(--tb-text);
  text-align: left;
  background: transparent;
  border: none;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.palette__item--active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.palette__item-kind {
  flex: none;
  min-width: 48px;
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.palette__item-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.palette__item-badge {
  flex: none;
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.palette__footer {
  display: flex;
  gap: var(--tb-space-4);
  padding: var(--tb-space-2) var(--tb-space-4);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  border-top: 1px solid var(--tb-border);
}

.palette__hint {
  display: inline-flex;
  align-items: center;
  gap: var(--tb-space-1);
}

@media (max-width: 768px) {
  .palette {
    padding: var(--tb-space-4);
  }

  .palette__panel {
    max-height: 80vh;
  }

  .palette__footer {
    display: none;
  }
}
</style>
