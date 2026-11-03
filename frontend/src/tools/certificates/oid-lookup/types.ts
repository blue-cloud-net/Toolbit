/**
 * OID 查询工具的前端类型。
 *
 * 与后端 `backend/internal/tools/pki/oidlookup.go` 的响应结构一一对应。
 */

/** OID 节点种类。 */
export type OidKind = 'root' | 'branch' | 'leaf'

/** OID 节点状态。 */
export type OidStatus = 'current' | 'reserved' | 'deprecated' | 'obsolete'

/** 一个 OID 节点。 */
export interface OidNode {
  /** 点分十进制对象标识符，如 `1.3.6.1.5.5.7.3.36`。 */
  oid: string
  /** 末弧号，如 `36`。 */
  arc: string
  /** 官方英文名。 */
  name: string
  /** 中文名。 */
  nameZh: string
  /** 官方英文说明。 */
  description: string
  /** 中文说明。 */
  descriptionZh: string
  /** 节点种类。 */
  kind: OidKind
  /** 节点状态。 */
  status: OidStatus
  /** 来源标注，如 `RFC 5280`。 */
  source: string
  /** 来源链接。 */
  sourceUrl: string
  /** 历史别名与常见写法。 */
  aliases: string[]
  /** 直接子节点数量。 */
  childCount: number
}

/** 一条搜索命中。 */
export interface OidMatch {
  /** 命中的节点。 */
  node: OidNode
  /** 祖先链（从根到父节点）。 */
  path: OidNode[]
}

/** 查询模式。 */
export type OidMode = 'root' | 'node' | 'search'

/** 后端 oid-lookup 工具的响应。 */
export interface OidLookupResult {
  /** 查询模式。 */
  mode: OidMode
  /** 规范化后的查询 OID。 */
  oid: string
  /** 规范化后的搜索词。 */
  query: string
  /** 是否命中精确节点。 */
  matched: boolean
  /** 命中的节点；未命中时为 null。 */
  node: OidNode | null
  /** 命中节点的祖先链。 */
  path: OidNode[]
  /** 祖先链各层的子节点。 */
  pathChildren: Record<string, OidNode[]>
  /** 命中节点的直接子节点。 */
  children: OidNode[]
  /** 子节点总数或命中总数。 */
  total: number
  /** 分页偏移量。 */
  offset: number
  /** 单页条数。 */
  limit: number
  /** 是否还有下一页。 */
  hasMore: boolean
  /** 搜索结果。 */
  matches: OidMatch[]
  /** 未命中时的最近祖先 OID。 */
  nearestOid: string
  /** 相对最近祖先缺失的弧后缀。 */
  missingSuffix: string
  /** 数据集版本。 */
  dataVersion: string
}
