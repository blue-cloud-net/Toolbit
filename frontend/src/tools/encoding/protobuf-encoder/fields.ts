/**
 * protobuf-encoder 的字段模型。
 *
 * 表格编辑与 JSON 文本编辑共用同一份节点模型：
 * 两种形态之间只做序列化 / 反序列化，不各自维护状态。
 */

/** 字段节点。 */
export interface ProtoFieldNode {
  /** 字段号（1–536870911）。 */
  field: number
  /** 语义类型名。 */
  type: string
  /** 值的文本表示；**一律是字符串**，避免 64 位整数在前端丢精度。 */
  value: string
  /** bytes 类型的值编码（utf8 / hex / base64 / base64url）。 */
  encoding: string
  /** message 类型的子字段。 */
  children: ProtoFieldNode[]
}

/** 后端请求里的字段规格。 */
export interface ProtoFieldPayload {
  field: number
  type: string
  value: string
  encoding: string
  fields: ProtoFieldPayload[]
}

/** 可选的语义类型名（顺序稳定，与后端一致）。 */
export const FIELD_TYPES: readonly string[] = [
  'int32',
  'int64',
  'uint32',
  'uint64',
  'sint32',
  'sint64',
  'bool',
  'enum',
  'fixed32',
  'sfixed32',
  'float',
  'fixed64',
  'sfixed64',
  'double',
  'string',
  'bytes',
  'message',
]

/** 需要子字段的类型。 */
export const NESTED_TYPE = 'message'

/** 新建一个空节点。 */
export function createNode(): ProtoFieldNode {
  return { field: 1, type: 'string', value: '', encoding: '', children: [] }
}

/** 深拷贝节点树。 */
export function cloneNodes(nodes: ProtoFieldNode[]): ProtoFieldNode[] {
  return nodes.map((node) => ({
    field: node.field,
    type: node.type,
    value: node.value,
    encoding: node.encoding,
    children: cloneNodes(node.children),
  }))
}

/** 节点树 → 后端请求字段。 */
export function toPayload(nodes: ProtoFieldNode[]): ProtoFieldPayload[] {
  return nodes.map((node) => ({
    field: node.field,
    type: node.type,
    value: node.value,
    encoding: node.encoding,
    fields: node.type === NESTED_TYPE ? toPayload(node.children) : [],
  }))
}

/** 节点树 → JSON 文本（两空格缩进）。 */
export function toJsonText(nodes: ProtoFieldNode[]): string {
  return JSON.stringify(toJsonItems(nodes), null, 2)
}

/** 节点树 → JSON 数组项。 */
function toJsonItems(nodes: ProtoFieldNode[]): unknown[] {
  return nodes.map((node) => {
    const item: Record<string, unknown> = {
      field: node.field,
      type: node.type,
      value: node.value,
    }
    if (node.encoding !== '') {
      item.encoding = node.encoding
    }
    if (node.type === NESTED_TYPE && node.children.length > 0) {
      item.fields = toJsonItems(node.children)
    }
    return item
  })
}

/** JSON 文本 → 节点树。 */
export function fromJsonText(text: string): ProtoFieldNode[] {
  const trimmed = text.trim()
  if (trimmed === '') {
    return []
  }

  let parsed: unknown
  try {
    parsed = JSON.parse(trimmed)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'invalid json')
  }
  return fromUnknown(parsed)
}

/** 任意 JSON 值 → 节点树。 */
function fromUnknown(parsed: unknown): ProtoFieldNode[] {
  if (!Array.isArray(parsed)) {
    throw new Error('字段规格必须是 JSON 数组')
  }
  return parsed.map((item) => nodeFromUnknown(item))
}

/** 单个 JSON 项 → 节点。 */
function nodeFromUnknown(item: unknown): ProtoFieldNode {
  if (item === null || typeof item !== 'object' || Array.isArray(item)) {
    throw new Error('字段规格的每一项都必须是对象')
  }

  const record = item as Record<string, unknown>
  const field = Number(record.field)
  if (!Number.isInteger(field)) {
    throw new Error('字段号必须是整数')
  }

  const type = typeof record.type === 'string' ? record.type : ''
  if (!FIELD_TYPES.includes(type)) {
    throw new Error(`未知类型：${type}`)
  }

  const rawValue = record.value
  let value = ''
  if (typeof rawValue === 'string') {
    value = rawValue
  } else if (typeof rawValue === 'number' || typeof rawValue === 'boolean') {
    value = String(rawValue)
  } else if (rawValue !== undefined && rawValue !== null) {
    throw new Error('字段值必须是字符串、数字或布尔值')
  }

  const encoding = typeof record.encoding === 'string' ? record.encoding : ''
  const children = record.fields === undefined ? [] : fromUnknown(record.fields)

  return { field, type, value, encoding, children }
}

/** 在节点树里按路径删除节点。 */
export function removeNode(nodes: ProtoFieldNode[], path: number[]): ProtoFieldNode[] {
  if (path.length === 0) {
    return nodes
  }
  const [head, ...rest] = path
  if (rest.length === 0) {
    return nodes.filter((_, index) => index !== head)
  }
  return nodes.map((node, index) => {
    if (index !== head) {
      return node
    }
    return { ...node, children: removeNode(node.children, rest) }
  })
}
