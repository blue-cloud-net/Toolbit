<script setup lang="ts">
/**
 * 中国行政区划工具页。
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
import { EXAMPLES } from '@/tools/data/cn-regions/example'

/** 一个行政区划条目。 */
interface Region {
  /** 六位行政区划代码。 */
  code: string
  /** 名称。 */
  name: string
  /** 层级：province / city / district。 */
  level: string
  /** 上级代码；省级为空。 */
  parent?: string
  /** 所属省级代码。 */
  province: string
}

/** 后端 cn-regions 工具的响应。 */
interface CNRegionsResult {
  /** 省级行政区。 */
  provinces: Region[]
  /** 地级 / 县级行政区。 */
  cities: Region[]
  /** 匹配条数。 */
  total: number
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const query = ref('')
const province = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CNRegionsResult>({
  slug: 'cn-regions',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：每行一条。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [...result.provinces, ...result.cities]
    .map((region) => `${region.code}  ${region.name}  ${region.level}`)
    .join('\n')
})

/** 层级显示文案。 */
function levelLabel(level: string): string {
  const key = `tools.cn-regions.options.level.${level}`
  const translated = t(key)
  return translated === key ? level : translated
}

const faq = computed<FaqItem[]>(() => tm('tools.cn-regions.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ query: query.value, province: province.value })
}

/** 重置表单与结果。 */
function reset(): void {
  query.value = ''
  province.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  query.value = example.query
  province.value = example.province
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cn-regions"
    :name="t('tools.cn-regions.name')"
    name-en="China Administrative Regions"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.cn-regions.fields.query') }}</span>
            <input
              v-model="query"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="杭州"
              :aria-label="t('tools.cn-regions.fields.query')"
              @keydown.enter.prevent="run"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cn-regions.fields.province') }}</span>
            <input
              v-model="province"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="浙江省"
              :aria-label="t('tools.cn-regions.fields.province')"
              @keydown.enter.prevent="run"
            >
          </label>

          <p class="field__hint">
            {{ t('tools.cn-regions.hints.query') }}
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
        :title="t('tools.cn-regions.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>
                {{ t('tools.cn-regions.results.total') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.total }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.cn-regions.results.dataVersion') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.dataVersion }}
              </dd>
            </div>
          </dl>

          <section
            v-if="(data?.provinces ?? []).length > 0"
            class="group"
          >
            <h3 class="group__title">
              {{ t('tools.cn-regions.results.provinces') }}
            </h3>
            <div class="table-wrap">
              <table class="table">
                <thead>
                  <tr>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.code') }}
                    </th>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.name') }}
                    </th>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.level') }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="region in data?.provinces ?? []"
                    :key="region.code"
                  >
                    <td class="tb-mono">
                      {{ region.code }}
                    </td>
                    <td>
                      {{ region.name }}
                    </td>
                    <td>
                      {{ levelLabel(region.level) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section
            v-if="(data?.cities ?? []).length > 0"
            class="group"
          >
            <h3 class="group__title">
              {{ t('tools.cn-regions.results.cities') }}
            </h3>
            <div class="table-wrap">
              <table class="table">
                <thead>
                  <tr>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.code') }}
                    </th>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.name') }}
                    </th>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.level') }}
                    </th>
                    <th scope="col">
                      {{ t('tools.cn-regions.results.provinceCode') }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="region in data?.cities ?? []"
                    :key="region.code"
                  >
                    <td class="tb-mono">
                      {{ region.code }}
                    </td>
                    <td>
                      {{ region.name }}
                    </td>
                    <td>
                      {{ levelLabel(region.level) }}
                    </td>
                    <td class="tb-mono">
                      {{ region.province }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cn-regions.hints.output') }}
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

.group {
  margin-top: var(--tb-space-3);
}

.group__title {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.table-wrap {
  max-height: 420px;
  overflow: auto;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.table {
  width: 100%;
  font-size: var(--tb-font-sm);
  border-collapse: collapse;
}

.table th,
.table td {
  padding: var(--tb-space-1) var(--tb-space-3);
  text-align: left;
  white-space: nowrap;
  border-bottom: 1px solid var(--tb-border);
}

.table th {
  position: sticky;
  top: 0;
  font-weight: 600;
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
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
