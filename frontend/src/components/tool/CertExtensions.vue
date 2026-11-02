<script setup lang="ts">
/**
 * 证书扩展展示。
 *
 * 分两段（像浏览器内置的证书查看器那样）：
 *
 * 1. **关键扩展** —— 证书的身份与能力：能否签发（BasicConstraints）、
 *    能干什么（KeyUsage / EKU）、访问哪些名字（SAN / IAN）、谁签的
 *    （SKID / AKID）、吊销与策略信息从哪拿（CRLDP / AIA / Policies）。
 *    只列出证书**实际带有**的项，值用可读文本。
 * 2. **全部扩展** —— 证书里所有扩展的完整清单（含未知扩展的十六进制原值）。
 *
 * 两段共用同一份 `CertificateExtension[]`；未知扩展只会出现在第二段。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import { allExtensions, extensionValue, keyExtensions } from '@/tools/certificates/certDetail'
import type { CertificateExtension } from '@/types/certificate'

const props = withDefaults(
  defineProps<{
    /** 证书的全部扩展。 */
    extensions: CertificateExtension[]
  }>(),
  { extensions: () => [] },
)

const { t, te } = useI18n()

/** 本地化标签；缺词条时回落到扩展自身的短名。 */
function label(name: string, fallback: string): string {
  const key = `tool.cert.ext.${name}`
  return te(key) ? t(key) : fallback
}

/** 关键扩展：只保留实际带有的，按固定顺序。 */
const keyRows = computed<KeyValueRow[]>(() =>
  keyExtensions(props.extensions).map((ext) => ({
    label: label(ext.keyName, ext.name || ext.oid),
    value: extensionValue(ext),
    level: ext.critical ? ('warning' as const) : undefined,
  })),
)

/** 全部扩展表格的列。 */
const extensionColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tool.cert.fields.extensionName') },
  { key: 'oid', label: t('tool.cert.fields.oid'), mono: true },
  { key: 'critical', label: t('tool.cert.fields.critical'), align: 'center' },
  { key: 'value', label: t('tool.cert.fields.extensionValue'), mono: true },
])

/** 全部扩展表格的行。 */
const extensionRows = computed<DataTableRow[]>(() =>
  allExtensions(props.extensions).map((ext) => ({
    key: `${ext.oid}-${ext.nid}`,
    cells: {
      name: ext.name || ext.oid,
      oid: ext.oid,
      critical: ext.critical ? t('common.states.yes') : '-',
      value: extensionValue(ext),
    },
  })),
)
</script>

<template>
  <div class="cert-extensions">
    <section
      v-if="keyRows.length > 0"
      class="cert-extensions__section"
    >
      <h3 class="cert-extensions__title">
        {{ t('tool.cert.sections.keyExtensions') }}
      </h3>
      <KeyValueTable :rows="keyRows" />
    </section>

    <section
      v-if="extensionRows.length > 0"
      class="cert-extensions__section"
    >
      <h3 class="cert-extensions__title">
        {{ t('tool.cert.sections.allExtensions') }}
      </h3>
      <DataTable
        :columns="extensionColumns"
        :rows="extensionRows"
        :copy-column="'value'"
      />
    </section>
  </div>
</template>

<style scoped>
.cert-extensions {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-5);
}

.cert-extensions__section {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.cert-extensions__title {
  margin: 0;
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}
</style>
