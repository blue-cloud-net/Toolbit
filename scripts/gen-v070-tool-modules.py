#!/usr/bin/env python3
"""一次性生成 v0.7.0 八个工具的 index.ts 与 example.ts。"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
SRC = ROOT / "frontend" / "src" / "tools"

INDEX_TEMPLATE = '''/**
 * {slug} 工具模块。
 */
import type {{ Component }} from 'vue'

import enUS from '@/locales/en-US/tools/{slug}.json'
import zhCN from '@/locales/zh-CN/tools/{slug}.json'

import Tool from './Tool.vue'

/** 工具词条。 */
export const messages: Record<'zh-CN' | 'en-US', Record<string, unknown>> = {{
  'zh-CN': zhCN,
  'en-US': enUS,
}}

export default Tool as Component
'''

EXAMPLES = {
    "network/cidr-calc": '''/**
 * cidr-calc 工具的示例数据。
 */
export interface CIDRExample {
  /** 示例名称。 */
  label: string
  /** 输入方式：cidr / range。 */
  mode: 'cidr' | 'range'
  /** 网段。 */
  cidr: string
  /** 起始地址。 */
  start: string
  /** 结束地址。 */
  end: string
}

/** 默认示例：IPv4 /24 网段。 */
export const DEFAULT_EXAMPLE: CIDRExample = {
  label: 'IPv4 /24 网段',
  mode: 'cidr',
  cidr: '192.168.1.0/24',
  start: '10.0.0.1',
  end: '10.0.0.6',
}

/** 全部示例。 */
export const EXAMPLES: readonly CIDRExample[] = [DEFAULT_EXAMPLE]
''',
    "network/ssl-check": '''/**
 * ssl-check 工具的示例数据。
 */
export interface SSLExample {
  /** 示例名称。 */
  label: string
  /** 目标主机。 */
  host: string
  /** 目标端口。 */
  port: number
  /** SNI。 */
  sni: string
}

/** 默认示例：检测公共站点。 */
export const DEFAULT_EXAMPLE: SSLExample = {
  label: '检测公共站点',
  host: 'example.com',
  port: 443,
  sni: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly SSLExample[] = [DEFAULT_EXAMPLE]
''',
    "network/dns-lookup": '''/**
 * dns-lookup 工具的示例数据。
 */
export interface DNSLookupExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  host: string
  /** 记录类型。 */
  type: string
  /** 自定义服务器（host:port）；为空表示用系统解析器。 */
  addr: string
}

/** 默认示例：查询 A 记录。 */
export const DEFAULT_EXAMPLE: DNSLookupExample = {
  label: '查询 A 记录',
  host: 'example.com',
  type: 'A',
  addr: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSLookupExample[] = [DEFAULT_EXAMPLE]
''',
    "network/dns-tester": '''/**
 * dns-tester 工具的示例数据。
 */
export interface DNSTesterExample {
  /** 示例名称。 */
  label: string
  /** 检测项：whois / dnssec / ptr。 */
  action: 'whois' | 'dnssec' | 'ptr'
  /** 域名。 */
  domain: string
  /** IP 地址。 */
  ip: string
}

/** 默认示例：查询 DNSSEC 状态。 */
export const DEFAULT_EXAMPLE: DNSTesterExample = {
  label: '查询 DNSSEC 状态',
  action: 'dnssec',
  domain: 'example.com',
  ip: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSTesterExample[] = [DEFAULT_EXAMPLE]
''',
    "network/websocket-test": '''/**
 * websocket-test 工具的示例数据。
 */
export interface WebSocketExample {
  /** 示例名称。 */
  label: string
  /** 目标地址。 */
  url: string
  /** 子协议。 */
  protocols: string
  /** 发送内容。 */
  message: string
}

/** 默认示例：公共回显服务。 */
export const DEFAULT_EXAMPLE: WebSocketExample = {
  label: '公共回显服务',
  url: 'wss://echo.websocket.org',
  protocols: '',
  message: 'hello toolbit',
}

/** 全部示例。 */
export const EXAMPLES: readonly WebSocketExample[] = [DEFAULT_EXAMPLE]
''',
    "certificates/cert-chain-fetch": '''/**
 * cert-chain-fetch 工具的示例数据。
 */
export interface ChainFetchExample {
  /** 示例名称。 */
  label: string
  /** 目标主机。 */
  host: string
  /** 目标端口。 */
  port: number
  /** SNI。 */
  serverName: string
}

/** 默认示例：抓取公共站点证书链。 */
export const DEFAULT_EXAMPLE: ChainFetchExample = {
  label: '抓取公共站点证书链',
  host: 'example.com',
  port: 443,
  serverName: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly ChainFetchExample[] = [DEFAULT_EXAMPLE]
''',
    "certificates/ct-search": '''/**
 * ct-search 工具的示例数据。
 */
export interface CTSearchExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  domain: string
  /** 匹配方式。 */
  matchType: 'exact' | 'contains' | 'wildcard'
  /** 返回条数。 */
  limit: number
}

/** 默认示例：查询 example.com。 */
export const DEFAULT_EXAMPLE: CTSearchExample = {
  label: '查询 example.com',
  domain: 'example.com',
  matchType: 'exact',
  limit: 20,
}

/** 全部示例。 */
export const EXAMPLES: readonly CTSearchExample[] = [DEFAULT_EXAMPLE]
''',
    "certificates/domain-verify": '''/**
 * domain-verify 工具的示例数据。
 */
export interface DomainVerifyExample {
  /** 示例名称。 */
  label: string
  /** 验证方式。 */
  type: 'dns-txt' | 'dns-cname' | 'http-file'
  /** 域名。 */
  domain: string
  /** 期望值。 */
  expected: string
  /** 站点文件路径。 */
  filePath: string
}

/** 默认示例：校验 TXT 记录。 */
export const DEFAULT_EXAMPLE: DomainVerifyExample = {
  label: '校验 TXT 记录',
  type: 'dns-txt',
  domain: 'example.com',
  expected: 'toolbit-verify=abc123',
  filePath: '/.well-known/toolbit-verify.txt',
}

/** 全部示例。 */
export const EXAMPLES: readonly DomainVerifyExample[] = [DEFAULT_EXAMPLE]
''',
}

for target, body in EXAMPLES.items():
    folder = SRC / target
    folder.mkdir(parents=True, exist_ok=True)
    slug = target.split("/")[-1]
    (folder / "index.ts").write_text(INDEX_TEMPLATE.format(slug=slug), encoding="utf-8")
    (folder / "example.ts").write_text(body, encoding="utf-8")
    print("wrote", folder.relative_to(ROOT))
