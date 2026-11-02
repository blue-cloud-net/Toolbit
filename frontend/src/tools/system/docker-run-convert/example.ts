/**
 * docker-run-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 转换方向。 */
export type ConvertDirection = 'runToCompose' | 'composeToRun'

/** 一个示例。 */
export interface DockerRunConvertExample {
  /** 示例名称。 */
  label: string
  /** 转换方向。 */
  direction: ConvertDirection
  /** 输入文本。 */
  input: string
}

/** 默认示例：docker run 命令转 Compose。 */
export const DEFAULT_EXAMPLE: DockerRunConvertExample = {
  label: 'docker run 转 Compose',
  direction: 'runToCompose',
  input: 'docker run -d --name web -p 8080:80 -e TZ=Asia/Shanghai nginx:alpine',
}

/** 全部示例。 */
export const EXAMPLES: readonly DockerRunConvertExample[] = [DEFAULT_EXAMPLE]
