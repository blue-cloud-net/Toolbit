/**
 * 材料文件导入的**纯函数**。
 *
 * 与 `MaterialInput.vue` 分开：文件读取（File API）留在组件里，
 * **判定**逻辑在这里，便于单测覆盖。
 *
 * 判定顺序与后端 `PemBundle` 的自动识别一致（PEM → base64 → hex），
 * 但二进制 DER 走「base64 承载」：十六进制会让请求体翻倍，
 * 更容易撞上请求体上限。
 */

/** 嗅探出的格式。 */
export type MaterialFileFormat = 'pem' | 'der' | 'base64' | 'hex'

/** 嗅探结果。 */
export interface DetectedMaterialFile {
  /** 嗅探出的格式。 */
  detected: MaterialFileFormat
  /** 可提交给后端的文本。 */
  content: string
}

/** 判定文本是否像十六进制串。 */
export function looksLikeHex(text: string): boolean {
  const clean = text.replace(/\s+/g, '')
  return clean.length >= 16 && clean.length % 2 === 0 && /^[0-9a-fA-F]+$/.test(clean)
}

/** 判定文本是否像 base64 串。 */
export function looksLikeBase64(text: string): boolean {
  const clean = text.replace(/\s+/g, '')
  return clean.length >= 16 && clean.length % 4 === 0 && /^[A-Za-z0-9+/]+={0,2}$/.test(clean)
}

/** 字节转 base64（分块处理，避免大文件爆栈）。 */
export function bytesToBase64(bytes: Uint8Array): string {
  let binary = ''
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
  }
  return btoa(binary)
}

/**
 * 嗅探文件内容。
 *
 * `bytes` 是文件的原始字节，`text` 是同一份内容的 UTF-8 文本视图
 * （二进制文件里会出现替换字符，属正常现象）。
 *
 * 判定顺序：
 *
 *  1. **PEM 文本** —— 含 `-----BEGIN `；
 *  2. **十六进制文本** —— 字符集是 base64 的子集，先判它才不会误标；
 *  3. **base64 文本**；
 *  4. **DER 二进制** —— 首字节 0x30（ASN.1 SEQUENCE）。
 *
 * 第 4 步必须排在文本判定**之后**：ASCII 字符 `'0'` 的字节值正好是 0x30，
 * 先判 DER 会把「以 0 开头的十六进制串」误认成二进制证书。
 *
 * 返回 null 表示「不像 PEM 族材料」：调用方应提示用户。
 */
export function detectMaterialFile(bytes: Uint8Array, text: string): DetectedMaterialFile | null {
  // PEM 文本：直接用原文。
  if (text.includes('-----BEGIN ')) {
    return { detected: 'pem', content: text }
  }
  // 十六进制 / base64 文本：原样保留，只标格式。
  if (looksLikeHex(text)) {
    return { detected: 'hex', content: text }
  }
  if (looksLikeBase64(text)) {
    return { detected: 'base64', content: text }
  }
  // DER 二进制：首字节 0x30（ASN.1 SEQUENCE）。
  if (bytes.length >= 2 && bytes[0] === 0x30) {
    return { detected: 'der', content: bytesToBase64(bytes) }
  }
  return null
}

/** 把字节数组按 UTF-8 解成文本（非法字节用替换字符，不抛错）。 */
export function decodeUtf8(bytes: Uint8Array): string {
  return new TextDecoder('utf-8', { fatal: false }).decode(bytes)
}
