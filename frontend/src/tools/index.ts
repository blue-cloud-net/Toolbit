/**
 * 工具组件静态登记表。
 *
 * slug → 组件是**编译期**映射：目录数据只提供展示元信息，
 * 不在运行期拼装组件。
 *
 * 组件用动态 `import()` 加载，Vite 会按此拆出独立的 chunk。
 */
import type { Component } from 'vue'

/** 工具组件加载器。 */
export type ToolLoader = () => Promise<{ default: Component }>

/** 工具登记表。 */
export const toolRegistry: Record<string, ToolLoader> = {
  // ── 编码 / Encoding ──
  'base64': () => import('./encoding/base64/index'),
  'number-base': () => import('./encoding/number-base/index'),
  'url-codec': () => import('./encoding/url-codec/index'),
  'jwt-codec': () => import('./encoding/jwt-codec/index'),
  'encoding-convert': () => import('./encoding/encoding-convert/index'),
  'gzip-deflate': () => import('./encoding/gzip-deflate/index'),

  // ── 格式化 / Formatting ──
  'json-format': () => import('./formatting/json-format/index'),
  'yaml-format': () => import('./formatting/yaml-format/index'),
  'toml-format': () => import('./formatting/toml-format/index'),
  'xml-format': () => import('./formatting/xml-format/index'),
  'html-format': () => import('./formatting/html-format/index'),
  'js-format': () => import('./formatting/js-format/index'),
  'sql-format': () => import('./formatting/sql-format/index'),
  'format-convert': () => import('./formatting/format-convert/index'),

  // ── 密码学 / Crypto ──
  'hash-calc': () => import('./crypto/hash-calc/index'),
  'hmac-gen': () => import('./crypto/hmac-gen/index'),
  'password-gen': () => import('./crypto/password-gen/index'),
  'symmetric-crypt': () => import('./crypto/symmetric-crypt/index'),
  'signature': () => import('./crypto/signature/index'),
  'asymmetric-crypt': () => import('./crypto/asymmetric-crypt/index'),
  'ecdh': () => import('./crypto/ecdh/index'),

  // ── 文本 / Text ──
  'regex-tester': () => import('./text/regex-tester/index'),
  'text-convert': () => import('./text/text-convert/index'),

  // ── 数据 / Data ──
  'timestamp': () => import('./data/timestamp/index'),
  'uuid-gen': () => import('./data/uuid-gen/index'),
  'date-calc': () => import('./data/date-calc/index'),
  'country-code': () => import('./data/country-code/index'),
  'cn-regions': () => import('./data/cn-regions/index'),
  'qr-code': () => import('./data/qr-code/index'),

  // ── 系统 / System ──
  'curl-generator': () => import('./system/curl-generator/index'),
  'crontab-generator': () => import('./system/crontab-generator/index'),
  'chmod-calc': () => import('./system/chmod-calc/index'),

  // ── 文档 / Documents ──
  'markdown-preview': () => import('./documents/markdown-preview/index'),
  'eml-parser': () => import('./documents/eml-parser/index'),
}

/** 已实现的工具 slug 集合。 */
export const IMPLEMENTED_SLUGS: readonly string[] = Object.keys(toolRegistry)

/** 判断工具是否已实现。 */
export function isImplemented(slug: string): boolean {
  return Object.prototype.hasOwnProperty.call(toolRegistry, slug)
}

/** 取工具组件加载器；未实现时返回 undefined。 */
export function getToolLoader(slug: string): ToolLoader | undefined {
  return toolRegistry[slug]
}
