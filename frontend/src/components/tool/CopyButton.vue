<script setup lang="ts">
/**
 * 复制按钮（图标式）。
 *
 * 纯图标 + 顶部 tooltip：文字按钮在表格里占宽、在标题栏里抢镜。
 * tooltip **固定**为「复制」；复制结果（已复制 / 复制失败）由 `ElMessage` 提示，
 * 图标则短暂切成打勾 / 警告做为局部反馈。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElTooltip } from 'element-plus/es/components/tooltip/index'
import 'element-plus/es/components/tooltip/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCopy } from '@/composables/useCopy'

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
const { copy: copyText } = useCopy()
const copied = ref(false)
const failed = ref(false)

/** tooltip 文案：固定为「复制」，不随结果切换。 */
const tipText = computed(() => props.label ?? t('common.actions.copy'))

/** 图标：默认复制，成功 / 失败短暂切换（全局提示由 `ElMessage` 负责）。 */
const iconName = computed<'copy' | 'check' | 'alert'>(() => {
  if (copied.value) {
    return 'check'
  }
  if (failed.value) {
    return 'alert'
  }
  return 'copy'
})

let timer: ReturnType<typeof setTimeout> | null = null

/** 执行复制。 */
async function onCopy(): Promise<void> {
  if (props.text === '') {
    return
  }
  const ok = await copyText(props.text)
  copied.value = ok
  failed.value = !ok

  if (timer !== null) {
    clearTimeout(timer)
  }
  timer = setTimeout(() => {
    copied.value = false
    failed.value = false
    timer = null
  }, 1600)
}
</script>

<template>
  <ElTooltip
    :content="tipText"
    placement="top"
    :show-after="200"
  >
    <ElButton
      class="copy-button"
      :class="{
        'copy-button--compact': size === 'small',
        'copy-button--ok': copied,
        'copy-button--failed': failed && !copied,
      }"
      :size="size === 'small' ? 'small' : 'default'"
      :aria-label="tipText"
      :disabled="text === ''"
      @click="onCopy"
    >
      <AppIcon
        :name="iconName"
        :size="size === 'small' ? 13 : 15"
      />
    </ElButton>
  </ElTooltip>
</template>

<style scoped>
/* 图标式按钮：固定正方形，表格里不占额外宽度 */
.copy-button {
  justify-content: center;
  width: 30px;
  min-width: 30px;
  height: 30px;
  padding: 0;
}

.copy-button--compact {
  width: 22px;
  min-width: 22px;
  height: 22px;
}

.copy-button--ok {
  color: var(--tb-success);
}

.copy-button--failed {
  color: var(--tb-danger);
}
</style>
