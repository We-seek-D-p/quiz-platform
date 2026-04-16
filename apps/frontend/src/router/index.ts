import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const APP_TITLE = 'Quiz Platform'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/public/Home.vue'),
    meta: {
      layout: 'DefaultLayout',
      title: 'Главная',
    },
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/auth/Login.vue'),
    meta: {
      layout: 'DefaultLayout',
      title: 'Вход',
      guestOnly: true,
    },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/auth/Register.vue'),
    meta: {
      layout: 'DefaultLayout',
      title: 'Регистрация',
      guestOnly: true,
    },
  },
  {
    path: '/host',
    name: 'dashboard',
    component: () => import('@/views/host/Dashboard.vue'),
    meta: {
      layout: 'DashboardLayout',
      title: 'Панель управления',
      requiresAuth: true,
    },
  },
  {
    path: '/quizzes',
    name: 'quiz-list',
    component: () => import('@/views/host/QuizList.vue'),
    meta: {
      layout: 'DashboardLayout',
      title: 'Список квизов',
      requiresAuth: true,
    },
  },
  {
    path: '/quizzes/editor',
    name: 'quiz-editor',
    component: () => import('@/views/host/QuizEditor.vue'),
    meta: {
      layout: 'DashboardLayout',
      title: 'Редактор квиза',
      requiresAuth: true,
    },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (!authStore.isSessionReady) {
    await authStore.initializeSession()
  }

  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth === true)
  const guestOnly = to.matched.some((record) => record.meta.guestOnly === true)

  if (requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath || '/host',
      },
    }
  }

  if (guestOnly && authStore.isAuthenticated) {
    return { name: 'dashboard' }
  }

  return true
})

router.afterEach((to) => {
  if (typeof document !== 'undefined') {
    const pageTitle = typeof to.meta.title === 'string' ? to.meta.title : undefined
    document.title = pageTitle ? `${pageTitle} · ${APP_TITLE}` : APP_TITLE
  }
})

export default router
