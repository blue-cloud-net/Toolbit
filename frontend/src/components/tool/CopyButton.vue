<script setup lang="ts">
/**
 * 复制按钮。
 *
 * 优先使用 `navigator.clipboard`，不可用时回退到隐藏 textarea + `execCommand`。
 */
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
  defineProps<{
    /** 要复制的文本。 */
    text: string
    /** 无障碍标签；缺省使用「复制」。 */
    label?: string
    /** 尺寸；small 用于表格内的紧凑场景。 */
    size?: 'default' | 'small'
  }>(),
  { label: undefined, size: 'default' },
)

const { t } = useI18n()
const copied = ref(false)
const failed = ref(false)

let timer: ReturnType<typeof setTimeout> | null = null

/** 回退复制实现。 */
function fallbackCopy(text: string): boolean {
  if (typeof document === 'undefined') {
    return false
  }
  const area = document.createElement('textarea')
  area.value = text
  area.setAttribute('readonly', '')
  area.style.position = 'fixed'
  area.style.top = '-1000px'
  document.body.appendChild(area)
  area.select()
  let ok = false
  try {
    ok = document.execCommand('copy')
  } catch {
    ok = false
  }
  document.body.removeChild(area)
  return ok
}

/** 执行复制。 */
async function copy(): Promise<void> {
  if (props.text === '') {
    return
  }
  let ok = false
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(props.text)
      ok = true
    } else {
      ok = fallbackCopy(props.text)
    }
  } catch {
    ok = fallbackCopy(props.text)
  }

  copied.value = ok
  failed.value = !ok

  if (timer !== null) {
    clearTimeout(timer)
  }
  timer = setTimeout(() => {
    copied.value = false
    failed.value = false
    timer = null
  }, 2000)
}
</script>

<template>
  <button
    type="button"
    class="copy-button"
    :class="{ 'copy-button--small': size === 'small' }"
    :aria-label="label ?? t('common.actions.copy')"
    :disabled="text === ''"
    @click="copy"
  >
    <span
      v-if="copied"
      class="copy-button--ok"
    >{{ t('common.actions.copied') }}</span>
    <span v-else-if="failed">{{ t('tool.copy.failed') }}</span>
    <span v-else>{{ t('common.actions.copy') }}</span>
  </button>
</template>

<style scoped>
.copy-button {
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: transparent;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: border-color var(--tb-transition), color var(--tb-transition);
}

.copy-button:hover:not(:disabled) {
  color: var(--tb-text);
  border-color: var(--tb-border-strong);
}

.copy-button--small {
  padding: 0 var(--tb-space-1);
  font-size: var(--tb-font-2xs);
}

.copy-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.copy-button--ok {
  color: var(--tb-success);
}
</style>
