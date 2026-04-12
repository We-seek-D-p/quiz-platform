export interface HostUser {
  id: string
  email: string
  nickname: string
  role: string
}

export interface AccessTokenPayload {
  accessToken: string
  tokenType: 'Bearer'
  expiresIn: number
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  nickname: string
  password: string
}

export interface LoginResponse extends AccessTokenPayload {
  user: HostUser
}
