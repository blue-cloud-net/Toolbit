<script setup lang="ts">
/**
 * 配置树节点编辑器。
 *
 * 递归渲染一棵配置树：每个节点可编辑类型、指令名、参数、目标文件与注释，
 * 块节点还可以继续挂子节点。
 *
 * 指令名既可从官方目录里选，也可直接输入新指令（第三方模块或新版本指令）。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import type { NginxDirectiveInfo, NginxNode, NginxNodeType } from '@/types/nginx'

const props = withDefaults(
  defineProps<{
    /** 该层的节点列表。 */
    nodes: NginxNode[]
    /** 该层的上下文。 */
    context: string
    /** 官方指令目录。 */
    catalog: NginxDirectiveInfo[]
    /** 嵌套深度；用于限制 UI 复杂度。 */
    depth?: number
  }>(),
  { depth: 0 },
)

const emit = defineEmits<{
  /** 节点列表变化。 */
  'update:nodes': [NginxNode[]]
}>()

const { t } = useI18n()

/** 块名 → 子节点上下文。 */
const BLOCK_CONTEXTS: Record<string, string> = {
  events: 'events',
  http: 'http',
  stream: 'stream',
  mail: 'mail',
  server: 'server',
  location: 'location',
  upstream: 'upstream',
  if: 'if',
  limit_except: 'limit_except',
  types: 'types',
}

/** 节点类型选项。 */
const NODE_TYPES: NginxNodeType[] = ['directive', 'block', 'raw']

/** 当前上下文下可选的官方指令。 */
const availableDirectives = computed(() =>
  props.catalog.filter((item) => item.contexts.some((ctx) => normalizeContext(ctx) === props.context)),
)

/** 归一化官方上下文写法。 */
function normalizeContext(context: string): string {
  return context === 'if in location' ? 'if' : context
}

/** 选中指令时同步节点类型：块类指令自动切成配置块。 */
function onNameChange(index: number, name: string): void {
  const found = props.catalog.find((item) => item.name === name)
  if (found?.block) {
    updateNode(index, { name, type: 'block' })
    return
  }
  updateNode(index, { name })
}

/** 选中指令的语法提示。 */
function syntaxHint(node: NginxNode): string {
  const found = props.catalog.find((item) => item.name === node.name)
  if (!found) {
    return ''
  }
  return found.syntax
}

/** 指令说明（按当前语言）。 */
function summaryOf(node: NginxNode): string {
  const found = props.catalog.find((item) => item.name === node.name)
  if (!found) {
    return ''
  }
  return found.summaryZh === '' ? found.summaryEn : found.summaryZh
}

/** 生成节点标识。 */
function newId(): string {
  return `n${Math.random().toString(36).slice(2, 10)}`
}

/** 创建一个空节点。 */
function createNode(): NginxNode {
  return {
    id: newId(),
    type: 'directive',
    context: props.context,
    name: '',
    args: '',
    body: '',
    file: '',
    comment: '',
    children: [],
  }
}

/** 子节点上下文。 */
function childContext(node: NginxNode): string {
  if (node.type !== 'block') {
    return props.context
  }
  return BLOCK_CONTEXTS[node.name] ?? props.context
}

/** 复制节点列表并向上通知。 */
function commit(nodes: NginxNode[]): void {
  emit('update:nodes', nodes)
}

/** 更新单个节点字段。 */
function updateNode(index: number, patch: Partial<NginxNode>): void {
  const next = props.nodes.map((node, i) => {
    if (i !== index) {
      return node
    }
    const merged = { ...node, ...patch }
    // 切回非块类型时清掉子节点，避免残留。
    if (patch.type && patch.type !== 'block') {
      merged.children = []
    }
    merged.context = props.context
    return merged
  })
  commit(next)
}

/** 追加子节点。 */
function addChild(index: number): void {
  const next = props.nodes.map((node, i) => {
    if (i !== index) {
      return node
    }
    const child = createNode()
    child.context = childContext(node)
    return { ...node, children: [...node.children, child] }
  })
  commit(next)
}

/** 追加同级节点。 */
function addSibling(): void {
  commit([...props.nodes, createNode()])
}

/** 删除节点。 */
function removeNode(index: number): void {
  commit(props.nodes.filter((_, i) => i !== index))
}

/** 上移节点。 */
function moveUp(index: number): void {
  if (index <= 0) {
    return
  }
  const next = [...props.nodes]
  const [item] = next.splice(index, 1)
  next.splice(index - 1, 0, item!)
  commit(next)
}

/** 下移节点。 */
function moveDown(index: number): void {
  if (index >= props.nodes.length - 1) {
    return
  }
  const next = [...props.nodes]
  const [item] = next.splice(index, 1)
  next.splice(index + 1, 0, item!)
  commit(next)
}

/** 更新子节点列表。 */
function updateChildren(index: number, children: NginxNode[]): void {
  updateNode(index, { children })
}
</script>

<template>
  <div class="node-editor">
    <p
      v-if="nodes.length === 0"
      class="node-editor__empty"
    >
      {{ t('tools.nginx-config.advanced.emptyContext') }}
    </p>

    <div
      v-for="(node, index) in nodes"
      :key="node.id"
      class="node-editor__node"
    >
      <div class="node-editor__row">
        <ElSelect
          :model-value="node.type"
          class="node-editor__type"
          :aria-label="t('tools.nginx-config.advanced.nodeType')"
          @update:model-value="(value: NginxNodeType) => updateNode(index, { type: value })"
        >
          <ElOption
            v-for="type in NODE_TYPES"
            :key="type"
            :label="t(`tools.nginx-config.advanced.types.${type}`)"
            :value="type"
          />
        </ElSelect>

        <ElSelect
          v-if="node.type !== 'raw'"
          :model-value="node.name"
          class="node-editor__name"
          filterable
          allow-create
          default-first-option
          :placeholder="t('tools.nginx-config.advanced.pickDirective')"
          :aria-label="t('tools.nginx-config.advanced.directiveName')"
          @update:model-value="(value: string) => onNameChange(index, value)"
        >
          <ElOption
            v-for="item in availableDirectives"
            :key="item.id"
            :label="item.name"
            :value="item.name"
          />
        </ElSelect>

        <ElInput
          v-if="node.type === 'raw'"
          :model-value="node.body"
          class="node-editor__body"
          type="textarea"
          :rows="3"
          :placeholder="t('tools.nginx-config.advanced.rawPlaceholder')"
          :aria-label="t('tools.nginx-config.advanced.rawBody')"
          @update:model-value="(value: string) => updateNode(index, { body: value })"
        />

        <ElInput
          v-else
          :model-value="node.args"
          class="node-editor__args"
          :placeholder="t('tools.nginx-config.advanced.argsPlaceholder')"
          :aria-label="t('tools.nginx-config.advanced.args')"
          @update:model-value="(value: string) => updateNode(index, { args: value })"
        />

        <ElInput
          :model-value="node.file"
          class="node-editor__file"
          :placeholder="t('tools.nginx-config.advanced.filePlaceholder')"
          :aria-label="t('tools.nginx-config.advanced.file')"
          @update:model-value="(value: string) => updateNode(index, { file: value })"
        />
      </div>

      <p
        v-if="syntaxHint(node)"
        class="node-editor__syntax"
      >
        <code>{{ syntaxHint(node) }}</code>
      </p>
      <p
        v-if="summaryOf(node)"
        class="node-editor__summary"
      >
        {{ summaryOf(node) }}
      </p>

      <div class="node-editor__actions">
        <ElButton
          size="small"
          text
          @click="addChild(index)"
        >
          {{ t('tools.nginx-config.advanced.addChild') }}
        </ElButton>
        <ElButton
          size="small"
          text
          @click="moveUp(index)"
        >
          {{ t('tools.nginx-config.advanced.moveUp') }}
        </ElButton>
        <ElButton
          size="small"
          text
          @click="moveDown(index)"
        >
          {{ t('tools.nginx-config.advanced.moveDown') }}
        </ElButton>
        <ElButton
          size="small"
          text
          type="danger"
          @click="removeNode(index)"
        >
          {{ t('tools.nginx-config.advanced.remove') }}
        </ElButton>
      </div>

      <NodeEditor
        v-if="node.type === 'block' && node.children.length > 0"
        :nodes="node.children"
        :context="childContext(node)"
        :catalog="catalog"
        :depth="depth + 1"
        @update:nodes="(value: NginxNode[]) => updateChildren(index, value)"
      />
    </div>

    <ElButton
      class="node-editor__add"
      size="small"
      @click="addSibling"
    >
      {{ t('tools.nginx-config.advanced.addNode') }}
    </ElButton>
  </div>
</template>

<style scoped>
.node-editor {
  display: grid;
  gap: var(--tb-space-2);
}

.node-editor__empty {
  margin: 0;
  color: var(--tb-text-tertiary);
  font-size: var(--tb-font-sm);
}

.node-editor__node {
  display: grid;
  gap: var(--tb-space-2);
  padding: var(--tb-space-2);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-sm);
}

.node-editor__row {
  display: grid;
  grid-template-columns: 8rem minmax(0, 1fr) minmax(0, 1.2fr) minmax(0, 0.8fr);
  gap: var(--tb-space-2);
  align-items: center;
}

.node-editor__type {
  width: 100%;
}

.node-editor__name,
.node-editor__args,
.node-editor__file,
.node-editor__body {
  width: 100%;
}

.node-editor__syntax,
.node-editor__summary {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  word-break: break-all;
}

.node-editor__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1);
}

.node-editor__add {
  justify-self: start;
}

@media (max-width: 900px) {
  .node-editor__row {
    grid-template-columns: 1fr;
  }
}
</style>
