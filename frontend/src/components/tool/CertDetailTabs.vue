<script setup lang="ts">
/**
 * 证书详情页签容器。
 *
 * 在**弹窗**里给一张证书提供分区页签：
 * 摘要 / 名称与用途 / 标识与指纹 / 公钥 / 扩展 / 证书PEM。
 *
 * `cert-parse` 的结果区用 `ResultPanel` 自带的页签，不经过本组件；
 * 证书链的详情弹窗用本组件 —— 两者的分区**完全一致**。
 */
import { ElTabPane, ElTabs } from 'element-plus/es/components/tabs/index'
import 'element-plus/es/components/tabs/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertDetail, { type CertDetailSection } from '@/components/tool/CertDetail.vue'
import type {
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
    /** 指纹列表。 */
    fingerprints?: CertificateFingerprint[]
    /** 公钥展开视图。 */
    publicKey?: PublicKeyDetail | null
    /** 证书 PEM。 */
    pem?: string
  }>(),
  { fingerprints: () => [], publicKey: null, pem: '' },
)

const { t } = useI18n()

/** 可选分区。 */
const sections = computed<{ id: CertDetailSection; label: string }[]>(() => [
  { id: 'summary', label: t('tool.cert.sections.summary') },
  { id: 'names', label: t('tool.cert.sections.namesAndUsages') },
  { id: 'identifiers', label: t('tool.cert.sections.identifiers') },
  { id: 'publicKey', label: t('tool.cert.sections.publicKey') },
  { id: 'extensions', label: t('tool.cert.sections.extensions') },
  { id: 'pem', label: t('tool.cert.sections.certificate') },
])

/** 当前分区。 */
const active = ref<CertDetailSection>('summary')
</script>

<template>
  <ElTabs
    v-model="active"
    class="cert-detail-tabs"
  >
    <ElTabPane
      v-for="item in sections"
      :key="item.id"
      :label="item.label"
      :name="item.id"
    >
      <CertDetail
        v-if="item.id === active"
        :info="props.info"
        :validity="props.validity"
        :section="item.id"
        :fingerprints="props.fingerprints"
        :public-key="props.publicKey"
        :pem="props.pem"
      />
    </ElTabPane>
  </ElTabs>
</template>

<style scoped>
.cert-detail-tabs {
  min-width: 0;
}
</style>
