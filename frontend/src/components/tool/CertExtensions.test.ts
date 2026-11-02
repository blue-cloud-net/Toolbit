/**
 * CertExtensions 组件测试。
 *
 * 覆盖两段式展示：**关键扩展**（只列实际带有的、按固定顺序）
 * 与 **全部扩展**（含未知扩展的完整清单）。
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
  it('关键扩展只列实际带有的项', () => {
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

  it('未知扩展只出现在全部扩展里', () => {
    const wrapper = mount(CertExtensions, {
      ...mountOptions,
      props: {
        extensions: [ext('1.2.3.4', 'unknownExt', '')],
      },
    })

    // 全部扩展段给出短名；关键扩展段不包含它。
    expect(wrapper.text()).toContain('unknownExt')
    expect(wrapper.text()).toContain('AB12')
  })

  it('没有扩展时不渲染任何段', () => {
    const wrapper = mount(CertExtensions, {
      ...mountOptions,
      props: { extensions: [] },
    })

    expect(wrapper.find('table').exists()).toBe(false)
  })
})
