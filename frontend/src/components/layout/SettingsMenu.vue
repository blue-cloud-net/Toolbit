<script setup lang="ts">
/**
 * 设置菜单。
 *
 * 主题与语言都收进这里；顶栏只保留一个齿轮图标。
 * Esc 关闭、点击外部关闭，关闭后焦点回到触发按钮。
 * 末尾的「清理缓存」用于把全部本地偏好恢复为默认值。
 */
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { ACCENT_PRESETS, type AccentKey } from '@/config/accents'
import { SUPPORTED_LOCALES, type Locale } from '@/i18n'
import { usePrefsStore, type ThemeMode } from '@/stores/prefs'

const { t } = useI18n()
const prefs = usePrefsStore()

const open = ref(false)
const trigger = ref<HTMLButtonElement | null>(null)
const panel = ref<HTMLElement | null>(null)
/** 二次确认区；展开后把焦点移进来，避免键盘用户丢失位置。 */
const confirmBox = ref<HTMLElement | null>(null)

/** 是否已展开「清理缓存」的二次确认。 */
const confirmingClear = ref(false)
/** 清理进行中，避免重复触发。 */
const clearing = ref(false)

const themeOptions = computed<{ value: ThemeMode; label: string }[]>(() => [
  { value: 'system', label: t('common.theme.system') },
  { value: 'light', label: t('common.theme.light') },
  { value: 'dark', label: t('common.theme.dark') },
])

/** 主色选项：色块网格，色值与 tokens.css 一致。 */
const accentOptions = computed(() =>
  ACCENT_PRESETS.map((preset) => ({
    value: preset.key,
    label: t(`common.accent.${preset.key}`),
    swatch: preset.swatch,
  })),
)

const LOCALE_LABELS: Record<Locale, string> = {
  'zh-CN': '简体中文',
  'en-US': 'English',
}

const localeOptions = SUPPORTED_LOCALES.map((value) => ({ value, label: LOCALE_LABELS[value] }))

const currentTheme = computed<ThemeMode>(() => prefs.theme)
const currentAccent = computed<AccentKey>(() => prefs.accent)
const currentLocale = computed<Locale>(() => prefs.locale)

function toggle(): void {
  open.value = !open.value
  if (!open.value) {
    confirmingClear.value = false
  }
}

function close(): void {
  open.value = false
  confirmingClear.value = false
  trigger.value?.focus()
}

function chooseTheme(value: ThemeMode): void {
  prefs.setTheme(value)
}

function chooseAccent(value: AccentKey): void {
  prefs.setAccent(value)
}

function chooseLocale(value: Locale): void {
  prefs.setLocale(value)
}

/** 展开二次确认：清理会丢掉收藏与最近使用，不能一点就执行。 */
function askClear(): void {
  confirmingClear.value = true
  void nextTick(() => confirmBox.value?.focus())
}

/** 收起二次确认。 */
function cancelClear(): void {
  confirmingClear.value = false
}

/** 确认清理：重置全部本地偏好后重新加载页面。 */
async function confirmClear(): Promise<void> {
  if (clearing.value) {
    return
  }
  clearing.value = true
  try {
    await prefs.clearLocalData()
  } finally {
    clearing.value = false
  }
  reloadPage()
}

/**
 * 重新加载页面。
 *
 * 页面内还残留工具输入与分栏比例等内存态，只有重载才能一并回到默认。
 */
function reloadPage(): void {
  if (typeof window === 'undefined') {
    return
  }
  window.location.reload()
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
          {{ t('common.settings.accent') }}
        </legend>
        <div class="settings__swatches">
          <button
            v-for="opt in accentOptions"
            :key="opt.value"
            type="button"
            class="settings__swatch"
            :class="{ 'settings__swatch--active': currentAccent === opt.value }"
            :style="{ '--swatch': opt.swatch }"
            :aria-pressed="currentAccent === opt.value"
            :title="opt.label"
            @click="chooseAccent(opt.value)"
          >
            <AppIcon
              v-if="currentAccent === opt.value"
              name="check"
              :size="14"
            />
            <span class="tb-sr-only">{{ opt.label }}</span>
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

      <fieldset class="settings__group settings__group--last">
        <legend class="settings__legend">
          {{ t('common.settings.data') }}
        </legend>
        <p class="settings__note">
          {{ t('common.settings.dataHint') }}
        </p>
        <button
          v-if="!confirmingClear"
          type="button"
          class="settings__option settings__option--danger"
          @click="askClear"
        >
          <AppIcon
            name="refresh"
            :size="15"
          />
          <span>{{ t('common.settings.clearCache') }}</span>
        </button>
        <div
          v-else
          ref="confirmBox"
          class="settings__confirm"
          tabindex="-1"
        >
          <p class="settings__note settings__note--warn">
            {{ t('common.settings.clearConfirm') }}
          </p>
          <div class="settings__confirm-actions">
            <button
              type="button"
              class="settings__action"
              @click="cancelClear"
            >
              {{ t('common.settings.clearCancel') }}
            </button>
            <button
              type="button"
              class="settings__action settings__action--danger"
              :disabled="clearing"
              @click="confirmClear"
            >
              {{ t('common.settings.clearOk') }}
            </button>
          </div>
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
  width: 260px;
  padding: var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  box-shadow: var(--tb-shadow-lg);
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

.settings__group--last {
  margin-bottom: 0;
  padding-top: var(--tb-space-2);
  border-top: 1px solid var(--tb-border);
}

.settings__note {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  line-height: 1.5;
  color: var(--tb-text-tertiary);
}

.settings__note--warn {
  margin-bottom: var(--tb-space-1);
  color: var(--tb-danger);
}

.settings__option--danger {
  color: var(--tb-danger);
}

.settings__option--danger:hover {
  color: var(--tb-danger);
  background: color-mix(in srgb, var(--tb-danger) 12%, transparent);
}

.settings__confirm-actions {
  display: flex;
  gap: var(--tb-space-2);
}

.settings__confirm:focus {
  outline: none;
}

.settings__action {
  flex: 1;
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.settings__action:hover {
  color: var(--tb-text);
}

.settings__action--danger {
  color: #ffffff;
  background: var(--tb-danger);
  border-color: var(--tb-danger);
}

.settings__action--danger:hover {
  color: #ffffff;
  filter: brightness(1.08);
}

.settings__action:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* 主色色块网格：4 列 × 3 行 */
.settings__swatches {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--tb-space-2);
}

.settings__swatch {
  display: grid;
  place-items: center;
  height: 34px;
  color: #ffffff;
  background: var(--swatch);
  border: 2px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition:
    border-color var(--tb-transition),
    transform var(--tb-transition);
}

.settings__swatch:hover {
  transform: translateY(-1px);
}

.settings__swatch--active {
  border-color: var(--tb-text);
  box-shadow: inset 0 0 0 2px var(--tb-bg-elevated);
}

.settings__swatch:focus-visible {
  outline: 3px solid var(--tb-primary);
  outline-offset: 2px;
}
</style>
