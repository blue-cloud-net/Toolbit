/**
 * 路由表。
 *
 * 全部路由级组件懒加载；工具页按 slug 拆成独立 chunk。
 *
 * 每个路由声明 `meta.title`，由 `afterEach` 钩子写入 `document.title`。
 */
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { i18n } from '@/i18n'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
        meta: { titleKey: 'common.nav.home' },
      },
      {
        path: 'recent',
        name: 'recent',
        component: () => import('@/views/RecentView.vue'),
        meta: { titleKey: 'common.nav.recent' },
      },
      {
        path: 'favorites',
        name: 'favorites',
        component: () => import('@/views/FavoritesView.vue'),
        meta: { titleKey: 'common.nav.favorites' },
      },
      {
        path: 'c/:category',
        name: 'category',
        component: () => import('@/views/CategoryView.vue'),
        props: true,
        meta: { titleKey: 'common.nav.categories' },
      },
      {
        path: 'tools/:slug',
        name: 'tool',
        component: () => import('@/views/ToolView.vue'),
        props: true,
        // 工具页的标题由目录元信息动态生成，见下方 afterEach。
        meta: { titleKey: null },
      },
      {
        path: ':pathMatch(.*)*',
        name: 'not-found',
        component: () => import('@/views/NotFoundView.vue'),
        meta: { titleKey: 'common.states.notFound' },
      },
    ],
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 }
  },
})

/**
 * 写入页面标题。
 *
 * 格式：`页面名 · Toolbit`。
 * 由路由 `afterEach` 与工具页在目录数据就绪后共同调用。
 */
export function setDocumentTitle(title?: string): void {
  if (typeof document === 'undefined') {
    return
  }
  const appName = i18n.global.t('common.appName')
  document.title = title ? `${title} · ${appName}` : appName
}

/** 路由切换后更新标题。 */
router.afterEach((to) => {
  const titleKey = to.meta.titleKey as string | null | undefined
  if (!titleKey) {
    // 工具页与分类页的标题依赖目录数据，由页面组件在数据就绪后设置。
    setDocumentTitle()
    return
  }
  setDocumentTitle(i18n.global.t(titleKey))
})
