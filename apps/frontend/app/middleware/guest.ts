import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware(async () => {
  const authStore = useAuthStore()

  await authStore.initializeSession()

  if (!authStore.isAuthenticated) {
    return
  }

  return navigateTo('/host')
})
