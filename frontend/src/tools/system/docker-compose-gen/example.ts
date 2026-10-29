/**
 * docker-compose-gen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface DockerComposeExample {
  /** 示例名称。 */
  label: string
  /** projectName。 */
  projectName: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: DockerComposeExample = {
  label: 'docker-compose-gen 示例',
  projectName: 'demo',
}

/** 全部示例。 */
export const EXAMPLES: readonly DockerComposeExample[] = [DEFAULT_EXAMPLE]
