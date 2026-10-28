<script setup lang="ts">
/**
 * 设置菜单。
 *
 * 主题与语言都收进这里；顶栏只保留一个齿轮图标。
 * Esc 关闭、点击外部关闭，关闭后焦点回到触发按钮。
 */
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { SUPPORTED_LOCALES, type Locale } from '@/i18n'
import { usePrefsStore, type ThemeMode } from '@/stores/prefs'

const { t } = useI18n()
const prefs = usePrefsStore()

const open = ref(false)
const trigger = ref<HTMLButtonElement | null>(null)
const panel = ref<HTMLElement | null>(null)

const themeOptions = computed<{ value: ThemeMode; label: string }[]>(() => [
  { value: 'system', label: t('common.theme.system') },
  { value: 'light', label: t('common.theme.light') },
  { value: 'dark', label: t('common.theme.dark') },
])

const LOCALE_LABELS: Record<Locale, string> = {
  'zh-CN': '简体中文',
  'en-US': 'English',
}

const localeOptions = SUPPORTED_LOCALES.map((value) => ({ value, label: LOCALE_LABELS[value] }))

const currentTheme = computed<ThemeMode>(() => prefs.theme)
const currentLocale = computed<Locale>(() => prefs.locale)

function toggle(): void {
  open.value = !open.value
}

function close(): void {
  open.value = false
  trigger.value?.focus()
}

function chooseTheme(value: ThemeMode): void {
  prefs.setTheme(value)
}

function chooseLocale(value: Locale): void {
  prefs.setLocale(value)
}

/** 点击面板外部时关闭。 */
function onDocumentPointerDown(event: PointerEvent): void {
  const target = event.target as Node | null
  if (!target) {
    return
  }
  if (panel.value?.contains(target) || trigger.value?.contains(target)) {
    return
  }
  open.value = false
}

/** Esc 关闭。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.stopPropagation()
    close()
  }
}

watch(open, (next) => {
  if (typeof document === 'undefined') {
    return
  }
  if (next) {
    document.addEventListener('pointerdown', onDocumentPointerDown)
    document.addEventListener('keydown', onKeydown)
  } else {
    document.removeEventListener('pointerdown', onDocumentPointerDown)
    document.removeEventListener('keydown', onKeydown)
  }
})

onBeforeUnmount(() => {
  if (typeof document === 'undefined') {
    return
  }
  document.removeEventListener('pointerdown', onDocumentPointerDown)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="settings">
    <button
      ref="trigger"
      type="button"
      class="settings__trigger"
      :aria-label="t('common.settings.open')"
      aria-haspopup="dialog"
      :aria-expanded="open"
      @click="toggle"
    >
      <AppIcon name="settings" />
    </button>

    <div
      v-if="open"
      ref="panel"
      class="settings__panel"
      role="dialog"
      :aria-label="t('common.settings.title')"
    >
      <div class="settings__head">
        <strong class="settings__title">{{ t('common.settings.title') }}</strong>
        <button
          type="button"
          class="settings__close"
          :aria-label="t('common.actions.close')"
          @click="close"
        >
          <AppIcon
            name="close"
            :size="16"
          />
        </button>
      </div>

      <fieldset class="settings__group">
        <legend class="settings__legend">
          {{ t('common.settings.theme') }}
        </legend>
        <div class="settings__options">
          <button
            v-for="opt in themeOptions"
            :key="opt.value"
            type="button"
            class="settings__option"
            :class="{ 'settings__option--active': currentTheme === opt.value }"
            :aria-pressed="currentTheme === opt.value"
            @click="chooseTheme(opt.value)"
          >
            <AppIcon
              v-if="opt.value !== 'system'"
              :name="opt.value === 'dark' ? 'moon' : 'sun'"
              :size="15"
            />
            <span>{{ opt.label }}</span>
          </button>
        </div>
      </fieldset>

      <fieldset class="settings__group">
        <legend class="settings__legend">
          {{ t('common.settings.locale') }}
        </legend>
        <div class="settings__options">
          <button
            v-for="opt in localeOptions"
            :key="opt.value"
            type="button"
            class="settings__option"
            :class="{ 'settings__option--active': currentLocale === opt.value }"
            :aria-pressed="currentLocale === opt.value"
            @click="chooseLocale(opt.value)"
          >
            <AppIcon
              v-if="currentLocale === opt.value"
              name="check"
              :size="15"
            />
            <span>{{ opt.label }}</span>
          </button>
        </div>
      </fieldset>
    </div>
  </div>
</template>

<style scoped>
.settings {
  position: relative;
}

.settings__trigger,
.settings__close {
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

.settings__trigger:hover,
.settings__close:hover {
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
}

.settings__panel {
  position: absolute;
  top: calc(100% + var(--tb-space-2));
  right: 0;
  z-index: 30;
  width: 240px;
  padding: var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  box-shadow: 0 12px 32px rgb(0 0 0 / 12%);
}

.settings__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--tb-space-2);
}

.settings__title {
  font-size: var(--tb-font-sm);
}

.settings__close {
  width: 26px;
  height: 26px;
}

.settings__group {
  margin: 0 0 var(--tb-space-3);
  padding: 0;
  border: 0;
}

.settings__legend {
  margin-bottom: var(--tb-space-1);
  padding: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.settings__options {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.settings__option {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  width: 100%;
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
  text-align: left;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.settings__option:hover {
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
}

.settings__option--active {
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
  border-color: color-mix(in srgb, var(--tb-primary) 30%, transparent);
}
</style>
