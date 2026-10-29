/**
 * 分类图标映射。
 *
 * catalog 的分类 `icon` 字段是字符串（如 `lock`、`code`），
 * 这里把它映射到 `AppIcon` 的图标名。侧栏与工具卡片共用此表，
 * 避免两处各写一份而漂移。
 */
import type { IconName } from '@/components/common/AppIcon.vue'

/** 分类 icon 字段 → 图标名。 */
export const CATEGORY_ICONS: Record<string, IconName> = {
  lock: 'lock',
  key: 'key',
  certificate: 'certificate',
  code: 'code',
  indent: 'indent',
  text: 'text',
  data: 'data',
  monitor: 'monitor',
  connection: 'connection',
  document: 'document',
}

/** 取分类图标名；未知值回落到 `document`。 */
export function categoryIcon(icon: string | undefined): IconName {
  if (!icon) {
    return 'document'
  }
  return CATEGORY_ICONS[icon] ?? 'document'
}
