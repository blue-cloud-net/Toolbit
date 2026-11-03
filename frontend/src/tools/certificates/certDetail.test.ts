/**
 * 证书工具的纯函数测试。
 *
 * 覆盖「数据 → 展示形态」的映射：位掩码、OID、剩余天数级别、
 * 有效期比例、链角色与断链原因的键映射。
 */
import { describe, expect, it } from 'vitest'

import {
  chainRoleKey,
  displayVersion,
  durationDays,
  extensionValue,
  extKeyUsageEntries,
  formatDateTime,
  formatDays,
  hexUpper,
  joinChainPem,
  keyUsageEntries,
  linkMethodKey,
  linkReasonKey,
  safetyLevel,
  safetyLevelColor,
  toList,
  validityLevel,
  validityPercent,
} from './certDetail'

describe('keyUsageEntries', () => {
  it('给出 RFC 5280 的位掩码', () => {
    const entries = keyUsageEntries(['digitalSignature', 'keyEncipherment'])
    expect(entries).toEqual([
      { name: 'digitalSignature', value: '0x80' },
      { name: 'keyEncipherment', value: '0x20' },
    ])
  })

  it('未知短名没有掩码', () => {
    expect(keyUsageEntries(['nope'])[0]?.value).toBe('')
  })
})

describe('extKeyUsageEntries', () => {
  it('给出点分 OID', () => {
    const entries = extKeyUsageEntries([{ name: 'serverAuth', oid: '1.3.6.1.5.5.7.3.1' }])
    expect(entries[0]?.value).toBe('1.3.6.1.5.5.7.3.1')
    expect(entries[0]?.name).toBe('serverAuth')
  })

  it('OpenSSL 没有登记的 OID 仍带 OID，名字为空', () => {
    const entries = extKeyUsageEntries([{ name: '', oid: '1.3.6.1.5.5.7.3.36' }])
    expect(entries[0]?.value).toBe('1.3.6.1.5.5.7.3.36')
    expect(entries[0]?.name).toBe('')
  })
})

describe('validityLevel', () => {
  it('映射有效性状态', () => {
    expect(validityLevel('valid')).toBe('success')
    expect(validityLevel('not_yet_valid')).toBe('warning')
    expect(validityLevel('expired')).toBe('danger')
    expect(validityLevel('weird')).toBe('muted')
  })
})

describe('safetyLevel', () => {
  it('按 30 天阈值分档', () => {
    expect(safetyLevel(365)).toBe('valid')
    expect(safetyLevel(29)).toBe('expiring')
    expect(safetyLevel(-1)).toBe('expired')
  })

  it('颜色与档位对应', () => {
    expect(safetyLevelColor('valid')).toBe('success')
    expect(safetyLevelColor('expiring')).toBe('warning')
    expect(safetyLevelColor('expired')).toBe('danger')
  })
})

describe('validityPercent', () => {
  const notBefore = '2026-01-01T00:00:00Z'
  const notAfter = '2027-01-01T00:00:00Z'

  it('按剩余天数推算已过去的比例', () => {
    expect(validityPercent(notBefore, notAfter, 365)).toBe(0)
    expect(validityPercent(notBefore, notAfter, 0)).toBe(100)
    expect(validityPercent(notBefore, notAfter, 182)).toBe(50)
  })

  it('比例钳在 0–100', () => {
    expect(validityPercent(notBefore, notAfter, 400)).toBe(0)
    expect(validityPercent(notBefore, notAfter, -10)).toBe(100)
  })

  it('时间非法时返回 0', () => {
    expect(validityPercent('bad', 'worse', 0)).toBe(0)
  })
})

describe('durationDays', () => {
  it('计算有效期总天数', () => {
    expect(durationDays('2026-01-01T00:00:00Z', '2027-01-01T00:00:00Z')).toBe(365)
  })

  it('不足 1 天不再显示成 0 天', () => {
    expect(durationDays('2026-01-01T00:00:00Z', '2026-01-01T12:00:00Z')).toBe(0.5)
  })

  it('时间非法时返回 0', () => {
    expect(durationDays('bad', 'worse')).toBe(0)
  })
})

describe('toList', () => {
  it('null 与 undefined 归一化为空数组', () => {
    expect(toList(null)).toEqual([])
    expect(toList(undefined)).toEqual([])
  })

  it('已有数组原样返回', () => {
    const items = ['a', 'b']
    expect(toList(items)).toEqual(items)
  })
})

describe('formatDays', () => {
  it('不足 10 天保留 1 位小数', () => {
    expect(formatDays(9.6)).toBe('9.6')
    expect(formatDays(0.4)).toBe('0.4')
  })

  it('10 天以上取整', () => {
    expect(formatDays(12)).toBe('12')
    expect(formatDays(365.4)).toBe('365')
  })

  it('负数同样保留 1 位小数', () => {
    expect(formatDays(-2.4)).toBe('-2.4')
    expect(formatDays(-30)).toBe('-30')
  })

  it('非法值返回 0', () => {
    expect(formatDays(Number.NaN)).toBe('0')
  })
})

describe('joinChainPem', () => {
  it('块间恰好一个换行，没有空行', () => {
    const joined = joinChainPem([
      '-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n',
      '-----BEGIN CERTIFICATE-----\nBBBB\n-----END CERTIFICATE-----\n',
    ])
    expect(joined).toBe(
      '-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n'
      + '-----BEGIN CERTIFICATE-----\nBBBB\n-----END CERTIFICATE-----\n',
    )
    expect(joined).not.toContain('\n\n')
  })

  it('过滤空块', () => {
    expect(joinChainPem(['', '  ', 'AAA'])).toBe('AAA\n')
  })

  it('空输入返回空串', () => {
    expect(joinChainPem([])).toBe('')
  })
})

describe('extensionValue', () => {
  it('优先用可读文本', () => {
    expect(extensionValue({
      nid: 1,
      oid: '2.5.29.19',
      name: 'basicConstraints',
      critical: true,
      value: '3000',
      text: 'CA:TRUE',
    })).toBe('CA:TRUE')
  })

  it('没有可读文本时回落到十六进制', () => {
    expect(extensionValue({
      nid: 1,
      oid: '1.2.3.4',
      name: 'unknown',
      critical: false,
      value: 'ab12',
      text: '',
    })).toBe('AB12')
  })
})

describe('formatDateTime', () => {
  it('格式化 RFC3339 时间', () => {
    expect(formatDateTime('2026-01-02T03:04:05Z', 'en-US')).toContain('2026')
  })

  it('无法解析时原样返回', () => {
    expect(formatDateTime('not a date', 'en-US')).toBe('not a date')
  })

  it('空串返回空串', () => {
    expect(formatDateTime('', 'en-US')).toBe('')
  })
})

describe('hexUpper', () => {
  it('转大写并处理空值', () => {
    expect(hexUpper('ab12')).toBe('AB12')
    expect(hexUpper(undefined)).toBe('')
  })
})

describe('displayVersion', () => {
  it('ASN.1 原始值转展示文本', () => {
    expect(displayVersion(2)).toBe('v3')
    expect(displayVersion(0)).toBe('v1')
  })
})

describe('链角色与连接诊断的键映射', () => {
  it('角色', () => {
    expect(chainRoleKey('leaf')).toBe('leaf')
    expect(chainRoleKey('intermediate')).toBe('intermediate')
    expect(chainRoleKey('root')).toBe('root')
    expect(chainRoleKey('unknown')).toBe('unlinked')
  })

  it('连接方式', () => {
    expect(linkMethodKey('keyid')).toBe('keyid')
    expect(linkMethodKey('name')).toBe('name')
    expect(linkMethodKey('self')).toBe('self')
    expect(linkMethodKey('other')).toBe('leaf')
  })

  it('断链原因', () => {
    expect(linkReasonKey('keyid_mismatch')).toBe('keyid_mismatch')
    expect(linkReasonKey('missing_intermediate')).toBe('missing_intermediate')
    expect(linkReasonKey('unlinked')).toBe('unlinked')
    expect(linkReasonKey('')).toBe('ok')
  })
})
