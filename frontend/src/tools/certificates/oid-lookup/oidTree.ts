/**
 * OID 树的纯函数工具。
 *
 * 与界面无关，便于脱离网络单测：
 *   - `normalizeOid`：归一化用户输入的 OID；
 *   - `isValidOid`：校验 OID 语法；
 *   - `arcOf` / `parentOf`：取末弧与父 OID；
 *   - `mergeChildList`：把一页子节点并进已有的子节点表；
 *   - `buildBreadcrumb`：把祖先链摊平成面包屑片段。
 */

/** OID 字符串的最大长度。 */
export const OID_MAX_LENGTH = 256

/** OID 最大的弧数。 */
export const OID_MAX_ARCS = 128

/**
 * 归一化用户输入的 OID。
 *
 * 去掉空白、`OID:` 前缀与首尾的点；空串表示「未指定」。
 */
export function normalizeOid(raw: string): string {
  let value = raw.trim()
  if (value.toLowerCase().startsWith('oid:')) {
    value = value.slice(4)
  }
  value = value.trim().replace(/^\.+|\.+$/g, '')
  return value
}

/** 校验 OID 语法：一串非负整数弧，用点分隔。 */
export function isValidOid(oid: string): boolean {
  if (oid === '') {
    return false
  }
  const arcs = oid.split('.')
  if (arcs.length > OID_MAX_ARCS) {
    return false
  }
  return arcs.every((arc) => /^\d+$/.test(arc))
}

/**
 * 判断输入是否是一个**绝对** OID（以 0 / 1 / 2 开头）。
 *
 * 只写 `840.113549` 这类片段时应当按关键词搜索，而不是把它当成
 * 一个从根开始的 OID 去导航。
 */
export function isAbsoluteOid(oid: string): boolean {
  if (!isValidOid(oid)) {
    return false
  }
  const first = oid.split('.')[0]
  return first === '0' || first === '1' || first === '2'
}

/** 取末弧号；单段 OID（如根节点 `0`）返回自身。 */
export function arcOf(oid: string): string {
  const index = oid.lastIndexOf('.')
  return index < 0 ? oid : oid.slice(index + 1)
}

/** 取父 OID；顶层节点返回空串。 */
export function parentOf(oid: string): string {
  const index = oid.lastIndexOf('.')
  return index < 0 ? '' : oid.slice(0, index)
}

/** 取末弧的数值；非数字弧返回 NaN。 */
export function arcNumber(oid: string): number {
  return Number.parseInt(arcOf(oid), 10)
}

/**
 * 把一页子节点并进已有的子节点表。
 *
 * 后端按弧号**数值**升序返回，因此新页直接接在旧页后面即可；
 * 重复的 OID 会被去重（分页边界可能重叠时也保持稳定）。
 */
export function mergeChildList(
  existing: readonly string[],
  incoming: readonly string[],
): string[] {
  const merged = [...existing]
  const seen = new Set(existing)
  for (const oid of incoming) {
    if (seen.has(oid)) {
      continue
    }
    seen.add(oid)
    merged.push(oid)
  }
  return merged
}

/** 面包屑片段。 */
export interface OidCrumb {
  /** 节点的 OID。 */
  oid: string
  /** 节点的本地化名称。 */
  label: string
}

/**
 * 把祖先链摊平成面包屑片段。
 *
 * `nodes` 是「祖先 → 子节点」有序的节点表；
 * `labelOf` 负责取本地化名称。
 */
export function buildBreadcrumb<T extends { oid: string }>(
  nodes: readonly T[],
  labelOf: (node: T) => string,
): OidCrumb[] {
  return nodes.map((node) => ({ oid: node.oid, label: labelOf(node) }))
}
