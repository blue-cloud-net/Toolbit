/**
 * 工具图标映射。
 *
 * 图标属于**展示层**选择：分类图标来自 catalog（`icon` 字段），
 * 单个工具的图标在这里按 slug 登记，未知 slug 回落到所属分类图标，
 * 因此新增工具不登记也不会出现空图标。
 *
 * 语义按「工具做什么」选取（解析 / 生成 / 转换 / 校验…），
 * 而不是按分类重复同一个图标 —— 侧栏子项与折叠态悬浮菜单都依赖它区分条目。
 */
import type { IconName } from '@/components/common/AppIcon.vue'

/** slug → 图标名。 */
export const TOOL_ICONS: Record<string, IconName> = {
  // ── 加密 / Cryptography ──
  signature: 'pen',
  'asymmetric-crypt': 'unlock',
  'symmetric-crypt': 'lock',
  ecdh: 'link',
  'sm4-file-crypt': 'document',
  'hmac-gen': 'fingerprint',
  'password-gen': 'asterisk',
  'hash-calc': 'hash',
  'password-policy': 'shield-check',
  'totp': 'clock',

  // ── 密钥 / Keys ──
  keygen: 'key-plus',
  'key-convert': 'swap',
  'key-parse': 'key',
  'key-ops': 'calculator',
  'key-match': 'link',
  'jwk-xml-convert': 'braces',
  'ssh-keygen': 'key-plus',

  // ── 证书 / Certificates ──
  'cert-parse': 'certificate',
  'csr-gen': 'file-plus',
  'csr-parse': 'file-search',
  'pem-der': 'swap',
  'asn1-view': 'sitemap',
  'cert-chain-parse': 'chain',
  'cert-chain-verify': 'shield-check',
  'cert-chain-repair': 'refresh',
  'revocation-check': 'ban',
  'crl-parse': 'list',
  'crl-gen': 'list-plus',
  'ocsp-request': 'send',
  'ocsp-response': 'inbox',
  'cert-gen': 'certificate',
  'pfx-pack': 'box',
  'pfx-parse': 'box-open',
  'pfx-change-password': 'swap',
  'p7b-pack': 'box',
  'p7b-extract': 'box-open',
  'caa-gen': 'globe',
  'cert-chain-fetch': 'cloud-download',
  'ct-search': 'search',
  'domain-verify': 'globe-check',
  'cert-lifecycle': 'calendar',

  // ── 编码 / Encoding ──
  base64: 'swap',
  'url-codec': 'link',
  'jwt-codec': 'token',
  'number-base': 'hash',
  'encoding-convert': 'swap',
  'gzip-deflate': 'archive',
  'html-escape': 'tag',

  // ── 格式化 / Formatting ──
  'json-format': 'braces',
  'js-format': 'code',
  'sql-format': 'data',
  'html-format': 'tag',
  'xml-format': 'tag',
  'toml-format': 'brackets',
  'yaml-format': 'list',
  'format-convert': 'swap',
  'color-convert': 'asterisk',

  // ── 文本 / Text ──
  'regex-tester': 'regex',
  'text-convert': 'text',
  'text-diff': 'swap',
  'regex-validate': 'shield-check',
  'regex-build': 'braces',
  'data-mask': 'asterisk',
  'text-stats': 'text',

  // ── 数据 / Data ──
  timestamp: 'clock',
  'uuid-gen': 'fingerprint',
  'date-calc': 'calendar',
  'country-code': 'globe',
  'cn-regions': 'map',
  'qr-code': 'qr',
  'id-validator': 'certificate',
  'csv-convert': 'data',
  'json-to-types': 'braces',
  'mock-data': 'list',
  'id-generator': 'fingerprint',
  'archive-inspect': 'archive',
  'file-analyze': 'file-search',

  // ── 系统 / System ──
  'curl-generator': 'terminal',
  'crontab-generator': 'clock',
  'chmod-calc': 'shield',
  'lookup-table': 'list',
  'nginx-config': 'terminal',
  'firewall-rules': 'shield',
  'docker-compose-gen': 'box',

  // ── 网络 / Network ──
  'cidr-calc': 'sitemap',
  'ssl-check': 'shield-check',
  'dns-lookup': 'globe',
  'dns-tester': 'pulse',
  'websocket-test': 'connection',
  'ip-range-expand': 'sitemap',
  'device-headers': 'monitor',
  'mime-lookup': 'tag',
  'http-status': 'globe-check',

  // ── 文档 / Documents ──
  'pdf-sign-verify': 'document-check',
  'markdown-preview': 'markdown',
  'eml-parser': 'mail',
  'markdown-convert': 'markdown',
}

/** 取工具图标名；未登记时回落到调用方给出的分类图标。 */
export function toolIcon(slug: string, fallback: IconName = 'document'): IconName {
  return TOOL_ICONS[slug] ?? fallback
}
