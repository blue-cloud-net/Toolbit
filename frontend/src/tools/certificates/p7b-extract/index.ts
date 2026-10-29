/**
 * p7b-extract 工具模块。
 */
import type { Component } from 'vue'

import enUS from '@/locales/en-US/tools/p7b-extract.json'
import zhCN from '@/locales/zh-CN/tools/p7b-extract.json'

import Tool from './Tool.vue'

/** 工具词条。 */
export const messages: Record<'zh-CN' | 'en-US', Record<string, unknown>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

export default Tool as Component
