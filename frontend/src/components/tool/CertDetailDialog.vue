<script setup lang="ts">
/**
 * 证书详情对话框。
 *
 * 以浮层展示单张证书的完整字段。
 * 内容由共享的 `CertDetail` 渲染，弹窗本身只管开关与尺寸。
 */
import { ElDialog } from 'element-plus/es/components/dialog/index'
import 'element-plus/es/components/dialog/style/css'
import { useI18n } from 'vue-i18n'

import CertDetail from '@/components/tool/CertDetail.vue'
import type {
  CertificateFingerprint,
  CertificateInfo,
  CertificateValidity,
  PublicKeyDetail,
} from '@/types/certificate'

/** 弹窗内容。 */
export interface CertDetailPayload {
  /** 证书元信息。 */
  info: CertificateInfo
  /** 有效性判定。 */
  validity: CertificateValidity
  /** 指纹列表。 */
  fingerprints?: CertificateFingerprint[]
  /** 公钥展开视图。 */
  publicKey?: PublicKeyDetail | null
  /** 证书 PEM；给出时在底部提供复制入口。 */
  pem?: string
}

defineProps<{
  /** 详情；为 null 时不渲染。 */
  detail: CertDetailPayload | null
  /** 关闭回调。 */
  close: () => void
}>()

const { t } = useI18n()
</script>

<template>
  <ElDialog
    :model-value="detail !== null"
    :title="t('tool.cert.detail')"
    width="900px"
    append-to-body
    destroy-on-close
    @close="close()"
  >
    <CertDetail
      v-if="detail"
      :info="detail.info"
      :validity="detail.validity"
      :fingerprints="detail.fingerprints ?? []"
      :public-key="detail.publicKey ?? null"
      :pem="detail.pem ?? ''"
      show-extensions
    />
  </ElDialog>
</template>
