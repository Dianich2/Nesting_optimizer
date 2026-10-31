
import { createContext } from 'react'

import type { LoginRequest, User } from './types'

export interface AuthContextValue {
    user: User | null
    isAuthenticated: boolean
    isInitializing: boolean
    logoutRedirectPending: boolean
    acknowledgeLogoutRedirect: () => void
    login: (credentials: LoginRequest) => Promise<void>
    logout: () => Promise<void>
    authorizedRequest: <T>(
        path: string,
        options?: RequestInit,
    ) => Promise<T>
}

export const AuthContext = createContext<AuthContextValue | null>(null)
