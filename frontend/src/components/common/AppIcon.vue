<script lang="ts">
/**
 * 可用图标名。
 *
 * 放在普通 `<script>` 块里以便导出类型：调用方（如侧栏）需要把
 * catalog 的 `icon` 字段映射到这个联合类型。
 */
export type IconName =
  | 'all-tools'
  | 'clock'
  | 'star'
  | 'star-filled'
  | 'chevron'
  | 'menu'
  | 'search'
  | 'focus'
  | 'sun'
  | 'moon'
  | 'settings'
  | 'close'
  | 'check'
  | 'alert'
  | 'external'
  | 'lock'
  | 'key'
  | 'certificate'
  | 'code'
  | 'indent'
  | 'text'
  | 'data'
  | 'monitor'
  | 'connection'
  | 'document'
</script>

<script setup lang="ts">
/**
 * 内联 SVG 图标。
 *
 * 不使用图标字体（规范禁止），图标以 `24×24` 的 path 内联渲染，
 * 颜色继承 `currentColor`。
 */

withDefaults(
  defineProps<{
    /** 图标名。 */
    name: IconName
    /** 边长（像素）。 */
    size?: number
  }>(),
  { size: 18 },
)

/** 图标路径表；每条都是一次或多次 path 的 `d` 集合。 */
const PATHS: Record<IconName, string[]> = {
  'all-tools': [
    'M4 5.5A1.5 1.5 0 0 1 5.5 4h4A1.5 1.5 0 0 1 11 5.5v4A1.5 1.5 0 0 1 9.5 11h-4A1.5 1.5 0 0 1 4 9.5v-4Z',
    'M13 5.5A1.5 1.5 0 0 1 14.5 4h4A1.5 1.5 0 0 1 20 5.5v4A1.5 1.5 0 0 1 18.5 11h-4A1.5 1.5 0 0 1 13 9.5v-4Z',
    'M4 14.5A1.5 1.5 0 0 1 5.5 13h4A1.5 1.5 0 0 1 11 14.5v4A1.5 1.5 0 0 1 9.5 20h-4A1.5 1.5 0 0 1 4 18.5v-4Z',
    'M13 14.5A1.5 1.5 0 0 1 14.5 13h4A1.5 1.5 0 0 1 20 14.5v4A1.5 1.5 0 0 1 18.5 20h-4A1.5 1.5 0 0 1 13 18.5v-4Z',
  ],
  clock: ['M12 7v5l3 2', 'M12 21a9 9 0 1 1 0-18 9 9 0 0 1 0 18Z'],
  star: [
    'M12 4.5l2.4 4.9 5.4.8-3.9 3.8.9 5.4-4.8-2.6-4.8 2.6.9-5.4L4.2 10.2l5.4-.8L12 4.5Z',
  ],
  'star-filled': [
    'M12 4.2l2.5 5.1 5.6.8-4.1 4 1 5.6-5-2.7-5 2.7 1-5.6-4.1-4 5.6-.8L12 4.2Z',
  ],
  chevron: ['M9 6l6 6-6 6'],
  menu: ['M4 7h16', 'M4 12h16', 'M4 17h16'],
  search: ['M10.5 4a6.5 6.5 0 1 1 0 13 6.5 6.5 0 0 1 0-13Z', 'M15.4 15.4 20 20'],
  focus: ['M4 9V5.5A1.5 1.5 0 0 1 5.5 4H9', 'M15 4h3.5A1.5 1.5 0 0 1 20 5.5V9', 'M20 15v3.5a1.5 1.5 0 0 1-1.5 1.5H15', 'M9 20H5.5A1.5 1.5 0 0 1 4 18.5V15'],
  sun: [
    'M12 6.5a5.5 5.5 0 1 1 0 11 5.5 5.5 0 0 1 0-11Z',
    'M12 2v2M12 20v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M2 12h2M20 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4',
  ],
  moon: ['M20 14.5A8.5 8.5 0 0 1 9.5 4 8.5 8.5 0 1 0 20 14.5Z'],
  settings: [
    'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z',
    'M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2v.2a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-2.9-1.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1A1.7 1.7 0 0 0 3 15a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.3-2.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1A1.7 1.7 0 0 0 10 4.1V4a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 2.9 1.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0 1.2 2.9H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1Z',
  ],
  close: ['M6 6l12 12', 'M18 6L6 18'],
  check: ['M5 12.5l4.5 4.5L19 7'],
  alert: ['M12 4.5 21 19H3l9-14.5Z', 'M12 10v3.5', 'M12 16.2v.3'],
  external: ['M14 4h6v6', 'M20 4l-9 9', 'M18 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h5'],
  lock: [
    'M8 10V7.5a4 4 0 0 1 8 0V10',
    'M6.5 10h11a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-7a1 1 0 0 1 1-1Z',
  ],
  key: [
    'M14 10a4 4 0 1 0-3.5 3.97L4 20.47V21h3v-2h2v-2l2.5-2.5A4 4 0 0 0 14 10Z',
    'M16 8h.01',
  ],
  certificate: [
    'M7 3h7l4 4v9a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1Z',
    'M14 3v4h4',
    'M9 11h6',
    'M9 14h4',
    'M10 18v3l2-1.2 2 1.2v-3',
  ],
  code: ['M9 8l-4 4 4 4', 'M15 8l4 4-4 4'],
  indent: ['M4 6h16', 'M10 12h10', 'M10 18h10', 'M4 10l2 2-2 2'],
  text: ['M5 6h14', 'M5 11h9', 'M5 16h13'],
  data: [
    'M4 7c0-1.7 3.6-3 8-3s8 1.3 8 3-3.6 3-8 3-8-1.3-8-3Z',
    'M4 7v10c0 1.7 3.6 3 8 3s8-1.3 8-3V7',
  ],
  monitor: [
    'M4 6h16a1 1 0 0 1 1 1v8a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1Z',
    'M9 20h6',
    'M12 16v4',
  ],
  connection: [
    'M8 8a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5Z',
    'M16 21a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5Z',
    'M8 8v5a3 3 0 0 0 3 3h5',
  ],
  document: [
    'M7 3h7l4 4v13a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1Z',
    'M14 3v5h5',
  ],
}
</script>

<template>
  <svg
    class="icon"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.7"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    focusable="false"
  >
    <path
      v-for="(d, index) in PATHS[name]"
      :key="index"
      :d="d"
      :fill="name === 'star-filled' ? 'currentColor' : 'none'"
    />
  </svg>
</template>

<style scoped>
.icon {
  display: block;
  flex: none;
}
</style>
