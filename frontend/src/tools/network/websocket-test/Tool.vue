<script setup lang="ts">
/**
 * WebSocket 测试工具页。
 *
 * **纯前端**：用浏览器原生 `WebSocket` 直连目标服务，后端不参与握手，
 * 因此本页不调用任何 `/api` 端点。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, onScopeDispose, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'

import { DEFAULT_EXAMPLE } from './example'

/** 连接阶段。 */
type Phase = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const url = ref(DEFAULT_EXAMPLE.url)
const protocols = ref(DEFAULT_EXAMPLE.protocols)
const message = ref(DEFAULT_EXAMPLE.message)

const phase = ref<Phase>('idle')
const log = ref<string[]>([])
const openMs = ref<number | null>(null)
const negotiated = ref('')
const closeInfo = ref('')

let socket: WebSocket | null = null
let openedAt = 0

const faq = computed<FaqItem[]>(() => tm('tools.websocket-test.faq') as FaqItem[])

const running = computed(() => phase.value === 'connecting')

const state = computed<'idle' | 'loading' | 'success' | 'error'>(() => {
  switch (phase.value) {
    case 'connecting':
      return 'loading'
    case 'open':
      return 'success'
    case 'error':
      return 'error'
    default:
      return 'idle'
  }
})

const resultText = computed(() => {
  const lines = [
    `connected: ${String(phase.value === 'open' || phase.value === 'closed')}`,
    `openMs: ${openMs.value === null ? '-' : String(openMs.value)}`,
    `protocol: ${negotiated.value === '' ? '-' : negotiated.value}`,
    `closed: ${closeInfo.value === '' ? '-' : closeInfo.value}`,
    '',
    ...log.value,
  ]
  return lines.join('\n')
})

function append(entry: string): void {
  log.value = [...log.value, `${new Date().toISOString()} ${entry}`]
}

function cleanup(): void {
  if (socket) {
    socket.onopen = null
    socket.onmessage = null
    socket.onerror = null
    socket.onclose = null
    socket.close(1000, 'reset')
    socket = null
  }
}

onScopeDispose(cleanup)

function connect(): void {
  cleanup()
  log.value = []
  closeInfo.value = ''
  negotiated.value = ''
  openMs.value = null
  phase.value = 'connecting'

  const list = protocols.value
    .split(',')
    .map((item) => item.trim())
    .filter((item) => item !== '')

  let opened: WebSocket
  try {
    opened = list.length > 0 ? new WebSocket(url.value, list) : new WebSocket(url.value)
  } catch {
    phase.value = 'error'
    append('! invalid url')
    return
  }
  socket = opened
  openedAt = performance.now()

  opened.onopen = () => {
    openMs.value = Math.round(performance.now() - openedAt)
    negotiated.value = opened.protocol
    phase.value = 'open'
    append(`> open (${openMs.value}ms)`)
    if (message.value !== '') {
      opened.send(message.value)
      append(`> ${message.value}`)
    }
  }

  opened.onmessage = (event: MessageEvent<unknown>) => {
    append(`< ${typeof event.data === 'string' ? event.data : '[binary]'}`)
  }

  opened.onerror = () => {
    if (phase.value !== 'closed') {
      phase.value = 'error'
    }
    append('! error')
  }

  opened.onclose = (event: CloseEvent) => {
    phase.value = 'closed'
    closeInfo.value = `${event.code} ${event.reason}`.trim()
    append(`- close ${closeInfo.value}`)
    socket = null
  }
}

function disconnect(): void {
  if (socket) {
    socket.close(1000, 'client close')
  }
  phase.value = 'closed'
}

function reset(): void {
  cleanup()
  url.value = DEFAULT_EXAMPLE.url
  protocols.value = DEFAULT_EXAMPLE.protocols
  message.value = DEFAULT_EXAMPLE.message
  phase.value = 'idle'
  log.value = []
  openMs.value = null
  negotiated.value = ''
  closeInfo.value = ''
}
</script>

<template>
  <ToolPageLayout
    slug="websocket-test"
    name="WebSocket 测试"
    name-en="WebSocket Tester"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.websocket-test.fields.url') }}</span>
            <ElInput
              v-model="url"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="wss://echo.websocket.org"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.websocket-test.fields.message') }}</span>
            <ElInput
              v-model="message"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.websocket-test.fields.protocols') }}</span>
            <ElInput
              v-model="protocols"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="chat, superchat"
            />
          </label>

          <p class="prose">
            {{ t('tools.websocket-test.hints.headers') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="url.trim() === ''"
            @run="connect"
            @reset="reset"
            @example="reset"
          />
          <ElButton
            v-if="phase === 'open'"
            class="field__input"
            @click="disconnect"
          >
            {{ t('tools.websocket-test.actions.disconnect') }}
          </ElButton>
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="resultText"
        :meta="null"
        :error="null"
        :title="t('tools.websocket-test.results.log')"
        @retry="connect"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.websocket-test.description') }}
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
