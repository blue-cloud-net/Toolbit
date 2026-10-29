/**
 * json-to-types 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface JSONToTypesExample {
  /** 示例名称。 */
  label: string
  /** input。 */
  input: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: JSONToTypesExample = {
  label: 'json-to-types 示例',
  input: '{\n  "id": 1,\n  "name": "Alice"\n}',
}

/** 全部示例。 */
export const EXAMPLES: readonly JSONToTypesExample[] = [DEFAULT_EXAMPLE]
