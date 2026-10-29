<script setup lang="ts">
/**
 * cURL 命令生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import {
  EXAMPLES,
  type AuthType,
  type BodyFormat,
  type CurlHeaderExample,
  type HttpMethod,
  type Shell,
} from '@/tools/system/curl-generator/example'

/** 后端 curl-generator 工具的响应。 */
interface CurlGeneratorResult {
  /** 生成的命令文本。 */
  command: string
  /** 目标 Shell。 */
  shell: string
  /** 被占位符替代的凭据字段名。 */
  placeholders?: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 受支持的 HTTP 方法。 */
const METHODS: readonly HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']

/** 受支持的请求体类型。 */
const BODY_FORMATS: readonly BodyFormat[] = ['json', 'form', 'raw']

/** 受支持的认证方式。 */
const AUTH_TYPES: readonly AuthType[] = ['none', 'basic', 'bearer']

/** 受支持的目标 Shell。 */
const SHELLS: readonly Shell[] = ['bash', 'sh', 'powershell', 'cmd']

/** 超时上限，与服务端保持一致。 */
const MAX_TIMEOUT = 3600

const method = ref<HttpMethod>('GET')
const url = ref('')
const headers = ref<CurlHeaderExample[]>([])
const body = ref('')
const bodyFormat = ref<BodyFormat>('json')
const auth = ref<AuthType>('none')
const username = ref('')
const password = ref('')
const token = ref('')
const insecure = ref(false)
const followRedirects = ref(false)
const timeout = ref(0)
const shell = ref<Shell>('bash')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CurlGeneratorResult>({
  slug: 'curl-generator',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的命令。 */
const resultText = computed(() => data.value?.command ?? '')

/** 被占位符替代的字段。 */
const placeholders = computed<string[]>(() => data.value?.placeholders ?? [])

/** 占位符的展示文本。 */
const placeholderText = computed(() => placeholders.value.map((name) => `<${name}>`).join(' '))

/** 超时秒数；空值或越界都归一到服务端允许的范围。 */
const timeoutValue = computed(() => {
  const value = Number(timeout.value)
  if (!Number.isFinite(value) || value <= 0) {
    return 0
  }
  return Math.min(Math.trunc(value), MAX_TIMEOUT)
})

const faq = computed<FaqItem[]>(() => tm('tools.curl-generator.faq') as FaqItem[])

/** 新增一个请求头。 */
function addHeader(): void {
  headers.value.push({ name: '', value: '', sensitive: false })
}

/** 删除指定请求头。 */
function removeHeader(index: number): void {
  headers.value.splice(index, 1)
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    method: method.value,
    url: url.value,
    headers: headers.value.map((header) => ({
      name: header.name,
      value: header.value,
      sensitive: header.sensitive,
    })),
    body: body.value,
    bodyFormat: bodyFormat.value,
    auth: auth.value,
    username: username.value,
    password: password.value,
    token: token.value,
    insecure: insecure.value,
    followRedirects: followRedirects.value,
    timeout: timeoutValue.value,
    shell: shell.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'curl-generator',
  run,
  watch: [method, url, headers, body, bodyFormat, auth, username, password, token, insecure, followRedirects, timeoutValue, shell],
  enabled: () => url.value !== '' || body.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  method.value = 'GET'
  url.value = ''
  headers.value = []
  body.value = ''
  bodyFormat.value = 'json'
  auth.value = 'none'
  username.value = ''
  password.value = ''
  token.value = ''
  insecure.value = false
  followRedirects.value = false
  timeout.value = 0
  shell.value = 'bash'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  method.value = example.method
  url.value = example.url
  headers.value = example.headers.map((header) => ({ ...header }))
  body.value = example.body
  bodyFormat.value = example.bodyFormat
  auth.value = example.auth
  username.value = example.username
  password.value = example.password
  token.value = example.token
  insecure.value = example.insecure
  followRedirects.value = example.followRedirects
  timeout.value = example.timeout
  shell.value = example.shell
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="curl-generator"
    :name="t('tools.curl-generator.name')"
    name-en="cURL Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.method') }}</span>
              <select
                v-model="method"
                class="field__select"
              >
                <option
                  v-for="item in METHODS"
                  :key="item"
                  :value="item"
                >
                  {{ item }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.url') }}</span>
              <input
                v-model="url"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                placeholder="https://api.example.com/v1/items"
                :aria-label="t('tools.curl-generator.fields.url')"
              >
            </label>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.curl-generator.fields.body') }}</span>
            <CodeEditor
              v-model="body"
              :label="t('tools.curl-generator.fields.body')"
              :max-rows="10"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.curl-generator.hints.url') }}
          </p>
        </template>

        <template #options>
          <fieldset class="field">
            <legend class="field__label">
              {{ t('tools.curl-generator.fields.headers') }}
            </legend>

            <div
              v-for="(header, index) in headers"
              :key="index"
              class="header"
            >
              <input
                v-model="header.name"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.curl-generator.fields.headerName')"
              >
              <input
                v-model="header.value"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.curl-generator.fields.headerValue')"
              >
              <label class="field field--inline">
                <input
                  v-model="header.sensitive"
                  type="checkbox"
                  class="field__checkbox"
                >
                <span>{{ t('tools.curl-generator.fields.headerSensitive') }}</span>
              </label>
              <button
                type="button"
                class="header__remove"
                @click="removeHeader(index)"
              >
                {{ t('tools.curl-generator.actions.removeHeader') }}
              </button>
            </div>

            <button
              type="button"
              class="header__add"
              @click="addHeader"
            >
              {{ t('tools.curl-generator.actions.addHeader') }}
            </button>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.curl-generator.hints.headers') }}
          </p>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.bodyFormat') }}</span>
              <select
                v-model="bodyFormat"
                class="field__select"
              >
                <option
                  v-for="item in BODY_FORMATS"
                  :key="item"
                  :value="item"
                >
                  {{ t(`tools.curl-generator.options.bodyFormat.${item}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.auth') }}</span>
              <select
                v-model="auth"
                class="field__select"
              >
                <option
                  v-for="item in AUTH_TYPES"
                  :key="item"
                  :value="item"
                >
                  {{ t(`tools.curl-generator.options.auth.${item}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.shell') }}</span>
              <select
                v-model="shell"
                class="field__select"
              >
                <option
                  v-for="item in SHELLS"
                  :key="item"
                  :value="item"
                >
                  {{ t(`tools.curl-generator.options.shell.${item}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.timeout') }}</span>
              <input
                v-model.number="timeout"
                type="number"
                class="field__input"
                min="0"
                :max="MAX_TIMEOUT"
                :aria-label="t('tools.curl-generator.fields.timeout')"
              >
            </label>
          </div>

          <template v-if="auth === 'basic'">
            <div class="field-row">
              <label class="field">
                <span class="field__label">{{ t('tools.curl-generator.fields.username') }}</span>
                <input
                  v-model="username"
                  type="text"
                  class="field__input"
                  autocomplete="off"
                  :aria-label="t('tools.curl-generator.fields.username')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.curl-generator.fields.password') }}</span>
                <input
                  v-model="password"
                  type="password"
                  class="field__input"
                  autocomplete="off"
                  :aria-label="t('tools.curl-generator.fields.password')"
                >
              </label>
            </div>
          </template>

          <template v-else-if="auth === 'bearer'">
            <label class="field">
              <span class="field__label">{{ t('tools.curl-generator.fields.token') }}</span>
              <input
                v-model="token"
                type="password"
                class="field__input"
                autocomplete="off"
                :aria-label="t('tools.curl-generator.fields.token')"
              >
            </label>
          </template>

          <p class="field__hint">
            {{ t('tools.curl-generator.hints.auth') }}
          </p>

          <label class="field field--inline">
            <input
              v-model="followRedirects"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.curl-generator.fields.followRedirects') }}</span>
          </label>

          <label class="field field--inline">
            <input
              v-model="insecure"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.curl-generator.fields.insecure') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.curl-generator.hints.insecure') }}
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
        :title="t('tools.curl-generator.results.command')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.curl-generator.results.shell') }}</dt>
              <dd class="tb-mono">
                {{ data?.shell }}
              </dd>
            </div>
            <div
              v-if="placeholders.length > 0"
              class="detail__row"
            >
              <dt>{{ t('tools.curl-generator.results.placeholders') }}</dt>
              <dd class="tb-mono">
                {{ placeholderText }}
              </dd>
            </div>
          </dl>

          <p
            v-if="placeholders.length > 0"
            class="notice"
          >
            {{ t('tools.curl-generator.hints.placeholders') }}
          </p>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.curl-generator.hints.output') }}
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
.header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-2);
}

.header__add,
.header__remove {
  padding: var(--tb-space-1) var(--tb-space-3);
  font-family: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  cursor: pointer;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.header__add:hover,
.header__remove:hover {
  color: var(--tb-text);
  border-color: var(--tb-text-tertiary);
}

.header__remove {
  color: var(--tb-danger);
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

.notice {
  margin: var(--tb-space-2) 0 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-warning);
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
