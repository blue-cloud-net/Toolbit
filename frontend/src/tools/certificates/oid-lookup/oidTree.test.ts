/**
 * OID 树纯函数的单元测试。
 *
 * 这些函数不依赖网络与界面，因此可以脱离组件单测。
 */
import { describe, expect, it } from 'vitest'

import {
  arcNumber,
  arcOf,
  buildBreadcrumb,
  isAbsoluteOid,
  isValidOid,
  mergeChildList,
  normalizeOid,
  OID_MAX_ARCS,
  parentOf,
} from './oidTree'

describe('normalizeOid', () => {
  it('去空白与首尾的点', () => {
    expect(normalizeOid('  1.2.3  ')).toBe('1.2.3')
    expect(normalizeOid('.1.2.3.')).toBe('1.2.3')
  })

  it('去掉 OID: 前缀', () => {
    expect(normalizeOid('OID:1.2.3')).toBe('1.2.3')
    expect(normalizeOid('oid: 1.2.3')).toBe('1.2.3')
  })

  it('空串仍是空串', () => {
    expect(normalizeOid('   ')).toBe('')
  })
})

describe('isValidOid', () => {
  it('接受点分十进制弧', () => {
    expect(isValidOid('1.3.6.1.5.5.7.3.36')).toBe(true)
    expect(isValidOid('0')).toBe(true)
    expect(isValidOid('2.5.29')).toBe(true)
  })

  it('拒绝非数字弧', () => {
    expect(isValidOid('1.2.3.x')).toBe(false)
    expect(isValidOid('1.2.')).toBe(false)
    expect(isValidOid('.1.2')).toBe(false)
    expect(isValidOid('1..2')).toBe(false)
    expect(isValidOid('')).toBe(false)
  })

  it('拒绝超长弧数', () => {
    const many = Array.from({ length: OID_MAX_ARCS + 1 }, () => '1').join('.')
    expect(isValidOid(many)).toBe(false)
    const ok = Array.from({ length: OID_MAX_ARCS }, () => '1').join('.')
    expect(isValidOid(ok)).toBe(true)
  })
})

describe('isAbsoluteOid', () => {
  it('接受以 0/1/2 开头的绝对 OID', () => {
    expect(isAbsoluteOid('1.3.6.1.5.5.7.3.36')).toBe(true)
    expect(isAbsoluteOid('0')).toBe(true)
    expect(isAbsoluteOid('2.5.29')).toBe(true)
  })

  it('把片段或非法输入留给关键词搜索', () => {
    expect(isAbsoluteOid('840.113549')).toBe(false)
    expect(isAbsoluteOid('subjectAltName')).toBe(false)
    expect(isAbsoluteOid('')).toBe(false)
  })
})

describe('arcOf / parentOf / arcNumber', () => {
  it('取末弧', () => {
    expect(arcOf('1.3.6.1.5.5.7.3.36')).toBe('36')
    expect(arcOf('0')).toBe('0')
  })

  it('取父 OID', () => {
    expect(parentOf('1.3.6.1.5.5.7.3.36')).toBe('1.3.6.1.5.5.7.3')
    expect(parentOf('0')).toBe('')
  })

  it('按数值解析末弧', () => {
    expect(arcNumber('1.3.6.1.5.5.7.3.10')).toBe(10)
    expect(arcNumber('1.3.6.1.5.5.7.3.2')).toBe(2)
    expect(Number.isNaN(arcNumber('1.3.x'))).toBe(true)
  })
})

describe('mergeChildList', () => {
  it('把新页接在旧页后面', () => {
    expect(mergeChildList(['1.2'], ['1.3', '1.4'])).toEqual(['1.2', '1.3', '1.4'])
  })

  it('去重（分页边界重叠时保持稳定）', () => {
    expect(mergeChildList(['1.2', '1.3'], ['1.3', '1.4'])).toEqual(['1.2', '1.3', '1.4'])
  })

  it('空列表也能并入', () => {
    expect(mergeChildList([], ['1.2'])).toEqual(['1.2'])
  })
})

describe('buildBreadcrumb', () => {
  it('把祖先链摊平成面包屑', () => {
    const crumbs = buildBreadcrumb([{ oid: '1' }, { oid: '1.3.6.1' }], (node) => node.oid)
    expect(crumbs).toEqual([
      { oid: '1', label: '1' },
      { oid: '1.3.6.1', label: '1.3.6.1' },
    ])
  })

  it('空链得到空面包屑', () => {
    expect(buildBreadcrumb([], () => '')).toEqual([])
  })
})
