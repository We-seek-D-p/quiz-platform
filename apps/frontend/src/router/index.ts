import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/public/Home.vue'),
    meta: {
      layout: 'DefaultLayout',
      title: '',
    },
  },
  // {
  //   path: '/join/:code',
  //   name: 'join',
  //   component: () => import('@/views/public/JoinRoom.vue'),
  //   meta: { layout: 'DefaultLayout' }
  // },
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

  // {
  //   path: '/host/quizzes',
  //   name: 'host-quizzes',
  //   component: () => import('@/views/host/Quizzes.vue'),
  //   meta: {
  //     layout: 'DashboardLayout',
  //     title: 'Мои квизы',
  //     requiresAuth: true,
  //   },
  // },
  // {
  //   path: '/host/quiz/new',
  //   name: 'create-quiz',
  //   component: () => import('@/views/host/QuizEditor.vue'),
  //   meta: {
  //     layout: 'DashboardLayout',
  //     title: 'Создать квиз',
  //     requiresAuth: true,
  //   },
  // },
  // {
  //   path: '/host/quiz/:id/edit',
  //   name: 'edit-quiz',
  //   component: () => import('@/views/host/QuizEditor.vue'),
  //   meta: {
  //     layout: 'DashboardLayout',
  //     title: 'Редактирование квиза',
  //     requiresAuth: true,
  //   },
  // },

  // {
  //   path: '/game/:code/lobby',
  //   name: 'lobby',
  //   component: () => import('@/views/game/Lobby.vue'),
  //   meta: {
  //     layout: 'GameLayout',
  //     title: 'Лобби',
  //   },
  // },
  // {
  //   path: '/game/:code/play',
  //   name: 'play',
  //   component: () => import('@/views/game/Play.vue'),
  //   meta: {
  //     layout: 'GameLayout',
  //     title: 'Игра',
  //   },
  // },
  // {
  //   path: '/game/:code/results',
  //   name: 'results',
  //   component: () => import('@/views/game/Results.vue'),
  //   meta: {
  //     layout: 'GameLayout',
  //     title: 'Результаты',
  //   },
  // },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
