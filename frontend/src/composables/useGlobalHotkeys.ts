/**
 * 全局快捷键。
 *
 * 规范 §12 约定的三个全局键：
 *
 *   - `/`     聚焦搜索（首页搜索框；命令面板已打开时聚焦其输入框）；
 *   - `?`     打开快捷键说明；
 *   - `Alt+↑/↓` 在「最近使用」列表中上下切换工具。
 *
 * `Ctrl/⌘ + K` 由 `CommandPalette` 独占注册，`Esc` 由各浮层自行处理，
 * 这里不重复实现，避免一次按键被处理两次。
 */
import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

/** 快捷键说明的打开状态（模块级单例，供说明面板读取）。 */
const shortcutHelpOpen = ref(false)

/** 回调注册表：让不同页面按需挂接行为。 */
const handlers = {
  /** 聚焦搜索：返回被聚焦的元素，或 null 表示当前无可聚焦目标。 */
  focusSearch: null as (() => HTMLElement | null) | null,
  /** 最近使用列表的上下移动。 */
  moveRecent: null as ((delta: number) => void) | null,
}

export interface GlobalHotkeyOptions {
  /** 聚焦搜索的实现；返回 null 表示当前页面没有搜索框。 */
  focusSearch?: () => HTMLElement | null
  /** 最近使用列表的上下移动。 */
  moveRecent?: (delta: number) => void
}

/** 注册全局快捷键；组件卸载时自动解绑。 */
export function useGlobalHotkeys(options: GlobalHotkeyOptions = {}): void {
  function onKeydown(event: KeyboardEvent): void {
    const target = event.target
    const inEditable =
      target instanceof HTMLElement &&
      (target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable)

    // `/` 聚焦搜索：输入控件内不拦截。
    if (event.key === '/' && !inEditable && !event.ctrlKey && !event.metaKey && !event.altKey) {
      const el = options.focusSearch?.() ?? handlers.focusSearch?.()
      if (el) {
        event.preventDefault()
        el.focus()
        if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
          el.select()
        }
      }
      return
    }

    // `?` 打开快捷键说明（需 Shift）。
    if (event.key === '?' && !inEditable && !event.ctrlKey && !event.metaKey && !event.altKey) {
      event.preventDefault()
      shortcutHelpOpen.value = !shortcutHelpOpen.value
      return
    }

    // Alt + ↑/↓ 在最近使用中切换。
    if (event.altKey && (event.key === 'ArrowUp' || event.key === 'ArrowDown')) {
      const delta = event.key === 'ArrowUp' ? -1 : 1
      const mover = options.moveRecent ?? handlers.moveRecent
      if (mover) {
        event.preventDefault()
        mover(delta)
      }
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', onKeydown)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', onKeydown)
  })
}

/** 快捷键说明面板是否打开。 */
export function useShortcutHelp(): { open: Ref<boolean> } {
  return { open: shortcutHelpOpen }
}

/** 注册全局行为（供跨页面共享）。 */
export function registerGlobalHotkeyHandlers(next: {
  focusSearch?: () => HTMLElement | null
  moveRecent?: (delta: number) => void
}): void {
  if (next.focusSearch) {
    handlers.focusSearch = next.focusSearch
  }
  if (next.moveRecent) {
    handlers.moveRecent = next.moveRecent
  }
}
