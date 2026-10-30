/**
 * 复制到剪贴板。
 *
 * 优先 `navigator.clipboard`，不可用时回退到隐藏 textarea + `execCommand`。
 *
 * 结果统一用 `ElMessage` 轻提示（「已复制」/「复制失败」，1.8 秒自动消失）：
 * 按钮自身的 tooltip 因此可以固定为「复制」，不会因为状态切换而闪动文案。
 */
import { ElMessage } from 'element-plus/es/components/message/index'
import 'element-plus/es/components/message/style/css'
import { useI18n } from 'vue-i18n'

/** 回退复制实现（老浏览器 / 非安全上下文）。 */
function fallbackCopy(text: string): boolean {
  if (typeof document === 'undefined') {
    return false
  }
  const area = document.createElement('textarea')
  area.value = text
  area.setAttribute('readonly', '')
  area.style.position = 'fixed'
  area.style.top = '-1000px'
  document.body.appendChild(area)
  area.select()
  let ok = false
  try {
    ok = document.execCommand('copy')
  } catch {
    ok = false
  }
  document.body.removeChild(area)
  return ok
}

/** 复制能力。 */
export interface CopyHandle {
  /** 复制文本；空串直接忽略。返回是否成功。 */
  copy: (text: string) => Promise<boolean>
}

/** 取得复制能力。 */
export function useCopy(): CopyHandle {
  const { t } = useI18n()

  async function copy(text: string): Promise<boolean> {
    if (text === '') {
      return false
    }
    let ok = false
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
        ok = true
      } else {
        ok = fallbackCopy(text)
      }
    } catch {
      ok = fallbackCopy(text)
    }

    ElMessage({
      message: ok ? t('common.actions.copied') : t('tool.copy.failed'),
      type: ok ? 'success' : 'error',
      duration: 1800,
      // 连续复制多行时合并成一条，避免消息堆成瀑布。
      grouping: true,
      showClose: false,
    })

    return ok
  }

  return { copy }
}
