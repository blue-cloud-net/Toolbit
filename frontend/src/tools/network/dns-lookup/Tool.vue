<script setup lang="ts">
/**
 * DNS 查询工具页。
 *
 * 逐台服务器展示解析结果与耗时；单台失败不影响整体。
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

/** 单条解析结果。 */
interface DNSAnswer {
  name: string
  type: string
  ttl: number
  value: string
}

/** 单台服务器的结果。 */
interface DNSServerResult {
  server: string
  proto: string
  latencyMs: number
  answers: DNSAnswer[]
  error: { code: string; message: string } | null
}

/** 后端 dns-lookup 工具的响应。 */
interface DNSLookupResult {
  host: string
  type: string
  perServer: DNSServerResult[]
}

interface FaqItem {
  q: string
  a: string
}

const RECORD_TYPES = [
  'A',
  'AAAA',
  'CNAME',
  'MX',
  'TXT',
  'NS',
  'SOA',
  'SRV',
  'CAA',
  'PTR',
  'ANY',
]
const PROTOS = ['udp', 'tcp', 'dot', 'doh']

const { t, tm } = useI18n()

const host = ref(DEFAULT_EXAMPLE.host)
const type = ref(DEFAULT_EXAMPLE.type)
const proto = ref('udp')
const addr = ref(DEFAULT_EXAMPLE.addr)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DNSLookupResult>({ slug: 'dns-lookup' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const lines: string[] = []
  for (const server of data.value.perServer) {
    lines.push(`# ${server.server} (${server.proto}) ${server.latencyMs}ms`)
    if (server.error) {
      lines.push(`  ! ${server.error.code}: ${server.error.message}`)
    }
    for (const answer of server.answers) {
      lines.push(`  ${answer.name} ${answer.ttl}s ${answer.type} ${answer.value}`)
    }
    lines.push('')
  }
  return lines.join('\n').trimEnd()
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-lookup.faq') as FaqItem[])

/** 汇总：成功数、最慢节点、失败节点。 */
const summary = computed(() => {
  const result = data.value
  if (!result || result.perServer.length === 0) {
    return null
  }
  const ok = result.perServer.filter((item) => item.error === null)
  const failed = result.perServer.length - ok.length
  let slowest = ''
  let slowestMs = -1
  for (const item of ok) {
    if (item.latencyMs > slowestMs) {
      slowest = `${item.latencyMs}ms (${item.server})`
      slowestMs = item.latencyMs
    }
  }
  return {
    succeeded: ok.length,
    failed,
    slowest,
    total: result.perServer.length,
  }
})

function run(): void {
  const payload: Record<string, unknown> = { host: host.value, type: type.value }
  if (addr.value.trim() !== '') {
    payload.servers = [{ proto: proto.value, addr: addr.value.trim() }]
  }
  void execute(payload)
}

function reset(): void {
  host.value = DEFAULT_EXAMPLE.host
  type.value = DEFAULT_EXAMPLE.type
  proto.value = 'udp'
  addr.value = DEFAULT_EXAMPLE.addr
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dns-lookup"
    name="DNS 查询"
    name-en="DNS Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.host') }}</span>
            <ElInput
              v-model="host"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.type') }}</span>
            <ElSelect
              v-model="type"
              class="field__select"
              :aria-label="t('tools.dns-lookup.fields.type')"
            >
              <ElOption
                v-for="name in RECORD_TYPES"
                :key="name"
                :label="String(name)"
                :value="name"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.proto') }}</span>
            <ElSelect
              v-model="proto"
              class="field__select"
              :aria-label="t('tools.dns-lookup.fields.proto')"
            >
              <ElOption
                v-for="name in PROTOS"
                :key="name"
                :label="String(name)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.addr') }}</span>
            <ElInput
              v-model="addr"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="1.1.1.1:53"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="host.trim() === ''"
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
        :title="t('tools.dns-lookup.results.output')"
        @retry="run"
      >
        <template
          v-if="summary"
          #summary
        >
          <span>{{ t('tool.result.summary.succeeded') }} {{ summary.succeeded }} / {{ summary.total }}</span>
          <span v-if="summary.slowest">{{ t('tool.result.summary.slowest') }} {{ summary.slowest }}</span>
          <span v-if="summary.failed > 0">{{ t('tool.result.summary.failed') }} {{ summary.failed }}</span>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-lookup.description') }}
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
