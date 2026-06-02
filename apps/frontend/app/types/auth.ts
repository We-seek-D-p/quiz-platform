export type HostUser = {
  id: string
  email: string
  nickname: string
  role: string
}

export type LoginRequest = {
  email: string
  password: string
}

export type RegisterRequest = {
  email: string
  nickname: string
  password: string
}

export type AccessTokenPayload = {
  accessToken: string
  expiresIn: number
  tokenType: 'Bearer'
}

export type LoginResponse = AccessTokenPayload & {
  user: HostUser
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
