/**
 * Base64 → 字节数组：单元测试。
 */
import { describe, expect, it } from 'vitest'

import { base64ToBytes } from '@/components/tool/base64'

describe('base64ToBytes', () => {
  it('空串得到空数组', () => {
    expect(base64ToBytes('')).toHaveLength(0)
  })

  it('解码 ASCII 文本', () => {
    // "SGVsbG8=" 是 "Hello" 的 Base64 编码。
    expect(Array.from(base64ToBytes('SGVsbG8='))).toEqual([
      0x48, 0x65, 0x6c, 0x6c, 0x6f,
    ])
  })

  it('忽略空白字符', () => {
    expect(Array.from(base64ToBytes('SGVs\nbG8='))).toEqual(
      Array.from(base64ToBytes('SGVsbG8=')),
    )
  })

  it('保留高位字节', () => {
    // 0xFF 0x00 0x80 的 Base64 编码。
    expect(Array.from(base64ToBytes('/wCA'))).toEqual([0xff, 0x00, 0x80])
  })

  it('跨块解码结果与一次性解码一致', () => {
    // 超过一个分块（4 × 8192 字符）的输入。
    const raw = new Uint8Array(70000)
    for (let i = 0; i < raw.length; i += 1) {
      raw[i] = i % 256
    }
    let binary = ''
    for (const byte of raw) {
      binary += String.fromCharCode(byte)
    }
    const encoded = btoa(binary)

    const decoded = base64ToBytes(encoded)
    expect(decoded).toHaveLength(raw.length)
    expect(decoded[0]).toBe(raw[0])
    expect(decoded[raw.length - 1]).toBe(raw[raw.length - 1])
  })

  it('非法内容抛错', () => {
    expect(() => base64ToBytes('not base64!!')).toThrow()
  })
})
