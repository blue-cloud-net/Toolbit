<script setup lang="ts">
/**
 * 证书扩展展示。
 *
 * 只保留一张「全部扩展」表：证书里每个扩展的 OID、是否关键与可读值。
 *
 * 此前还有一段「关键扩展」（BasicConstraints / KeyUsage / EKU / SAN …），
 * 但那些内容在「名称与用途」与「标识与指纹」页已经展开过一遍 ——
 * 同一份信息在同一页出现两次，读者只会怀疑哪一份才是对的。
 * 这里改为只做**完整清单**，扩展值用后端解出的可读文本。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import { allExtensions, extensionValue } from '@/tools/certificates/certDetail'
import type { CertificateExtension } from '@/types/certificate'

const props = withDefaults(
  defineProps<{
    /** 证书的全部扩展。 */
    extensions: CertificateExtension[]
  }>(),
  { extensions: () => [] },
)

const { t } = useI18n()

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
