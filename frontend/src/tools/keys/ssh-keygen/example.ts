/**
 * ssh-keygen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface SSHKeygenExample {
  /** 示例名称。 */
  label: string
  /** algorithm。 */
  algorithm: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: SSHKeygenExample = {
  label: 'ssh-keygen 示例',
  algorithm: 'Ed25519',
}

/** 全部示例。 */
export const EXAMPLES: readonly SSHKeygenExample[] = [DEFAULT_EXAMPLE]
