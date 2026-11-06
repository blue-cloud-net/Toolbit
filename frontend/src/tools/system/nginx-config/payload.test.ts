/**
 * Nginx 配置工具请求体构造的测试。
 */
import { describe, expect, it } from 'vitest'

import type { NginxNode, NginxSite } from '@/types/nginx'

import { buildNginxPayload, createNginxSite, parseUpstreams } from './payload'

/** 构造一个站点。 */
function site(overrides: Partial<NginxSite> = {}): NginxSite {
  return { ...createNginxSite(), ...overrides }
}

/** 构造一个节点。 */
function node(overrides: Partial<NginxNode> = {}): NginxNode {
  return {
    id: 'n1',
    type: 'directive',
    context: 'main',
    name: 'worker_processes',
    args: '4',
    body: '',
    file: '',
    comment: '',
    children: [],
    ...overrides,
  }
}

describe('nginx 配置请求体', () => {
  it('简单模式提交预设与站点', () => {
    const payload = buildNginxPayload({
      mode: 'simple',
      output: 'bundle',
      targetVersion: '',
      preset: 'static',
      sites: [site({ serverName: 'example.com' })],
      nodes: [],
      responseShape: 'files',
    })

    expect(payload.mode).toBe('simple')
    expect(payload.output).toBe('bundle')
    expect(payload.preset).toBe('static')
    expect(Array.isArray(payload.sites)).toBe(true)
    expect(payload.nodes).toBeUndefined()
  })

  it('高级模式只提交配置树', () => {
    const payload = buildNginxPayload({
      mode: 'advanced',
      output: 'single',
      targetVersion: '1.30.6',
      preset: 'static',
      sites: [site()],
      nodes: [node()],
      responseShape: 'files',
    })

    expect(payload.mode).toBe('advanced')
    expect(payload.nodes).toHaveLength(1)
    expect(payload.sites).toBeUndefined()
    expect(payload.targetVersion).toBe('1.30.6')
  })

  it('未选目标版本时不提交该字段', () => {
    const payload = buildNginxPayload({
      mode: 'simple',
      output: 'bundle',
      targetVersion: '',
      preset: 'static',
      sites: [],
      nodes: [],
      responseShape: 'files',
    })

    expect('targetVersion' in payload).toBe(false)
  })

  it('打包请求切换响应形态', () => {
    const payload = buildNginxPayload({
      mode: 'simple',
      output: 'bundle',
      targetVersion: '',
      preset: 'static',
      sites: [],
      nodes: [],
      responseShape: 'bundle',
    })

    expect(payload.responseShape).toBe('bundle')
  })
})

describe('后端地址解析', () => {
  it('按逗号切分并忽略空白项', () => {
    expect(parseUpstreams('127.0.0.1:8081, 127.0.0.1:8082 ,,')).toEqual([
      '127.0.0.1:8081',
      '127.0.0.1:8082',
    ])
  })

  it('空输入返回空列表', () => {
    expect(parseUpstreams('   ')).toEqual([])
  })
})

describe('站点工厂', () => {
  it('新站点带唯一标识与默认端口', () => {
    const first = createNginxSite()
    const second = createNginxSite()

    expect(first.id).not.toBe(second.id)
    expect(first.listenPort).toBe(80)
    expect(first.serverName).toBe('example.com')
  })
})
