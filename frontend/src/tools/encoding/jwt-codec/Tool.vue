<script setup lang="ts">
/**
 * JWT 编解码工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/jwt-codec/example'

/** 后端 jwt-codec 工具的响应。 */
interface JwtCodecResult {
  header?: Record<string, unknown>
  payload?: Record<string, unknown>
  signature?: string
  token?: string
  valid?: boolean
  status?: JwtStatus
}

/** 令牌的时间状态。 */
interface JwtStatus {
  expired: boolean
  notYetValid: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Operation = 'decode' | 'encode' | 'verify'
type Algorithm = 'HS256' | 'HS384' | 'HS512'

const { t, tm } = useI18n()

const operation = ref<Operation>('decode')
const token = ref('')
const headerText = ref('')
const payloadText = ref('')
const secret = ref('')
const algorithm = ref<Algorithm>('HS256')
const jsonError = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<JwtCodecResult>({
  slug: 'jwt-codec',
})

const running = computed(() => state.value === 'loading')
const faq = computed<FaqItem[]>(() => tm('tools.jwt-codec.faq') as FaqItem[])

/** 结果正文：生成时给出令牌，解析与校验时给出头部和载荷。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  if (result.token) {
    return result.token
  }
  const view: Record<string, unknown> = {
    header: result.header ?? {},
    payload: result.payload ?? {},
  }
  if (result.valid !== undefined) {
    view.valid = result.valid
  }
  return JSON.stringify(view, null, 2)
})

/** 解析 JSON 对象；内容不是对象时返回 undefined。 */
function parseJsonObject(text: string): Record<string, unknown> | undefined {
  try {
    const parsed: unknown = JSON.parse(text)
    if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return undefined
    }
    return parsed as Record<string, unknown>
  } catch {
    return undefined
  }
}

/** 运行当前输入。 */
function run(): void {
  jsonError.value = ''

  if (operation.value === 'encode') {
    const header = parseJsonObject(headerText.value)
    const payload = parseJsonObject(payloadText.value)
    if (!header || !payload) {
      jsonError.value = t('tools.jwt-codec.hints.invalidJson')
      return
    }
    void execute({
      operation: operation.value,
      header,
      payload,
      secret: secret.value,
      algorithm: algorithm.value,
    })
    return
  }

  void execute({
    operation: operation.value,
    token: token.value,
    secret: operation.value === 'verify' ? secret.value : '',
  })
}

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'decode'
  token.value = ''
  headerText.value = ''
  payloadText.value = ''
  secret.value = ''
  algorithm.value = 'HS256'
  jsonError.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  token.value = example.token
  headerText.value = example.header
  payloadText.value = example.payload
  secret.value = example.secret
  algorithm.value = example.algorithm
  jsonError.value = ''
  resetRun()
}

/** Ctrl / ⌘ + Enter 触发运行。 */
function onKeydown(event: KeyboardEvent): void {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    run()
  }
}
</script>

<template>
  <ToolPageLayout
    slug="jwt-codec"
    :name="t('tools.jwt-codec.name')"
    name-en="JWT Codec"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label
            v-if="operation !== 'encode'"
            class="field"
          >
            <span class="field__label">{{ t('tools.jwt-codec.fields.token') }}</span>
            <ElInput
              v-model="token"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="3"
              spellcheck="false"
              :aria-label="t('tools.jwt-codec.fields.token')"
              @keydown="(event) => onKeydown(event as KeyboardEvent)"
            />
          </label>

          <label
            v-if="operation === 'encode'"
            class="field"
          >
            <span class="field__label">{{ t('tools.jwt-codec.fields.header') }}</span>
            <CodeEditor
              v-model="headerText"
              :label="t('tools.jwt-codec.fields.header')"
              :max-rows="8"
              @run="run"
            />
          </label>

          <label
            v-if="operation === 'encode'"
            class="field"
          >
            <span class="field__label">{{ t('tools.jwt-codec.fields.payload') }}</span>
            <CodeEditor
              v-model="payloadText"
              :label="t('tools.jwt-codec.fields.payload')"
              :max-rows="10"
              @run="run"
            />
          </label>

          <p
            v-if="jsonError"
            class="input-error"
          >
            {{ jsonError }}
          </p>

          <p class="field__hint">
            {{ t('tools.jwt-codec.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.jwt-codec.fields.operation') }}</span>
              <ElSelect
                v-model="operation"
                class="field__select"
                :aria-label="t('tools.jwt-codec.fields.operation')"
              >
                <ElOption
                  :label="t('tools.jwt-codec.options.operation.decode')"
                  value="decode"
                />
                <ElOption
                  :label="t('tools.jwt-codec.options.operation.encode')"
                  value="encode"
                />
                <ElOption
                  :label="t('tools.jwt-codec.options.operation.verify')"
                  value="verify"
                />
              </ElSelect>
            </label>

            <label
              v-if="operation === 'encode'"
              class="field"
            >
              <span class="field__label">{{ t('tools.jwt-codec.fields.algorithm') }}</span>
              <ElSelect
                v-model="algorithm"
                class="field__select"
                :aria-label="t('tools.jwt-codec.fields.algorithm')"
              >
                <ElOption
                  label="HS256"
                  value="HS256"
                />
                <ElOption
                  label="HS384"
                  value="HS384"
                />
                <ElOption
                  label="HS512"
                  value="HS512"
                />
              </ElSelect>
            </label>
          </div>

          <label
            v-if="operation !== 'decode'"
            class="field"
          >
            <span class="field__label">{{ t('tools.jwt-codec.fields.secret') }}</span>
            <ElInput
              v-model="secret"
              type="password"
              class="field__input"
              autocomplete="off"
              spellcheck="false"
            />
          </label>
          <p
            v-if="operation !== 'decode'"
            class="field__hint"
          >
            {{ t('tools.jwt-codec.hints.secret') }}
          </p>

          <p class="field__hint">
            {{ t('tools.jwt-codec.hints.algorithm') }}
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
        :title="t('tools.jwt-codec.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div
              v-if="data?.valid !== undefined"
              class="detail__row"
            >
              <dt>{{ t('tools.jwt-codec.results.validity') }}</dt>
              <dd :class="data?.valid ? 'detail__ok' : 'detail__bad'">
                {{ data?.valid ? t('tools.jwt-codec.results.valid') : t('tools.jwt-codec.results.invalid') }}
              </dd>
            </div>
            <div
              v-if="data?.signature"
              class="detail__row"
            >
              <dt>{{ t('tools.jwt-codec.results.signature') }}</dt>
              <dd class="tb-mono">
                {{ data?.signature }}
              </dd>
            </div>
            <div
              v-if="data?.status?.expired"
              class="detail__row"
            >
              <dt>{{ t('tools.jwt-codec.results.expired') }}</dt>
              <dd>
                {{ t('tools.jwt-codec.results.expiredHint') }}
              </dd>
            </div>
            <div
              v-if="data?.status?.notYetValid"
              class="detail__row"
            >
              <dt>{{ t('tools.jwt-codec.results.notYetValid') }}</dt>
              <dd>
                {{ t('tools.jwt-codec.results.notYetValidHint') }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.jwt-codec.hints.output') }}
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
.input-error {
  margin: 0;
  font-size: var(--tb-font-xs);
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

.detail__ok {
  color: var(--tb-success);
}

.detail__bad {
  color: var(--tb-danger);
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
