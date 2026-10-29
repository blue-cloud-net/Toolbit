<script setup lang="ts">
/**
 * 流式进度条。
 *
 * 服务端按阶段推送进度时展示：阶段文案 + 进度条 + 计数。
 */
import { ElProgress } from 'element-plus/es/components/progress/index'
import 'element-plus/es/components/progress/style/css'
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 阶段文案。 */
    phase?: string
    /** 已完成数。 */
    done?: number
    /** 总数。 */
    total?: number
  }>(),
  { phase: '', done: 0, total: 0 },
)

/** 进度百分比（0 ~ 100）。 */
const percent = computed(() => {
  if (props.total <= 0) {
    return 0
  }
  return Math.min(100, Math.round((props.done / props.total) * 100))
})
</script>

<template>
  <div class="progress-stream">
    <div class="progress-stream__head">
      <span
        v-if="phase"
        class="progress-stream__phase"
      >{{ phase }}</span>
      <span
        v-if="total > 0"
        class="progress-stream__count"
      >{{ done }} / {{ total }}</span>
    </div>
    <ElProgress
      :percentage="percent"
      :stroke-width="4"
      :show-text="false"
      :aria-label="phase"
    />
  </div>
</template>

<style scoped>
.progress-stream {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.progress-stream__head {
  display: flex;
  justify-content: space-between;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}
</style>
