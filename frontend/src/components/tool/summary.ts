/**
 * 结果摘要构建器。
 *
 * 工具结果里有大量「若干个标量字段」——时间戳的 4 种表示、文件分析的 9 项元数据、
 * 文本统计的 10 个计数、证书解析的若干项属性。
 *
 * 逐个工具手写 `StatGridItem[]` 会重复取值、判空、类型转换这三段逻辑，
 * 因此把它固化成一份**声明式**描述：工具只声明「字段名 + 标签」，
 * 取值与格式化交给本模块。
 *
 * 用法：
 *
 * ```ts
 * const summary = computed(() =>
 *   buildSummary(data.value, [
 *     { key: 'mimeType', label: t('tools.file-analyze.results.mimeType') },
 *     { key: 'isTextual', label: t('tools.file-analyze.results.isTextual'), booleanLevel: { true: 'success', false: 'muted' } },
 *   ], { yes: t('common.states.yes'), no: t('common.states.no') }),
 * )
 * ```
 *
 * 语义色只能来自共享组件的取值范围（`success` / `danger` / `warning` / `info` / `muted`），
 * **禁止**在工具页自造颜色。
 */
import type { StatGridItem, StatGridLevel } from '@/components/tool/StatGrid.vue'

/** 摘要字段描述。 */
export interface SummaryField {
  /** 响应里的字段名；支持点号路径（如 `timestamp.s`）。 */
  key: string
  /** 字段标签（已本地化）。 */
  label: string
  /** 是否用等宽字体；缺省 `true`。 */
  mono?: boolean
  /** 固定语义级别。 */
  level?: StatGridLevel
  /** 按取值推导语义级别；返回 `undefined` 表示不着色。 */
  levelBy?: (value: unknown) => StatGridLevel | undefined
  /** 布尔字段的语义级别映射。 */
  booleanLevel?: { true: StatGridLevel; false: StatGridLevel }
  /** 自定义格式化；缺省按类型转字符串。 */
  format?: (value: unknown) => string
  /** 整行展示（长值不折行）。 */
  wide?: boolean
}

/** 构建摘要时的通用文案。 */
export interface SummaryOptions {
  /** 「是」的文案（布尔字段）。 */
  yes: string
  /** 「否」的文案（布尔字段）。 */
  no: string
  /** 空值占位符；缺省 `—`。 */
  empty?: string
}

/** 读取点号路径；任一层缺失返回 `undefined`。 */
function readPath(source: Record<string, unknown>, path: string): unknown {
  if (!path.includes('.')) {
    return source[path]
  }
  let current: unknown = source
  for (const part of path.split('.')) {
    if (typeof current !== 'object' || current === null) {
      return undefined
    }
    current = (current as Record<string, unknown>)[part]
  }
  return current
}

/**
 * 按描述构建摘要卡片。
 *
 * `source` 为空时返回空数组（调用方据此决定是否渲染摘要区）。
 */
export function buildSummary(
  source: Record<string, unknown> | null | undefined,
  fields: readonly SummaryField[],
  options: SummaryOptions,
): StatGridItem[] {
  if (!source) {
    return []
  }
  const empty = options.empty ?? '—'
  const items: StatGridItem[] = []

  for (const field of fields) {
    const raw = readPath(source, field.key)
    const missing = raw === undefined || raw === null || raw === ''
    const value = missing ? '' : raw

    let text: string
    if (field.format !== undefined) {
      text = field.format(raw)
    } else if (typeof value === 'boolean') {
      text = value ? options.yes : options.no
    } else {
      text = String(value)
    }
    if (missing && field.format === undefined) {
      text = empty
    }

    let level = field.level
    if (level === undefined && field.booleanLevel !== undefined && typeof value === 'boolean') {
      level = value ? field.booleanLevel.true : field.booleanLevel.false
    }
    if (level === undefined && field.levelBy !== undefined) {
      level = field.levelBy(raw)
    }

    const item: StatGridItem = { label: field.label, value: text }
    if (level !== undefined) {
      item.level = level
    }
    if (field.mono === false) {
      item.mono = false
    }
    if (field.wide === true) {
      item.wide = true
    }
    items.push(item)
  }

  return items
}

/**
 * 把字符串数组转成单列表格行。
 *
 * 多记录结果（UUID 列表、ID 列表、防火墙规则、密码列表）都是「一列一串」的形态，
 * 共用这一份转换。
 *
 * @param values  记录数组
 * @param key     列键（同时用作行键前缀）
 */
export function listRows(
  values: readonly string[] | null | undefined,
  key: string,
): { key: string; cells: Record<string, string> }[] {
  if (!values) {
    return []
  }
  return values.map((value, index) => ({
    key: `${key}-${index}`,
    cells: { [key]: value },
  }))
}
