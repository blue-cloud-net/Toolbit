<script setup lang="ts">
/**
 * UUID 生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/uuid-gen/example'
import type { UUIDVersion } from '@/tools/data/uuid-gen/example'

/** 后端 uuid-gen 工具的响应。 */
interface UUIDGenResult {
  /** 生成的 UUID 列表。 */
  uuids: string[]
  /** 生成条数。 */
  count: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 单次生成数量上限；与服务端一致。 */
const MAX_COUNT = 1000

const { t, tm } = useI18n()

const version = ref<UUIDVersion>(4)
const count = ref(1)
const uppercase = ref(false)
const hyphens = ref(true)
const braces = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<UUIDGenResult>({
  slug: 'uuid-gen',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：每条一行。 */
const resultText = computed(() => (data.value?.uuids ?? []).join('\n'))

/** 结果条数。 */
const resultCount = computed(() => data.value?.count ?? 0)

const faq = computed<FaqItem[]>(() => tm('tools.uuid-gen.faq') as FaqItem[])

/** 解析整数输入；非法时返回 undefined。 */
function parseIntOrUndefined(raw: number): number | undefined {
  return Number.isFinite(raw) ? Math.trunc(raw) : undefined
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    version: version.value,
    count: parseIntOrUndefined(count.value) ?? 1,
    uppercase: uppercase.value,
    hyphens: hyphens.value,
    braces: braces.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  version.value = 4
  count.value = 1
  uppercase.value = false
  hyphens.value = true
  braces.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  version.value = example.version
  count.value = example.count
  uppercase.value = example.uppercase
  hyphens.value = example.hyphens
  braces.value = example.braces
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="uuid-gen"
    :name="t('tools.uuid-gen.name')"
    name-en="UUID Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.uuid-gen.fields.version') }}</span>
              <select
                v-model="version"
                class="field__select"
              >
                <option :value="1">
                  {{ t('tools.uuid-gen.options.version.1') }}
                </option>
                <option :value="4">
                  {{ t('tools.uuid-gen.options.version.4') }}
                </option>
                <option :value="7">
                  {{ t('tools.uuid-gen.options.version.7') }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.uuid-gen.fields.count') }}</span>
              <input
                v-model.number="count"
                type="number"
                class="field__input tb-mono"
                min="1"
                :max="MAX_COUNT"
                step="1"
                :aria-label="t('tools.uuid-gen.fields.count')"
              >
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.uuid-gen.hints.version') }}
          </p>
        </template>

        <template #options>
          <label class="field field--inline">
            <input
              v-model="uppercase"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.uuid-gen.fields.uppercase') }}</span>
          </label>

          <label class="field field--inline">
            <input
              v-model="hyphens"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.uuid-gen.fields.hyphens') }}</span>
          </label>

          <label class="field field--inline">
            <input
              v-model="braces"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.uuid-gen.fields.braces') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.uuid-gen.hints.format') }}
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
        :title="t('tools.uuid-gen.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>
                {{ t('tools.uuid-gen.results.count') }}
              </dt>
              <dd class="tb-mono">
                {{ resultCount }}
              </dd>
            </div>
          </dl>

          <ol class="uuids">
            <li
              v-for="(uuid, index) in data?.uuids ?? []"
              :key="uuid + String(index)"
              class="uuids__item"
            >
              <span class="uuids__label">
                {{ t('tools.uuid-gen.results.item', { index: index + 1 }) }}
              </span>
              <span class="uuids__value tb-mono">{{ uuid }}</span>
            </li>
          </ol>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.uuid-gen.hints.output') }}
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

.uuids {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
  padding-left: var(--tb-space-5);
  font-size: var(--tb-font-sm);
}

.uuids__item {
  word-break: break-all;
}

.uuids__label {
  margin-right: var(--tb-space-2);
  color: var(--tb-text-tertiary);
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
