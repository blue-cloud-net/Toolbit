/**
 * DNS 服务器预设的纯数据校验。
 *
 * 预设表是页面上「主用 / 备用 / DoH」三组标签的唯一来源，
 * 一旦分组或地址写错，页面会向后端发出非法请求。
 */
import { describe, expect, it } from 'vitest'

import {
  DEFAULT_PRESET_IDS,
  DNS_PRESET_GROUPS,
  DNS_SERVER_PRESETS,
  MAX_SERVERS,
} from './presets'

describe('DNS 服务器预设', () => {
  it('预设标识唯一', () => {
    const ids = DNS_SERVER_PRESETS.map((preset) => preset.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('每个分组都有预设', () => {
    for (const group of DNS_PRESET_GROUPS) {
      const count = DNS_SERVER_PRESETS.filter((preset) => preset.group === group).length
      expect(count, `分组 ${group} 没有预设`).toBeGreaterThan(0)
    }
  })

  it('预设的分组都在声明的分组列表里', () => {
    for (const preset of DNS_SERVER_PRESETS) {
      expect(DNS_PRESET_GROUPS).toContain(preset.group)
    }
  })

  it('DoH 分组只用 https 端点', () => {
    for (const preset of DNS_SERVER_PRESETS.filter((item) => item.group === 'doh')) {
      expect(preset.proto).toBe('doh')
      expect(preset.addr.startsWith('https://')).toBe(true)
    }
  })

  it('主用与备用分组是 host:port 形式的明文地址', () => {
    for (const preset of DNS_SERVER_PRESETS.filter(
      (item) => item.group === 'primary' || item.group === 'backup',
    )) {
      expect(preset.proto).toBe('udp')
      expect(preset.addr).toMatch(/^\d{1,3}(\.\d{1,3}){3}:\d+$/)
    }
  })

  it('默认选中的预设都存在', () => {
    const ids = new Set(DNS_SERVER_PRESETS.map((preset) => preset.id))
    for (const id of DEFAULT_PRESET_IDS) {
      expect(ids.has(id), `默认预设 ${id} 不存在`).toBe(true)
    }
  })

  it('默认预设不超过服务器数量上限', () => {
    expect(DEFAULT_PRESET_IDS.length).toBeLessThanOrEqual(MAX_SERVERS)
  })
})
