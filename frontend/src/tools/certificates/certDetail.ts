/**
 * 证书类工具的纯函数工具集。
 *
 * 只做「数据 → 展示形态」的转换，不依赖 Vue：
 * 便于在 vitest 中独立覆盖，也避免各工具页各写一份。
 */

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

/** 扩展密钥用途的点分 OID。 */
export const EKU_OID: Record<string, string> = {
  anyExtendedKeyUsage: '2.5.29.37.0',
  serverAuth: '1.3.6.1.5.5.7.3.1',
  clientAuth: '1.3.6.1.5.5.7.3.2',
  codeSigning: '1.3.6.1.5.5.7.3.3',
  emailProtection: '1.3.6.1.5.5.7.3.4',
  timeStamping: '1.3.6.1.5.5.7.3.8',
  OCSPSigning: '1.3.6.1.5.5.7.3.9',
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
 * 未知短名没有登记 OID，`value` 为空串。
 */
export function extKeyUsageEntries(names: string[]): UsageEntry[] {
  return names.map((name) => ({
    name,
    value: EKU_OID[name] ?? '',
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
 * 阈值与 cert-lifecycle 一致：不足 30 天即视为临期。
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
 * 有效期总天数。
 */
export function durationDays(notBefore: string, notAfter: string): number {
  const start = Date.parse(notBefore)
  const end = Date.parse(notAfter)
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) {
    return 0
  }
  return Math.round((end - start) / 86_400_000)
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
