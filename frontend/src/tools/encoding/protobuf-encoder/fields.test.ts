/**
 * protobuf-encoder 字段模型：单元测试。
 */
import { describe, expect, it } from 'vitest'

import {
  cloneNodes,
  createNode,
  fromJsonText,
  removeNode,
  toJsonText,
  toPayload,
  type ProtoFieldNode,
} from '@/tools/encoding/protobuf-encoder/fields'

/** 构造一个字段节点。 */
function node(overrides: Partial<ProtoFieldNode> = {}): ProtoFieldNode {
  return { ...createNode(), ...overrides }
}

describe('toPayload', () => {
  it('输出后端字段规格的完整形状', () => {
    expect(toPayload([node({ field: 1, type: 'int32', value: '150' })])).toEqual([
      { field: 1, type: 'int32', value: '150', encoding: '', fields: [] },
    ])
  })

  it('message 类型才带子字段', () => {
    const nodes = [
      node({
        field: 1,
        type: 'message',
        children: [node({ field: 1, type: 'int32', value: '1' })],
      }),
    ]
    const payload = toPayload(nodes)
    expect(payload[0]?.fields).toHaveLength(1)
    expect(payload[0]?.fields[0]?.value).toBe('1')
  })

  it('非 message 类型的子字段被忽略', () => {
    const nodes = [node({ field: 1, type: 'string', children: [node()] })]
    expect(toPayload(nodes)[0]?.fields).toEqual([])
  })
})

describe('toJsonText / fromJsonText', () => {
  it('往返后字段数据不变', () => {
    const nodes = [
      node({ field: 1, type: 'string', value: 'Alice' }),
      node({ field: 2, type: 'int32', value: '7' }),
    ]
    expect(fromJsonText(toJsonText(nodes))).toEqual(nodes)
  })

  it('嵌套消息可以往返', () => {
    const nodes = [
      node({
        field: 1,
        type: 'message',
        children: [node({ field: 1, type: 'int32', value: '1' })],
      }),
    ]
    expect(fromJsonText(toJsonText(nodes))).toEqual(nodes)
  })

  it('数字与布尔值会被转成字符串', () => {
    const parsed = fromJsonText('[{"field":1,"type":"int32","value":150}]')
    expect(parsed[0]?.value).toBe('150')
  })

  it('空文本解析成空数组', () => {
    expect(fromJsonText('   ')).toEqual([])
  })

  it('非法 JSON 抛错', () => {
    expect(() => fromJsonText('{')).toThrow()
  })

  it('不是数组时抛错', () => {
    expect(() => fromJsonText('{"field":1}')).toThrow()
  })

  it('未知类型抛错', () => {
    expect(() => fromJsonText('[{"field":1,"type":"int7","value":"1"}]')).toThrow()
  })

  it('字段号不是整数时抛错', () => {
    expect(() => fromJsonText('[{"field":"x","type":"int32","value":"1"}]')).toThrow()
  })
})

describe('cloneNodes', () => {
  it('深拷贝后互不影响', () => {
    const nodes = [node({ type: 'message', children: [node({ value: '1' })] })]
    const copy = cloneNodes(nodes)
    copy[0]!.children[0]!.value = '2'
    expect(nodes[0]!.children[0]!.value).toBe('1')
  })
})

describe('removeNode', () => {
  it('按路径删除顶层节点', () => {
    const nodes = [node({ field: 1 }), node({ field: 2 })]
    expect(removeNode(nodes, [0])).toHaveLength(1)
  })

  it('按路径删除嵌套节点', () => {
    const nodes = [
      node({
        field: 1,
        type: 'message',
        children: [node({ field: 1 }), node({ field: 2 })],
      }),
    ]
    const result = removeNode(nodes, [0, 0])
    expect(result[0]?.children).toHaveLength(1)
  })
})
