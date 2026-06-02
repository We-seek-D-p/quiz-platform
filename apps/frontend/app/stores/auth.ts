import { defineStore } from 'pinia'
import type { HostUser, LoginRequest, LoginResponse, RegisterRequest } from '~/types/auth'
import { useAuthApi } from '~/composables/api/useAuthApi'

export const useAuthStore = defineStore('auth', () => {
  const authApi = useAuthApi()

  const user = ref<HostUser | null>(null)
  const accessToken = ref<string | null>(null)
  const isSessionReady = ref(false)

  let refreshPromise: Promise<boolean> | null = null

  const isAuthenticated = computed(() => {
    return Boolean(user.value && accessToken.value)
  })

  const clearSession = () => {
    user.value = null
    accessToken.value = null
  }

  const setAccessToken = (token: string) => {
    accessToken.value = token
  }

  const refreshAccessToken = async (): Promise<boolean> => {
    if (refreshPromise) {
      return refreshPromise
    }

    const pendingRefresh = authApi
      .refresh()
      .then((payload) => {
        setAccessToken(payload.accessToken)
        return true
      })
      .catch(() => {
        clearSession()
        return false
      })

    refreshPromise = pendingRefresh

    try {
      return await pendingRefresh
    } finally {
      if (refreshPromise === pendingRefresh) {
        refreshPromise = null
      }
    }
  }

  const initializeSession = async (): Promise<void> => {
    if (isSessionReady.value) {
      return
    }

    try {
      const refreshed = await refreshAccessToken()
      if (!refreshed || !accessToken.value) {
        clearSession()
        return
      }

      user.value = await authApi.me(accessToken.value)
    } catch {
      clearSession()
    } finally {
      isSessionReady.value = true
    }
  }

  const login = async (payload: LoginRequest): Promise<LoginResponse> => {
    const data = await authApi.login(payload)

    user.value = data.user
    setAccessToken(data.accessToken)

    return data
  }

  const register = async (payload: RegisterRequest): Promise<HostUser> => {
    return authApi.register(payload)
  }

  const logout = async (): Promise<void> => {
    try {
      await authApi.logout(accessToken.value)
    } finally {
      clearSession()
      isSessionReady.value = true
    }
  }

  return {
    user,
    accessToken,
    isSessionReady,
    isAuthenticated,
    initializeSession,
    refreshAccessToken,
    login,
    register,
    logout,
    clearSession,
  }
})
