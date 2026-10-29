<script setup lang="ts">
/**
 * 加载骨架。
 *
 * 结果区 / 列表区加载中的占位：若干条 shimmering 灰块，
 * 避免纯文字「加载中」造成的布局跳动。
 */
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
    <span
      v-for="n in rows"
      :key="n"
      class="loading-skeleton__row"
      :style="{ width: `${100 - (n - 1) * 12}%` }"
    />
    <span class="tb-sr-only">…</span>
  </div>
</template>

<style scoped>
.loading-skeleton {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  padding: var(--tb-space-4) 0;
}

.loading-skeleton__row {
  display: block;
  height: 12px;
  background: var(--tb-bg-hover);
  border-radius: var(--tb-radius-control);
  animation: loading-skeleton-pulse 1.4s ease-in-out infinite;
}

.loading-skeleton--block .loading-skeleton__row {
  height: 56px;
}

@keyframes loading-skeleton-pulse {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.45;
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-skeleton__row {
    animation: none;
  }
}
</style>
