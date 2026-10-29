<script setup lang="ts">
/**
 * 国家代码工具页。
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
import { EXAMPLES } from '@/tools/data/country-code/example'

/** 一个国家与地区条目。 */
interface Country {
  /** 两位字母代码。 */
  alpha2: string
  /** 三位字母代码。 */
  alpha3: string
  /** 三位数字代码。 */
  numeric: string
  /** 中文名。 */
  nameZh: string
  /** 英文名。 */
  nameEn: string
}

/** 后端 country-code 工具的响应。 */
interface CountryCodeResult {
  /** 匹配的国家与地区。 */
  countries: Country[]
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

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CountryCodeResult>({
  slug: 'country-code',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：每行一条。 */
const resultText = computed(() =>
  (data.value?.countries ?? [])
    .map((item) => `${item.alpha2}  ${item.alpha3}  ${item.numeric}  ${item.nameZh}  ${item.nameEn}`)
    .join('\n'),
)

const faq = computed<FaqItem[]>(() => tm('tools.country-code.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ query: query.value })
}

/** 重置表单与结果。 */
function reset(): void {
  query.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  query.value = example.query
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="country-code"
    :name="t('tools.country-code.name')"
    name-en="Country Codes"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.country-code.fields.query') }}</span>
            <input
              v-model="query"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="中国"
              :aria-label="t('tools.country-code.fields.query')"
              @keydown.enter.prevent="run"
            >
          </label>

          <p class="field__hint">
            {{ t('tools.country-code.hints.query') }}
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
        :title="t('tools.country-code.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>
                {{ t('tools.country-code.results.total') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.total }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.country-code.results.dataVersion') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.dataVersion }}
              </dd>
            </div>
          </dl>

          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr>
                  <th scope="col">
                    {{ t('tools.country-code.results.alpha2') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.country-code.results.alpha3') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.country-code.results.numeric') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.country-code.results.nameZh') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.country-code.results.nameEn') }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="country in data?.countries ?? []"
                  :key="country.alpha2"
                >
                  <td class="tb-mono">
                    {{ country.alpha2 }}
                  </td>
                  <td class="tb-mono">
                    {{ country.alpha3 }}
                  </td>
                  <td class="tb-mono">
                    {{ country.numeric }}
                  </td>
                  <td>
                    {{ country.nameZh }}
                  </td>
                  <td>
                    {{ country.nameEn }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.country-code.hints.output') }}
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

.table-wrap {
  max-height: 420px;
  margin-top: var(--tb-space-3);
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
