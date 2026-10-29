<script setup lang="ts">
/**
 * 加载骨架。
 *
 * 结果区 / 列表区加载中的占位：用 Element Plus 的骨架屏，
 * 避免纯文字「加载中」造成的布局跳动。
 */
import { ElSkeleton } from 'element-plus/es/components/skeleton/index'
import 'element-plus/es/components/skeleton/style/css'

withDefaults(
  defineProps<{
    /** 骨架行数。 */
    rows?: number
    /** 是否用多行文本块形态（否则用段落块）。 */
    variant?: 'text' | 'block'
  }>(),
  { rows: 3, variant: 'text' },
)
</script>

<template>
  <div
    class="loading-skeleton"
    :class="`loading-skeleton--${variant}`"
    role="status"
    aria-busy="true"
  >
    <ElSkeleton
      :rows="rows"
      animated
    />
    <span class="tb-sr-only">…</span>
  </div>
</template>

<style scoped>
.loading-skeleton {
  padding: var(--tb-space-4) 0;
}

.loading-skeleton--block :deep(.el-skeleton__p) {
  height: 56px;
}

@media (prefers-reduced-motion: reduce) {
  .loading-skeleton :deep(.el-skeleton__item) {
    animation: none;
  }
}
</style>
