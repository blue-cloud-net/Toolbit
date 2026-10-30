<script setup lang="ts">
/**
 * 前缀位图。
 *
 * 把网段的前缀长度画成一串比特方块：网络位（1）与主机位（0）两色。
 *
 * - IPv4：32 位，按 8 位一组（4 组）；
 * - IPv6：128 位，按 16 位一组（8 组），组内 4 位一节便于阅读。
 *
 * 纯 CSS，不引图表库。
 */
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 地址族：4 或 6。 */
    version: number
    /** 前缀长度。 */
    prefix: number
  }>(),
  {},
)

/** 每组位数：IPv4 每 8 位一组，IPv6 每 16 位一组。 */
const groupSize = computed(() => (props.version === 4 ? 8 : 16))

/** 总位数。 */
const totalBits = computed(() => (props.version === 4 ? 32 : 128))

/** 分组结果：每组含若干比特。 */
const groups = computed(() => {
  const out: { index: number; bits: { value: 0 | 1; index: number }[] }[] = []
  for (let start = 0; start < totalBits.value; start += groupSize.value) {
    const bits: { value: 0 | 1; index: number }[] = []
    for (let offset = 0; offset < groupSize.value; offset++) {
      const index = start + offset
      bits.push({ value: index < props.prefix ? 1 : 0, index })
    }
    out.push({ index: start / groupSize.value, bits })
  }
  return out
})
</script>

<template>
  <div class="bitmap">
    <div
      v-for="group in groups"
      :key="group.index"
      class="bitmap__group"
    >
      <span
        v-for="bit in group.bits"
        :key="bit.index"
        class="bitmap__bit"
        :class="bit.value === 1 ? 'bitmap__bit--net' : 'bitmap__bit--host'"
      />
    </div>
  </div>
</template>

<style scoped>
.bitmap {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
}

.bitmap__group {
  display: flex;
  gap: 2px;
  padding: 2px;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.bitmap__bit {
  width: 8px;
  height: 14px;
  border-radius: 1px;
}

.bitmap__bit--net {
  background: var(--tb-primary);
}

.bitmap__bit--host {
  background: var(--tb-border);
}
</style>
