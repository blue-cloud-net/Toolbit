/**
 * 工具组件静态登记表。
 *
 * slug → 组件是**编译期**映射：目录数据只提供展示元信息，
 * 不在运行期拼装组件。
 *
 * 组件用动态 `import()` 加载，Vite 会按此拆出独立的 chunk。
 */
import type { Component } from 'vue'

/** 工具组件加载器。 */
export type ToolLoader = () => Promise<{ default: Component }>

/** 工具登记表。 */
export const toolRegistry: Record<string, ToolLoader> = {
  'base64': () => import('./encoding/base64/index'),
  'number-base': () => import('./encoding/number-base/index'),
  'hash-calc': () => import('./crypto/hash-calc/index'),
  'hmac-gen': () => import('./crypto/hmac-gen/index'),
  'password-gen': () => import('./crypto/password-gen/index'),
  'symmetric-crypt': () => import('./crypto/symmetric-crypt/index'),
  'signature': () => import('./crypto/signature/index'),
  'asymmetric-crypt': () => import('./crypto/asymmetric-crypt/index'),
  'ecdh': () => import('./crypto/ecdh/index'),
}

/** 已实现的工具 slug 集合。 */
export const IMPLEMENTED_SLUGS: readonly string[] = Object.keys(toolRegistry)

/** 判断工具是否已实现。 */
export function isImplemented(slug: string): boolean {
  return Object.prototype.hasOwnProperty.call(toolRegistry, slug)
}

/** 取工具组件加载器；未实现时返回 undefined。 */
export function getToolLoader(slug: string): ToolLoader | undefined {
  return toolRegistry[slug]
}
