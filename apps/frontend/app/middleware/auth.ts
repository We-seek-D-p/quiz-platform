import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware(async (to) => {
  const authStore = useAuthStore()

  await authStore.initializeSession()

  if (authStore.isAuthenticated) {
    return
  }

  return navigateTo({
    path: '/login',
    query: {
      redirect: to.fullPath || '/host',
    },
  })
})
