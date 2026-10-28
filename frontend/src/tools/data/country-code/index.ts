/**
 * country-code 工具模块。
 *
 * 默认导出工具组件；`messages` 是该工具的双语词条，
 * 由路由容器在组件加载后登记，避免进入初始包。
 */
import type { Component } from 'vue'

import enUS from '@/locales/en-US/tools/country-code.json'
import zhCN from '@/locales/zh-CN/tools/country-code.json'

import Tool from './Tool.vue'

/** 工具词条。 */
export const messages: Record<'zh-CN' | 'en-US', Record<string, unknown>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

export default Tool as Component
