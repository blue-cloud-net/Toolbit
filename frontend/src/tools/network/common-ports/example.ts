/**
 * common-ports 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface CommonPortExample {
  /** 示例名称。 */
  label: string
  /** 查询词：端口号或关键字。 */
  query: string
}

/** 默认示例：查 HTTPS 端口。 */
export const DEFAULT_EXAMPLE: CommonPortExample = {
  label: 'HTTPS 443',
  query: '443',
}

/** 服务名示例。 */
export const SERVICE_EXAMPLE: CommonPortExample = {
  label: '按服务名查 SSH',
  query: 'ssh',
}

/** 全部示例。 */
export const EXAMPLES: readonly CommonPortExample[] = [DEFAULT_EXAMPLE, SERVICE_EXAMPLE]
