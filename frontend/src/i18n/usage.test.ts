/**
 * 词条引用完整性测试。
 *
 * 工具页里 `t('tools.<slug>.<path>')` 的**字面量**引用，
 * 必须在对应语言包中真实存在——否则界面会直接显示原始键名
 * （如 `tools.keygen.views.summary`），而不是译文。
 *
 * 词条目录用 `import.meta.glob` 静态收集，
 * 工具页源码用 `import.meta.glob` 静态读取，
 * 新增工具时无需修改本文件。
 */
import { describe, expect, it } from 'vitest'

/** 工具页源码。 */
const TOOL_SOURCES = import.meta.glob<string>('/src/tools/*/*/Tool.vue', {
  eager: true,
  query: '?raw',
  import: 'default',
})

/** 工具词条：按语言分组的 slug → 词条。 */
const TOOL_MESSAGES = import.meta.glob<Record<string, unknown>>('/src/locales/*/tools/*.json', {
  eager: true,
})

/** 字面量键引用：`t('tools.<slug>.<path>')`。 */
const LITERAL_KEY = /t\(\s*'tools\.([a-z0-9-]+)\.([A-Za-z0-9_.-]+)'\s*\)/g

/** 模板字面量键引用：`t('tools.<slug>.' + name)` 这类无法静态求值，跳过。 */
const DYNAMIC_KEY = /t\(\s*`tools\./

/** 读取词条中的叶子路径集合。 */
function leafPaths(value: unknown, prefix = ''): Set<string> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return new Set([prefix])
  }
  const paths = new Set<string>()
  for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
    for (const path of leafPaths(child, prefix === '' ? key : `${prefix}.${key}`)) {
      paths.add(path)
    }
  }
  return paths
}

/** 收集工具页里所有字面量键引用：slug + 路径。 */
function literalUsages(): Array<{ slug: string; path: string; file: string }> {
  const usages: Array<{ slug: string; path: string; file: string }> = []
  for (const [file, source] of Object.entries(TOOL_SOURCES)) {
    if (DYNAMIC_KEY.test(source)) {
      // 含模板字面量键（依赖运行时变量），静态校验不适用。
      continue
    }
    for (const match of source.matchAll(LITERAL_KEY)) {
      const slug = match[1]
      const path = match[2]
      if (slug !== undefined && path !== undefined) {
        usages.push({ slug, path, file })
      }
    }
  }
  return usages
}

describe('工具词条引用', () => {
  const usages = literalUsages()

  it('工具页存在字面量键引用', () => {
    expect(usages.length).toBeGreaterThan(500)
  })

  it.each(
    (() => {
      const seen = new Map<string, { slug: string; path: string; file: string }>()
      for (const usage of usages) {
        seen.set(`${usage.slug}.${usage.path}`, usage)
      }
      return [...seen.entries()].map(([key, usage]) => ({ key, ...usage }))
    })(),
  )('$key 在语言包中存在', ({ slug, path }) => {
    const messages = TOOL_MESSAGES[`/src/locales/zh-CN/tools/${slug}.json`]
    expect(messages, `工具 ${slug} 缺少 zh-CN 词条`).toBeDefined()
    const paths = leafPaths(messages)
    expect(
      paths.has(path),
      `zh-CN/tools/${slug}.json 缺少键 "${path}"，界面会显示原始键名`,
    ).toBe(true)
  })

  it('每个工具都有对应词条', () => {
    const slugs = new Set(usages.map((usage) => usage.slug))
    for (const slug of slugs) {
      expect(
        TOOL_MESSAGES[`/src/locales/zh-CN/tools/${slug}.json`],
        `工具 ${slug} 缺少 zh-CN 词条`,
      ).toBeDefined()
    }
  })
})
