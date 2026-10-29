<script setup lang="ts">
/**
 * 证书透明度查询工具页。
 *
 * 本版**未**接入 CT 日志：后端不注册该 slug，调用返回 `NOT_IMPLEMENTED`。
 * 页面保留参数区以便展示预期契约，但运行按钮禁用并给出明确说明。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const domain = ref(DEFAULT_EXAMPLE.domain)
const matchType = ref<'exact' | 'contains' | 'wildcard'>(DEFAULT_EXAMPLE.matchType)
const limit = ref(DEFAULT_EXAMPLE.limit)

const { state, meta, error, run: execute, reset: resetRun } = useToolRun({
  slug: 'ct-search',
})

const running = computed(() => state.value === 'loading')

const faq = computed<FaqItem[]>(() => tm('tools.ct-search.faq') as FaqItem[])

function run(): void {
  void execute({ domain: domain.value, matchType: matchType.value, limit: limit.value })
}

function reset(): void {
  domain.value = DEFAULT_EXAMPLE.domain
  matchType.value = DEFAULT_EXAMPLE.matchType
  limit.value = DEFAULT_EXAMPLE.limit
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ct-search"
    name="证书透明度查询"
    name-en="Certificate Transparency Search"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <p class="prose">
            {{ t('tools.ct-search.hints.unavailable') }}
          </p>

          <label class="field">
            <span class="field__label">{{ t('tools.ct-search.fields.domain') }}</span>
            <input
              v-model="domain"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.ct-search.fields.matchType') }}</span>
            <select
              v-model="matchType"
              class="field__select"
            >
              <option value="exact">
                {{ t('tools.ct-search.options.exact') }}
              </option>
              <option value="contains">
                {{ t('tools.ct-search.options.contains') }}
              </option>
              <option value="wildcard">
                {{ t('tools.ct-search.options.wildcard') }}
              </option>
            </select>
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ct-search.fields.limit') }}</span>
            <input
              v-model.number="limit"
              type="number"
              class="field__input"
              min="1"
              max="100"
            >
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            disabled
            :show-example="false"
            @run="run"
            @reset="reset"
            @example="reset"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="''"
        :meta="meta"
        :error="error"
        :title="t('tools.ct-search.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.ct-search.description') }}
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
