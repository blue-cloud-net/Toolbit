<script setup lang="ts">
/**
 * OID 树节点（递归）。
 *
 * 只渲染**已加载**的子节点；未加载的分支由展开箭头触发懒加载。
 * 本组件**不**自行发请求：加载与状态都由工具页统一管理，
 * 这样「树」与「右侧详情」始终看到同一份数据。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import type { OidNode } from './types'

const props = withDefaults(
  defineProps<{
    /** 当前节点。 */
    node: OidNode
    /** 取子节点；返回空数组表示尚未加载。 */
    childrenOf: (oid: string) => readonly OidNode[]
    /** 某节点是否处于展开态。 */
    expandedOf: (oid: string) => boolean
    /** 某节点是否还有更多子节点可加载。 */
    hasMoreOf: (oid: string) => boolean
    /** 当前正在加载子节点的 OID；空串表示无。 */
    loadingOid: string
    /** 当前选中的 OID。 */
    selectedOid: string
    /** 层级（用于缩进与层级标记）。 */
    depth?: number
  }>(),
  { depth: 0 },
)

const emit = defineEmits<{
  /** 点击节点名：选中并展开。 */
  select: [node: OidNode]
  /** 点击展开箭头：展开 / 收起。 */
  toggle: [node: OidNode]
  /** 点击「加载更多」。 */
  loadMore: [node: OidNode]
}>()

/** 子节点列表。 */
const children = computed<readonly OidNode[]>(() => props.childrenOf(props.node.oid))

/** 是否可展开：后端已告知有子节点，或已加载出子节点。 */
const expandable = computed(
  () => props.node.childCount > 0 || children.value.length > 0 || props.hasMoreOf(props.node.oid),
)

const { locale, t } = useI18n()

/** 中文界面优先展示中文名，英文界面优先展示官方英文名。 */
const displayName = computed(() => {
  if (locale.value === 'zh-CN') {
    return props.node.nameZh || props.node.name
  }
  return props.node.name || props.node.nameZh
})

/** 状态文案走 i18n，避免在界面里裸露 `current` 这类枚举值。 */
const statusText = computed(() => t(`tools.oid-lookup.options.status.${props.node.status}`))

/** 状态语义级别。 */
const statusLevel = computed(() => {
  switch (props.node.status) {
    case 'deprecated':
      return 'warning'
    case 'obsolete':
    case 'reserved':
      return 'muted'
    default:
      return 'success'
  }
})
</script>

<template>
  <li class="oid-node">
    <div
      class="oid-node__row"
      :class="{ 'oid-node__row--selected': node.oid === selectedOid }"
    >
      <button
        type="button"
        class="oid-node__toggle"
        :aria-label="expandedOf(node.oid) ? $t('tools.oid-lookup.results.collapse') : $t('tools.oid-lookup.results.expand')"
        :aria-expanded="expandedOf(node.oid)"
        :disabled="!expandable || loadingOid === node.oid"
        @click="emit('toggle', node)"
      >
        <span
          v-if="loadingOid === node.oid"
          class="oid-node__spinner"
          aria-hidden="true"
        />
        <span
          v-else
          class="oid-node__chevron"
          :class="{ 'oid-node__chevron--open': expandedOf(node.oid) }"
          aria-hidden="true"
        >
          ›
        </span>
      </button>

      <button
        type="button"
        class="oid-node__label"
        @click="emit('select', node)"
      >
        <span
          class="oid-node__arc tb-mono"
          aria-hidden="true"
        >{{ node.arc }}</span>
        <span class="oid-node__name">{{ displayName }}</span>
      </button>

      <span
        class="oid-node__status"
        :class="`oid-node__status--${statusLevel}`"
      >{{ statusText }}</span>

      <span
        v-if="node.childCount > 0"
        class="oid-node__count tb-mono"
      >{{ node.childCount }}</span>
    </div>

    <ul
      v-if="expandedOf(node.oid) && children.length > 0"
      class="oid-node__children"
    >
      <OidTreeNode
        v-for="child in children"
        :key="child.oid"
        :node="child"
        :children-of="childrenOf"
        :expanded-of="expandedOf"
        :has-more-of="hasMoreOf"
        :loading-oid="loadingOid"
        :selected-oid="selectedOid"
        :depth="depth + 1"
        @select="emit('select', $event)"
        @toggle="emit('toggle', $event)"
        @load-more="emit('loadMore', $event)"
      />
      <li
        v-if="hasMoreOf(node.oid)"
        class="oid-node__more"
      >
        <button
          type="button"
          class="oid-node__more-btn"
          @click="emit('loadMore', node)"
        >
          {{ $t('tools.oid-lookup.results.loadMore') }}
        </button>
      </li>
    </ul>
  </li>
</template>

<style scoped>
.oid-node {
  list-style: none;
}

.oid-node__row {
  display: flex;
  align-items: center;
  gap: var(--tb-space-1);
  padding: 1px 0;
  border-radius: var(--tb-radius-control);
}

.oid-node__row--selected {
  background: var(--tb-primary-soft);
}

.oid-node__toggle {
  display: grid;
  place-items: center;
  width: 18px;
  height: 18px;
  padding: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  cursor: pointer;
  background: none;
  border: none;
}

.oid-node__toggle:disabled {
  cursor: default;
  opacity: 0.35;
}

.oid-node__chevron {
  display: inline-block;
  transition: transform 120ms ease;
}

.oid-node__chevron--open {
  transform: rotate(90deg);
}

.oid-node__spinner {
  width: 10px;
  height: 10px;
  border: 2px solid var(--tb-border-strong);
  border-top-color: var(--tb-primary);
  border-radius: 50%;
  animation: oid-node-spin 700ms linear infinite;
}

@keyframes oid-node-spin {
  to {
    transform: rotate(360deg);
  }
}

.oid-node__label {
  display: flex;
  flex: 1;
  gap: var(--tb-space-2);
  align-items: baseline;
  min-width: 0;
  padding: 0;
  font: inherit;
  color: inherit;
  text-align: left;
  cursor: pointer;
  background: none;
  border: none;
}

.oid-node__arc {
  color: var(--tb-text-tertiary);
}

.oid-node__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.oid-node__status {
  flex-shrink: 0;
  padding: 0 var(--tb-space-1);
  font-size: var(--tb-font-xs);
  border: 1px solid currentcolor;
  border-radius: var(--tb-radius-control);
  opacity: 0.75;
}

.oid-node__status--success {
  color: var(--tb-tone-green);
}

.oid-node__status--warning {
  color: var(--tb-tone-amber);
}

.oid-node__status--muted {
  color: var(--tb-text-tertiary);
}

.oid-node__count {
  flex-shrink: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.oid-node__children {
  margin: 0;
  padding: 0 0 0 var(--tb-space-3);
  list-style: none;
  border-left: 1px solid var(--tb-border);
}

.oid-node__more-btn {
  padding: 2px var(--tb-space-2);
  font: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-primary);
  cursor: pointer;
  background: none;
  border: 1px dashed var(--tb-border-strong);
  border-radius: var(--tb-radius-control);
}
</style>
