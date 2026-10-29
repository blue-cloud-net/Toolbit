/**
 * 工具一句话描述（按需加载）。
 *
 * 折叠态侧栏的悬浮子菜单需要在**未进入**工具页时展示描述，
 * 而工具词条是随工具 chunk 按需登记的（见 `i18n/index.ts`），此时还拿不到。
 * 这里用 `import.meta.glob` 为每个词条文件建一条轻量加载器：
 * 只在悬浮某个分类时加载该分类的工具词条，且结果按 locale + slug 缓存。
 *
 * **不复制文案**：读的仍是 `locales/<lang>/tools/<slug>.json`，
 * 单一事实来源不变。
 */
import { reactive } from 'vue'

import { DEFAULT_LOCALE, isLocale, type Locale } from '@/i18n'

/** 词条文件中被用到的字段。 */
interface ToolMessages {
  description?: string
}

/** 词条加载器表：locale → `词条路径 → 加载器`。 */
const LOADERS: Record<Locale, Record<string, () => Promise<{ default: ToolMessages }>>> = {
  'zh-CN': import.meta.glob<{ default: ToolMessages }>('../locales/zh-CN/tools/*.json'),
  'en-US': import.meta.glob<{ default: ToolMessages }>('../locales/en-US/tools/*.json'),
}

/** 已加载的描述（响应式，命中后菜单立刻刷新）。 */
const loaded = reactive(new Map<string, string>())

/** 正在加载的键，避免重复请求。 */
const pending = new Set<string>()

/** 词条文件路径 → slug。 */
function slugOf(key: string): string {
  const name = key.slice(key.lastIndexOf('/') + 1)
  return name.endsWith('.json') ? name.slice(0, -5) : name
}

/** 语言不可用时回退到默认语言（vue-i18n 的 locale 是 string）。 */
function resolveLocale(locale: string): Locale {
  return isLocale(locale) ? locale : DEFAULT_LOCALE
}

/** 取已加载的描述；未加载时返回空串（不阻塞渲染）。 */
export function toolDescription(locale: string, slug: string): string {
  return loaded.get(`${locale}:${slug}`) ?? ''
}

/** 确保给定工具的描述已加载；失败静默忽略（描述缺失不影响导航）。 */
export async function ensureToolDescriptions(
  locale: string,
  slugs: readonly string[],
): Promise<void> {
  const loaders = LOADERS[resolveLocale(locale)]
  const tasks: Promise<void>[] = []

  for (const [key, load] of Object.entries(loaders)) {
    if (!slugs.includes(slugOf(key))) {
      continue
    }
    const cacheKey = `${locale}:${slugOf(key)}`
    if (loaded.has(cacheKey) || pending.has(cacheKey)) {
      continue
    }
    pending.add(cacheKey)
    tasks.push(
      load()
        .then((mod) => {
          loaded.set(cacheKey, mod.default.description ?? '')
        })
        .catch(() => {
          // 词条缺失或加载失败时保持空串。
        })
        .finally(() => {
          pending.delete(cacheKey)
        }),
    )
  }

  await Promise.all(tasks)
}
