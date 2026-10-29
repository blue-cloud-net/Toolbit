<script setup lang="ts">
/**
 * 快速参考面板。
 *
 * 工具页侧边的速查表：列出可选值、固定语法与常用片段，
 * 点击条目可把内容插入输入区（由调用方通过 `insert` 回调接管）。
 *
 * 纯前端交互，不发起网络请求，不写任何存储。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'

withDefaults(
  defineProps<{
    /** 面板标题。 */
    title: string
    /** 条目列表。 */
    items: { label: string; value: string; hint?: string }[]
    /** 插入回调；不传时条目不可点击。 */
    insert?: (text: string) => void
  }>(),
  { insert: undefined },
)
</script>

<template>
  <aside class="quick-ref">
    <h3 class="quick-ref__title">
      {{ title }}
    </h3>
    <ul class="quick-ref__list">
      <li
        v-for="item in items"
        :key="item.value"
        class="quick-ref__row"
      >
        <ElButton
          class="quick-ref__item"
          text
          :disabled="!insert"
          :title="item.hint"
          @click="insert?.(item.value)"
        >
          <code class="quick-ref__value">{{ item.value }}</code>
          <span class="quick-ref__label">{{ item.label }}</span>
        </ElButton>
      </li>
    </ul>
  </aside>
</template>

<style scoped>
.quick-ref {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  padding: var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.quick-ref__title {
  margin: 0;
  font-size: var(--tb-font-xs);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.quick-ref__list {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
  padding: 0;
  list-style: none;
}

.quick-ref__item {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
  width: 100%;
  height: auto;
  padding: var(--tb-space-1) var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  text-align: left;
  justify-content: flex-start;
}

.quick-ref__value {
  flex: none;
  padding: 0 var(--tb-space-1);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
  border-radius: var(--tb-radius-control);
}

.quick-ref__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
