/**
 * Base64 → 字节数组。
 *
 * 二进制产物（PNG 等）以 Base64 文本在接口里传递，下载时需要还原成字节。
 * 这里按 4 字符一块分块解码，避免整串一次性 `atob` 造成的内存峰值。
 *
 * 纯函数，无副作用，便于单测。
 */

/** 单块解码的字符数；必须是 4 的倍数（Base64 以 4 字符为一组）。 */
const CHUNK_CHARS = 4 * 8192

/**
 * 把 Base64 文本解成字节数组。
 *
 * 忽略空白字符（换行等）。内容非法时抛错，由调用方决定如何提示。
 */
export function base64ToBytes(base64: string): Uint8Array {
  const compact = base64.replace(/\s+/g, '')
  if (compact === '') {
    return new Uint8Array(0)
  }

  const parts: Uint8Array[] = []
  for (let offset = 0; offset < compact.length; offset += CHUNK_CHARS) {
    const binary = atob(compact.slice(offset, offset + CHUNK_CHARS))
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i)
    }
    parts.push(bytes)
  }

  const total = parts.reduce((sum, part) => sum + part.length, 0)
  const out = new Uint8Array(total)
  let cursor = 0
  for (const part of parts) {
    out.set(part, cursor)
    cursor += part.length
  }
  return out
}
