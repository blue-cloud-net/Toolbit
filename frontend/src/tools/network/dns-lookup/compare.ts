/**
 * DNS 查询结果的**记录一致性**比较。
 *
 * 抽成独立模块（原逻辑内联在 `Tool.vue` 里），便于单测覆盖三种结论：
 *
 * - `na` —— 成功返回的服务器**不足两台**：没有可比对象，
 *   此时**不能**报「存在差异」（只有一台时谈不上差异）；
 * - `consistent` —— 各台成功服务器的记录集合完全相同；
 * - `inconsistent` —— 存在只出现在部分服务器上的记录。
 *
 * 参与比较的只有**成功**的服务器：失败服务器的记录集合不可信，
 * 把它算进来会把「超时」误报成「记录缺失」。
 */

/** 单条解析结果。 */
export interface DNSAnswer {
  name: string
  type: string
  ttl: number
  value: string
}

/** 单台服务器的结果。 */
export interface DNSServerResult {
  server: string
  proto: string
  latencyMs: number
  answers: DNSAnswer[]
  error: { code: string; message: string } | null
}

/** 一致性结论。 */
export type ConsistencyState = 'na' | 'consistent' | 'inconsistent'

/** 一条差异：某条记录在哪些服务器上缺失。 */
export interface AnswerDifference {
  /** 记录（类型 + 值）。 */
  value: string
  /** 缺少该记录的服务器标签。 */
  missing: string[]
}

/** 比较报告。 */
export interface ConsistencyReport {
  /** 结论。 */
  state: ConsistencyState
  /** 参与比较的成功服务器数量。 */
  compared: number
  /** 差异明细；仅 `inconsistent` 时非空。 */
  differences: AnswerDifference[]
}

/** 协议标签。 */
const PROTO_LABELS: Record<string, string> = {
  udp: 'UDP',
  tcp: 'TCP',
  dot: 'DoT',
  doh: 'DoH',
}

/**
 * 服务器短名：去掉协议前缀、路径与端口。
 *
 * IPv6 要**先**去掉方括号再截端口，否则 `[2606:4700:4700::1111]:53`
 * 会被截成 `[2606`。
 */
export function shortServerLabel(server: string): string {
  const withoutScheme = server.replace(/^https?:\/\//, '')
  const hostPart = withoutScheme.split('/')[0] ?? withoutScheme
  if (hostPart.startsWith('[')) {
    const end = hostPart.indexOf(']')
    return end > 0 ? hostPart.slice(1, end) : hostPart.slice(1)
  }
  const colon = hostPart.indexOf(':')
  return colon > 0 ? hostPart.slice(0, colon) : hostPart
}

/**
 * 服务器标签：短名 + 协议。
 *
 * 带上协议是必需的：同一个地址既有明文 UDP 又有 TCP 通道
 * （如 `1.1.1.1:53`），只显示地址时分不清是哪一路。
 */
export function serverLabel(server: DNSServerResult): string {
  const proto = PROTO_LABELS[server.proto] ?? server.proto.toUpperCase()
  return `${shortServerLabel(server.server)} · ${proto}`
}

/**
 * 比较各台服务器的记录集合。
 *
 * 记录按「类型 + 值」比较：TTL 不同不算差异（TTL 单独在汇总里给区间），
 * 同一条记录在一台服务器上重复出现只算一条。
 */
export function compareAnswers(servers: DNSServerResult[]): ConsistencyReport {
  const ok = servers.filter((item) => item.error === null)
  if (ok.length < 2) {
    // 不足两台：没有可比对象，只能说「不适用」，不能说「有差异」。
    return { state: 'na', compared: ok.length, differences: [] }
  }

  const sets = ok.map((item) => new Set(item.answers.map((a) => `${a.type} ${a.value}`)))
  const union = new Set<string>()
  for (const set of sets) {
    for (const value of set) {
      union.add(value)
    }
  }

  const differences: AnswerDifference[] = []
  for (const value of union) {
    const missing = ok
      .filter((_, index) => !sets[index]?.has(value))
      .map((item) => serverLabel(item))
    if (missing.length > 0) {
      differences.push({ value, missing })
    }
  }
  differences.sort((a, b) => a.value.localeCompare(b.value))

  return {
    state: differences.length === 0 ? 'consistent' : 'inconsistent',
    compared: ok.length,
    differences,
  }
}
