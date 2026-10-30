<script setup lang="ts">
/**
 * 密码生成工具页。
 *
 * 生成结果只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
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
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/password-gen/example'

/** 后端 password-gen 工具的响应。 */
interface PasswordGenResult {
  passwords: string[]
  entropyBits: number
  template?: string
  poolSize?: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 策略模板；空串表示不套模板。 */
type Template = '' | 'simple' | 'strong' | 'pin' | 'passphrase'

const { t, tm } = useI18n()

const CHARSETS = ['lower', 'upper', 'digit', 'symbol'] as const
const TEMPLATES: Exclude<Template, ''>[] = ['simple', 'strong', 'pin', 'passphrase']
const SEPARATORS = ['-', '.', '_'] as const

const length = ref(16)
const count = ref(1)
const charsets = ref<string[]>(['lower', 'upper', 'digit', 'symbol'])
const excludeAmbiguous = ref(false)
const requireEachClass = ref(true)
const customCharset = ref('')
const template = ref<Template>('')
const wordCount = ref(3)
const separator = ref<string>('-')
const excludeChars = ref('')

/** 只有「不套模板且未自定义字符集」时才由界面决定字符集。 */
const charsetSelectable = computed(() => template.value === '' && customCharset.value === '')
const isPassphrase = computed(() => template.value === 'passphrase')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<PasswordGenResult>({ slug: 'password-gen' })

const running = computed(() => state.value === 'loading')
const resultText = computed(() => (data.value?.passwords ?? []).join('\n'))
const entropyText = computed(() => {
  const parts: string[] = []
  if (typeof data.value?.entropyBits === 'number' && data.value.entropyBits > 0) {
    parts.push(`${t('tools.password-gen.results.entropy')}: ${data.value.entropyBits.toFixed(1)} bit`)
  }
  if (data.value?.template) {
    parts.push(`${t('tools.password-gen.results.template')}: ${t(`tools.password-gen.options.templates.${data.value.template}`)}`)
  }
  if (typeof data.value?.poolSize === 'number' && data.value.poolSize > 0) {
    parts.push(`${t('tools.password-gen.results.poolSize')}: ${data.value.poolSize}`)
  }
  return parts.join(' · ')
})

const faq = computed<FaqItem[]>(() => tm('tools.password-gen.faq') as FaqItem[])

/** 切换某个字符集是否启用。 */
function toggleCharset(name: string): void {
  const index = charsets.value.indexOf(name)
  if (index >= 0) {
    charsets.value.splice(index, 1)
  } else {
    charsets.value.push(name)
  }
}

/** 运行当前输入。 */
function run(): void {
  const passphrase = isPassphrase.value
  void execute({
    length: length.value,
    count: count.value,
    // 套用模板时把字符集与「每类至少一个」交给后端决定（模板只填充未显式给出的值）。
    charsets: charsetSelectable.value ? charsets.value : [],
    excludeAmbiguous: excludeAmbiguous.value,
    requireEachClass: charsetSelectable.value ? requireEachClass.value : false,
    customCharset: template.value === '' ? customCharset.value : '',
    template: template.value,
    wordCount: passphrase ? wordCount.value : 0,
    separator: passphrase ? separator.value : '',
    excludeChars: excludeChars.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  length.value = 16
  count.value = 1
  charsets.value = ['lower', 'upper', 'digit', 'symbol']
  excludeAmbiguous.value = false
  requireEachClass.value = true
  customCharset.value = ''
  template.value = ''
  wordCount.value = 3
  separator.value = '-'
  excludeChars.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  length.value = DEFAULT_EXAMPLE.length
  count.value = DEFAULT_EXAMPLE.count
  charsets.value = [...DEFAULT_EXAMPLE.charsets]
  excludeAmbiguous.value = DEFAULT_EXAMPLE.excludeAmbiguous
  requireEachClass.value = DEFAULT_EXAMPLE.requireEachClass
  customCharset.value = ''
  template.value = ''
  wordCount.value = 3
  separator.value = '-'
  excludeChars.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="password-gen"
    :name="t('tools.password-gen.name')"
    name-en="Password Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.password-gen.fields.length') }}</span>
              <ElInputNumber
                v-model="length"
                class="field__input"
                :min="4"
                :max="128"
                :aria-label="t('tools.password-gen.fields.length')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.password-gen.fields.count') }}</span>
              <ElInputNumber
                v-model="count"
                class="field__input"
                :min="1"
                :max="100"
                :aria-label="t('tools.password-gen.fields.count')"
              />
            </label>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.password-gen.fields.template') }}</span>
            <ElSelect
              v-model="template"
              class="field__select"
              :disabled="customCharset !== ''"
              :aria-label="t('tools.password-gen.fields.template')"
            >
              <ElOption
                :label="t('tools.password-gen.options.templates.none')"
                value=""
              />
              <ElOption
                v-for="name in TEMPLATES"
                :key="name"
                :label="t(`tools.password-gen.options.templates.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>
          <p class="field__hint">
            {{ t('tools.password-gen.hints.template') }}
          </p>

          <label class="field">
            <span class="field__label">{{ t('tools.password-gen.fields.customCharset') }}</span>
            <ElInput
              v-model="customCharset"
              class="field__input tb-mono"
              autocomplete="off"
              :disabled="template !== ''"
              :aria-label="t('tools.password-gen.fields.customCharset')"
            />
          </label>
          <p class="field__hint">
            {{ t('tools.password-gen.hints.customCharset') }}
          </p>

          <label class="field">
            <span class="field__label">{{ t('tools.password-gen.fields.excludeChars') }}</span>
            <ElInput
              v-model="excludeChars"
              class="field__input tb-mono"
              autocomplete="off"
              :maxlength="64"
              :aria-label="t('tools.password-gen.fields.excludeChars')"
            />
          </label>
          <p class="field__hint">
            {{ t('tools.password-gen.hints.excludeChars') }}
          </p>

          <div
            v-if="isPassphrase"
            class="field-row"
          >
            <label class="field">
              <span class="field__label">{{ t('tools.password-gen.fields.wordCount') }}</span>
              <ElInputNumber
                v-model="wordCount"
                class="field__input"
                :min="1"
                :max="12"
                :aria-label="t('tools.password-gen.fields.wordCount')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.password-gen.fields.separator') }}</span>
              <ElSelect
                v-model="separator"
                class="field__select"
                :aria-label="t('tools.password-gen.fields.separator')"
              >
                <ElOption
                  v-for="name in SEPARATORS"
                  :key="name"
                  :label="name"
                  :value="name"
                />
              </ElSelect>
            </label>
          </div>
        </template>

        <template #options>
          <fieldset
            class="field"
            :disabled="!charsetSelectable"
          >
            <legend class="field__label">
              {{ t('tools.password-gen.fields.charsets') }}
            </legend>
            <div class="charset-list">
              <ElCheckbox
                v-for="name in CHARSETS"
                :key="name"
                class="field__checkbox"
                :model-value="charsets.includes(name)"
                @change="toggleCharset(name)"
              >
                {{ t(`tools.password-gen.options.charsets.${name}`) }}
              </ElCheckbox>
            </div>
          </fieldset>

          <ElCheckbox
            v-model="excludeAmbiguous"
            class="field__checkbox"
          >
            {{ t('tools.password-gen.fields.excludeAmbiguous') }}
          </ElCheckbox>

          <ElCheckbox
            v-model="requireEachClass"
            class="field__checkbox"
          >
            {{ t('tools.password-gen.fields.requireEachClass') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.password-gen.hints.requireEachClass') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample"
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
        :title="t('tools.password-gen.results.passwords')"
        @retry="run"
      />
      <p
        v-if="entropyText"
        class="prose"
      >
        {{ entropyText }}
      </p>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.password-gen.description') }}
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
.charset-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
}
</style>
