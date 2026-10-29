<script setup lang="ts">
/**
 * ASN.1 结构树。
 *
 * 展示解析出的 ASN.1 节点：标签、类、长度与内容。
 */
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

export interface Asn1Node {
  /** 节点标签，如 SEQUENCE。 */
  tag: string
  /** 节点类，如 UNIVERSAL。 */
  cls?: string
  /** 是否为复合节点（有子节点）。 */
  composite?: boolean
  /** 节点内容。 */
  value?: string
  /** 子节点。 */
  children?: Asn1Node[]
}

defineOptions({ name: 'Asn1Tree' })

withDefaults(
  defineProps<{
    /** 根节点列表。 */
    nodes: Asn1Node[]
  }>(),
  {},
)

const { t } = useI18n()
</script>

<template>
  <div class="asn1-tree">
    <p
      v-if="nodes.length === 0"
      class="asn1-tree__empty"
    >
      {{ t('common.states.empty') }}
    </p>
    <ul
      v-else
      class="asn1-tree__list"
    >
      <li
        v-for="node in nodes"
        :key="`${node.tag}-${node.value ?? ''}`"
        class="asn1-tree__item"
      >
        <div class="asn1-tree__row">
          <span class="asn1-tree__tag">{{ node.tag }}</span>
          <span
            v-if="node.cls"
            class="asn1-tree__cls"
          >{{ node.cls }}</span>
          <span
            v-if="node.value"
            class="asn1-tree__value"
          >{{ node.value }}</span>
          <CopyButton
            v-if="node.value"
            :text="node.value"
            size="small"
          />
        </div>
        <Asn1Tree
          v-if="node.children && node.children.length > 0"
          :nodes="node.children"
        />
      </li>
    </ul>
  </div>
</template>

<style scoped>
.asn1-tree {
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
}

.asn1-tree__list {
  margin: 0;
  padding: 0 0 0 var(--tb-space-3);
  list-style: none;
  border-left: 1px solid var(--tb-border);
}

.asn1-tree__item {
  padding: 1px 0;
}

.asn1-tree__row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: var(--tb-space-2);
}

.asn1-tree__tag {
  color: var(--tb-primary);
}

.asn1-tree__cls {
  color: var(--tb-text-tertiary);
}

.asn1-tree__value {
  overflow-wrap: anywhere;
  color: var(--tb-text);
}

.asn1-tree__empty {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
