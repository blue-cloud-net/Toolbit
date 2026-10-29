/**
 * eml-parser 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料；发件人与收件人使用示例域名。
 */

/** 一个 EML 解析示例。 */
export interface EMLParserExample {
  /** 示例名称。 */
  label: string
  /** 邮件源文。 */
  text: string
  /** 源文字符编码。 */
  encoding: string
  /** 单附件字节上限；0 表示用服务端默认值。 */
  maxAttachmentBytes: number
}

/** 默认示例：带纯文本正文的简单邮件。 */
export const DEFAULT_EXAMPLE: EMLParserExample = {
  label: '纯文本邮件',
  text: [
    'From: sender@example.com',
    'To: receiver@example.com',
    'Subject: 测试邮件',
    'MIME-Version: 1.0',
    'Content-Type: text/plain; charset=utf-8',
    '',
    '这是一封用于演示的纯文本邮件。',
  ].join('\r\n'),
  encoding: 'utf-8',
  maxAttachmentBytes: 0,
}

/** 带附件的示例。 */
export const ATTACHMENT_EXAMPLE: EMLParserExample = {
  label: '带附件的邮件',
  text: [
    'Subject: 附件',
    'MIME-Version: 1.0',
    'Content-Type: multipart/mixed; boundary="B"',
    '',
    '--B',
    'Content-Type: text/plain; charset=utf-8',
    '',
    '正文内容。',
    '--B',
    'Content-Type: application/octet-stream',
    'Content-Disposition: attachment; filename="a.bin"',
    '',
    'BYTES',
    '--B--',
  ].join('\r\n'),
  encoding: 'utf-8',
  maxAttachmentBytes: 0,
}

/** 全部示例。 */
export const EXAMPLES: readonly EMLParserExample[] = [DEFAULT_EXAMPLE, ATTACHMENT_EXAMPLE]
