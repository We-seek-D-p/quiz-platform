export interface HostUser {
    id: string;
    email: string;
    nickname: string;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
    lastLoginAt: string | null;
}

export interface AuthTokens {
    accessToken: string;
    refreshToken: string;
    tokenType: 'Bearer';
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface RegisterRequest {
    email: string;
    nickname: string;
    password: string;
    confirmPassword: string;
}