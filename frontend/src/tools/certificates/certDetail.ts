/**
 * 证书类工具的纯函数工具集。
 *
 * 只做「数据 → 展示形态」的转换，不依赖 Vue：
 * 便于在 vitest 中独立覆盖，也避免各工具页各写一份。
 */
import type { CertificateExtension, ExtKeyUsageInfo } from '@/types/certificate'

/** 语义级别；与共享组件的 `level` 取值一致。 */
export type CertLevel = 'success' | 'warning' | 'danger' | 'info' | 'muted'

/** 剩余天数的紧急度。 */
export type SafetyLevel = 'valid' | 'expiring' | 'expired'

/** Key Usage 位掩码（RFC 5280 §4.2.1.3 的九个位）。 */
export const KU_BIT_HEX: Record<string, string> = {
  digitalSignature: '80',
  nonRepudiation: '40',
  keyEncipherment: '20',
  dataEncipherment: '10',
  keyAgreement: '08',
  keyCertSign: '04',
  cRLSign: '02',
  encipherOnly: '01',
  decipherOnly: '80',
}

/** 一个用途条目。 */
export interface UsageEntry {
  /** 上游给出的短名（如 digitalSignature）。 */
  name: string
  /** 附加信息：Key Usage 是位掩码，扩展密钥用途是点分 OID。 */
  value: string
}

/**
 * Key Usage 条目：每个置位用途附带其位掩码。
 *
 * 未知短名没有掩码，`value` 为空串。
 */
export function keyUsageEntries(names: string[]): UsageEntry[] {
  return names.map((name) => ({
    name,
    value: KU_BIT_HEX[name] ? `0x${KU_BIT_HEX[name]}` : '',
  }))
}

/**
 * 扩展密钥用途条目：每个用途附带其点分 OID。
 *
 * OID 由**后端**给出（它从扩展的原始 DER 里解出），
 * 因此 OpenSSL 对象库里没有登记的 OID 也能显示 —— 只是没有名字。
 *
 * 没有名字时 `name` 为空串，调用方应回落到 OID 展示。
 */
export function extKeyUsageEntries(items: ExtKeyUsageInfo[]): UsageEntry[] {
  return items.map((item) => ({
    name: item.name,
    value: item.oid,
  }))
}

/**
 * 有效性状态 → 语义级别。
 *
 * `valid` 为 success，`not_yet_valid` 为 warning，`expired` 为 danger；
 * 未知状态按 muted 处理（不误导）。
 */
export function validityLevel(status: string): CertLevel {
  switch (status) {
    case 'valid':
      return 'success'
    case 'not_yet_valid':
      return 'warning'
    case 'expired':
      return 'danger'
    default:
      return 'muted'
  }
}

/**
 * 剩余天数 → 紧急度。
 *
 * 阈值 30 天：不足 30 天即视为临期。
 */
export function safetyLevel(daysRemaining: number): SafetyLevel {
  if (daysRemaining < 0) {
    return 'expired'
  }
  if (daysRemaining < 30) {
    return 'expiring'
  }
  return 'valid'
}

/**
 * 紧急度 → 语义级别。
 */
export function safetyLevelColor(level: SafetyLevel): CertLevel {
  switch (level) {
    case 'valid':
      return 'success'
    case 'expiring':
      return 'warning'
    default:
      return 'danger'
  }
}

/**
 * 有效期已过去的比例（0–100）。
 *
 * 用后端给的剩余天数推算，保证与「判断时刻」一致：
 * 填了历史时刻时，进度条也停在那一刻。
 */
export function validityPercent(
  notBefore: string,
  notAfter: string,
  daysRemaining: number,
): number {
  const start = Date.parse(notBefore)
  const end = Date.parse(notAfter)
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) {
    return 0
  }
  const totalDays = (end - start) / 86_400_000
  if (totalDays <= 0) {
    return 0
  }
  const elapsedDays = totalDays - daysRemaining
  const percent = Math.round((elapsedDays / totalDays) * 100)
  if (percent < 0) {
    return 0
  }
  if (percent > 100) {
    return 100
  }
  return percent
}

/**
 * 有效期总天数（保留 1 位小数）。
 *
 * 短命证书（不足 1 天）不再显示成「0 天」。
 */
export function durationDays(notBefore: string, notAfter: string): number {
  const start = Date.parse(notBefore)
  const end = Date.parse(notAfter)
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    return 0
  }
  return Math.round(((end - start) / 86_400_000) * 10) / 10
}

/**
 * 数组归一化：`null` / `undefined` → 空数组。
 *
 * 后端对「已查询但无该项」编码成 `null`（见 types/certificate.ts 的说明），
 * 直接读 `.length` / `.map` 会抛 TypeError 让整页白屏。
 * 所有读取后端数组的地方都必须过这一层。
 */
export function toList<T>(value: T[] | null | undefined): T[] {
  return value ?? []
}

/**
 * 天数渲染。
 *
 * 不足 10 天时保留 1 位小数（「9.6 天」），否则取整（「12 天」）。
 * 临期证书需要区分「还剩 9.6 天」与「还剩 9 天」。
 */
export function formatDays(value: number): string {
  if (!Number.isFinite(value)) {
    return '0'
  }
  if (Math.abs(value) < 10) {
    return value.toFixed(1)
  }
  return String(Math.trunc(value))
}

/**
 * 按 RFC 7468 严格形式拼接多段 PEM。
 *
 * §3 Figure 3（Strict）：`stricttextualmsg = preeb eol strictbase64text posteb eol`
 * ⇒ 每个块以 `posteb eol` 收尾，**块与块之间恰好一个换行、没有空行**。
 * §2 要求解析侧忽略空白，因此**输入**时块间空行可有可无。
 */
export function joinChainPem(blocks: string[]): string {
  const parts = blocks.map((block) => block.trim()).filter((block) => block !== '')
  if (parts.length === 0) {
    return ''
  }
  return `${parts.join('\n')}\n`
}

/**
 * 全部扩展（保持证书里的原始顺序）。
 */
export function allExtensions(extensions: CertificateExtension[]): CertificateExtension[] {
  return [...extensions]
}

/**
 * 扩展的可读值。
 *
 * 优先用后端解出的可读文本（OpenSSL `x509 -text` 体例）；
 * 未知扩展回落十六进制原值。
 */
export function extensionValue(ext: CertificateExtension): string {
  return ext.text || hexUpper(ext.value)
}

/**
 * 按当前语言格式化时间。
 *
 * 输入是后端给的 RFC3339 字符串；解析失败时原样返回。
 */
export function formatDateTime(iso: string, locale: string): string {
  if (!iso) {
    return ''
  }
  const value = Date.parse(iso)
  if (!Number.isFinite(value)) {
    return iso
  }
  try {
    return new Date(value).toLocaleString(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })
  } catch {
    return new Date(value).toISOString()
  }
}

/**
 * 十六进制统一大写；空值返回空串。
 */
export function hexUpper(text: string | undefined): string {
  return text ? text.toUpperCase() : ''
}

/**
 * 链上角色 → 词条键后缀。
 *
 * 词条在共享命名空间 `tool.cert.role.*` 下。
 */
export function chainRoleKey(role: string): string {
  switch (role) {
    case 'leaf':
      return 'leaf'
    case 'intermediate':
      return 'intermediate'
    case 'root':
      return 'root'
    default:
      return 'unlinked'
  }
}

/**
 * 断链原因 → 词条键后缀。
 *
 * 词条在共享命名空间 `tool.cert.link.*` 下。
 */
export function linkReasonKey(reason: string): string {
  switch (reason) {
    case 'keyid_mismatch':
      return 'keyid_mismatch'
    case 'missing_intermediate':
      return 'missing_intermediate'
    case 'unlinked':
      return 'unlinked'
    default:
      return 'ok'
  }
}

/**
 * 连接方式 → 词条键后缀。
 */
export function linkMethodKey(method: string): string {
  switch (method) {
    case 'keyid':
      return 'keyid'
    case 'name':
      return 'name'
    case 'self':
      return 'self'
    default:
      return 'leaf'
  }
}

/**
 * 证书版本号 → 展示文本（后端给的是 ASN.1 原始值 0/1/2）。
 */
export function displayVersion(version: number): string {
  return `v${version + 1}`
}
