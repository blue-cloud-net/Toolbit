<script setup lang="ts">
/**
 * MIME 类型查询工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES, type MimeQueryKind } from '@/tools/network/mime-lookup/example'

/** 后端 mime-lookup 工具的响应。 */
interface MimeLookupResult {
  /** 规范化后的查询词。 */
  query: string
  /** 查询方式：extension / mimeType。 */
  kind: string
  /** 是否命中数据集。 */
  matched: boolean
  /** 规范媒体类型。 */
  mimeType: string
  /** 常用扩展名。 */
  extensions: string[]
  /** 顶层类型。 */
  category: string
  /** 是否是文本类媒体类型。 */
  isTextual: boolean
  /** 默认字符集。 */
  charset: string
  /** 用途说明。 */
  description: string
  /** 历史别名与常见变体。 */
  aliases: string[]
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const kind = ref<MimeQueryKind>('extension')
const input = ref('png')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<MimeLookupResult>({
  slug: 'mime-lookup',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：媒体类型本身。 */
const resultText = computed(() => data.value?.mimeType ?? '')

/** 结论徽章：命中与否。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.matched
    ? { level: 'success' as const, label: t('tools.mime-lookup.results.matched') }
    : { level: 'info' as const, label: t('tools.mime-lookup.results.notMatched') }
})

const faq = computed<FaqItem[]>(() => tm('tools.mime-lookup.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  const value = input.value.trim()
  if (value === '') {
    return
  }
  if (kind.value === 'extension') {
    void execute({ extension: value })
    return
  }
  void execute({ mimeType: value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'mime-lookup',
  run,
  watch: [kind, input],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  kind.value = 'extension'
  input.value = 'png'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  kind.value = example.kind
  input.value = example.kind === 'extension' ? example.extension : example.mimeType
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="mime-lookup"
    :name="t('tools.mime-lookup.name')"
    name-en="MIME Type Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="row">
            <label class="field field--kind">
              <span class="field__label">{{ t('tools.mime-lookup.fields.kind') }}</span>
              <ElSelect
                v-model="kind"
                class="field__input"
                :aria-label="t('tools.mime-lookup.fields.kind')"
              >
                <ElOption
                  :label="t('tools.mime-lookup.options.kind.extension')"
                  value="extension"
                />
                <ElOption
                  :label="t('tools.mime-lookup.options.kind.mimeType')"
                  value="mimeType"
                />
              </ElSelect>
            </label>

            <label class="field field--grow">
              <span class="field__label">{{ t('tools.mime-lookup.fields.query') }}</span>
              <ElInput
                v-model="input"
                class="field__input tb-mono"
                autocomplete="off"
                :placeholder="kind === 'extension' ? 'png' : 'image/png'"
                :aria-label="t('tools.mime-lookup.fields.query')"
                @keydown.enter.prevent="run"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.mime-lookup.hints.query') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample(0)"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :verdict="verdict"
        :title="t('tools.mime-lookup.results.mimeType')"
        @retry="run"
      >
        <template
          v-if="data?.matched"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.mime-lookup.results.category') }}</dt>
              <dd class="tb-mono">
                {{ data?.category }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.mime-lookup.results.charset') }}</dt>
              <dd class="tb-mono">
                {{ data?.charset || t('tools.mime-lookup.results.none') }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.mime-lookup.results.description') }}</dt>
              <dd>
                {{ data?.description }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.mime-lookup.results.dataVersion') }}</dt>
              <dd class="tb-mono">
                {{ data?.dataVersion }}
              </dd>
            </div>
          </dl>

          <div
            v-if="data?.extensions?.length"
            class="chips"
          >
            <span class="chips__label">{{ t('tools.mime-lookup.results.extensions') }}</span>
            <span
              v-for="ext in data?.extensions ?? []"
              :key="ext"
              class="chip tb-mono"
            >
              .{{ ext }}
            </span>
          </div>

          <div
            v-if="data?.aliases?.length"
            class="chips"
          >
            <span class="chips__label">{{ t('tools.mime-lookup.results.aliases') }}</span>
            <span
              v-for="alias in data?.aliases ?? []"
              :key="alias"
              class="chip tb-mono"
            >
              {{ alias }}
            </span>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.mime-lookup.hints.output') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="(item, index) in faq"
          :key="index"
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
.row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-3);
}

.field--kind {
  min-width: 160px;
}

.field--grow {
  flex: 1 1 200px;
}

.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
  word-break: break-all;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--tb-space-1) var(--tb-space-2);
  margin-top: var(--tb-space-3);
}

.chips__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.chip {
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
