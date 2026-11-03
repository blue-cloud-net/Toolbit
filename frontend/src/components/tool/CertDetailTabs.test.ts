/**
 * CertDetailTabs 组件测试。
 *
 * 证书链的详情弹窗靠它分区；分区必须与 `cert-parse` 结果区**一致**，
 * 否则两个入口看到的结构会不一样。
 */
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CertDetailTabs from '@/components/tool/CertDetailTabs.vue'
import { i18n } from '@/i18n'
import type { CertificateInfo, CertificateValidity } from '@/types/certificate'

/** 统一的挂载选项。 */
const mountOptions = { global: { plugins: [i18n] } }

/** 一份最小的证书信息。 */
const info: CertificateInfo = {
  version: 2,
  subject: [{ field: 'CN', value: 'leaf.example.com' }],
  subjectText: '/CN=leaf.example.com',
  issuer: [{ field: 'CN', value: 'leaf.example.com' }],
  issuerText: '/CN=leaf.example.com',
  serial: '01',
  notBefore: '2026-01-01T00:00:00Z',
  notAfter: '2027-01-01T00:00:00Z',
  publicKey: { algorithm: 'RSA', bits: 2048, curve: '' },
  signatureAlgorithm: 'RSA-SHA256',
  signatureAlgorithmOid: '1.2.840.113549.1.1.11',
  isCa: false,
  pathLen: -1,
  san: null,
  keyUsage: null,
  extendedKeyUsage: null,
  subjectKeyId: '',
  authorityKeyId: '',
  subjectHash: '01020304',
  issuerHash: '01020304',
  extensions: [],
  selfSigned: false,
  signatureHex: '00',
  policies: [],
}

/** 有效性判定。 */
const validity: CertificateValidity = {
  status: 'valid',
  daysRemaining: 100,
}

describe('CertDetailTabs', () => {
  it('提供 6 个分区页签', () => {
    const wrapper = mount(CertDetailTabs, {
      ...mountOptions,
      props: { info, validity },
    })

    const tabs = wrapper.findAll('.el-tabs__item')
    expect(tabs).toHaveLength(6)
  })

  it('默认停在摘要区', () => {
    const wrapper = mount(CertDetailTabs, {
      ...mountOptions,
      props: { info, validity },
    })

    expect(wrapper.text()).toContain('leaf.example.com')
  })
})
