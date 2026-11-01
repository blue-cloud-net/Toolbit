/**
 * 材料文件导入的纯函数测试。
 *
 * 覆盖 `detectMaterialFile` 的四条判定路径与两个「像不像」的启发式判定。
 */
import { describe, expect, it } from 'vitest'

import {
  bytesToBase64,
  decodeUtf8,
  detectMaterialFile,
  looksLikeBase64,
  looksLikeHex,
} from '@/components/tool/materialFile'

/** 一段示例 PEM。 */
const PEM = `-----BEGIN CERTIFICATE-----
MIIDzTCCArWgAwIBAgIUbA85LjhezNVJkBzBCm0Ku3kj04wwDQYJKoZIhvcNAQEL
BQAwSzELMAkGA1UEBhMCQ04xFTATBgNVBAoMDFRvb2xiaXQgVGVzdDElMCMGA1UE
-----END CERTIFICATE-----
`

/** 一段示例 DER（首字节 0x30 = SEQUENCE）。 */
const DER = new Uint8Array([0x30, 0x82, 0x01, 0x22, 0x00, 0xff, 0xfe])

describe('材料文件导入', () => {
  describe('looksLikeHex', () => {
    it('接受偶数长度的十六进制串', () => {
      expect(looksLikeHex('0123456789abcdef')).toBe(true)
    })

    it('接受带换行的十六进制串', () => {
      expect(looksLikeHex('0123 4567\n89ab cdef')).toBe(true)
    })

    it('拒绝奇数长度', () => {
      expect(looksLikeHex('0123456789abcde')).toBe(false)
    })

    it('拒绝过短的串', () => {
      expect(looksLikeHex('0123')).toBe(false)
    })

    it('拒绝含非十六进制字符', () => {
      expect(looksLikeHex('0123456789abcdefz')).toBe(false)
    })
  })

  describe('looksLikeBase64', () => {
    it('接受 4 的倍数长度', () => {
      expect(looksLikeBase64('QUJDREVGR0hJSktMTU5PUA==')).toBe(true)
    })

    it('接受带换行的 base64', () => {
      expect(looksLikeBase64('QUJDREVG\nR0hJSktM\nTU5PUFFS')).toBe(true)
    })

    it('拒绝长度不是 4 的倍数', () => {
      expect(looksLikeBase64('QUJDREVG')).toBe(false)
    })

    it('拒绝过短的串', () => {
      expect(looksLikeBase64('QUJD')).toBe(false)
    })

    it('拒绝含非法字符', () => {
      expect(looksLikeBase64('QUJD*REVG')).toBe(false)
    })
  })

  describe('bytesToBase64', () => {
    it('编码结果可被标准 base64 解回', () => {
      expect(bytesToBase64(new Uint8Array([1, 2, 3, 4, 5]))).toBe('AQIDBAU=')
    })

    it('空数组得到空串', () => {
      expect(bytesToBase64(new Uint8Array())).toBe('')
    })
  })

  describe('detectMaterialFile', () => {
    it('识别 PEM 文本', () => {
      const text = PEM
      const result = detectMaterialFile(new TextEncoder().encode(text), text)
      expect(result?.detected).toBe('pem')
      expect(result?.content).toBe(text)
    })

    it('识别 DER 二进制并转成 base64', () => {
      const result = detectMaterialFile(DER, decodeUtf8(DER))
      expect(result?.detected).toBe('der')
      expect(result?.content).toBe(bytesToBase64(DER))
    })

    it('识别十六进制文本', () => {
      const text = '0123456789abcdef'
      const result = detectMaterialFile(new TextEncoder().encode(text), text)
      expect(result?.detected).toBe('hex')
      expect(result?.content).toBe(text)
    })

    it('不会把以 0 开头的十六进制串误认成 DER', () => {
      // ASCII '0' 的字节值正好是 0x30（ASN.1 SEQUENCE），是最容易踩的坑。
      const text = '0a0b0c0d0e0f1011'
      const result = detectMaterialFile(new TextEncoder().encode(text), text)
      expect(result?.detected).toBe('hex')
    })

    it('识别 base64 文本', () => {
      const text = 'QUJDREVGR0hJSktMTU5PUA=='
      const result = detectMaterialFile(new TextEncoder().encode(text), text)
      expect(result?.detected).toBe('base64')
      expect(result?.content).toBe(text)
    })

    it('拒绝不像 PEM 族材料的文件', () => {
      const text = 'hello world, this is just a note'
      const bytes = new TextEncoder().encode(text)
      expect(detectMaterialFile(bytes, text)).toBeNull()
    })

    it('拒绝空文件', () => {
      expect(detectMaterialFile(new Uint8Array(), '')).toBeNull()
    })
  })

  describe('decodeUtf8', () => {
    it('非法字节用替换字符，不抛错', () => {
      expect(decodeUtf8(new Uint8Array([0xff, 0xfe]))).toHaveLength(2)
    })
  })
})
