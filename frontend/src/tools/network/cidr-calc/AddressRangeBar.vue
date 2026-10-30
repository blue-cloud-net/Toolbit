<script setup lang="ts">
/**
 * 地址区间条。
 *
 * 把一个网段的地址空间画成一条横向比例条：
 *
 * ```text
 * ├──────── 网络地址 ────────┼──────── 可用地址区间 ────────┼── 广播 ──┤
 * ```
 *
 * 宽度按**地址个数**等比分配，因此大网段里可用区会明显短于 /30。
 * 纯 CSS，无图表库。
 */
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 地址族：4 或 6。 */
    version: number
    /** 网络地址（含前缀长度），如 `192.168.1.0/24`。 */
    network: string
    /** 网段首地址。 */
    firstAddress: string
    /** 网段末地址。 */
    lastAddress: string
    /** 广播地址；IPv6 或无广播概念的网段为空串。 */
    broadcast?: string
    /** 首个可用地址。 */
    firstUsable: string
    /** 末个可用地址。 */
    lastUsable: string
  }>(),
  { broadcast: '' },
)

/** 把 IPv4 / IPv6 地址转成 BigInt。 */
function toBig(addr: string): bigint {
  if (addr.includes(':')) {
    const [head = '', tail = ''] = addr.split('::')
    const headParts = head === '' ? [] : head.split(':')
    const tailParts = tail === '' ? [] : tail.split(':')
    const missing = 8 - headParts.length - tailParts.length
    const parts = [...headParts, ...Array<string>(Math.max(missing, 0)).fill('0'), ...tailParts]
    return parts.reduce((acc, part) => (acc << 16n) + BigInt(parseInt(part || '0', 16)), 0n)
  }
  return addr
    .split('.')
    .reduce((acc, part) => (acc << 8n) + BigInt(Number(part || 0)), 0n)
}

/** 地址总数（2 的幂，用 BigInt 承载 IPv6）。 */
const total = computed<bigint>(() => {
  const bits = props.version === 4 ? 32 : 128
  const prefix = Number(props.network.split('/')[1] ?? '0')
  return 1n << BigInt(bits - prefix)
})

/** 各区间的占比（百分比，保留两位小数）。 */
const segments = computed(() => {
  const base = toBig(props.firstAddress)
  const span = total.value
  const percent = (value: bigint): number => {
    if (span <= 0n) {
      return 0
    }
    const offset = value - base
    if (offset < 0n) {
      return 0
    }
    const raw = Number((offset * 10000n) / span) / 100
    return Math.min(Math.max(raw, 0), 100)
  }
  const usableStart = percent(toBig(props.firstUsable))
  const usableEnd = percent(toBig(props.lastUsable))
  const broadcastStart = props.broadcast === '' ? 100 : percent(toBig(props.broadcast))
  return [
    { key: 'network', size: usableStart, kind: 'reserved' },
    { key: 'usable', size: Math.max(usableEnd - usableStart, 0), kind: 'usable' },
    { key: 'tail', size: Math.max(broadcastStart - usableEnd, 0), kind: 'reserved' },
  ]
})
</script>

<template>
  <div class="range-bar">
    <div
      class="range-bar__track"
      role="img"
      :aria-label="network"
    >
      <span
        v-for="segment in segments"
        :key="segment.key"
        class="range-bar__segment"
        :class="`range-bar__segment--${segment.kind}`"
        :style="{ width: `${segment.size}%` }"
      />
    </div>
    <div class="range-bar__labels">
      <span class="range-bar__label tb-mono">{{ firstAddress }}</span>
      <span class="range-bar__label range-bar__label--mid tb-mono">{{ lastUsable }}</span>
      <span class="range-bar__label tb-mono">{{ lastAddress }}</span>
    </div>
  </div>
</template>

<style scoped>
.range-bar {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.range-bar__track {
  display: flex;
  width: 100%;
  height: 12px;
  overflow: hidden;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.range-bar__segment {
  height: 100%;
}

.range-bar__segment--usable {
  background: var(--tb-primary);
}

.range-bar__segment--reserved {
  background: repeating-linear-gradient(
    45deg,
    var(--tb-border) 0,
    var(--tb-border) 3px,
    transparent 3px,
    transparent 6px
  );
}

.range-bar__labels {
  display: flex;
  justify-content: space-between;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.range-bar__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.range-bar__label--mid {
  color: var(--tb-primary);
}
</style>
