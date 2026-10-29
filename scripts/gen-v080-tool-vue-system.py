#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的 Tool.vue（系统 / 网络 / 格式化 / 编码 / 文档类）。"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
BASE = ROOT / "frontend" / "src" / "tools"

COMPONENTS: dict[str, str] = {}

COMPONENTS["system/nginx-config"] = '''<script setup lang="ts">
/**
 * Nginx 配置生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
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
import { EXAMPLES } from '@/tools/system/nginx-config/example'

/** 后端 nginx-config 工具的响应。 */
interface NginxConfigResult {
  /** 配置模式。 */
  mode: string
  /** 配置文本。 */
  config: string
  /** 配置段标题。 */
  sections: string[]
  /** 配置提示。 */
  warnings: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('site')
const serverName = ref('example.com')
const listenPort = ref(80)
const root = ref('/var/www/html')
const enableSsl = ref(false)
const sslCertificate = ref('/etc/ssl/certs/example.com.pem')
const sslPrivateKey = ref('/etc/ssl/private/example.com.key')
const enableProxy = ref(false)
const proxyPass = ref('http://127.0.0.1:8080')
const proxyPath = ref('/')
const enableGzip = ref(false)
const enableCors = ref(false)
const allowOrigin = ref('')
const enableWebSocket = ref(false)
const enableLoadBalance = ref(false)
const upstreams = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<NginxConfigResult>({
  slug: 'nginx-config',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：配置文本。 */
const resultText = computed(() => data.value?.config ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.nginx-config.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    serverName: serverName.value,
    listenPort: listenPort.value,
    root: root.value,
    enableSsl: enableSsl.value,
    sslCertificate: sslCertificate.value,
    sslPrivateKey: sslPrivateKey.value,
    enableProxy: enableProxy.value,
    proxyPass: proxyPass.value,
    proxyPath: proxyPath.value,
    enableGzip: enableGzip.value,
    enableCORS: enableCors.value,
    allowOrigin: allowOrigin.value,
    enableWebSocket: enableWebSocket.value,
    enableLoadBalance: enableLoadBalance.value,
    upstreams: upstreams.value
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item !== ''),
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'nginx-config',
  run,
  watch: [
    mode, serverName, listenPort, root, enableSsl, enableProxy,
    enableGzip, enableCors, enableWebSocket, enableLoadBalance, upstreams,
  ],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'site'
  serverName.value = 'example.com'
  listenPort.value = 80
  root.value = '/var/www/html'
  enableSsl.value = false
  enableProxy.value = false
  enableGzip.value = false
  enableCors.value = false
  enableWebSocket.value = false
  enableLoadBalance.value = false
  upstreams.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  serverName.value = example.serverName
  root.value = example.root
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="nginx-config"
    :name="t('tools.nginx-config.name')"
    name-en="Nginx Config Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.nginx-config.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.nginx-config.fields.mode')"
            >
              <ElOption
                :label="t('tools.nginx-config.options.mode.global')"
                value="global"
              />
              <ElOption
                :label="t('tools.nginx-config.options.mode.site')"
                value="site"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'site'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.serverName') }}</span>
            <ElInput
              v-model="serverName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.serverName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="mode === 'site'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.root') }}</span>
            <ElInput
              v-model="root"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.root')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.nginx-config.fields.listenPort') }}</span>
            <ElInputNumber
              v-model="listenPort"
              class="field__select"
              :min="0"
              :max="65535"
              :aria-label="t('tools.nginx-config.fields.listenPort')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableSsl"
              :aria-label="t('tools.nginx-config.fields.enableSsl')"
            >
              {{ t('tools.nginx-config.fields.enableSsl') }}
            </ElCheckbox>
          </label>

          <label
            v-if="enableSsl"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.sslCertificate') }}</span>
            <ElInput
              v-model="sslCertificate"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.sslCertificate')"
            />
          </label>

          <label
            v-if="enableSsl"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.sslPrivateKey') }}</span>
            <ElInput
              v-model="sslPrivateKey"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.sslPrivateKey')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableProxy"
              :aria-label="t('tools.nginx-config.fields.enableProxy')"
            >
              {{ t('tools.nginx-config.fields.enableProxy') }}
            </ElCheckbox>
          </label>

          <label
            v-if="enableProxy"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.proxyPass') }}</span>
            <ElInput
              v-model="proxyPass"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.proxyPass')"
            />
          </label>

          <label
            v-if="enableProxy"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.proxyPath') }}</span>
            <ElInput
              v-model="proxyPath"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.nginx-config.fields.proxyPath')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableWebSocket"
              :aria-label="t('tools.nginx-config.fields.enableWebSocket')"
            >
              {{ t('tools.nginx-config.fields.enableWebSocket') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableLoadBalance"
              :aria-label="t('tools.nginx-config.fields.enableLoadBalance')"
            >
              {{ t('tools.nginx-config.fields.enableLoadBalance') }}
            </ElCheckbox>
          </label>

          <label
            v-if="enableLoadBalance"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.upstreams') }}</span>
            <ElInput
              v-model="upstreams"
              class="field__input tb-mono"
              placeholder="127.0.0.1:8081,127.0.0.1:8082"
              :aria-label="t('tools.nginx-config.fields.upstreams')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableGzip"
              :aria-label="t('tools.nginx-config.fields.enableGzip')"
            >
              {{ t('tools.nginx-config.fields.enableGzip') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="enableCors"
              :aria-label="t('tools.nginx-config.fields.enableCors')"
            >
              {{ t('tools.nginx-config.fields.enableCors') }}
            </ElCheckbox>
          </label>

          <label
            v-if="enableCors"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.nginx-config.fields.allowOrigin') }}</span>
            <ElInput
              v-model="allowOrigin"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="*"
              :aria-label="t('tools.nginx-config.fields.allowOrigin')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.nginx-config.hints.config') }}
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
        :title="t('tools.nginx-config.results.config')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-if="data && data.sections.length > 0"
              class="detail__row"
            >
              <dt>{{ t('tools.nginx-config.results.sections') }}</dt>
              <dd class="tb-mono">
                {{ data.sections.join(' · ') }}
              </dd>
            </div>
            <div
              v-for="(warning, index) in data?.warnings ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ warning }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
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
'''

COMPONENTS["system/firewall-rules"] = '''<script setup lang="ts">
/**
 * 防火墙规则生成工具页。
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
import { EXAMPLES } from '@/tools/system/firewall-rules/example'

/** 一条规则。 */
interface FirewallRule {
  /** 动作。 */
  action: string
  /** 端口。 */
  port: string
  /** 协议。 */
  protocol: string
  /** 来源地址。 */
  source: string
  /** 说明。 */
  description: string
}

/** 后端 firewall-rules 工具的响应。 */
interface FirewallRulesResult {
  /** 防火墙工具。 */
  tool: string
  /** 场景。 */
  scenario: string
  /** 规则列表。 */
  rules: FirewallRule[]
  /** 可执行命令。 */
  commands: string[]
  /** 使用提示。 */
  notes: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const tool = ref('ufw')
const scenario = ref('https')
const ports = ref('')
const protocol = ref('tcp')
const source = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<FirewallRulesResult>({
  slug: 'firewall-rules',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：可执行命令。 */
const resultText = computed(() => data.value?.commands.join('\\n') ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.firewall-rules.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    tool: tool.value,
    scenario: scenario.value,
    ports: ports.value
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item !== ''),
    protocol: protocol.value,
    source: source.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'firewall-rules',
  run,
  watch: [tool, scenario, ports, protocol, source],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  tool.value = 'ufw'
  scenario.value = 'https'
  ports.value = ''
  protocol.value = 'tcp'
  source.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  scenario.value = example.scenario
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="firewall-rules"
    :name="t('tools.firewall-rules.name')"
    name-en="Firewall Rule Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.firewall-rules.fields.tool') }}</span>
            <ElSelect
              v-model="tool"
              class="field__select"
              :aria-label="t('tools.firewall-rules.fields.tool')"
            >
              <ElOption
                :label="t('tools.firewall-rules.options.tool.ufw')"
                value="ufw"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.tool.firewalld')"
                value="firewalld"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.tool.iptables')"
                value="iptables"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.tool.nftables')"
                value="nftables"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.firewall-rules.fields.scenario') }}</span>
            <ElSelect
              v-model="scenario"
              class="field__select"
              :aria-label="t('tools.firewall-rules.fields.scenario')"
            >
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.http')"
                value="http"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.https')"
                value="https"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.ssh')"
                value="ssh"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.database')"
                value="database"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.denyAll')"
                value="denyAll"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.scenario.allowAll')"
                value="allowAll"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.firewall-rules.fields.ports') }}</span>
            <ElInput
              v-model="ports"
              class="field__input tb-mono"
              placeholder="8080,8000-8100"
              autocomplete="off"
              :aria-label="t('tools.firewall-rules.fields.ports')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.firewall-rules.fields.protocol') }}</span>
            <ElSelect
              v-model="protocol"
              class="field__select"
              :aria-label="t('tools.firewall-rules.fields.protocol')"
            >
              <ElOption
                :label="t('tools.firewall-rules.options.protocol.tcp')"
                value="tcp"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.protocol.udp')"
                value="udp"
              />
              <ElOption
                :label="t('tools.firewall-rules.options.protocol.any')"
                value="any"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.firewall-rules.fields.source') }}</span>
            <ElInput
              v-model="source"
              class="field__input tb-mono"
              placeholder="203.0.113.0/24"
              autocomplete="off"
              :aria-label="t('tools.firewall-rules.fields.source')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.firewall-rules.hints.commands') }}
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
        :title="t('tools.firewall-rules.results.commands')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-for="(note, index) in data?.notes ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ note }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
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
'''

COMPONENTS["system/lookup-table"] = '''<script setup lang="ts">
/**
 * 速查表查询工具页。
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
import { EXAMPLES } from '@/tools/system/lookup-table/example'

/** 一个速查条目。 */
interface LookupEntry {
  /** 类别。 */
  kind: string
  /** 查询键。 */
  key: string
  /** 简短名称。 */
  label: string
  /** 中文说明。 */
  description: string
  /** 附加信息。 */
  extra: string[]
}

/** 后端 lookup-table 工具的响应。 */
interface LookupTableResult {
  /** 类别。 */
  kind: string
  /** 查询词。 */
  query: string
  /** 是否精确命中。 */
  matched: boolean
  /** 命中的条目。 */
  entry: LookupEntry | null
  /** 模糊检索的候选。 */
  candidates: LookupEntry[]
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const kind = ref('port')
const key = ref('443')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<LookupTableResult>({
  slug: 'lookup-table',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：命中条目或候选列表。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  if (data.value.entry) {
    return `${data.value.entry.label}\\t${data.value.entry.description}`
  }
  return data.value.candidates
    .map((candidate) => `${candidate.label}\\t${candidate.description}`)
    .join('\\n')
})

/** 结论徽章：是否精确命中。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.matched
    ? { level: 'success' as const, label: t('tools.lookup-table.results.matched') }
    : { level: 'info' as const, label: t('tools.lookup-table.results.notMatched') }
})

const faq = computed<FaqItem[]>(() => tm('tools.lookup-table.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (key.value.trim() === '') {
    return
  }
  void execute({ kind: kind.value, key: key.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'lookup-table',
  run,
  watch: [kind, key],
  enabled: () => key.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  kind.value = 'port'
  key.value = '443'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  kind.value = example.kind
  key.value = example.key
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="lookup-table"
    :name="t('tools.lookup-table.name')"
    name-en="Lookup Table"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.lookup-table.fields.kind') }}</span>
            <ElSelect
              v-model="kind"
              class="field__select"
              :aria-label="t('tools.lookup-table.fields.kind')"
            >
              <ElOption
                :label="t('tools.lookup-table.options.kind.port')"
                value="port"
              />
              <ElOption
                :label="t('tools.lookup-table.options.kind.exitcode')"
                value="exitcode"
              />
              <ElOption
                :label="t('tools.lookup-table.options.kind.timestamp')"
                value="timestamp"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.lookup-table.fields.key') }}</span>
            <ElInput
              v-model="key"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.lookup-table.fields.key')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.lookup-table.hints.key') }}
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
        :title="t('tools.lookup-table.results.entry')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-if="data?.entry"
              class="detail__row"
            >
              <dt>{{ data.entry.label }}</dt>
              <dd class="tb-mono">
                {{ data.entry.extra.join(' · ') }}
              </dd>
            </div>
            <div
              v-for="candidate in data?.candidates ?? []"
              :key="candidate.key"
              class="detail__row"
            >
              <dt class="tb-mono">
                {{ candidate.key }}
              </dt>
              <dd>{{ candidate.label }}</dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.lookup-table.results.dataVersion') }}</dt>
              <dd class="tb-mono">
                {{ data?.dataVersion ?? '' }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
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
'''

COMPONENTS["system/docker-compose-gen"] = '''<script setup lang="ts">
/**
 * Compose 配置生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/system/docker-compose-gen/example'

/** 一个服务定义。 */
interface ComposeService {
  /** 服务名。 */
  name: string
  /** 镜像名。 */
  image: string
  /** 端口映射。 */
  ports: string[]
  /** 环境变量。 */
  environment: Record<string, string>
  /** 卷挂载。 */
  volumes: string[]
  /** 依赖的服务。 */
  dependsOn: string[]
  /** 重启策略。 */
  restart: string
  /** 启动命令。 */
  command: string
}

/** 后端 docker-compose-gen 工具的响应。 */
interface DockerComposeResult {
  /** 配置内容。 */
  compose: string
  /** 服务数量。 */
  serviceCount: number
  /** 配置提示。 */
  warnings: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const projectName = ref('demo')
const servicesJson = ref(
  '[\\n  {\\n    "name": "web",\\n    "image": "nginx:alpine",\\n    "ports": ["8080:80"]\\n  }\\n]',
)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DockerComposeResult>({
  slug: 'docker-compose-gen',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：配置内容。 */
const resultText = computed(() => data.value?.compose ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.docker-compose-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (projectName.value.trim() === '') {
    return
  }
  let services: ComposeService[]
  try {
    services = JSON.parse(servicesJson.value) as ComposeService[]
  } catch {
    services = []
  }
  void execute({
    projectName: projectName.value,
    services,
    includeNetwork: true,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'docker-compose-gen',
  run,
  watch: [projectName, servicesJson],
  enabled: () => projectName.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  projectName.value = 'demo'
  servicesJson.value =
    '[\\n  {\\n    "name": "web",\\n    "image": "nginx:alpine",\\n    "ports": ["8080:80"]\\n  }\\n]'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  projectName.value = example.projectName
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="docker-compose-gen"
    :name="t('tools.docker-compose-gen.name')"
    name-en="Docker Compose Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.docker-compose-gen.fields.projectName') }}</span>
            <ElInput
              v-model="projectName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.docker-compose-gen.fields.projectName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.docker-compose-gen.fields.services') }}</span>
            <ElInput
              v-model="servicesJson"
              class="field__input tb-mono"
              type="textarea"
              :rows="10"
              autocomplete="off"
              :aria-label="t('tools.docker-compose-gen.fields.services')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.docker-compose-gen.hints.services') }}
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
        :title="t('tools.docker-compose-gen.results.compose')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.docker-compose-gen.results.serviceCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.serviceCount ?? 0 }}
              </dd>
            </div>
            <div
              v-for="(warning, index) in data?.warnings ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ warning }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
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
'''


def main() -> None:
    for relative, content in COMPONENTS.items():
        target = BASE / relative / "Tool.vue"
        target.write_text(content, encoding="utf-8")
        print(f"写入 {relative}/Tool.vue")


if __name__ == "__main__":
    main()
