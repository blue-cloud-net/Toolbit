<script setup lang="ts">
/**
 * 证书详情。
 *
 * 把一张证书的**全部**结构化字段分区呈现：
 * 有效期 / 身份 / 名称 / 公钥 / 用途 / 标识与指纹 / 策略 / 扩展。
 *
 * 供 `cert-parse` 的摘要页签与证书链的详情弹窗**共用**，
 * 因此只依赖 `CertificateInfo` 这类共享类型，不认识任何具体工具。
 *
 * 文案走共享命名空间 `tool.cert.*`：不与单个 slug 耦合。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ValidityTrack from '@/components/tool/ValidityTrack.vue'
import {
  displayVersion,
  durationDays,
  extKeyUsageEntries,
  formatDateTime,
  hexUpper,
  keyUsageEntries,
  safetyLevel,
  safetyLevelColor,
  validityLevel,
  validityPercent,
} from '@/tools/certificates/certDetail'
import type {
  CertificateExtension,
  CertificateFingerprint,
  CertificateInfo,
  CertificateValidity,
  PublicKeyDetail,
} from '@/types/certificate'

const props = withDefaults(
  defineProps<{
    /** 证书元信息。 */
    info: CertificateInfo
    /** 有效性判定。 */
    validity: CertificateValidity
    /** 指纹列表；不给时不渲染该区。 */
    fingerprints?: CertificateFingerprint[]
    /** 公钥展开视图；不给时不渲染该区。 */
    publicKey?: PublicKeyDetail | null
    /** 证书 PEM；给出时提供复制入口。 */
    pem?: string
    /** 是否渲染扩展区。 */
    showExtensions?: boolean
  }>(),
  { fingerprints: () => [], publicKey: null, pem: '', showExtensions: false },
)

const { t, te, locale } = useI18n()

/** 公钥明细是否展开。 */
const publicKeyExpanded = ref(false)

/** 本地化标签；缺词条时回落到原始短名。 */
function label(namespace: string, name: string): string {
  const key = `tool.cert.${namespace}.${name}`
  return te(key) ? t(key) : name
}

/** 格式化时间。 */
function formatTime(iso: string): string {
  return formatDateTime(iso, locale.value)
}

/** 有效期比例。 */
const percent = computed(() =>
  validityPercent(props.info.notBefore, props.info.notAfter, props.validity.daysRemaining),
)

/** 有效期的语义级别。 */
const validityTone = computed(() => validityLevel(props.validity.status))

/** 剩余天数的紧急度。 */
const safety = computed(() => safetyLevel(props.validity.daysRemaining))

/** 概览卡片。 */
const overview = computed<StatGridItem[]>(() => {
  const info = props.info
  const items: StatGridItem[] = [
    { label: t('tool.cert.fields.version'), value: displayVersion(info.version) },
    {
      label: t('tool.cert.fields.status'),
      value: label('status', props.validity.status),
      level: validityTone.value,
    },
    {
      label: t('tool.cert.fields.daysRemaining'),
      value: `${props.validity.daysRemaining} ${t('tool.cert.units.days')}`,
      level: safetyLevelColor(safety.value),
    },
    {
      label: t('tool.cert.fields.duration'),
      value: `${durationDays(info.notBefore, info.notAfter)} ${t('tool.cert.units.days')}`,
    },
    {
      label: t('tool.cert.fields.isCa'),
      value: info.isCa
        ? t('tool.cert.status.ca')
        : t('tool.cert.status.notCa'),
      level: info.isCa ? ('warning' as const) : undefined,
    },
    {
      label: t('tool.cert.fields.selfSigned'),
      value: info.selfSigned
        ? t('common.states.yes')
        : t('common.states.no'),
      level: info.selfSigned ? ('warning' as const) : undefined,
    },
  ]
  if (info.isCa && info.pathLen >= 0) {
    items.push({
      label: t('tool.cert.fields.pathLen'),
      value: String(info.pathLen),
    })
  }
  return items
})

/** 身份信息表。 */
const identityRows = computed<KeyValueRow[]>(() => {
  const info = props.info
  return [
    { label: t('tool.cert.fields.serial'), value: info.serial },
    { label: t('tool.cert.fields.signatureAlgorithm'), value: info.signatureAlgorithm },
    { label: t('tool.cert.fields.signatureAlgorithmOid'), value: info.signatureAlgorithmOid },
    { label: t('tool.cert.fields.subjectHash'), value: info.subjectHash },
    { label: t('tool.cert.fields.issuerHash'), value: info.issuerHash },
  ]
})

/** 主体 RDN 表。 */
const subjectRows = computed<KeyValueRow[]>(() =>
  props.info.subject.map((entry) => ({ label: entry.field, value: entry.value })),
)

/** 签发者 RDN 表。 */
const issuerRows = computed<KeyValueRow[]>(() =>
  props.info.issuer.map((entry) => ({ label: entry.field, value: entry.value })),
)

/** 公钥摘要表。 */
const publicKeyRows = computed<KeyValueRow[]>(() => {
  const detail = props.publicKey
  if (!detail) {
    return []
  }
  const rows: KeyValueRow[] = [
    { label: t('tool.cert.fields.algorithm'), value: detail.algorithm },
  ]
  if (detail.bits > 0) {
    rows.push({ label: t('tool.cert.fields.bits'), value: String(detail.bits) })
  }
  if (detail.curve) {
    rows.push({ label: t('tool.cert.fields.curve'), value: detail.curve })
  }
  rows.push({
    label: t('tool.cert.fields.publicKeyFingerprint'),
    value: detail.fingerprint,
  })
  return rows
})

/** 公钥明细（展开后）。 */
const publicKeyDetailRows = computed<KeyValueRow[]>(() => {
  const detail = props.publicKey
  if (!detail) {
    return []
  }
  const rows: KeyValueRow[] = []
  if (detail.modulus) {
    rows.push({ label: t('tool.cert.fields.modulus'), value: detail.modulus })
  }
  if (detail.exponent) {
    rows.push({ label: t('tool.cert.fields.exponent'), value: detail.exponent })
  }
  if (detail.x) {
    rows.push({ label: t('tool.cert.fields.pointX'), value: detail.x })
  }
  if (detail.y) {
    rows.push({ label: t('tool.cert.fields.pointY'), value: detail.y })
  }
  if (detail.spkiHex) {
    rows.push({ label: t('tool.cert.fields.spki'), value: detail.spkiHex })
  }
  return rows
})

/** 主体备用名表格列。 */
const sanColumns = computed<DataTableColumn[]>(() => [
  { key: 'type', label: t('tool.cert.fields.sanType') },
  { key: 'value', label: t('tool.cert.fields.sanValue'), mono: true },
])

/** 主体备用名表格行。 */
const sanRows = computed<DataTableRow[]>(() =>
  props.info.san.map((entry, index) => {
    const at = entry.indexOf(':')
    return {
      key: `${entry}-${index}`,
      cells:
        at > 0
          ? { type: entry.slice(0, at), value: entry.slice(at + 1) }
          : { type: '', value: entry },
    }
  }),
)

/** 密钥用途表格列。 */
const kuColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tool.cert.fields.keyUsage') },
  { key: 'value', label: t('tool.cert.fields.bitMask'), mono: true },
])

/** 密钥用途表格行。 */
const kuRows = computed<DataTableRow[]>(() =>
  keyUsageEntries(props.info.keyUsage).map((entry) => ({
    key: entry.name,
    cells: { name: label('ku', entry.name), value: entry.value || '-' },
  })),
)

/** 扩展密钥用途表格列。 */
const ekuColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tool.cert.fields.extKeyUsage') },
  { key: 'value', label: t('tool.cert.fields.oid'), mono: true },
])

/** 扩展密钥用途表格行。 */
const ekuRows = computed<DataTableRow[]>(() =>
  extKeyUsageEntries(props.info.extendedKeyUsage).map((entry) => ({
    key: entry.name,
    cells: { name: label('eku', entry.name), value: entry.value || '-' },
  })),
)

/** 标识符表。 */
const identifierRows = computed<KeyValueRow[]>(() => {
  const rows: KeyValueRow[] = []
  if (props.info.subjectKeyId) {
    rows.push({
      label: t('tool.cert.fields.subjectKeyId'),
      value: props.info.subjectKeyId,
    })
  }
  if (props.info.authorityKeyId) {
    rows.push({
      label: t('tool.cert.fields.authorityKeyId'),
      value: props.info.authorityKeyId,
    })
  }
  return rows
})

/** 指纹表格列。 */
const fingerprintColumns = computed<DataTableColumn[]>(() => [
  { key: 'algorithm', label: t('tool.cert.fields.algorithm'), mono: true },
  { key: 'value', label: t('tool.cert.fields.fingerprint'), mono: true },
])

/** 指纹表格行。 */
const fingerprintRows = computed<DataTableRow[]>(() =>
  props.fingerprints.map((fp) => ({
    key: fp.algorithm,
    cells: { algorithm: fp.algorithm, value: fp.value },
  })),
)

/** 策略表格列。 */
const policyColumns = computed<DataTableColumn[]>(() => [
  { key: 'oid', label: t('tool.cert.fields.policyOid'), mono: true },
  { key: 'cps', label: t('tool.cert.fields.cps') },
  { key: 'notice', label: t('tool.cert.fields.notice') },
])

/** 策略表格行。 */
const policyRows = computed<DataTableRow[]>(() =>
  props.info.policies.map((policy) => {
    const noticeParts: string[] = []
    if (policy.organization) {
      noticeParts.push(policy.organization)
    }
    if (policy.noticeNumbers.length > 0) {
      noticeParts.push(policy.noticeNumbers.join(', '))
    }
    if (policy.explicitText) {
      noticeParts.push(policy.explicitText)
    }
    return {
      key: policy.oid,
      cells: {
        oid: policy.oid,
        cps: policy.cpsUris.join('\n') || '-',
        notice: noticeParts.join(' · ') || '-',
      },
    }
  }),
)

/** 扩展表格列。 */
const extensionColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tool.cert.fields.extensionName') },
  { key: 'oid', label: t('tool.cert.fields.oid'), mono: true },
  { key: 'critical', label: t('tool.cert.fields.critical'), align: 'center' },
  { key: 'value', label: t('tool.cert.fields.extensionValue'), mono: true },
])

/** 扩展表格行。 */
const extensionRows = computed<DataTableRow[]>(() =>
  props.info.extensions.map((ext: CertificateExtension) => ({
    key: `${ext.oid}-${ext.nid}`,
    cells: {
      name: ext.name || ext.oid,
      oid: ext.oid,
      critical: ext.critical ? t('common.states.yes') : '-',
      value: ext.text || hexUpper(ext.value),
    },
  })),
)
</script>

<template>
  <div class="cert-detail">
    <section class="cert-detail__section">
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.validity') }}
      </h3>
      <StatGrid :items="overview" />
      <ValidityTrack
        class="cert-detail__track"
        :percent="percent"
        :level="validityTone === 'success' ? 'success' : validityTone === 'warning' ? 'warning' : 'danger'"
        :start="formatTime(info.notBefore)"
        :end="formatTime(info.notAfter)"
      />
    </section>

    <section class="cert-detail__section">
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.identity') }}
      </h3>
      <KeyValueTable :rows="identityRows" />
    </section>

    <section class="cert-detail__section">
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.names') }}
      </h3>
      <div class="cert-detail__names">
        <div class="cert-detail__name">
          <h4 class="cert-detail__subtitle">
            {{ t('tool.cert.fields.subject') }}
          </h4>
          <p class="cert-detail__dn">
            {{ info.subjectText }}
          </p>
          <KeyValueTable :rows="subjectRows" />
        </div>
        <div class="cert-detail__name">
          <h4 class="cert-detail__subtitle">
            {{ t('tool.cert.fields.issuer') }}
          </h4>
          <p class="cert-detail__dn">
            {{ info.issuerText }}
          </p>
          <KeyValueTable :rows="issuerRows" />
        </div>
      </div>
    </section>

    <section
      v-if="publicKey"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.publicKey') }}
      </h3>
      <KeyValueTable :rows="publicKeyRows" />
      <button
        type="button"
        class="cert-detail__toggle"
        :aria-expanded="publicKeyExpanded"
        @click="publicKeyExpanded = !publicKeyExpanded"
      >
        {{ publicKeyExpanded ? t('tool.cert.actions.collapse') : t('tool.cert.actions.expand') }}
      </button>
      <template v-if="publicKeyExpanded">
        <KeyValueTable
          class="cert-detail__extra"
          :rows="publicKeyDetailRows"
        />
        <div class="cert-detail__pem">
          <div class="cert-detail__pem-head">
            <span class="cert-detail__pem-label">PEM</span>
            <CopyButton :text="publicKey.pem" />
          </div>
          <pre class="cert-detail__pem-body"><code>{{ publicKey.pem }}</code></pre>
        </div>
      </template>
    </section>

    <section
      v-if="info.san.length > 0 || info.keyUsage.length > 0 || info.extendedKeyUsage.length > 0"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.usages') }}
      </h3>
      <DataTable
        v-if="sanRows.length > 0"
        class="cert-detail__table"
        :columns="sanColumns"
        :rows="sanRows"
        :copy-column="'value'"
      />
      <DataTable
        v-if="kuRows.length > 0"
        class="cert-detail__table"
        :columns="kuColumns"
        :rows="kuRows"
        :copy-column="'name'"
      />
      <DataTable
        v-if="ekuRows.length > 0"
        class="cert-detail__table"
        :columns="ekuColumns"
        :rows="ekuRows"
        :copy-column="'name'"
      />
    </section>

    <section
      v-if="identifierRows.length > 0 || fingerprintRows.length > 0"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.identifiers') }}
      </h3>
      <KeyValueTable :rows="identifierRows" />
      <DataTable
        v-if="fingerprintRows.length > 0"
        class="cert-detail__table"
        :columns="fingerprintColumns"
        :rows="fingerprintRows"
        :copy-column="'value'"
      />
    </section>

    <section
      v-if="policyRows.length > 0"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.policies') }}
      </h3>
      <DataTable
        :columns="policyColumns"
        :rows="policyRows"
        :copy-column="'oid'"
      />
    </section>

    <section
      v-if="showExtensions && extensionRows.length > 0"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.extensions') }}
      </h3>
      <DataTable
        :columns="extensionColumns"
        :rows="extensionRows"
        :copy-column="'value'"
      />
    </section>

    <section
      v-if="pem"
      class="cert-detail__section"
    >
      <h3 class="cert-detail__title">
        {{ t('tool.cert.sections.certificate') }}
      </h3>
      <div class="cert-detail__pem">
        <div class="cert-detail__pem-head">
          <span class="cert-detail__pem-label">PEM</span>
          <CopyButton :text="pem" />
        </div>
        <pre class="cert-detail__pem-body"><code>{{ pem }}</code></pre>
      </div>
    </section>
  </div>
</template>

<style scoped>
.cert-detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-5);
}

.cert-detail__section {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.cert-detail__title {
  margin: 0;
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.cert-detail__subtitle {
  margin: 0;
  font-size: var(--tb-font-xs);
  font-weight: 600;
  color: var(--tb-text-tertiary);
}

.cert-detail__names {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: var(--tb-space-3);
}

.cert-detail__name {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  min-width: 0;
}

.cert-detail__dn {
  margin: 0;
  overflow-wrap: anywhere;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-secondary);
}

.cert-detail__track {
  margin-top: var(--tb-space-1);
}

.cert-detail__toggle {
  align-self: flex-start;
  padding: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-primary);
  cursor: pointer;
  background: none;
  border: none;
}

.cert-detail__extra {
  margin-top: var(--tb-space-1);
}

.cert-detail__table {
  margin-top: var(--tb-space-2);
}

.cert-detail__pem {
  margin-top: var(--tb-space-2);
}

.cert-detail__pem-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-1);
}

.cert-detail__pem-label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.cert-detail__pem-body {
  margin: 0;
  padding: var(--tb-space-3);
  max-height: 260px;
  overflow: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
