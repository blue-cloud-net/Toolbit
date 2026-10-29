<script setup lang="ts">
/**
 * 搜索触发器。
 *
 * 顶栏里的搜索按钮，点击后打开命令面板；右侧提示 `Ctrl/⌘ + K` 快捷键。
 * 移动端隐藏快捷键提示，只保留图标。
 */
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCommandPalette } from '@/composables/useCommandPalette'

const { t } = useI18n()
const palette = useCommandPalette()

/** 触发按钮；打开面板时用于归还焦点。 */
const trigger = ref<HTMLButtonElement | null>(null)

/** 是否显示快捷键提示（macOS 显示 ⌘，其余显示 Ctrl）。 */
const hint = ref('Ctrl K')

/** 读取平台，判断展示 ⌘ 还是 Ctrl。 */
function detectHint(): void {
  if (typeof navigator === 'undefined') {
    return
  }
  const platform = navigator.platform || navigator.userAgent || ''
  const isMac = /mac|iphone|ipad|ipod/i.test(platform)
  hint.value = isMac ? '⌘ K' : 'Ctrl K'
}

onMounted(detectHint)
onBeforeUnmount(() => {
  hint.value = 'Ctrl K'
})

/** 打开命令面板。 */
function openPalette(): void {
  palette.show()
}
</script>

<template>
  <button
    ref="trigger"
    type="button"
    class="search-trigger"
    :aria-label="t('common.search.placeholder')"
    :title="t('common.search.hint')"
    @click="openPalette"
  >
    <AppIcon
      name="search"
      :size="16"
    />
    <span class="search-trigger__text">{{ t('common.actions.search') }}</span>
    <kbd class="search-trigger__kbd">{{ hint }}</kbd>
  </button>
</template>

<style scoped>
.search-trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--tb-space-2);
  width: 100%;
  height: 34px;
  padding: 0 var(--tb-space-2) 0 var(--tb-space-3);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-hover);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease;
}

.search-trigger:hover {
  color: var(--tb-text-secondary);
  border-color: var(--tb-border-strong);
}

/* 文字占满中间，快捷键提示靠右 —— 按钮宽度由顶栏中段给定。 */
.search-trigger__text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-trigger__kbd {
  padding: 1px var(--tb-space-2);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-sm);
}

@media (max-width: 768px) {
  .search-trigger__text,
  .search-trigger__kbd {
    display: none;
  }

  .search-trigger {
    width: 34px;
    padding: 0;
    justify-content: center;
  }
}
</style>
