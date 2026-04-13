import type {
  AccessTokenPayload,
  HostUser,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
} from '../types'

const AUTH_API_PREFIX = '/api/v1/auth'

type AuthRequestOptions = {
  accessToken?: string | null
}

export type ApiAccessToken = {
  access_token: string
  expires_in: number
}

export type ApiLoginResponse = ApiAccessToken & {
  user: HostUser
}

export type ApiErrorPayload = {
  detail?: string
}

const buildRequestInit = (
  init: RequestInit = {},
  options: AuthRequestOptions = {},
): RequestInit => {
  const headers = new Headers(init.headers)

  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (options.accessToken) {
    headers.set('Authorization', `Bearer ${options.accessToken}`)
  }

  return {
    ...init,
    headers,
    credentials: 'include',
  }
}

const request = async (
  path: string,
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
  options: AuthRequestOptions = {},
  body?: unknown,
): Promise<Response> => {
  const init: RequestInit = {
    method,
    ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
  }

  return fetch(`${AUTH_API_PREFIX}${path}`, buildRequestInit(init, options))
}

export const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const payload = (await response.json()) as ApiErrorPayload
    if (typeof payload.detail === 'string' && payload.detail.trim().length > 0) {
      return payload.detail
    }
  } catch {
    return `${response.status} ${response.statusText}`.trim()
  }

  return `${response.status} ${response.statusText}`.trim()
}

export const authRequest = async (
  path: string,
  init: RequestInit = {},
  options: AuthRequestOptions = {},
): Promise<Response> => {
  return fetch(`${AUTH_API_PREFIX}${path}`, buildRequestInit(init, options))
}

export const loginRequest = async (payload: LoginRequest): Promise<Response> => {
  return request('/login', 'POST', {}, payload)
}

export const registerRequest = async (payload: RegisterRequest): Promise<Response> => {
  return request('/register', 'POST', {}, payload)
}

export const refreshRequest = async (): Promise<Response> => {
  return request('/refresh', 'POST')
}

export const logoutRequest = async (accessToken: string | null): Promise<Response> => {
  return request('/logout', 'POST', { accessToken })
}

export const toAccessTokenPayload = (payload: ApiAccessToken): AccessTokenPayload => {
  return {
    accessToken: payload.access_token,
    tokenType: 'Bearer',
    expiresIn: payload.expires_in,
  }
}

export const toLoginResponse = (payload: ApiLoginResponse): LoginResponse => {
  return {
    ...toAccessTokenPayload(payload),
    user: payload.user,
  }
}
