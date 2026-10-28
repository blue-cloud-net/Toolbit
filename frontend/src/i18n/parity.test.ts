/**
 * 词条一致性测试。
 *
 * `zh-CN` 与 `en-US` 的 key 集合**必须**完全一致；
 * 工具词条**必须**覆盖 i18n.md §4 规定的十个字段。
 */
import { describe, expect, it } from 'vitest'

import enUSCommon from '@/locales/en-US/common.json'
import enUSErrors from '@/locales/en-US/errors.json'
import enUSHome from '@/locales/en-US/home.json'
import enUSTool from '@/locales/en-US/tool.json'
import enUSBase64 from '@/locales/en-US/tools/base64.json'
import enUSNumberBase from '@/locales/en-US/tools/number-base.json'
import zhCNCommon from '@/locales/zh-CN/common.json'
import zhCNErrors from '@/locales/zh-CN/errors.json'
import zhCNHome from '@/locales/zh-CN/home.json'
import zhCNTool from '@/locales/zh-CN/tool.json'
import zhCNBase64 from '@/locales/zh-CN/tools/base64.json'
import zhCNNumberBase from '@/locales/zh-CN/tools/number-base.json'

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
  { name: 'tools.base64', zh: zhCNBase64, en: enUSBase64 },
  { name: 'tools.number-base', zh: zhCNNumberBase, en: enUSNumberBase },
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

describe('工具词条字段', () => {
  it.each([
    ['base64', zhCNBase64, enUSBase64],
    ['number-base', zhCNNumberBase, enUSNumberBase],
  ])('%s 覆盖十个必备字段', (_slug, zh, en) => {
    for (const field of TOOL_FIELDS) {
      expect(Object.keys(zh as Record<string, unknown>)).toContain(field)
      expect(Object.keys(en as Record<string, unknown>)).toContain(field)
    }
  })

  it.each([
    ['base64', zhCNBase64, enUSBase64],
    ['number-base', zhCNNumberBase, enUSNumberBase],
  ])('%s 的示例与 FAQ 非空', (_slug, zh, en) => {
    for (const messages of [zh, en] as { examples: unknown[]; faq: unknown[] }[]) {
      expect(messages.examples.length).toBeGreaterThan(0)
      expect(messages.faq.length).toBeGreaterThan(0)
    }
  })
})

describe('错误码词条', () => {
  it('覆盖全部错误码与警告码', () => {
    // 与 backend/internal/toolerr 的注册表条目数量保持一致（67 + 14）。
    expect(Object.keys(zhCNErrors)).toHaveLength(81)
    expect(Object.keys(enUSErrors)).toHaveLength(81)
  })
})
