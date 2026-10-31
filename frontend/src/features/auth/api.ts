
import { apiRequest } from '../../shared/api/http'

import type {
    LoginRequest,
    RegisterRequest,
    TokenPair,
    User,
} from './types'

export function registerUser(data: RegisterRequest) {
    return apiRequest<User>('/users', {
        method: 'POST',
        body: JSON.stringify(data),
    })
}

export function loginUser(data: LoginRequest) {
    return apiRequest<TokenPair>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(data),
    })
}

export function refreshTokens(refreshToken: string) {
    return apiRequest<TokenPair>('/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({
            refresh_token: refreshToken,
        }),
    })
}

export function getCurrentUser(accessToken: string) {
    return apiRequest<User>('/users/me', {
        headers: {
            Authorization: `Bearer ${accessToken}`,
        },
    })
}

export function logoutUser(accessToken: string) {
    return apiRequest<void>('/auth/logout', {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${accessToken}`,
        },
    })
}
