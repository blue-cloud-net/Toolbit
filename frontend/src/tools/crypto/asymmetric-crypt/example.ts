/**
 * asymmetric-crypt 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface AsymmetricCryptExample {
  /** 示例名称。 */
  label: string
  /** 算法。 */
  algorithm: 'SM2' | 'RSA'
  /** 操作。 */
  operation: 'encrypt' | 'decrypt'
  /** RSA 填充。 */
  rsaPadding: 'oaep' | 'pkcs1'
  /** OAEP 摘要。 */
  oaepHash: 'SHA256' | 'SHA384' | 'SHA512'
  /** 输入内容。 */
  input: string
  /** 输入编码。 */
  inputEncoding: 'utf8' | 'hex'
}

/** 默认示例：SM2 加密。 */
export const DEFAULT_EXAMPLE: AsymmetricCryptExample = {
  label: 'SM2 加密一段中文',
  algorithm: 'SM2',
  operation: 'encrypt',
  rsaPadding: 'oaep',
  oaepHash: 'SHA256',
  input: '待加密内容',
  inputEncoding: 'utf8',
}

/** RSA 示例。 */
export const RSA_EXAMPLE: AsymmetricCryptExample = {
  label: 'RSA-OAEP 加密一段中文',
  algorithm: 'RSA',
  operation: 'encrypt',
  rsaPadding: 'oaep',
  oaepHash: 'SHA256',
  input: '待加密内容',
  inputEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly AsymmetricCryptExample[] = [DEFAULT_EXAMPLE, RSA_EXAMPLE]
