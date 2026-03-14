import {createRouter, createWebHistory, type RouteRecordRaw} from 'vue-router'

const routes: Array<RouteRecordRaw> = [
  // Public Routes (Default Layout)
  {
    path: '/',
    name: 'home',
    component: () => import('../views/public/Home.vue'),
    meta: {layout: 'DefaultLayout'}
  },
  // {
  //   path: '/join/:code',
  //   name: 'join',
  //   component: () => import('../views/public/JoinRoom.vue'),
  //   meta: { layout: 'DefaultLayout' }
  // },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/auth/Login.vue'),
    meta: {layout: 'DefaultLayout', guestOnly: true}
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/auth/Register.vue'),
    meta: {layout: 'DefaultLayout', guestOnly: true}
  },

  // Host Routes (Dashboard Layout - Requires Auth)
  // {
  //   path: '/host',
  //   name: 'dashboard',
  //   component: () => import('../views/host/Dashboard.vue'),
  //   meta: { layout: 'DashboardLayout', requiresAuth: true }
  // },
  // {
  //   path: '/host/quiz/new',
  //   name: 'create-quiz',
  //   component: () => import('../views/host/QuizEditor.vue'),
  //   meta: { layout: 'DashboardLayout', requiresAuth: true }
  // },
  // {
  //   path: '/host/quiz/:id/edit',
  //   name: 'edit-quiz',
  //   component: () => import('../views/host/QuizEditor.vue'),
  //   meta: { layout: 'DashboardLayout', requiresAuth: true }
  // },

  // Game Routes (Game Layout)
  // {
  //   path: '/game/:code/lobby',
  //   name: 'lobby',
  //   component: () => import('../views/game/Lobby.vue'),
  //   meta: { layout: 'GameLayout' } // Доступ проверяется по WS
  // },
  // {
  //   path: '/game/:code/play',
  //   name: 'play',
  //   component: () => import('../views/game/Play.vue'),
  //   meta: { layout: 'GameLayout' }
  // },
  // {
  //   path: '/game/:code/results',
  //   name: 'results',
  //   component: () => import('../views/game/Results.vue'),
  //   meta: { layout: 'GameLayout' }
  // },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
