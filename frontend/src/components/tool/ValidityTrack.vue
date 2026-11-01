<script setup lang="ts">
/**
 * 有效期时间轴。
 *
 * 在 [生效时间, 过期时间] 上标出「当前时刻」的位置：
 * 一眼看出证书是刚开始、临近过期还是已经过期。
 *
 * 语义色只由 `level` 决定（success / warning / danger），
 * 颜色取自 `tokens.css`，**不在组件里写死颜色**。
 */
import { computed } from 'vue'

/** 语义级别。 */
type TrackLevel = 'success' | 'warning' | 'danger'

const props = withDefaults(
  defineProps<{
    /** 已过去的比例（0–100）。 */
    percent: number
    /** 语义级别。 */
    level: TrackLevel
    /** 左端标签（通常是生效时间）。 */
    start: string
    /** 右端标签（通常是过期时间）。 */
    end: string
  }>(),
  {},
)

/** 实际填充比例：钳到 0–100。 */
const fill = computed(() => {
  if (!Number.isFinite(props.percent)) {
    return 0
  }
  return Math.min(100, Math.max(0, props.percent))
})
</script>

<template>
  <div class="track">
    <div
      class="track__bar"
      role="img"
      :aria-label="`${start} → ${end}`"
    >
      <div
        class="track__fill"
        :class="`track__fill--${level}`"
        :style="{ width: `${fill}%` }"
      />
      <div
        class="track__marker"
        :class="`track__marker--${level}`"
        :style="{ left: `${fill}%` }"
      />
    </div>
    <div class="track__labels">
      <span class="track__label">{{ start }}</span>
      <span class="track__percent">{{ fill }}%</span>
      <span class="track__label">{{ end }}</span>
    </div>
  </div>
</template>

<style scoped>
.track {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.track__bar {
  position: relative;
  height: 8px;
  overflow: hidden;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-pill);
}

.track__fill {
  height: 100%;
  border-radius: var(--tb-radius-pill);
  transition: width 0.2s ease;
}

.track__fill--success {
  background: var(--tb-success);
}

.track__fill--warning {
  background: var(--tb-warning);
}

.track__fill--danger {
  background: var(--tb-danger);
}

.track__marker {
  position: absolute;
  top: 50%;
  width: 10px;
  height: 10px;
  transform: translate(-50%, -50%);
  background: var(--tb-bg-elevated);
  border: 2px solid var(--tb-border-strong);
  border-radius: 50%;
}

.track__marker--success {
  border-color: var(--tb-success);
}

.track__marker--warning {
  border-color: var(--tb-warning);
}

.track__marker--danger {
  border-color: var(--tb-danger);
}

.track__labels {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.track__percent {
  font-family: var(--tb-font-mono);
  color: var(--tb-text-secondary);
}
</style>
