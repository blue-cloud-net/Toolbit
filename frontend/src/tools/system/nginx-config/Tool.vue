<script setup lang="ts">
/**
 * Nginx 配置生成工具页。
 *
 * 两种模式：
 *   - **简单模式**：按场景预设填写常用字段，可配置多个站点；
 *   - **高级模式**：直接编辑配置树，指令来自官方目录（按需拉取）。
 *
 * 输出默认是多文件配置包，也可切换为单个完整 `nginx.conf`。
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElCollapse, ElCollapseItem } from 'element-plus/es/components/collapse/index'
import 'element-plus/es/components/collapse/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { ElRadioButton, ElRadioGroup } from 'element-plus/es/components/radio/index'
import 'element-plus/es/components/radio/style/css'
import { computed, ref, shallowRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { callTool, fetchNginxDirectives } from '@/api/client'
import { base64ToBytes } from '@/components/tool/base64'
import CopyButton from '@/components/tool/CopyButton.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import WarningList from '@/components/tool/WarningList.vue'
import { buildSummary } from '@/components/tool/summary'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import type {
  NginxConfigResult,
  NginxDiagnostic,
  NginxDirectivesData,
  NginxNode,
  NginxSite,
} from '@/types/nginx'

import NodeEditor from './NodeEditor.vue'
import { EXAMPLES } from './example'
import { buildNginxPayload, createNginxSite, parseUpstreams } from './payload'

/** 场景预设。 */
const PRESETS = [
  'static',
  'spa',
  'tls',
  'proxy',
  'loadbalance',
  'compress',
  'security',
  'errorpage',
  'php',
] as const

/** TLS 档位。 */
const TLS_PROFILES = ['modern', 'intermediate'] as const

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref<'simple' | 'advanced'>('simple')
const output = ref<'bundle' | 'single'>('bundle')
const targetVersion = ref('')
const preset = ref<string>('static')

const sites = ref<NginxSite[]>([createNginxSite()])
const nodes = ref<NginxNode[]>([])

const catalog = shallowRef<NginxDirectivesData | null>(null)
const catalogState = ref<'idle' | 'loading' | 'error'>('idle')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<NginxConfigResult>({
  slug: 'nginx-config',
})

const running = computed(() => state.value === 'loading')
const archiveLoading = ref(false)

/** 结果文本：主配置内容。 */
const resultText = computed(() => {
  const files = data.value?.files ?? []
  const main = files.find((file) => file.path === 'nginx.conf') ?? files[0]
  return main?.content ?? ''
})

/** 诊断按级别分组，便于一眼看出问题严重程度。 */
const diagnostics = computed(() => {
  const items = data.value?.diagnostics ?? []
  const order: NginxDiagnostic['level'][] = ['error', 'warning', 'info']
  return order
    .map((level) => ({
      level,
      items: items.filter((item) => item.level === level),
    }))
    .filter((group) => group.items.length > 0)
})

/** 诊断级别的本地化标签。 */
function levelLabel(level: NginxDiagnostic['level']): string {
  return t(`tools.nginx-config.advanced.levels.${level}`)
}

/** 摘要卡片：关键指标。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'mode', label: t('tools.nginx-config.results.mode') },
      { key: 'output', label: t('tools.nginx-config.results.output') },
      { key: 'targetVersion', label: t('tools.nginx-config.results.targetVersion') },
      {
        key: 'summary',
        label: t('tools.nginx-config.results.fileCount'),
        format: (value) => {
          const summary = value as { fileCount?: number } | undefined
          return String(summary?.fileCount ?? 0)
        },
      },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

/** 结果页签：摘要 + 每个文件。 */
const views = computed<ResultView[]>(() => {
  const list: ResultView[] = [{ id: 'summary', label: t('tools.nginx-config.views.summary') }]
  for (const file of data.value?.files ?? []) {
    list.push({ id: `file:${file.path}`, label: file.path })
  }
  return list
})

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

const faq = computed<FaqItem[]>(() => tm('tools.nginx-config.faq') as FaqItem[])

/** 组装请求体。 */
function buildPayload(responseShape: 'files' | 'bundle' = 'files'): Record<string, unknown> {
  return buildNginxPayload({
    mode: mode.value,
    output: output.value,
    targetVersion: targetVersion.value,
    preset: preset.value,
    sites: sites.value,
    nodes: nodes.value,
    responseShape,
  })
}

/** 运行当前输入。 */
function run(): void {
  void execute(buildPayload())
}

/** 下载打包产物。 */
async function downloadArchive(): Promise<void> {
  archiveLoading.value = true
  try {
    const result = await callTool<NginxConfigResult>('nginx-config', buildPayload('bundle'))
    const archive = result.data.archive
    if (!archive) {
      return
    }
    triggerDownload(archive.filename, archive.base64, archive.contentType)
  } catch {
    // 打包失败沿用主结果区的错误展示，这里不重复提示。
  } finally {
    archiveLoading.value = false
  }
}

/** 触发浏览器下载。 */
function triggerDownload(filename: string, base64: string, mime: string): void {
  if (typeof document === 'undefined') {
    return
  }
  const blob = new Blob([base64ToBytes(base64)], { type: mime })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

// 输入变化后自动运行：仅简单模式自动触发，
// 高级模式每次按键都会重排配置树，交给用户显式点击「生成」。
useAutoRun({
  slug: 'nginx-config',
  run,
  watch: [mode, output, targetVersion, preset, sites],
  enabled: () => mode.value === 'simple',
})

// 切换到高级模式时按需拉取指令目录。
//
// 目录同时提供**可选版本列表**，因此简单模式也需要它：
// 挂载时先拉一次，之后只在首次进入高级模式时兜底重试。
watch(
  mode,
  (value) => {
    if (value === 'advanced') {
      void loadCatalog()
    }
  },
  { immediate: true },
)

// 版本下拉需要目录里的版本列表，挂载即加载。
void loadCatalog()

/** 拉取指令目录。 */
async function loadCatalog(): Promise<void> {
  if (catalog.value !== null || catalogState.value === 'loading') {
    return
  }
  catalogState.value = 'loading'
  try {
    const result = await fetchNginxDirectives(targetVersion.value)
    catalog.value = result.data
    catalogState.value = 'idle'
    if (targetVersion.value === '' && result.data.version !== '') {
      targetVersion.value = result.data.version
    }
  } catch {
    catalogState.value = 'error'
  }
}

/** 添加站点。 */
function addSite(): void {
  sites.value = [...sites.value, createNginxSite()]
}

/** 删除站点。 */
function removeSite(index: number): void {
  sites.value = sites.value.filter((_, i) => i !== index)
}

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'simple'
  output.value = 'bundle'
  targetVersion.value = ''
  preset.value = 'static'
  sites.value = [createNginxSite()]
  nodes.value = []
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  mode.value = 'simple'
  preset.value = example.preset
  const site = createNginxSite()
  site.serverName = example.serverName
  site.root = example.root
  sites.value = [site]
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
            <ElRadioGroup
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.nginx-config.fields.mode')"
            >
              <ElRadioButton value="simple">
                {{ t('tools.nginx-config.options.mode.simple') }}
              </ElRadioButton>
              <ElRadioButton value="advanced">
                {{ t('tools.nginx-config.options.mode.advanced') }}
              </ElRadioButton>
            </ElRadioGroup>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.nginx-config.fields.output') }}</span>
            <ElRadioGroup
              v-model="output"
              class="field__select"
              :aria-label="t('tools.nginx-config.fields.output')"
            >
              <ElRadioButton value="bundle">
                {{ t('tools.nginx-config.options.output.bundle') }}
              </ElRadioButton>
              <ElRadioButton value="single">
                {{ t('tools.nginx-config.options.output.single') }}
              </ElRadioButton>
            </ElRadioGroup>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.nginx-config.fields.targetVersion') }}</span>
            <ElSelect
              v-model="targetVersion"
              class="field__select"
              :aria-label="t('tools.nginx-config.fields.targetVersion')"
            >
              <ElOption
                :label="t('tools.nginx-config.options.version.default')"
                value=""
              />
              <ElOption
                v-for="version in catalog?.versions ?? []"
                :key="version"
                :label="version"
                :value="version"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <template v-if="mode === 'simple'">
            <label class="field">
              <span class="field__label">{{ t('tools.nginx-config.fields.preset') }}</span>
              <ElSelect
                v-model="preset"
                class="field__select"
                :aria-label="t('tools.nginx-config.fields.preset')"
              >
                <ElOption
                  v-for="item in PRESETS"
                  :key="item"
                  :label="t(`tools.nginx-config.options.preset.${item}`)"
                  :value="item"
                />
              </ElSelect>
            </label>

            <ElCollapse class="tb-sites">
              <ElCollapseItem
                v-for="(site, index) in sites"
                :key="site.id"
                :name="site.id"
              >
                <template #title>
                  <span class="tb-sites__title">
                    {{ site.serverName || t('tools.nginx-config.sites.unnamed') }}
                  </span>
                  <ElButton
                    class="tb-sites__remove"
                    size="small"
                    text
                    type="danger"
                    :aria-label="t('tools.nginx-config.sites.remove')"
                    @click.stop="removeSite(index)"
                  >
                    {{ t('tools.nginx-config.sites.remove') }}
                  </ElButton>
                </template>

                <label class="field field--grow">
                  <span class="field__label">{{ t('tools.nginx-config.fields.serverName') }}</span>
                  <ElInput
                    v-model="site.serverName"
                    class="field__input tb-mono"
                    autocomplete="off"
                    :aria-label="t('tools.nginx-config.fields.serverName')"
                  />
                </label>

                <label class="field field--grow">
                  <span class="field__label">{{ t('tools.nginx-config.fields.sitePreset') }}</span>
                  <ElSelect
                    v-model="site.preset"
                    class="field__select"
                    :aria-label="t('tools.nginx-config.fields.sitePreset')"
                  >
                    <ElOption
                      :label="t('tools.nginx-config.options.sitePreset.inherit')"
                      value=""
                    />
                    <ElOption
                      v-for="item in PRESETS"
                      :key="item"
                      :label="t(`tools.nginx-config.options.preset.${item}`)"
                      :value="item"
                    />
                  </ElSelect>
                </label>

                <div class="tb-sites__row">
                  <label class="field">
                    <span class="field__label">{{ t('tools.nginx-config.fields.listenPort') }}</span>
                    <ElInputNumber
                      v-model="site.listenPort"
                      class="field__select"
                      :min="0"
                      :max="65535"
                      :aria-label="t('tools.nginx-config.fields.listenPort')"
                    />
                  </label>

                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.root') }}</span>
                    <ElInput
                      v-model="site.root"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :aria-label="t('tools.nginx-config.fields.root')"
                    />
                  </label>
                </div>

                <div class="tb-sites__row">
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.listenIPv6"
                      :aria-label="t('tools.nginx-config.fields.listenIPv6')"
                    >
                      {{ t('tools.nginx-config.fields.listenIPv6') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableSsl"
                      :aria-label="t('tools.nginx-config.fields.enableSsl')"
                    >
                      {{ t('tools.nginx-config.fields.enableSsl') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableProxy"
                      :aria-label="t('tools.nginx-config.fields.enableProxy')"
                    >
                      {{ t('tools.nginx-config.fields.enableProxy') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableLoadBalance"
                      :aria-label="t('tools.nginx-config.fields.enableLoadBalance')"
                    >
                      {{ t('tools.nginx-config.fields.enableLoadBalance') }}
                    </ElCheckbox>
                  </label>
                </div>

                <template v-if="site.enableSsl">
                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.sslCertificate') }}</span>
                    <ElInput
                      v-model="site.sslCertificate"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :aria-label="t('tools.nginx-config.fields.sslCertificate')"
                    />
                  </label>
                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.sslPrivateKey') }}</span>
                    <ElInput
                      v-model="site.sslPrivateKey"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :aria-label="t('tools.nginx-config.fields.sslPrivateKey')"
                    />
                  </label>
                  <label class="field">
                    <span class="field__label">{{ t('tools.nginx-config.fields.tlsProfile') }}</span>
                    <ElSelect
                      v-model="site.tlsProfile"
                      class="field__select"
                      :aria-label="t('tools.nginx-config.fields.tlsProfile')"
                    >
                      <ElOption
                        v-for="item in TLS_PROFILES"
                        :key="item"
                        :label="t(`tools.nginx-config.options.tlsProfile.${item}`)"
                        :value="item"
                      />
                    </ElSelect>
                  </label>
                </template>

                <template v-if="site.enableProxy">
                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.proxyPass') }}</span>
                    <ElInput
                      v-model="site.proxyPass"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :aria-label="t('tools.nginx-config.fields.proxyPass')"
                    />
                  </label>
                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.proxyPath') }}</span>
                    <ElInput
                      v-model="site.proxyPath"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :aria-label="t('tools.nginx-config.fields.proxyPath')"
                    />
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableWebSocket"
                      :aria-label="t('tools.nginx-config.fields.enableWebSocket')"
                    >
                      {{ t('tools.nginx-config.fields.enableWebSocket') }}
                    </ElCheckbox>
                  </label>
                </template>

                <template v-if="site.enableLoadBalance">
                  <label class="field field--grow">
                    <span class="field__label">{{ t('tools.nginx-config.fields.upstreams') }}</span>
                    <ElInput
                      :model-value="site.upstreams.join(', ')"
                      class="field__input tb-mono"
                      autocomplete="off"
                      :placeholder="t('tools.nginx-config.options.upstreamsPlaceholder')"
                      :aria-label="t('tools.nginx-config.fields.upstreams')"
                      @update:model-value="(value: string) => {
                        site.upstreams = parseUpstreams(value)
                      }"
                    />
                  </label>
                </template>

                <div class="tb-sites__row">
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableGzip"
                      :aria-label="t('tools.nginx-config.fields.enableGzip')"
                    >
                      {{ t('tools.nginx-config.fields.enableGzip') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableCache"
                      :aria-label="t('tools.nginx-config.fields.enableCache')"
                    >
                      {{ t('tools.nginx-config.fields.enableCache') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableSecurityHeaders"
                      :aria-label="t('tools.nginx-config.fields.enableSecurityHeaders')"
                    >
                      {{ t('tools.nginx-config.fields.enableSecurityHeaders') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enableErrorPages"
                      :aria-label="t('tools.nginx-config.fields.enableErrorPages')"
                    >
                      {{ t('tools.nginx-config.fields.enableErrorPages') }}
                    </ElCheckbox>
                  </label>
                  <label class="field field--check">
                    <ElCheckbox
                      v-model="site.enablePhp"
                      :aria-label="t('tools.nginx-config.fields.enablePhp')"
                    >
                      {{ t('tools.nginx-config.fields.enablePhp') }}
                    </ElCheckbox>
                  </label>
                </div>

                <label class="field field--grow">
                  <span class="field__label">{{ t('tools.nginx-config.fields.fileName') }}</span>
                  <ElInput
                    v-model="site.fileName"
                    class="field__input tb-mono"
                    autocomplete="off"
                    :placeholder="t('tools.nginx-config.options.fileNamePlaceholder')"
                    :aria-label="t('tools.nginx-config.fields.fileName')"
                  />
                </label>
              </ElCollapseItem>
            </ElCollapse>

            <ElButton
              class="tb-sites__add"
              @click="addSite"
            >
              {{ t('tools.nginx-config.sites.add') }}
            </ElButton>
          </template>

          <template v-else>
            <p
              v-if="catalogState === 'loading'"
              class="tb-advanced__status"
            >
              {{ t('tools.nginx-config.advanced.loading') }}
            </p>
            <p
              v-else-if="catalogState === 'error'"
              class="tb-advanced__status"
            >
              {{ t('tools.nginx-config.advanced.loadFailed') }}
            </p>
            <p
              v-else-if="catalog"
              class="tb-advanced__status"
            >
              {{ t('tools.nginx-config.advanced.catalogSummary', {
                count: catalog.directives.length,
                version: catalog.docsVersion,
              }) }}
            </p>

            <NodeEditor
              v-model:nodes="nodes"
              context="main"
              :catalog="catalog?.directives ?? []"
            />
          </template>

          <p class="field__hint">
            {{ t('tools.nginx-config.hints.config') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :examples="EXAMPLES"
            @run="run"
            @reset="reset"
            @fill="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :views="views"
        :title="t('tools.nginx-config.results.config')"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />

            <section
              v-for="group in diagnostics"
              :key="group.level"
              class="tb-diagnostics"
            >
              <h3 class="tb-diagnostics__title">
                {{ levelLabel(group.level) }}
              </h3>
              <WarningList
                :items="group.items.map((item) => item.message)"
                :label="levelLabel(group.level)"
              />
            </section>

            <div class="tb-result__actions">
              <ElButton
                :loading="archiveLoading"
                @click="downloadArchive"
              >
                {{ t('tools.nginx-config.actions.downloadArchive') }}
              </ElButton>
            </div>
          </template>

          <template v-else-if="viewId.startsWith('file:') && data">
            <div
              v-for="file in data.files.filter((item) => `file:${item.path}` === viewId)"
              :key="file.path"
              class="tb-result__file"
            >
              <div class="tb-result__file-head">
                <code class="tb-mono">{{ file.path }}</code>
                <div class="tb-result__file-actions">
                  <CopyButton
                    :text="file.content"
                    :label="t('tools.nginx-config.actions.copyFile')"
                  />
                  <DownloadButton
                    :filename="file.path.split('/').pop() ?? file.path"
                    :content="file.content"
                    :mime="'text/plain;charset=utf-8'"
                    :label="t('tools.nginx-config.actions.downloadFile')"
                  />
                </div>
              </div>
              <ProductText
                :label="file.path"
                :content="file.content"
              />
            </div>
          </template>
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

<style scoped>
.tb-sites {
  display: grid;
  gap: var(--tb-space-2);
}

.tb-sites__title {
  font-weight: 600;
  margin-right: var(--tb-space-2);
}

.tb-sites__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-3);
  align-items: center;
}

.tb-sites__add {
  justify-self: start;
}

.tb-advanced__status {
  margin: 0;
  color: var(--tb-text-tertiary);
  font-size: var(--tb-font-sm);
}

.tb-result__actions {
  margin-top: var(--tb-space-3);
}

.tb-diagnostics {
  display: grid;
  gap: var(--tb-space-2);
  margin-top: var(--tb-space-3);
}

.tb-diagnostics__title {
  margin: 0;
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.tb-result__file {
  display: grid;
  gap: var(--tb-space-2);
}

.tb-result__file-head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  align-items: center;
  justify-content: space-between;
}

.tb-result__file-actions {
  display: flex;
  gap: var(--tb-space-2);
}
</style>
