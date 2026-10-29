/**
 * 预设主色板。
 *
 * 色值与 `frontend/src/styles/tokens.css` 中的 `[data-accent]` 规则**必须一致**；
 * 两者由 `scripts/check-accent.mjs` 校验（对比度 + 与 tokens.css 的一致性）。
 *
 * 这里的色值仅用于**色板选择器的色块预览**，
 * 实际生效的样式始终来自 CSS 变量。
 */

/** 预设主色标识。 */
export type AccentKey =
  | 'indigo'
  | 'blue'
  | 'gitlab-blue'
  | 'sky'
  | 'teal'
  | 'green'
  | 'lime'
  | 'amber'
  | 'orange'
  | 'rose'
  | 'violet'
  | 'slate'

/** 主色定义：标识、预览色值（浅色模式下的 base 色）。 */
export interface AccentPreset {
  /** 标识，同时是 `data-accent` 的值。 */
  key: AccentKey
  /** 预览色值（浅色模式 base 色）。 */
  swatch: string
}

/** 全部预设主色，顺序即选择器展示顺序。 */
export const ACCENT_PRESETS: readonly AccentPreset[] = [
  { key: 'indigo', swatch: '#4f46e5' },
  { key: 'blue', swatch: '#2563eb' },
  { key: 'gitlab-blue', swatch: '#1f75cb' },
  { key: 'sky', swatch: '#0369a1' },
  { key: 'teal', swatch: '#0f766e' },
  { key: 'green', swatch: '#15803d' },
  { key: 'lime', swatch: '#4d7c0f' },
  { key: 'amber', swatch: '#b45309' },
  { key: 'orange', swatch: '#c2410c' },
  { key: 'rose', swatch: '#e11d48' },
  { key: 'violet', swatch: '#7c3aed' },
  { key: 'slate', swatch: '#475569' },
]
