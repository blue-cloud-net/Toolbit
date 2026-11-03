/**
 * CertDetail 组件测试。
 *
 * **回归守卫**：后端对「证书没有该扩展」编码成 `null`（`san` / `keyUsage` /
 * `extendedKeyUsage` / `cpsUris` / `noticeNumbers`），
 * 组件必须显式处理这些 `null`，否则 `null.length` 会抛 TypeError
 * 让整页白屏（摘要页签与链详情弹窗一起挂掉）。
 *
 * 组件按 `section` **只渲染一个区**，因此每个用例都要指定要测的区。
 */
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CertDetail from '@/components/tool/CertDetail.vue'
import { i18n } from '@/i18n'
import type { CertificateInfo, CertificateValidity } from '@/types/certificate'

/** 统一的挂载选项。 */
const mountOptions = { global: { plugins: [i18n] } }

/** 一个「什么扩展都没有」的证书信息：四个数组字段全是 `null`。 */
const emptyInfo: CertificateInfo = {
  version: 2,
  subject: [{ field: 'CN', value: 'root.example.com' }],
  subjectText: '/C=CN/O=Toolbit Test/CN=root.example.com',
  issuer: [{ field: 'CN', value: 'root.example.com' }],
  issuerText: '/C=CN/O=Toolbit Test/CN=root.example.com',
  serial: '01',
  notBefore: '2026-01-01T00:00:00Z',
  notAfter: '2027-01-01T00:00:00Z',
  publicKey: { algorithm: 'RSA', bits: 2048, curve: '' },
  signatureAlgorithm: 'RSA-SHA256',
  signatureAlgorithmOid: '1.2.840.113549.1.1.11',
  isCa: true,
  pathLen: -1,
  san: null,
  keyUsage: null,
  extendedKeyUsage: null,
  subjectKeyId: 'AB:CD',
  authorityKeyId: '',
  subjectHash: '01020304',
  issuerHash: '01020304',
  extensions: [],
  selfSigned: true,
  signatureHex: '00',
  policies: [],
}

/** 有效性判定。 */
const validity: CertificateValidity = {
  status: 'valid',
  daysRemaining: 9.6,
}

/** 一份公钥展开视图。 */
const publicKey = {
  algorithm: 'RSA',
  bits: 2048,
  curve: '',
  pem: '-----BEGIN PUBLIC KEY-----\nAA\n-----END PUBLIC KEY-----',
  spkiHex: '3000',
  fingerprint: 'b'.repeat(64),
  modulus: 'C0',
  exponent: '010001',
  x: '',
  y: '',
}

describe('CertDetail', () => {
  it('数组字段全为 null 时不抛错（回归守卫）', () => {
    for (const section of ['summary', 'names', 'identifiers', 'publicKey', 'extensions', 'pem'] as const) {
      expect(() =>
        mount(CertDetail, {
          ...mountOptions,
          props: { info: emptyInfo, validity, section },
        }),
      ).not.toThrow()
    }
  })

  it('摘要区渲染有效期、指纹与公钥', () => {
    const wrapper = mount(CertDetail, {
      ...mountOptions,
      props: {
        info: emptyInfo,
        validity,
        section: 'summary',
        fingerprints: [{ algorithm: 'sha256', value: 'a'.repeat(64) }],
        publicKey,
      },
    })

    const text = wrapper.text()
    expect(text).toContain('9.6')
    expect(text).toContain('RSA')
    expect(text).toContain('a'.repeat(64))
  })

  it('摘要区只给要紧信息，不渲染扩展与策略', () => {
    const wrapper = mount(CertDetail, {
      ...mountOptions,
      props: {
        info: {
          ...emptyInfo,
          extensions: [
            {
              nid: 19,
              oid: '2.5.29.19',
              name: 'basicConstraints',
              critical: true,
              value: '3000',
              text: 'CA:TRUE',
            },
          ],
        },
        validity,
        section: 'summary',
      },
    })

    expect(wrapper.text()).not.toContain('CA:TRUE')
  })

  it('有策略扩展但没有 CPS 时不崩', () => {
    const info: CertificateInfo = {
      ...emptyInfo,
      policies: [
        {
          oid: '1.2.3.4',
          cpsUris: null,
          organization: 'Example Org',
          noticeNumbers: null,
          explicitText: '',
        },
      ],
    }

    expect(() =>
      mount(CertDetail, { ...mountOptions, props: { info, validity, section: 'identifiers' } }),
    ).not.toThrow()
  })

  it('扩展区渲染扩展内容', () => {
    const wrapper = mount(CertDetail, {
      ...mountOptions,
      props: {
        info: {
          ...emptyInfo,
          extensions: [
            {
              nid: 19,
              oid: '2.5.29.19',
              name: 'basicConstraints',
              critical: true,
              value: '3000',
              text: 'CA:TRUE',
            },
          ],
        },
        validity,
        section: 'extensions',
      },
    })

    expect(wrapper.text()).toContain('CA:TRUE')
  })

  it('公钥区渲染公钥摘要', () => {
    const wrapper = mount(CertDetail, {
      ...mountOptions,
      props: { info: emptyInfo, validity, section: 'publicKey', publicKey },
    })

    expect(wrapper.text()).toContain('RSA')
    expect(wrapper.text()).toContain('2048')
  })
})
