import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type {
  AccessTokenPayload,
  HostUser,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
} from '../types'
import {
  type ApiAccessToken,
  type ApiLoginResponse,
  authRequest,
  loginRequest,
  logoutRequest,
  parseErrorMessage,
  refreshRequest,
  registerRequest,
  toAccessTokenPayload,
  toLoginResponse,
} from '../api/auth'

type RequestWithRetryOptions = {
  skipRefresh?: boolean
}

type StoreHttpError = Error & { status?: number }

export const useAuthStore = defineStore('auth', () => {
  const user = ref<HostUser | null>(null)
  const accessToken = ref<string | null>(null)
  const isSessionReady = ref(false)

  let refreshPromise: Promise<boolean> | null = null

  const isAuthenticated = computed(() => {
    return Boolean(user.value && accessToken.value)
  })

  const setAccessToken = (payload: AccessTokenPayload) => {
    accessToken.value = payload.accessToken
  }

  const clearSession = () => {
    user.value = null
    accessToken.value = null
  }

  const requestWithRetry = async (
    path: string,
    init: RequestInit = {},
    options: RequestWithRetryOptions = {},
  ): Promise<Response> => {
    const response = await authRequest(path, init, { accessToken: accessToken.value })

    if (response.status !== 401 || options.skipRefresh) {
      return response
    }

    const refreshed = await refreshAccessToken()
    if (!refreshed) {
      clearSession()
      return response
    }

    return authRequest(path, init, { accessToken: accessToken.value })
  }

  const refreshAccessToken = async (): Promise<boolean> => {
    if (refreshPromise) {
      return refreshPromise
    }

    refreshPromise = (async () => {
      const response = await refreshRequest()

      if (!response.ok) {
        clearSession()
        return false
      }

      const payload = (await response.json()) as ApiAccessToken
      if (!payload.access_token) {
        clearSession()
        return false
      }

      setAccessToken(toAccessTokenPayload(payload))
      return true
    })().catch(() => {
      clearSession()
      return false
    })

    try {
      return await refreshPromise
    } finally {
      refreshPromise = null
    }
  }

  const initializeSession = async (): Promise<void> => {
    isSessionReady.value = false

    const refreshed = await refreshAccessToken()
    if (!refreshed) {
      clearSession()
      isSessionReady.value = true
      return
    }

    try {
      if (!accessToken.value) {
        clearSession()
        return
      }

      const response = await requestWithRetry('/me', { method: 'GET' }, { skipRefresh: true })
      if (!response.ok) {
        clearSession()
        return
      }

      user.value = (await response.json()) as HostUser
    } catch {
      clearSession()
    } finally {
      isSessionReady.value = true
    }
  }

  const login = async (payload: LoginRequest): Promise<LoginResponse> => {
    const response = await loginRequest(payload)

    if (!response.ok) {
      const error = new Error(await parseErrorMessage(response)) as StoreHttpError
      error.status = response.status
      throw error
    }

    const data = toLoginResponse((await response.json()) as ApiLoginResponse)
    setAccessToken(data)
    user.value = data.user
    return data
  }

  const register = async (payload: RegisterRequest): Promise<HostUser> => {
    const response = await registerRequest(payload)

    if (!response.ok) {
      const error = new Error(await parseErrorMessage(response)) as StoreHttpError
      error.status = response.status
      throw error
    }

    return (await response.json()) as HostUser
  }

  const logout = async (): Promise<void> => {
    try {
      await logoutRequest(accessToken.value)
    } finally {
      clearSession()
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
