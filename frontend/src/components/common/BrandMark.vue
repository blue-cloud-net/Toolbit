<script setup lang="ts">
/**
 * 品牌标识：T 型手柄 + 六角批头。
 *
 * 语义取自品牌名 Toolbit —— `T` 既是首字母，也构成一把 T 型螺丝刀的整体轮廓；
 * 底端的尖头六角是批头（bit），与工具语义合起来正好是「Tool + Bit」。
 *
 * 配色走 `--tb-brand-*` 令牌，**固定靛蓝**，不随主色切换 ——
 * 徽标是品牌锚点，在任意主色与明暗主题下都保持同一识别度。
 */
import { useId } from 'vue'

withDefaults(
  defineProps<{
    /** 边长（像素）。 */
    size?: number
  }>(),
  { size: 28 },
)

/** 渐变 id 必须按实例唯一，否则同页多处渲染会互相覆盖。 */
const gradId = `tb-brand-${useId()}`
</script>

<template>
  <svg
    class="brand-mark"
    :width="size"
    :height="size"
    viewBox="0 0 64 64"
    aria-hidden="true"
    focusable="false"
  >
    <defs>
      <linearGradient
        :id="gradId"
        x1="0"
        y1="0"
        x2="1"
        y2="1"
      >
        <stop
          offset="0"
          stop-color="var(--tb-brand-from)"
        />
        <stop
          offset="1"
          stop-color="var(--tb-brand-to)"
        />
      </linearGradient>
    </defs>
    <rect
      width="64"
      height="64"
      rx="14"
      :fill="`url(#${gradId})`"
    />
    <g fill="var(--tb-brand-fg)">
      <rect
        x="10"
        y="10"
        width="44"
        height="10"
        rx="5"
      />
      <rect
        x="27"
        y="16"
        width="10"
        height="22"
        rx="5"
      />
      <polygon points="32,33.3 41.09,38.55 41.09,49.05 32,54.3 22.91,49.05 22.91,38.55" />
    </g>
  </svg>
</template>

<style scoped>
.brand-mark {
  display: block;
  flex: none;
}
</style>
