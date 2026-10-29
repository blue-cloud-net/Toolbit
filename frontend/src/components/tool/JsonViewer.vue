<script setup lang="ts">
/**
 * JSON 折叠树查看器。
 *
 * 结构化结果的多视图之一：对象 / 数组可逐层折叠，
 * 标量直接展示，每个标量可单独复制。
 */
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

export interface JsonNode {
  key: string
  value: unknown
}

defineOptions({ name: 'JsonViewer' })

withDefaults(
  defineProps<{
    /** 节点列表。 */
    nodes: JsonNode[]
    /** 默认展开层级；0 表示全部折叠。 */
    defaultExpandDepth?: number
  }>(),
  { defaultExpandDepth: 1 },
)

const { t } = useI18n()

/** 记录被手动切换展开状态的路径；未记录时按默认层级决定。 */
const toggled = defineModel<Record<string, boolean>>('toggled', { default: () => ({}) })

/** 判断某个路径当前是否展开。 */
function isExpanded(path: string, depth: number, defaultDepth: number): boolean {
  const override = toggled.value[path]
  if (override !== undefined) {
    return override
  }
  return depth < defaultDepth
}

/** 切换展开状态。 */
function toggle(path: string, depth: number, defaultDepth: number): void {
  toggled.value = { ...toggled.value, [path]: !isExpanded(path, depth, defaultDepth) }
}

/** 取值的类型标签。 */
function typeTag(value: unknown): string {
  if (value === null) {
    return 'null'
  }
  if (Array.isArray(value)) {
    return 'array'
  }
  return typeof value
}

/** 是否为可折叠的复合值。 */
function isComposite(value: unknown): boolean {
  return value !== null && typeof value === 'object'
}

/** 标量的展示文本。 */
function scalarText(value: unknown): string {
  if (value === null) {
    return 'null'
  }
  if (typeof value === 'string') {
    return value
  }
  return JSON.stringify(value)
}

/** 子节点列表。 */
function childrenOf(node: JsonNode): JsonNode[] {
  const value = node.value
  if (Array.isArray(value)) {
    return value.map((item, index) => ({ key: String(index), value: item }))
  }
  if (value !== null && typeof value === 'object') {
    return Object.entries(value as Record<string, unknown>).map(([key, item]) => ({
      key,
      value: item,
    }))
  }
  return []
}

/** 复合值的概览文本。 */
function summaryOf(node: JsonNode): string {
  const children = childrenOf(node)
  const kind = typeTag(node.value)
  return kind === 'array' ? `${children.length} 项` : `${children.length} 个字段`
}
</script>

<template>
  <div class="json-viewer">
    <p
      v-if="nodes.length === 0"
      class="json-viewer__empty"
    >
      {{ t('common.states.empty') }}
    </p>
    <ul
      v-else
      class="json-viewer__list"
    >
      <li
        v-for="node in nodes"
        :key="node.key"
        class="json-viewer__item"
      >
        <template v-if="isComposite(node.value)">
          <button
            type="button"
            class="json-viewer__toggle"
            :aria-expanded="isExpanded(node.key, 0, defaultExpandDepth)"
            @click="toggle(node.key, 0, defaultExpandDepth)"
          >
            <span
              class="json-viewer__caret"
              :class="{ 'json-viewer__caret--open': isExpanded(node.key, 0, defaultExpandDepth) }"
            >▸</span>
            <span class="json-viewer__key">{{ node.key }}</span>
            <span class="json-viewer__summary">{{ summaryOf(node) }}</span>
          </button>
          <JsonViewer
            v-if="isExpanded(node.key, 0, defaultExpandDepth)"
            v-model:toggled="toggled"
            :nodes="childrenOf(node)"
            :default-expand-depth="defaultExpandDepth"
          />
        </template>
        <template v-else>
          <span class="json-viewer__key">{{ node.key }}</span>
          <span
            class="json-viewer__scalar"
            :class="`json-viewer__scalar--${typeTag(node.value)}`"
          >{{ scalarText(node.value) }}</span>
          <CopyButton
            :text="scalarText(node.value)"
            size="small"
          />
        </template>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.json-viewer {
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
}

.json-viewer__list {
  margin: 0;
  padding: 0 0 0 var(--tb-space-3);
  list-style: none;
  border-left: 1px solid var(--tb-border);
}

.json-viewer__item {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--tb-space-1);
  padding: 1px 0;
}

.json-viewer__toggle {
  display: inline-flex;
  align-items: baseline;
  gap: var(--tb-space-1);
  padding: 0;
  font-family: inherit;
  font-size: inherit;
  color: var(--tb-text-secondary);
  background: transparent;
  border: none;
  cursor: pointer;
}

.json-viewer__caret {
  display: inline-block;
  transition: transform var(--tb-transition);
}

.json-viewer__caret--open {
  transform: rotate(90deg);
}

.json-viewer__key {
  color: var(--tb-primary);
}

.json-viewer__summary,
.json-viewer__scalar {
  color: var(--tb-text-secondary);
}

.json-viewer__scalar--string {
  color: var(--tb-success);
}

.json-viewer__scalar--number {
  color: var(--tb-info);
}

.json-viewer__scalar--boolean {
  color: var(--tb-warning);
}

.json-viewer__scalar--null {
  color: var(--tb-text-tertiary);
}

.json-viewer__empty {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
