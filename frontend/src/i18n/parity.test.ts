/**
 * 词条一致性测试。
 *
 * `zh-CN` 与 `en-US` 的 key 集合**必须**完全一致；
 * 工具词条**必须**覆盖 i18n.md §4 规定的十个字段。
 *
 * 工具词条目录用 `import.meta.glob` 静态收集，
 * 新增工具时无需修改本文件。
 */
import { describe, expect, it } from 'vitest'

import enUSCommon from '@/locales/en-US/common.json'
import enUSErrors from '@/locales/en-US/errors.json'
import enUSHome from '@/locales/en-US/home.json'
import enUSTool from '@/locales/en-US/tool.json'
import zhCNCommon from '@/locales/zh-CN/common.json'
import zhCNErrors from '@/locales/zh-CN/errors.json'
import zhCNHome from '@/locales/zh-CN/home.json'
import zhCNTool from '@/locales/zh-CN/tool.json'

/** 工具词条：按语言分组的 slug → 词条。 */
const TOOL_MESSAGES = import.meta.glob<Record<string, unknown>>('/src/locales/*/tools/*.json', {
  eager: true,
})

/** 收集某个 slug 在两种语言下的词条。 */
function toolMessages(slug: string): { zh: unknown; en: unknown } {
  const zh = TOOL_MESSAGES[`/src/locales/zh-CN/tools/${slug}.json`]
  const en = TOOL_MESSAGES[`/src/locales/en-US/tools/${slug}.json`]
  if (!zh || !en) {
    throw new Error(`工具 ${slug} 的词条缺失（zh 或 en）`)
  }
  return { zh, en }
}

/** 已实现的工具 slug（由词条目录推导）。 */
const TOOL_SLUGS = [...new Set(
  Object.keys(TOOL_MESSAGES)
    .map((path) => path.match(/^\/src\/locales\/zh-CN\/tools\/(.+)\.json$/)?.[1])
    .filter((slug): slug is string => Boolean(slug)),
)].sort()

/** 递归收集叶子 key 路径。 */
function leafKeys(value: unknown, prefix = ''): string[] {
  if (value === null || typeof value !== 'object') {
    return [prefix]
  }
  if (Array.isArray(value)) {
    return [prefix]
  }
  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) =>
    leafKeys(child, prefix === '' ? key : `${prefix}.${key}`),
  )
}

/** 命名空间对照表。 */
const NAMESPACES: { name: string; zh: unknown; en: unknown }[] = [
  { name: 'common', zh: zhCNCommon, en: enUSCommon },
  { name: 'home', zh: zhCNHome, en: enUSHome },
  { name: 'tool', zh: zhCNTool, en: enUSTool },
  { name: 'errors', zh: zhCNErrors, en: enUSErrors },
]

/** 工具词条必备字段。 */
const TOOL_FIELDS = [
  'name',
  'description',
  'category',
  'actions',
  'fields',
  'options',
  'results',
  'examples',
  'hints',
  'faq',
]

describe.each(NAMESPACES)('$name', ({ zh, en }) => {
  it('双语 key 集合一致', () => {
    expect(leafKeys(zh).sort()).toEqual(leafKeys(en).sort())
  })
})

describe('工具词条', () => {
  it('zh-CN 与 en-US 的工具 slug 集合一致', () => {
    const zhSlugs = Object.keys(TOOL_MESSAGES)
      .filter((path) => path.includes('/zh-CN/'))
      .map((path) => path.split('/').pop()!.replace(/\.json$/, ''))
      .sort()
    const enSlugs = Object.keys(TOOL_MESSAGES)
      .filter((path) => path.includes('/en-US/'))
      .map((path) => path.split('/').pop()!.replace(/\.json$/, ''))
      .sort()
    expect(zhSlugs).toEqual(enSlugs)
  })

  it.each(TOOL_SLUGS)('%s 覆盖十个必备字段', (slug) => {
    const { zh, en } = toolMessages(slug)
    for (const field of TOOL_FIELDS) {
      expect(Object.keys(zh as Record<string, unknown>)).toContain(field)
      expect(Object.keys(en as Record<string, unknown>)).toContain(field)
    }
  })

  it.each(TOOL_SLUGS)('%s 的示例与 FAQ 非空', (slug) => {
    const { zh, en } = toolMessages(slug)
    for (const messages of [zh, en] as { examples: unknown[]; faq: unknown[] }[]) {
      expect(messages.examples.length).toBeGreaterThan(0)
      expect(messages.faq.length).toBeGreaterThan(0)
    }
  })

  it.each(TOOL_SLUGS)('%s 的双语 key 集合一致', (slug) => {
    const { zh, en } = toolMessages(slug)
    expect(leafKeys(zh).sort()).toEqual(leafKeys(en).sort())
  })
})

describe('错误码词条', () => {
  it('覆盖全部错误码与警告码', () => {
    // 与 backend/internal/toolerr 的注册表条目数量保持一致（67 + 14）。
    expect(Object.keys(zhCNErrors)).toHaveLength(82)
    expect(Object.keys(enUSErrors)).toHaveLength(82)
  })
})
