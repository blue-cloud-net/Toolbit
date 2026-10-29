<script setup lang="ts">
/**
 * 十六进制转储视图。
 *
 * 按 16 字节一行展示二进制内容，左侧为偏移、中间为十六进制、
 * 右侧为可打印 ASCII。
 */
import { computed } from 'vue'

export interface ByteViewProps {
  /** 字节数组。 */
  bytes: number[]
  /** 每行字节数，默认 16。 */
  bytesPerLine?: number
}

const props = withDefaults(defineProps<ByteViewProps>(), { bytesPerLine: 16 })

/** 转储行。 */
interface DumpLine {
  offset: string
  hex: string[]
  ascii: string
}

const lines = computed<DumpLine[]>(() => {
  const result: DumpLine[] = []
  for (let i = 0; i < props.bytes.length; i += props.bytesPerLine) {
    const chunk = props.bytes.slice(i, i + props.bytesPerLine)
    result.push({
      offset: i.toString(16).padStart(8, '0'),
      hex: chunk.map((b) => b.toString(16).padStart(2, '0')),
      ascii: chunk.map((b) => (b >= 0x20 && b <= 0x7e ? String.fromCharCode(b) : '.')).join(''),
    })
  }
  return result
})
</script>

<template>
  <div class="byte-view">
    <p
      v-if="bytes.length === 0"
      class="byte-view__empty"
    >
      —
    </p>
    <div
      v-for="line in lines"
      :key="line.offset"
      class="byte-view__line"
    >
      <span class="byte-view__offset">{{ line.offset }}</span>
      <span class="byte-view__hex">
        <span
          v-for="(byte, i) in line.hex"
          :key="i"
          class="byte-view__byte"
        >{{ byte }}</span>
      </span>
      <span class="byte-view__ascii">{{ line.ascii }}</span>
    </div>
  </div>
</template>

<style scoped>
.byte-view {
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  line-height: 1.7;
  overflow-x: auto;
}

.byte-view__line {
  display: flex;
  gap: var(--tb-space-3);
  white-space: pre;
}

.byte-view__offset {
  flex: none;
  color: var(--tb-text-tertiary);
}

.byte-view__hex {
  display: flex;
  gap: var(--tb-space-1);
  color: var(--tb-text);
}

.byte-view__byte {
  min-width: 18px;
}

.byte-view__ascii {
  flex: none;
  color: var(--tb-success);
}

.byte-view__empty {
  margin: 0;
  color: var(--tb-text-tertiary);
}
</style>
