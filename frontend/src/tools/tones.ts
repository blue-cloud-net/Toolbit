/**
 * 图标色调映射。
 *
 * 图标颜色是**展示层**选择：与「主色（accent）」正交 —— 主色负责交互高亮，
 * 色调负责图标识别，因此图标不再统一成一个主色，而是每个分类、每个工具
 * 各有各的颜色。
 *
 * 色调只对应 `tokens.css` 里的 `--tb-tone-<名>` / `--tb-tone-<名>-soft`
 * 两组令牌，**不写死颜色**；新增色调时先在 `tokens.css` 补令牌，
 * 再在这里登记。
 */

/** 图标色调名；与 `tokens.css` 的 `--tb-tone-<名>` 一一对应。 */
export type Tone =
  | 'indigo'
  | 'blue'
  | 'sky'
  | 'cyan'
  | 'teal'
  | 'green'
  | 'lime'
  | 'amber'
  | 'orange'
  | 'rose'
  | 'violet'
  | 'slate'

/** 分类 slug → 色调；分类下的工具未单独登记时回落到这里。 */
export const CATEGORY_TONES: Record<string, Tone> = {
  crypto: 'violet',
  keys: 'amber',
  certificates: 'green',
  encoding: 'indigo',
  formatting: 'teal',
  text: 'sky',
  data: 'lime',
  system: 'slate',
  network: 'cyan',
  documents: 'orange',
}

/**
 * 工具 slug → 色调。
 *
 * 选取贴合「工具做什么」的语义色：解析 / 查看偏冷色，生成偏暖色，
 * 校验偏绿，打包 / 转换偏紫。同一分类下的工具因此彼此可区分，
 * 整屏图标也不会挤成同一个颜色。
 */
export const TOOL_TONES: Record<string, Tone> = {
  // ── 加密 / Cryptography ──
  signature: 'violet',
  'asymmetric-crypt': 'rose',
  'symmetric-crypt': 'amber',
  ecdh: 'teal',
  'sm4-file-crypt': 'orange',
  'hmac-gen': 'green',
  'password-gen': 'lime',
  'hash-calc': 'indigo',
  'password-policy': 'sky',
  totp: 'cyan',

  // ── 密钥 / Keys ──
  keygen: 'amber',
  'key-convert': 'violet',
  'key-parse': 'blue',
  'key-ops': 'teal',
  'key-match': 'rose',
  'jwk-xml-convert': 'green',
  'ssh-keygen': 'orange',

  // ── 证书 / Certificates ──
  'cert-parse': 'blue',
  'csr-gen': 'green',
  'csr-parse': 'teal',
  'pem-der': 'violet',
  'asn1-view': 'sky',
  'cert-chain-parse': 'indigo',
  'cert-chain-verify': 'green',
  'cert-chain-repair': 'orange',
  'revocation-check': 'rose',
  'crl-parse': 'slate',
  'crl-gen': 'amber',
  'ocsp-request': 'cyan',
  'ocsp-response': 'lime',
  'cert-gen': 'teal',
  'pfx-pack': 'violet',
  'pfx-parse': 'indigo',
  'pfx-change-password': 'rose',
  'p7b-pack': 'sky',
  'p7b-extract': 'blue',
  'caa-gen': 'green',
  'cert-chain-fetch': 'cyan',
  'ct-search': 'amber',
  'domain-verify': 'teal',
  'cert-lifecycle': 'orange',

  // ── 编码 / Encoding ──
  base64: 'indigo',
  'url-codec': 'blue',
  'jwt-codec': 'violet',
  'number-base': 'sky',
  'encoding-convert': 'teal',
  'gzip-deflate': 'amber',
  'html-escape': 'orange',

  // ── 格式化 / Formatting ──
  'json-format': 'amber',
  'js-format': 'lime',
  'sql-format': 'sky',
  'html-format': 'orange',
  'xml-format': 'teal',
  'toml-format': 'blue',
  'yaml-format': 'violet',
  'format-convert': 'green',
  'color-convert': 'rose',

  // ── 文本 / Text ──
  'regex-tester': 'violet',
  'text-convert': 'blue',
  'text-diff': 'lime',
  'regex-validate': 'green',
  'regex-build': 'teal',
  'data-mask': 'rose',
  'text-stats': 'amber',

  // ── 数据 / Data ──
  timestamp: 'sky',
  'uuid-gen': 'indigo',
  'date-calc': 'blue',
  'country-code': 'green',
  'cn-regions': 'teal',
  'qr-code': 'violet',
  'id-validator': 'orange',
  'csv-convert': 'amber',
  'json-to-types': 'cyan',
  'mock-data': 'rose',
  'archive-inspect': 'lime',
  'file-analyze': 'slate',
  'wifi-qr': 'green',
  'id-generator': 'violet',

  // ── 系统 / System ──
  'curl-generator': 'cyan',
  'crontab-generator': 'indigo',
  'chmod-calc': 'amber',
  'nginx-config': 'teal',
  'firewall-rules': 'rose',
  'docker-compose-gen': 'blue',
  'lookup-table': 'violet',

  // ── 网络 / Network ──
  'cidr-calc': 'sky',
  'ssl-check': 'green',
  'dns-lookup': 'blue',
  'dns-tester': 'teal',
  'websocket-test': 'violet',
  'device-headers': 'amber',
  'ip-range-expand': 'orange',
  'mime-lookup': 'indigo',
  'http-status': 'cyan',

  // ── 文档 / Documents ──
  'pdf-sign-verify': 'rose',
  'markdown-preview': 'indigo',
  'eml-parser': 'cyan',
  'markdown-convert': 'lime',
}

/** 取工具色调；未登记时回落到所属分类，再回落到 `slate`。 */
export function toolTone(slug: string, category?: string): Tone {
  const own = TOOL_TONES[slug]
  if (own) {
    return own
  }
  if (category) {
    const fallback = CATEGORY_TONES[category]
    if (fallback) {
      return fallback
    }
  }
  return 'slate'
}

/** 取分类色调；未登记时回落到 `slate`。 */
export function categoryTone(category: string | undefined): Tone {
  if (!category) {
    return 'slate'
  }
  return CATEGORY_TONES[category] ?? 'slate'
}

/**
 * 色调对应的 CSS 变量。
 *
 * 组件只需把结果绑到 `:style` 上，再用 `var(--tb-tone)` / `var(--tb-tone-soft)`
 * 取色，**组件内不出现具体颜色值**。
 */
export function toneStyle(tone: Tone): Record<string, string> {
  return {
    '--tb-tone': `var(--tb-tone-${tone})`,
    '--tb-tone-soft': `var(--tb-tone-${tone}-soft)`,
  }
}
