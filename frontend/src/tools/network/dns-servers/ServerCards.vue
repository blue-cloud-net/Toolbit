<script setup lang="ts">
/**
 * DNS 服务器卡片视图。
 *
 * 卡片模式下按**服务商**分组展示：一家服务商一张卡，卡内按地址族列出端点，
 * 同一个地址的多个协议（UDP / TCP）合并成一行 —— 地址只出现一次，
 * 协议以小标签并列，右侧是复制与「实测」。
 *
 * 与列表模式（`DataTable`）的取舍：
 *   - 卡片模式**一眼看全**服务商与端点，适合浏览、挑选、复制；
 *   - 列表模式逐行成表，适合按列比对、整行复制。
 *
 * 布局用 `auto-fill + minmax()` 网格：容器变宽时列数自动增加，
 * 卡片宽度**不会**被拉伸到超出可读范围。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

import type { DNSCardEndpoint, DNSCardProvider, DNSServerEntry } from './types'

const props = defineProps<{
  /** 命中筛选的端点列表（已按目录顺序排好）。 */
  entries: DNSServerEntry[]
  /** 协议标签。 */
  protoLabels: Record<string, string>
  /** 地址族标签。 */
  familyLabels: Record<string, string>
  /** 地区标签。 */
  regionLabels: Record<string, string>
}>()

const { t } = useI18n()

/**
 * 按服务商分组，并把同一地址的多协议合并成一行。
 *
 * 顺序沿用传入顺序（后端已按「服务商 → 地址族 → 协议 → 地址」排好），
 * 因此卡片顺序与地址顺序都与目录一致，不做二次排序。
 */
const providers = computed<DNSCardProvider[]>(() => {
  const groups: DNSCardProvider[] = []
  const byProvider = new Map<string, DNSCardProvider>()
  const byAddr = new Map<string, Map<string, DNSCardEndpoint>>()

  for (const entry of props.entries) {
    let group = byProvider.get(entry.provider)
    if (!group) {
      group = {
        provider: entry.provider,
        providerEn: entry.providerEn,
        region: entry.region,
        endpoints: [],
      }
      byProvider.set(entry.provider, group)
      byAddr.set(entry.provider, new Map())
      groups.push(group)
    }
    const addrMap = byAddr.get(entry.provider)!
    let endpoint = addrMap.get(entry.addr)
    if (!endpoint) {
      endpoint = {
        addr: entry.addr,
        family: entry.family,
        protos: [],
        note: entry.note,
        primaryProto: entry.proto,
      }
      addrMap.set(entry.addr, endpoint)
      group.endpoints.push(endpoint)
    }
    if (!endpoint.protos.includes(entry.proto)) {
      endpoint.protos.push(entry.proto)
    }
  }

  return groups
})

/** 协议的语义级别：明文中性，TLS 警告色，HTTPS 成功色。 */
function protoLevel(proto: string): 'success' | 'warning' | 'info' | 'muted' {
  switch (proto) {
    case 'doh':
      return 'success'
    case 'dot':
      return 'warning'
    case 'tcp':
      return 'info'
    default:
      return 'muted'
  }
}

/** 协议标签文案。 */
function protoLabel(proto: string): string {
  return props.protoLabels[proto] ?? proto
}

/** 地址族标签文案。 */
function familyLabel(family: string): string {
  return props.familyLabels[family] ?? family
}

/** 地区标签文案。 */
function regionLabel(region: string): string {
  return props.regionLabels[region] ?? region
}

/** 「实测」跳转：带上协议与地址，跳到 DNS 查询页。 */
function lookupLink(endpoint: DNSCardEndpoint) {
  return {
    name: 'tool',
    params: { slug: 'dns-lookup' },
    query: { proto: endpoint.primaryProto, addr: endpoint.addr },
  }
}
</script>

<template>
  <div class="server-cards">
    <article
      v-for="group in providers"
      :key="group.provider"
      class="server-card"
    >
      <header class="server-card__head">
        <h4 class="server-card__name">
          {{ group.provider }}
        </h4>
        <span class="server-card__name-en">{{ group.providerEn }}</span>
        <span class="server-card__region">{{ regionLabel(group.region) }}</span>
      </header>

      <ul class="server-card__list">
        <li
          v-for="endpoint in group.endpoints"
          :key="endpoint.addr"
          class="server-card__row"
        >
          <div class="server-card__main">
            <div class="server-card__line">
              <span class="server-card__family">{{ familyLabel(endpoint.family) }}</span>
              <span class="server-card__addr tb-mono">{{ endpoint.addr }}</span>
              <span class="server-card__protos">
                <span
                  v-for="proto in endpoint.protos"
                  :key="proto"
                  class="server-card__proto"
                  :class="`server-card__proto--${protoLevel(proto)}`"
                >{{ protoLabel(proto) }}</span>
              </span>
            </div>
            <p class="server-card__note">
              {{ endpoint.note }}
            </p>
          </div>

          <div class="server-card__actions">
            <CopyButton
              :text="endpoint.addr"
              size="small"
            />
            <RouterLink
              class="server-card__link"
              :to="lookupLink(endpoint)"
            >
              {{ t('tools.dns-servers.actions.lookup') }}
            </RouterLink>
          </div>
        </li>
      </ul>
    </article>
  </div>
</template>

<style scoped>
/* 卡片网格：列数随容器宽度自适应，卡片本身不被拉得过宽。 */
.server-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--tb-space-3);
}

.server-card {
  display: flex;
  flex-direction: column;
  padding: var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
}

.server-card__head {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  padding-bottom: var(--tb-space-2);
  border-bottom: 1px solid var(--tb-border);
}

.server-card__name {
  margin: 0;
  font-size: var(--tb-font-base);
  font-weight: 600;
}

.server-card__name-en {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.server-card__region {
  margin-left: auto;
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-secondary);
  white-space: nowrap;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.server-card__list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.server-card__row {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-2) 0;
  border-bottom: 1px dashed var(--tb-border);
}

.server-card__row:last-child {
  border-bottom: none;
}

.server-card__main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1 1 auto;
  min-width: 0;
}

.server-card__line {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  min-width: 0;
}

.server-card__family {
  flex: none;
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.server-card__addr {
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  overflow-wrap: anywhere;
}

.server-card__protos {
  display: inline-flex;
  flex: none;
  gap: var(--tb-space-1);
}

.server-card__proto {
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  font-weight: 600;
  line-height: 1.7;
  white-space: nowrap;
  border: 1px solid;
  border-radius: var(--tb-radius-pill);
}

.server-card__proto--success {
  color: var(--tb-success);
  background: color-mix(in srgb, var(--tb-success) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-success) 40%, var(--tb-border));
}

.server-card__proto--warning {
  color: var(--tb-warning);
  background: color-mix(in srgb, var(--tb-warning) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-warning) 40%, var(--tb-border));
}

.server-card__proto--info {
  color: var(--tb-info);
  background: color-mix(in srgb, var(--tb-info) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-info) 40%, var(--tb-border));
}

.server-card__proto--muted {
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border-color: var(--tb-border);
}

.server-card__note {
  margin: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
  overflow-wrap: anywhere;
}

.server-card__actions {
  display: flex;
  align-items: center;
  flex: none;
  gap: var(--tb-space-2);
}

.server-card__link {
  font-size: var(--tb-font-xs);
  color: var(--tb-primary);
  text-decoration: none;
  white-space: nowrap;
}

.server-card__link:hover {
  text-decoration: underline;
}
</style>
