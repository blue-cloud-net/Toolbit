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
  'html-escape': () => import('./encoding/html-escape/index'),

  // ── 格式化 / Formatting ──
  'json-format': () => import('./formatting/json-format/index'),
  'yaml-format': () => import('./formatting/yaml-format/index'),
  'toml-format': () => import('./formatting/toml-format/index'),
  'xml-format': () => import('./formatting/xml-format/index'),
  'html-format': () => import('./formatting/html-format/index'),
  'js-format': () => import('./formatting/js-format/index'),
  'sql-format': () => import('./formatting/sql-format/index'),
  'format-convert': () => import('./formatting/format-convert/index'),
  'color-convert': () => import('./formatting/color-convert/index'),

  // ── 密码学 / Crypto ──
  'hash-calc': () => import('./crypto/hash-calc/index'),
  'hmac-gen': () => import('./crypto/hmac-gen/index'),
  'password-gen': () => import('./crypto/password-gen/index'),
  'symmetric-crypt': () => import('./crypto/symmetric-crypt/index'),
  'signature': () => import('./crypto/signature/index'),
  'asymmetric-crypt': () => import('./crypto/asymmetric-crypt/index'),
  'ecdh': () => import('./crypto/ecdh/index'),
  'sm4-file-crypt': () => import('./crypto/sm4-file-crypt/index'),
  'password-policy': () => import('./crypto/password-policy/index'),
  'totp': () => import('./crypto/totp/index'),
  'keygen': () => import('./keys/keygen/index'),
  'key-convert': () => import('./keys/key-convert/index'),
  'key-parse': () => import('./keys/key-parse/index'),
  'key-ops': () => import('./keys/key-ops/index'),
  'key-match': () => import('./keys/key-match/index'),
  'jwk-xml-convert': () => import('./keys/jwk-xml-convert/index'),
  'ssh-keygen': () => import('./keys/ssh-keygen/index'),
  'cert-parse': () => import('./certificates/cert-parse/index'),
  'csr-parse': () => import('./certificates/csr-parse/index'),
  'pem-der': () => import('./certificates/pem-der/index'),
  'asn1-view': () => import('./certificates/asn1-view/index'),
  'cert-chain-parse': () => import('./certificates/cert-chain-parse/index'),
  'cert-chain-verify': () => import('./certificates/cert-chain-verify/index'),
  'cert-chain-repair': () => import('./certificates/cert-chain-repair/index'),
  'crl-parse': () => import('./certificates/crl-parse/index'),
  'csr-gen': () => import('./certificates/csr-gen/index'),
  'cert-gen': () => import('./certificates/cert-gen/index'),
  'crl-gen': () => import('./certificates/crl-gen/index'),
  'revocation-check': () => import('./certificates/revocation-check/index'),
  'ocsp-request': () => import('./certificates/ocsp-request/index'),
  'ocsp-response': () => import('./certificates/ocsp-response/index'),
  'pfx-pack': () => import('./certificates/pfx-pack/index'),
  'pfx-parse': () => import('./certificates/pfx-parse/index'),
  'pfx-change-password': () => import('./certificates/pfx-change-password/index'),
  'p7b-pack': () => import('./certificates/p7b-pack/index'),
  'p7b-extract': () => import('./certificates/p7b-extract/index'),
  'caa-gen': () => import('./certificates/caa-gen/index'),
  'cert-chain-fetch': () => import('./certificates/cert-chain-fetch/index'),
  'ct-search': () => import('./certificates/ct-search/index'),
  'domain-verify': () => import('./certificates/domain-verify/index'),
  'cert-lifecycle': () => import('./certificates/cert-lifecycle/index'),
  'cidr-calc': () => import('./network/cidr-calc/index'),
  'ssl-check': () => import('./network/ssl-check/index'),
  'dns-lookup': () => import('./network/dns-lookup/index'),
  'dns-tester': () => import('./network/dns-tester/index'),
  'websocket-test': () => import('./network/websocket-test/index'),
  'mime-lookup': () => import('./network/mime-lookup/index'),
  'http-status': () => import('./network/http-status/index'),
  'ip-range-expand': () => import('./network/ip-range-expand/index'),
  'device-headers': () => import('./network/device-headers/index'),

  // ── 文本 / Text ──
  'regex-tester': () => import('./text/regex-tester/index'),
  'text-convert': () => import('./text/text-convert/index'),
  'text-diff': () => import('./text/text-diff/index'),
  'regex-validate': () => import('./text/regex-validate/index'),
  'regex-build': () => import('./text/regex-build/index'),
  'data-mask': () => import('./text/data-mask/index'),
  'text-stats': () => import('./text/text-stats/index'),

  // ── 数据 / Data ──
  'timestamp': () => import('./data/timestamp/index'),
  'uuid-gen': () => import('./data/uuid-gen/index'),
  'date-calc': () => import('./data/date-calc/index'),
  'country-code': () => import('./data/country-code/index'),
  'cn-regions': () => import('./data/cn-regions/index'),
  'qr-code': () => import('./data/qr-code/index'),
  'id-validator': () => import('./data/id-validator/index'),
  'csv-convert': () => import('./data/csv-convert/index'),
  'json-to-types': () => import('./data/json-to-types/index'),
  'mock-data': () => import('./data/mock-data/index'),
  'id-generator': () => import('./data/id-generator/index'),
  'archive-inspect': () => import('./data/archive-inspect/index'),
  'file-analyze': () => import('./data/file-analyze/index'),
  'wifi-qr': () => import('./data/wifi-qr/index'),

  // ── 系统 / System ──
  'curl-generator': () => import('./system/curl-generator/index'),
  'crontab-generator': () => import('./system/crontab-generator/index'),
  'chmod-calc': () => import('./system/chmod-calc/index'),
  'lookup-table': () => import('./system/lookup-table/index'),
  'nginx-config': () => import('./system/nginx-config/index'),
  'firewall-rules': () => import('./system/firewall-rules/index'),
  'docker-compose-gen': () => import('./system/docker-compose-gen/index'),

  // ── 文档 / Documents ──
  'markdown-preview': () => import('./documents/markdown-preview/index'),
  'eml-parser': () => import('./documents/eml-parser/index'),
  'markdown-convert': () => import('./documents/markdown-convert/index'),
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
