<script setup lang="ts">
/**
 * 快捷键说明。
 *
 * `?` 唤起，列出全局快捷键。
 */
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useShortcutHelp } from '@/composables/useGlobalHotkeys'

const { t } = useI18n()
const { open } = useShortcutHelp()

/** 快捷键清单。 */
const ITEMS = [
  { keys: ['Ctrl', '⌘', 'K'], label: 'common.shortcuts.search' },
  { keys: ['/'], label: 'common.shortcuts.focusSearch' },
  { keys: ['Ctrl', '⌘', 'Enter'], label: 'common.shortcuts.run' },
  { keys: ['Alt', '↑', '↓'], label: 'common.shortcuts.moveRecent' },
  { keys: ['?'], label: 'common.shortcuts.help' },
  { keys: ['Esc'], label: 'common.shortcuts.close' },
]
</script>

<template>
  <div
    v-if="open"
    class="shortcut-help"
    @click="open = false"
  >
    <div
      class="shortcut-help__panel"
      role="dialog"
      aria-modal="true"
      :aria-label="t('common.shortcuts.title')"
      @click.stop
    >
      <header class="shortcut-help__head">
        <strong class="shortcut-help__title">{{ t('common.shortcuts.title') }}</strong>
        <button
          type="button"
          class="shortcut-help__close"
          :aria-label="t('common.actions.close')"
          @click="open = false"
        >
          <AppIcon
            name="close"
            :size="16"
          />
        </button>
      </header>

      <dl class="shortcut-help__list">
        <div
          v-for="item in ITEMS"
          :key="item.label"
          class="shortcut-help__row"
        >
          <dt class="shortcut-help__keys">
            <kbd
              v-for="key in item.keys"
              :key="key"
              class="shortcut-help__kbd"
            >{{ key }}</kbd>
          </dt>
          <dd class="shortcut-help__label">
            {{ t(item.label) }}
          </dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<style scoped>
.shortcut-help {
  position: fixed;
  inset: 0;
  z-index: 2100;
  display: grid;
  place-items: center;
  padding: var(--tb-space-4);
  background: rgb(0 0 0 / 45%);
}

.shortcut-help__panel {
  width: 100%;
  max-width: 420px;
  padding: var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
  box-shadow: var(--tb-shadow-lg);
}

.shortcut-help__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-3);
}

.shortcut-help__title {
  font-size: var(--tb-font-base);
}

.shortcut-help__close {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  color: var(--tb-text-secondary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.shortcut-help__close:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
}

.shortcut-help__list {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  margin: 0;
}

.shortcut-help__row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--tb-space-3);
}

.shortcut-help__keys {
  display: flex;
  gap: var(--tb-space-1);
  margin: 0;
}

.shortcut-help__kbd {
  padding: 1px var(--tb-space-2);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-hover);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.shortcut-help__label {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  text-align: right;
}
</style>
