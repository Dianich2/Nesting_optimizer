
import { apiRequest } from '../../shared/api/http'

import type {
    AccessTokenResponse,
    LoginRequest,
    RegisterRequest,
    User,
} from './types'

export function registerUser(data: RegisterRequest) {
    return apiRequest<User>('/users', {
        method: 'POST',
        body: JSON.stringify(data),
    })
}

export function loginUser(data: LoginRequest) {
    return apiRequest<AccessTokenResponse>('/auth/browser/login', {
        method: 'POST',
        credentials: 'same-origin',
        body: JSON.stringify(data),
    })
}

let refreshPromise: Promise<AccessTokenResponse> | null = null

export function refreshBrowserSession() {
    if (!refreshPromise) {
        refreshPromise = apiRequest<AccessTokenResponse>(
            '/auth/browser/refresh',
            {
                method: 'POST',
                credentials: 'same-origin',
            },
        ).finally(() => {
            refreshPromise = null
        })
    }

    return refreshPromise
}

export function getCurrentUser(accessToken: string) {
    return apiRequest<User>('/users/me', {
        headers: {
            Authorization: `Bearer ${accessToken}`,
        },
    })
}

export function logoutUser(accessToken: string) {
    return apiRequest<void>('/auth/browser/logout', {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            Authorization: `Bearer ${accessToken}`,
        },
    })
}
