<script setup lang="ts">
/**
 * 证书链表格。
 *
 * 展示链上每张证书的颁发者、主题、有效期与校验结论。
 */
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

export interface ChainCertRow {
  /** 主体。 */
  subject: string
  /** 颁发者。 */
  issuer?: string
  /** 生效时间。 */
  notBefore?: string
  /** 失效时间。 */
  notAfter?: string
  /** 该级是否校验通过。 */
  valid?: boolean
  /** 校验失败原因。 */
  reason?: string
}

defineOptions({ name: 'CertChainTable' })

withDefaults(
  defineProps<{
    /** 证书行。 */
    rows: ChainCertRow[]
  }>(),
  {},
)

const { t } = useI18n()
</script>

<template>
  <div class="cert-chain">
    <p
      v-if="rows.length === 0"
      class="cert-chain__empty"
    >
      {{ t('common.states.empty') }}
    </p>
    <ol
      v-else
      class="cert-chain__list"
    >
      <li
        v-for="(row, index) in rows"
        :key="`${row.subject}-${index}`"
        class="cert-chain__item"
      >
        <div class="cert-chain__head">
          <span
            v-if="row.valid !== undefined"
            class="cert-chain__verdict"
            :class="row.valid ? 'cert-chain__verdict--ok' : 'cert-chain__verdict--bad'"
          >{{ row.valid ? '✓' : '✗' }}</span>
          <span class="cert-chain__subject">{{ row.subject }}</span>
          <CopyButton
            :text="row.subject"
            size="small"
          />
        </div>
        <dl class="cert-chain__fields">
          <div
            v-if="row.issuer"
            class="cert-chain__field"
          >
            <dt>{{ t('tool.cert.issuer') }}</dt>
            <dd>{{ row.issuer }}</dd>
          </div>
          <div
            v-if="row.notBefore || row.notAfter"
            class="cert-chain__field"
          >
            <dt>{{ t('tool.cert.validity') }}</dt>
            <dd>{{ row.notBefore }} → {{ row.notAfter }}</dd>
          </div>
          <div
            v-if="row.reason"
            class="cert-chain__field"
          >
            <dt>{{ t('tool.cert.reason') }}</dt>
            <dd>{{ row.reason }}</dd>
          </div>
        </dl>
      </li>
    </ol>
  </div>
</template>

<style scoped>
.cert-chain {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.cert-chain__list {
  margin: 0;
  padding: 0;
  list-style: none;
  counter-reset: cert-chain;
}

.cert-chain__item {
  padding: var(--tb-space-2);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.cert-chain__head {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
}

.cert-chain__verdict {
  flex: none;
  font-weight: 700;
}

.cert-chain__verdict--ok {
  color: var(--tb-success);
}

.cert-chain__verdict--bad {
  color: var(--tb-danger);
}

.cert-chain__subject {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
  color: var(--tb-text);
}

.cert-chain__fields {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: var(--tb-space-1) 0 0;
}

.cert-chain__field {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-2xs);
}

.cert-chain__field dt {
  flex: none;
  min-width: 60px;
  color: var(--tb-text-tertiary);
}

.cert-chain__field dd {
  margin: 0;
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--tb-text-secondary);
}

.cert-chain__empty {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
