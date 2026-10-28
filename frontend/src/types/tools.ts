/**
 * 与后端 catalog DTO 对齐的类型。
 *
 * 字段名必须与 `backend/internal/api/catalog_handler.go` 逐字一致。
 */

/** 工具目录响应。 */
export interface CatalogResponse {
  version: number
  categories: CategoryMeta[]
  tools: ToolMeta[]
}

/** 分类展示元信息。 */
export interface CategoryMeta {
  slug: string
  name: string
  nameEn: string
  order: number
  icon: string
}

/** 工具展示元信息。 */
export interface ToolMeta {
  slug: string
  category: string
  order: number
  name: string
  nameEn: string
  enabled: boolean
  transport: 'json' | 'multipart' | 'websocket' | 'client'
  rateClass: 'light' | 'standard' | 'heavy' | 'file' | 'network'
  maxBodyBytes: number
  timeoutMs: number
  sensitive: boolean
  requiresNetwork: boolean
  autoRun: boolean
  wide: boolean
  i18nKey: string
}

/** 传输方式。 */
export type Transport = ToolMeta['transport']

/** 限流档位。 */
export type RateClass = ToolMeta['rateClass']
