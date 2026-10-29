<script setup lang="ts">
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
    enableCors: enableCors.value,
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
