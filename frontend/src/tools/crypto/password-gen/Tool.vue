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
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const CHARSETS = ['lower', 'upper', 'digit', 'symbol'] as const

const length = ref(16)
const count = ref(1)
const charsets = ref<string[]>(['lower', 'upper', 'digit', 'symbol'])
const excludeAmbiguous = ref(false)
const requireEachClass = ref(true)
const customCharset = ref('')

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
const entropyText = computed(() =>
  typeof data.value?.entropyBits === 'number'
    ? `${t('tools.password-gen.results.entropy')}: ${data.value.entropyBits.toFixed(1)} bit`
    : '',
)

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
  void execute({
    length: length.value,
    count: count.value,
    charsets: charsets.value,
    excludeAmbiguous: excludeAmbiguous.value,
    requireEachClass: requireEachClass.value,
    customCharset: customCharset.value,
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
            <span class="field__label">{{ t('tools.password-gen.fields.customCharset') }}</span>
            <ElInput
              v-model="customCharset"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.password-gen.fields.customCharset')"
            />
          </label>
          <p class="field__hint">
            {{ t('tools.password-gen.hints.customCharset') }}
          </p>
        </template>

        <template #options>
          <fieldset
            class="field"
            :disabled="customCharset !== ''"
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
