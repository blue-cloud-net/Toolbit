/**
 * 记录一致性比较的单元测试。
 *
 * 覆盖三种结论，尤其是「只有一台服务器」——
 * 那种情况下没有可比对象，必须报「不适用」而不是「存在差异」。
 */
import { describe, expect, it } from 'vitest'

import {
  compareAnswers,
  serverLabel,
  shortServerLabel,
  type DNSServerResult,
} from './compare'

/** 构造一台成功的服务器。 */
function ok(server: string, proto: string, values: string[]): DNSServerResult {
  return {
    server,
    proto,
    latencyMs: 10,
    answers: values.map((value) => ({
      name: 'example.com',
      type: value.split(' ')[0] ?? 'A',
      ttl: 300,
      value: value.split(' ').slice(1).join(' '),
    })),
    error: null,
  }
}

/** 构造一台失败的服务器。 */
function failed(server: string, proto = 'udp'): DNSServerResult {
  return { server, proto, latencyMs: 5000, answers: [], error: { code: 'UPSTREAM_TIMEOUT', message: 'timeout' } }
}

describe('服务器短名', () => {
  it('去掉端口与协议前缀', () => {
    expect(shortServerLabel('1.1.1.1:53')).toBe('1.1.1.1')
    expect(shortServerLabel('one.one.one.one:853')).toBe('one.one.one.one')
    expect(shortServerLabel('https://dns.google/dns-query')).toBe('dns.google')
  })

  it('IPv6 不会被截断成半个地址', () => {
    expect(shortServerLabel('[2606:4700:4700::1111]:53')).toBe('2606:4700:4700::1111')
  })
})

describe('服务器标签', () => {
  it('带上协议以区分同一地址的不同通道', () => {
    expect(serverLabel(ok('1.1.1.1:53', 'udp', ['A 1.1.1.1']))).toBe('1.1.1.1 · UDP')
    expect(serverLabel(ok('1.1.1.1:53', 'tcp', ['A 1.1.1.1']))).toBe('1.1.1.1 · TCP')
    expect(serverLabel(ok('one.one.one.one:853', 'dot', []))).toBe('one.one.one.one · DoT')
    expect(serverLabel(ok('https://dns.google/dns-query', 'doh', []))).toBe('dns.google · DoH')
  })
})

describe('记录一致性比较', () => {
  it('只有一台服务器时报「不适用」而不是「存在差异」', () => {
    const report = compareAnswers([ok('1.1.1.1:53', 'udp', ['A 1.1.1.1'])])
    expect(report.state).toBe('na')
    expect(report.differences).toHaveLength(0)
    expect(report.compared).toBe(1)
  })

  it('一台都没有时报「不适用」', () => {
    expect(compareAnswers([]).state).toBe('na')
    expect(compareAnswers([failed('1.1.1.1:53')]).state).toBe('na')
  })

  it('两台返回相同记录时报「一致」', () => {
    const report = compareAnswers([
      ok('1.1.1.1:53', 'udp', ['A 1.1.1.1', 'A 1.0.0.1']),
      ok('8.8.8.8:53', 'udp', ['A 1.1.1.1', 'A 1.0.0.1']),
    ])
    expect(report.state).toBe('consistent')
    expect(report.differences).toHaveLength(0)
    expect(report.compared).toBe(2)
  })

  it('TTL 不同不算差异', () => {
    const a = ok('1.1.1.1:53', 'udp', ['A 1.1.1.1'])
    const b = ok('8.8.8.8:53', 'udp', ['A 1.1.1.1'])
    b.answers[0]!.ttl = 60
    expect(compareAnswers([a, b]).state).toBe('consistent')
  })

  it('记录缺失时报「存在差异」并指出缺哪台', () => {
    const report = compareAnswers([
      ok('1.1.1.1:53', 'udp', ['A 1.1.1.1']),
      ok('8.8.8.8:53', 'udp', ['A 1.1.1.1', 'A 8.8.8.8']),
    ])
    expect(report.state).toBe('inconsistent')
    expect(report.differences).toHaveLength(1)
    expect(report.differences[0]?.value).toBe('A 8.8.8.8')
    expect(report.differences[0]?.missing).toEqual(['1.1.1.1 · UDP'])
  })

  it('失败的服务器不参与比较', () => {
    const report = compareAnswers([
      ok('1.1.1.1:53', 'udp', ['A 1.1.1.1']),
      ok('8.8.8.8:53', 'udp', ['A 1.1.1.1']),
      failed('9.9.9.9:53'),
    ])
    expect(report.state).toBe('consistent')
    expect(report.compared).toBe(2)
  })

  it('一台成功一台失败时报「不适用」', () => {
    const report = compareAnswers([ok('1.1.1.1:53', 'udp', ['A 1.1.1.1']), failed('8.8.8.8:53')])
    expect(report.state).toBe('na')
    expect(report.compared).toBe(1)
  })

  it('同一地址的不同协议通道分别参与比较', () => {
    const report = compareAnswers([
      ok('1.1.1.1:53', 'udp', ['A 1.1.1.1']),
      ok('1.1.1.1:53', 'tcp', ['A 1.0.0.1']),
    ])
    expect(report.state).toBe('inconsistent')
    // 两条记录各自缺一台：UDP 缺 A 1.0.0.1，TCP 缺 A 1.1.1.1。
    expect(report.differences.map((row) => [row.value, row.missing.join()])).toEqual([
      ['A 1.0.0.1', '1.1.1.1 · UDP'],
      ['A 1.1.1.1', '1.1.1.1 · TCP'],
    ])
  })
})
