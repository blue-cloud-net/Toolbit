/**
 * hash-calc 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface HashCalcExample {
  /** 示例名称。 */
  label: string
  /** 要计算的摘要算法。 */
  algorithms: readonly string[]
  /** 输入内容。 */
  input: string
  /** 输入编码。 */
  inputEncoding: 'utf8' | 'hex'
}

/** 默认示例：对一段中文计算 SM3 与 SHA-256。 */
export const DEFAULT_EXAMPLE: HashCalcExample = {
  label: '计算中文文本的 SM3 与 SHA-256',
  algorithms: ['SM3', 'SHA256'],
  input: '蓝云工具箱 Toolbit',
  inputEncoding: 'utf8',
}

/** 二进制示例。 */
export const HEX_EXAMPLE: HashCalcExample = {
  label: '计算十六进制内容的摘要',
  algorithms: ['SM3', 'SHA256'],
  input: '616263',
  inputEncoding: 'hex',
}

/** 全部示例。 */
export const EXAMPLES: readonly HashCalcExample[] = [DEFAULT_EXAMPLE, HEX_EXAMPLE]
