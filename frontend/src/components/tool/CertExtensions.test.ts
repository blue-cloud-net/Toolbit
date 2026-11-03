/**
 * CertExtensions 组件测试。
 *
 * 覆盖「全部扩展」清单：已知与未知扩展都在表里，
 * 值用后端解出的可读文本，没有可读文本时回落十六进制。
 */
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CertExtensions from '@/components/tool/CertExtensions.vue'
import { i18n } from '@/i18n'
import type { CertificateExtension } from '@/types/certificate'

/** 统一的挂载选项。 */
const mountOptions = { global: { plugins: [i18n] } }

/** 构造一个扩展条目。 */
function ext(
  oid: string,
  name: string,
  text: string,
  critical = false,
): CertificateExtension {
  return { nid: 1, oid, name, critical, value: 'ab12', text }
}

describe('CertExtensions', () => {
  it('列出全部扩展及其可读值', () => {
    const wrapper = mount(CertExtensions, {
      ...mountOptions,
      props: {
        extensions: [
          ext('2.5.29.17', 'subjectAltName', 'DNS:example.com'),
          ext('2.5.29.19', 'basicConstraints', 'CA:TRUE', true),
        ],
      },
    })

    const text = wrapper.text()
    expect(text).toContain('CA:TRUE')
    expect(text).toContain('DNS:example.com')
  })

  it('未知扩展也在清单里，值回落十六进制', () => {
    const wrapper = mount(CertExtensions, {
      ...mountOptions,
      props: {
        extensions: [ext('1.2.3.4', 'unknownExt', '')],
      },
    })

    expect(wrapper.text()).toContain('unknownExt')
    expect(wrapper.text()).toContain('AB12')
  })

  it('没有扩展时不渲染表格', () => {
    const wrapper = mount(CertExtensions, {
      ...mountOptions,
      props: { extensions: [] },
    })

    expect(wrapper.find('table').exists()).toBe(false)
  })
})
