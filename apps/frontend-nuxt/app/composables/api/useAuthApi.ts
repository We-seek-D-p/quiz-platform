import type {
  AccessTokenPayload,
  ApiAccessToken,
  ApiLoginResponse,
  HostUser,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
} from '~/types/auth'
import { useApiClient } from '~/composables/api/useApiClient'

const toAccessTokenPayload = (payload: ApiAccessToken): AccessTokenPayload => {
  return {
    accessToken: payload.access_token,
    expiresIn: payload.expires_in,
    tokenType: 'Bearer',
  }
}

const toLoginResponse = (payload: ApiLoginResponse): LoginResponse => {
  return {
    ...toAccessTokenPayload(payload),
    user: payload.user,
  }
}

export const useAuthApi = () => {
  const config = useRuntimeConfig()
  const { request } = useApiClient()

  const authBase = config.public.authApiBase

  return {
    login: async (payload: LoginRequest): Promise<LoginResponse> => {
      const data = await request<ApiLoginResponse>(`${authBase}/login`, {
        method: 'POST',
        body: payload,
      })

      return toLoginResponse(data)
    },

    register: async (payload: RegisterRequest): Promise<HostUser> => {
      return request<HostUser>(`${authBase}/register`, {
        method: 'POST',
        body: payload,
      })
    },

    refresh: async (): Promise<AccessTokenPayload> => {
      const data = await request<ApiAccessToken>(`${authBase}/refresh`, {
        method: 'POST',
      })

      return toAccessTokenPayload(data)
    },

    logout: async (accessToken: string | null): Promise<void> => {
      await request<void>(`${authBase}/logout`, {
        method: 'POST',
        accessToken,
      })
    },

    me: async (accessToken: string): Promise<HostUser> => {
      return request<HostUser>(`${authBase}/me`, {
        method: 'GET',
        accessToken,
      })
    },
  }
}
