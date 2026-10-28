/**
 * 路由表。
 *
 * 全部路由级组件懒加载；工具页按 slug 拆成独立 chunk。
 */
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
      },
      {
        path: 'recent',
        name: 'recent',
        component: () => import('@/views/RecentView.vue'),
      },
      {
        path: 'favorites',
        name: 'favorites',
        component: () => import('@/views/FavoritesView.vue'),
      },
      {
        path: 'c/:category',
        name: 'category',
        component: () => import('@/views/CategoryView.vue'),
        props: true,
      },
      {
        path: 'tools/:slug',
        name: 'tool',
        component: () => import('@/views/ToolView.vue'),
        props: true,
      },
      {
        path: ':pathMatch(.*)*',
        name: 'not-found',
        component: () => import('@/views/NotFoundView.vue'),
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
