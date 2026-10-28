import presetUno from '@unocss/preset-uno'
import { defineConfig } from 'unocss'

/**
 * UnoCSS 配置。
 *
 * 颜色一律引用 `--tb-*` 变量，不在 preset 里写死色值，
 * 主题切换由 CSS 变量驱动。
 */
export default defineConfig({
  presets: [presetUno()],
  theme: {
    colors: {
      primary: 'var(--tb-primary)',
      'primary-soft': 'var(--tb-primary-soft)',
      bg: 'var(--tb-bg)',
      'bg-elevated': 'var(--tb-bg-elevated)',
      'bg-sunken': 'var(--tb-bg-sunken)',
      border: 'var(--tb-border)',
      'border-strong': 'var(--tb-border-strong)',
      text: 'var(--tb-text)',
      'text-secondary': 'var(--tb-text-secondary)',
      'text-tertiary': 'var(--tb-text-tertiary)',
      success: 'var(--tb-success)',
      warning: 'var(--tb-warning)',
      danger: 'var(--tb-danger)',
      info: 'var(--tb-info)',
    },
    fontSize: {
      xs: '12px',
      sm: '13px',
      base: '14px',
      md: '16px',
      lg: '18px',
      xl: '20px',
      '2xl': '24px',
      '3xl': '30px',
    },
    borderRadius: {
      control: '6px',
      card: '10px',
      panel: '14px',
      pill: '999px',
    },
  },
  shortcuts: {
    'tb-card': 'rounded-card bg-bg-elevated border border-border',
    'tb-focus': 'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary',
    'tb-mono': 'font-mono text-sm',
  },
})
