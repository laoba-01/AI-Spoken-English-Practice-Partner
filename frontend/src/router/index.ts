import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/pages/HomePage.vue'),
    meta: { title: 'AI 英语口语陪练' },
  },
  {
    path: '/chat/:id',
    name: 'Chat',
    component: () => import('@/pages/ChatPage.vue'),
    meta: { title: '对话中' },
    props: true,
  },
  {
    path: '/history',
    name: 'History',
    component: () => import('@/pages/HistoryPage.vue'),
    meta: { title: '历史记录' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 动态设置页面标题
router.beforeEach((to) => {
  document.title = (to.meta.title as string) || 'AI 英语口语陪练'
})

export default router
