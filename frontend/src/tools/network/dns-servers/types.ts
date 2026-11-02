/**
 * dns-servers 工具的共享类型。
 *
 * 字段名与后端 `DNSServersResult` / `DNSServerEntry` **逐字对齐**
 * （见 `docs/api/tools/network.md` 的 dns-servers 一节）。
 */

/** 一条 DNS 服务器端点记录。 */
export interface DNSServerEntry {
  /** 服务商中文名。 */
  provider: string
  /** 服务商英文名。 */
  providerEn: string
  /** 归属地区：`china` / `asia` / `global`。 */
  region: string
  /** 地址族：`ipv4` / `ipv6` / `domain`。 */
  family: string
  /** 查询协议：`udp` / `tcp` / `dot` / `doh`。 */
  proto: string
  /** 服务器地址；DoH 时是完整 URL。 */
  addr: string
  /** 端点说明。 */
  note: string
}

/** 后端 dns-servers 工具的响应。 */
export interface DNSServersResult {
  /** 命中的端点。 */
  servers: DNSServerEntry[]
  /** 命中条数。 */
  total: number
  /** 命中的服务商数。 */
  providers: number
  /** 数据集版本。 */
  dataVersion: string
}

/** 卡片模式下的一条端点：**同一个地址**的多个协议合并成一行。 */
export interface DNSCardEndpoint {
  /** 地址字面量（DoH 时是完整 URL）。 */
  addr: string
  /** 地址族。 */
  family: string
  /** 该地址支持的协议（保持目录顺序：`udp` → `tcp` → `dot` → `doh`）。 */
  protos: string[]
  /** 端点说明（取该地址下第一条记录的说明）。 */
  note: string
  /** 「实测」跳转时使用的协议（取该地址下第一条记录的协议）。 */
  primaryProto: string
}

/** 卡片模式下的一家服务商。 */
export interface DNSCardProvider {
  /** 服务商中文名。 */
  provider: string
  /** 服务商英文名。 */
  providerEn: string
  /** 归属地区。 */
  region: string
  /** 该服务商的端点（按目录顺序）。 */
  endpoints: DNSCardEndpoint[]
}
