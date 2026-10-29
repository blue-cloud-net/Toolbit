/**
 * websocket-test 工具的示例数据。
 */
export interface WebSocketExample {
  /** 示例名称。 */
  label: string
  /** 目标地址。 */
  url: string
  /** 子协议。 */
  protocols: string
  /** 发送内容。 */
  message: string
}

/** 默认示例：公共回显服务。 */
export const DEFAULT_EXAMPLE: WebSocketExample = {
  label: '公共回显服务',
  url: 'wss://echo.websocket.org',
  protocols: '',
  message: 'hello toolbit',
}

/** 全部示例。 */
export const EXAMPLES: readonly WebSocketExample[] = [DEFAULT_EXAMPLE]
