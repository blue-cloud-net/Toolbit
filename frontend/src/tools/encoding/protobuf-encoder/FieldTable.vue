<script setup lang="ts">
/**
 * protobuf-encoder 的字段表格编辑器。
 *
 * 字段树被拍平成带层级的行渲染：`message` 类型字段的行下方
 * 缩进显示它的子字段，操作按钮按路径回写整棵树。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import {
  FIELD_TYPES,
  NESTED_TYPE,
  createNode,
  removeNode,
  type ProtoFieldNode,
} from '@/tools/encoding/protobuf-encoder/fields'

const nodes = defineModel<ProtoFieldNode[]>({ required: true })

const { t } = useI18n()

/** 可选的字节编码。 */
const ENCODING_OPTIONS = ['utf8', 'hex', 'base64', 'base64url'] as const

/** 拍平后的一行。 */
interface FlatRow {
  /** 节点本身。 */
  node: ProtoFieldNode
  /** 从根到该节点的子节点下标路径。 */
  path: number[]
  /** 层级深度。 */
  depth: number
}

/** 拍平字段树。 */
const rows = computed<FlatRow[]>(() => {
  const out: FlatRow[] = []
  const walk = (list: ProtoFieldNode[], path: number[], depth: number): void => {
    list.forEach((node, index) => {
      const current = [...path, index]
      out.push({ node, path: current, depth })
      if (node.type === NESTED_TYPE && node.children.length > 0) {
        walk(node.children, current, depth + 1)
      }
    })
  }
  walk(nodes.value, [], 0)
  return out
})

/** 删除一行。 */
function remove(path: number[]): void {
  nodes.value = removeNode(nodes.value, path)
}

/** 按路径找到节点。 */
function nodeAt(path: number[]): ProtoFieldNode | null {
  let list = nodes.value
  let current: ProtoFieldNode | null = null
  for (const index of path) {
    current = list[index] ?? null
    if (!current) {
      return null
    }
    list = current.children
  }
  return current
}

/** 给 message 字段添加子字段。 */
function addChild(path: number[]): void {
  const parent = nodeAt(path)
  if (!parent) {
    return
  }
  parent.children.push(createNode())
}
</script>

<template>
  <div class="pf">
    <div class="pf__head">
      <span class="pf__cell pf__cell--number">
        {{ t('tools.protobuf-encoder.fields.fieldNumber') }}
      </span>
      <span class="pf__cell pf__cell--type">
        {{ t('tools.protobuf-encoder.fields.type') }}
      </span>
      <span class="pf__cell pf__cell--value">
        {{ t('tools.protobuf-encoder.fields.value') }}
      </span>
      <span class="pf__cell pf__cell--encoding">
        {{ t('tools.protobuf-encoder.fields.encoding') }}
      </span>
      <span class="pf__cell pf__cell--action" />
    </div>

    <div
      v-for="row in rows"
      :key="row.path.join('.')"
      class="pf__row"
    >
      <div
        class="pf__cell pf__cell--number"
        :style="{ paddingLeft: `${row.depth * 16}px` }"
      >
        <ElInputNumber
          v-model="row.node.field"
          class="pf__number"
          :min="1"
          :max="536870911"
          :controls="false"
        />
      </div>

      <div class="pf__cell pf__cell--type">
        <ElSelect
          v-model="row.node.type"
          class="pf__select"
          :aria-label="t('tools.protobuf-encoder.fields.type')"
        >
          <ElOption
            v-for="item in FIELD_TYPES"
            :key="item"
            :label="item"
            :value="item"
          />
        </ElSelect>
      </div>

      <div class="pf__cell pf__cell--value">
        <ElInput
          v-model="row.node.value"
          class="pf__value tb-mono"
          :aria-label="t('tools.protobuf-encoder.fields.value')"
        />
      </div>

      <div class="pf__cell pf__cell--encoding">
        <ElSelect
          v-model="row.node.encoding"
          class="pf__select"
          :disabled="row.node.type !== 'bytes'"
          :aria-label="t('tools.protobuf-encoder.fields.encoding')"
        >
          <ElOption
            v-for="item in ENCODING_OPTIONS"
            :key="item"
            :label="item"
            :value="item"
          />
        </ElSelect>
      </div>

      <div class="pf__cell pf__cell--action">
        <ElButton
          v-if="row.node.type === NESTED_TYPE"
          size="small"
          text
          @click="addChild(row.path)"
        >
          {{ t('tools.protobuf-encoder.actions.addChild') }}
        </ElButton>
        <ElButton
          size="small"
          text
          @click="remove(row.path)"
        >
          {{ t('tools.protobuf-encoder.actions.removeRow') }}
        </ElButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pf {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.pf__head,
.pf__row {
  display: grid;
  grid-template-columns: 96px 132px minmax(0, 1fr) 120px auto;
  gap: 6px;
  align-items: center;
}

.pf__head {
  color: var(--tb-text-secondary);
  font-size: 0.75rem;
}

.pf__cell {
  min-width: 0;
}

.pf__number,
.pf__select,
.pf__value {
  width: 100%;
}
</style>
