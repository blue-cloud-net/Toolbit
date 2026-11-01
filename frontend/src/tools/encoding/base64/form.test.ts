/**
 * base64 工具表单纯函数：单元测试。
 */
import { describe, expect, it } from 'vitest'

import {
  buildPayload,
  normalizeForm,
  reverseConversion,
  type Base64FormState,
} from '@/tools/encoding/base64/form'

/** 构造一份表单状态。 */
function form(overrides: Partial<Base64FormState> = {}): Base64FormState {
  return {
    operation: 'encode',
    input: 'Hello',
    inputEncoding: 'utf8',
    variant: 'standard',
    padding: true,
    outputEncoding: 'utf8',
    targetVariant: 'url',
    ...overrides,
  }
}

describe('buildPayload', () => {
  it('encode 下发 outputEncoding', () => {
    expect(buildPayload(form({ operation: 'encode' }))).toEqual({
      operation: 'encode',
      input: { data: 'Hello', encoding: 'utf8' },
      variant: 'standard',
      padding: true,
      outputEncoding: 'utf8',
    })
  })

  it('decode 下发 outputEncoding', () => {
    const payload = buildPayload(form({ operation: 'decode', outputEncoding: 'hex' }))
    expect(payload.outputEncoding).toBe('hex')
    expect(payload).not.toHaveProperty('targetVariant')
  })

  it('convert 下发 targetVariant', () => {
    const payload = buildPayload(
      form({ operation: 'convert', variant: 'auto', targetVariant: 'url' }),
    )
    expect(payload.targetVariant).toBe('url')
    expect(payload.variant).toBe('auto')
    expect(payload).not.toHaveProperty('outputEncoding')
  })

  it('convert 不下发 outputEncoding', () => {
    expect(buildPayload(form({ operation: 'convert' }))).not.toHaveProperty('outputEncoding')
  })
})

describe('normalizeForm', () => {
  it('切到 encode 时把 auto 变体回落为 standard', () => {
    const next = normalizeForm(form({ operation: 'encode', variant: 'auto' }))
    expect(next.variant).toBe('standard')
  })

  it('切到 convert 时把 hex 输入编码回落为 utf8', () => {
    const next = normalizeForm(form({ operation: 'convert', inputEncoding: 'hex' }))
    expect(next.inputEncoding).toBe('utf8')
  })

  it('convert 保留 auto 变体', () => {
    const next = normalizeForm(form({ operation: 'convert', variant: 'auto' }))
    expect(next.variant).toBe('auto')
  })

  it('不修改传入的对象', () => {
    const original = form({ operation: 'convert', inputEncoding: 'hex' })
    normalizeForm(original)
    expect(original.inputEncoding).toBe('hex')
  })
})

describe('reverseConversion', () => {
  it('交换 standard 与 url', () => {
    const next = reverseConversion(form({ variant: 'standard', targetVariant: 'url' }), '')
    expect(next.variant).toBe('url')
    expect(next.targetVariant).toBe('standard')
  })

  it('auto 变体落到目标变体的对面', () => {
    const next = reverseConversion(form({ variant: 'auto', targetVariant: 'url' }), '')
    expect(next.variant).toBe('url')
    expect(next.targetVariant).toBe('standard')
  })

  it('auto 变体且目标为 standard 时落到 url', () => {
    const next = reverseConversion(form({ variant: 'auto', targetVariant: 'standard' }), '')
    expect(next.variant).toBe('standard')
    expect(next.targetVariant).toBe('url')
  })

  it('把结果回填为新输入', () => {
    const next = reverseConversion(form({ input: '++//Pw==' }), '--__Pw==')
    expect(next.input).toBe('--__Pw==')
  })

  it('没有结果时保持输入不变', () => {
    const next = reverseConversion(form({ input: '++//Pw==' }), '')
    expect(next.input).toBe('++//Pw==')
  })

  it('反向两次回到原状', () => {
    const once = reverseConversion(form({ variant: 'standard', targetVariant: 'url' }), '')
    const twice = reverseConversion(once, '')
    expect(twice.variant).toBe('standard')
    expect(twice.targetVariant).toBe('url')
  })
})
