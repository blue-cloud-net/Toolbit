<script setup lang="ts">
/**
 * ASN.1 视图工具页。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import CodeViewer from '@/components/tool/CodeViewer.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/asn1-view/example'

/** ASN.1 节点（与后端响应一致）。 */
interface ASN1Node {
  offset: number
  tag: number
  class: number
  number: number
  constructed: boolean
  length: number
  value: string
  truncated: boolean
  children?: ASN1Node[]
}

/** 后端 asn1-view 工具的响应。 */
interface ASN1ViewResult {
  length: number
  tree: ASN1Node
  nodeCount: number
  depth: number
  maxDepth: number
  hexDump: string
}

/** 树节点（供 CodeViewer 渲染）。 */
interface TreeNode {
  label: string
  detail?: string
  children?: TreeNode[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('')
const inputFormat = ref('auto')
const maxDepth = ref(0)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<ASN1ViewResult>({ slug: 'asn1-view' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 标签类别的可读名。 */
const CLASS_NAMES: Record<number, string> = {
  0: 'universal',
  1: 'application',
  2: 'context',
  3: 'private',
}

/** 把 ASN.1 节点转成 CodeEditor 的树。 */
function toTree(node: ASN1Node): TreeNode {
  const cls = CLASS_NAMES[node.class] ?? `class-${node.class}`
  const label = `${node.tag === 0 ? '' : `#${node.number} `}${cls}`
  const detail = node.constructed
    ? `length=${node.length} offset=${node.offset}`
    : `length=${node.length} ${node.value.slice(0, 32)}`
  const tree: TreeNode = { label, detail }
  if (node.children && node.children.length > 0) {
    tree.children = node.children.map((child) => toTree(child))
  }
  if (node.truncated) {
    tree.children = [...(tree.children ?? []), { label: '…', detail: 'truncated' }]
  }
  return tree
}

/** 结果树。 */
const resultTree = computed<TreeNode | null>(() => (data.value ? toTree(data.value.tree) : null))

/** 概要文本。 */
const summary = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.asn1-view.results.length')}: ${result.length}`,
    `${t('tools.asn1-view.results.nodeCount')}: ${result.nodeCount}`,
    `${t('tools.asn1-view.results.depth')}: ${result.depth}`,
    `${t('tools.asn1-view.results.maxDepth')}: ${result.maxDepth}`,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.asn1-view.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value, format: inputFormat.value },
    maxDepth: maxDepth.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  inputFormat.value = 'auto'
  maxDepth.value = 0
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  input.value = DEFAULT_EXAMPLE.data
  inputFormat.value = DEFAULT_EXAMPLE.format
  maxDepth.value = 0
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="asn1-view"
    :name="t('tools.asn1-view.name')"
    name-en="ASN.1 Viewer"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="input"
            v-model:format="inputFormat"
            :label="t('tools.asn1-view.fields.input')"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.asn1-view.fields.maxDepth') }}</span>
            <input
              v-model.number="maxDepth"
              type="number"
              class="field__input"
              min="0"
              max="64"
            >
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="input === ''"
            @run="run"
            @reset="reset"
            @example="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <div class="asn1-view">
        <p
          v-if="summary"
          class="asn1-view__summary tb-mono"
        >
          {{ summary }}
        </p>
        <CodeViewer
          :tree="resultTree"
          :text="data?.hexDump ?? ''"
          :empty="t('tools.asn1-view.results.empty')"
        />
        <ResultPanel
          :state="state"
          :meta="meta"
          :error="error"
          :title="t('tools.asn1-view.results.tree')"
          :use-as-input="useAsInput"
          @retry="run"
        />
      </div>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.asn1-view.description') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
        >
          <dt class="faq__q">
            {{ item.q }}
          </dt>
          <dd class="faq__a">
            {{ item.a }}
          </dd>
        </template>
      </dl>
    </template>
  </ToolPageLayout>
</template>

<style scoped>
.asn1-view {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.asn1-view__summary {
  margin: 0;
  color: var(--tb-text-secondary);
  font-size: 0.85rem;
  white-space: pre-wrap;
}
</style>
