<script setup lang="ts">
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
const resultText = computed(() => data.value?.commands.join('\n') ?? '')

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
