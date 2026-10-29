<script setup lang="ts">
/**
 * 只读代码 / 树视图。
 *
 * **自渲染**：不引入 codemirror / monaco 之类的编辑器依赖 ——
 * 那会直接击穿「单工具 chunk < 40 KB gzip」的预算
 * （见 docs/roadmap.md §8.3 L-T）。
 *
 * 支持两种模式：
 *   - text —— 等宽文本（十六进制转储、错误说明等）；
 *   - tree —— 可折叠的节点树（ASN.1 结构）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

/** 树节点。 */
interface TreeNode {
  label: string
  detail?: string
  children?: TreeNode[]
}

const props = withDefaults(
  defineProps<{
    /** 文本模式的内容。 */
    text?: string
    /** 树模式的根节点。 */
    tree?: TreeNode | null
    /** 空状态提示。 */
    empty?: string
  }>(),
  { text: '', tree: null, empty: undefined },
)

const { t } = useI18n()

/** 折叠状态：按节点路径索引。 */
const collapsed = ref<Record<string, boolean>>({})

/** 树的扁平行列表（用于渲染）。 */
interface FlatRow {
  path: string
  node: TreeNode
  depth: number
  hasChildren: boolean
}

const rows = computed<FlatRow[]>(() => {
  if (!props.tree) {
    return []
  }
  const out: FlatRow[] = []
  const walk = (node: TreeNode, depth: number, path: string): void => {
    out.push({
      path,
      node,
      depth,
      hasChildren: Boolean(node.children && node.children.length > 0),
    })
    if (collapsed.value[path]) {
      return
    }
    node.children?.forEach((child, index) => {
      walk(child, depth + 1, `${path}.${index}`)
    })
  }
  walk(props.tree, 0, 'root')
  return out
})

/** 切换某个节点的折叠状态。 */
function toggle(path: string): void {
  collapsed.value = { ...collapsed.value, [path]: !collapsed.value[path] }
}

/** 是否有可显示的内容。 */
const hasContent = computed(() => rows.value.length > 0 || props.text !== '')
</script>

<template>
  <div class="code-editor">
    <p
      v-if="!hasContent"
      class="code-editor__empty"
    >
      {{ empty ?? t('tool.cert.empty') }}
    </p>

    <pre
      v-else-if="rows.length === 0"
      class="code-editor__text tb-mono"
    >{{ text }}</pre>

    <ul
      v-else
      class="code-editor__tree"
    >
      <li
        v-for="row in rows"
        :key="row.path"
        class="code-editor__node"
        :style="{ paddingLeft: `${row.depth * 16}px` }"
      >
        <button
          v-if="row.hasChildren"
          type="button"
          class="code-editor__toggle"
          :aria-expanded="!collapsed[row.path]"
          @click="toggle(row.path)"
        >
          {{ collapsed[row.path] ? '▸' : '▾' }}
        </button>
        <span
          v-else
          class="code-editor__toggle code-editor__toggle--empty"
        >·</span>

        <span class="code-editor__label tb-mono">{{ row.node.label }}</span>
        <span
          v-if="row.node.detail"
          class="code-editor__detail tb-mono"
        >{{ row.node.detail }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.code-editor {
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  background: var(--tb-bg-sunken);
  max-height: 480px;
  overflow: auto;
}

.code-editor__empty {
  padding: var(--tb-space-3);
  color: var(--tb-text-secondary);
  margin: 0;
}

.code-editor__text {
  margin: 0;
  padding: var(--tb-space-3);
  font-size: 0.85rem;
  line-height: 1.5;
  white-space: pre;
}

.code-editor__tree {
  margin: 0;
  padding: var(--tb-space-2) 0;
  list-style: none;
}

.code-editor__node {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-1);
  padding-right: var(--tb-space-3);
  font-size: 0.85rem;
  line-height: 1.6;
}

.code-editor__toggle {
  border: none;
  background: none;
  color: var(--tb-text-secondary);
  cursor: pointer;
  padding: 0;
  width: 1em;
  font-size: inherit;
}

.code-editor__toggle--empty {
  cursor: default;
}

.code-editor__label {
  color: var(--tb-text-primary);
}

.code-editor__detail {
  color: var(--tb-text-secondary);
}
</style>
