/**
 * i18n 装配。
 *
 * 词条按 [docs/i18n.md](docs/i18n.md) §2 的目录结构落地：
 * `common.json` / `home.json` / `tool.json` / `errors.json` / `tools/<slug>.json`。
 */
import { createI18n } from 'vue-i18n'

import enUS from '@/locales/en-US/common.json'
import enUSErrors from '@/locales/en-US/errors.json'
import enUSHome from '@/locales/en-US/home.json'
import enUSTool from '@/locales/en-US/tool.json'
import zhCN from '@/locales/zh-CN/common.json'
import zhCNErrors from '@/locales/zh-CN/errors.json'
import zhCNHome from '@/locales/zh-CN/home.json'
import zhCNTool from '@/locales/zh-CN/tool.json'

/** 支持的语言。 */
export const SUPPORTED_LOCALES = ['zh-CN', 'en-US'] as const

/** 界面语言类型。 */
export type Locale = (typeof SUPPORTED_LOCALES)[number]

/** 默认语言。 */
export const DEFAULT_LOCALE: Locale = 'zh-CN'

/** 语言在 localStorage 中的键。 */
export const LOCALE_STORAGE_KEY = 'toolbit.locale'

/** 读取存储的语言；无效值回退到默认语言。 */
export function readStoredLocale(): Locale {
  if (typeof localStorage === 'undefined') {
    return DEFAULT_LOCALE
  }
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  return isLocale(stored) ? stored : DEFAULT_LOCALE
}

/** 判断是否为受支持的语言。 */
export function isLocale(value: unknown): value is Locale {
  return typeof value === 'string' && (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

const messages = {
  'zh-CN': {
    common: zhCN,
    home: zhCNHome,
    tool: zhCNTool,
    errors: zhCNErrors,
  },
  'en-US': {
    common: enUS,
    home: enUSHome,
    tool: enUSTool,
    errors: enUSErrors,
  },
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: DEFAULT_LOCALE,
  fallbackLocale: 'en-US',
  missingWarn: false,
  fallbackWarn: false,
  messages,
})

/**
 * 登记单个工具的词条。
 *
 * 工具词条是运行期按需加载的（配合路由懒加载）。
 */
export function registerToolMessages(locale: Locale, slug: string, messages_: Record<string, unknown>): void {
  const target = i18n.global.messages.value[locale] as Record<string, Record<string, unknown>>
  if (!target.tools) {
    target.tools = {}
  }
  target.tools[slug] = messages_
}
