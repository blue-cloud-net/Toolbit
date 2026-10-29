<script setup lang="ts">
/**
 * DNS 综合检测工具页。
 *
 * 三个检测项：whois（RDAP / whois）、dnssec、ptr。
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
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

/** 后端 dns-tester 工具的响应。 */
interface DNSTesterResult {
  action: string
  whois: Record<string, unknown> | null
  dnssec: Record<string, unknown> | null
  ptr: Record<string, unknown> | null
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const action = ref<'whois' | 'dnssec' | 'ptr'>(DEFAULT_EXAMPLE.action)
const domain = ref(DEFAULT_EXAMPLE.domain)
const ip = ref(DEFAULT_EXAMPLE.ip)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DNSTesterResult>({ slug: 'dns-tester' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const payload =
    data.value.whois ?? data.value.dnssec ?? data.value.ptr ?? data.value
  return JSON.stringify(payload, null, 2)
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-tester.faq') as FaqItem[])

const canRun = computed(() =>
  action.value === 'ptr' ? ip.value.trim() !== '' || domain.value.trim() !== '' : domain.value.trim() !== '',
)

function run(): void {
  const payload: Record<string, unknown> = { action: action.value }
  if (domain.value.trim() !== '') {
    payload.domain = domain.value.trim()
  }
  if (ip.value.trim() !== '') {
    payload.ip = ip.value.trim()
  }
  void execute(payload)
}

function reset(): void {
  action.value = DEFAULT_EXAMPLE.action
  domain.value = DEFAULT_EXAMPLE.domain
  ip.value = DEFAULT_EXAMPLE.ip
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dns-tester"
    name="DNS 多记录测试"
    name-en="DNS Record Tester"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.action') }}</span>
            <ElSelect
              v-model="action"
              class="field__select"
              :aria-label="t('tools.dns-tester.fields.action')"
            >
              <ElOption
                :label="t('tools.dns-tester.options.whois')"
                value="whois"
              />
              <ElOption
                :label="t('tools.dns-tester.options.dnssec')"
                value="dnssec"
              />
              <ElOption
                :label="t('tools.dns-tester.options.ptr')"
                value="ptr"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.ip') }}</span>
            <ElInput
              v-model="ip"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="8.8.8.8"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="!canRun"
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
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.dns-tester.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-tester.description') }}
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
