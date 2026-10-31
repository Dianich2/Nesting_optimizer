
import { useState, useCallback } from 'react'
import type { ReactNode } from 'react'

import {
    getCurrentUser,
    loginUser,
    logoutUser,
} from './api'

import { AuthContext } from './AuthContext'

import type { LoginRequest, TokenPair, User } from './types'

interface AuthProviderProps {
    children: ReactNode
}

interface Session {
    user: User
    tokens: TokenPair
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [session, setSession] = useState<Session | null>(null)
    const [logoutRedirectPending, setLogoutRedirectPending] = useState(false)

    const acknowledgeLogoutRedirect = useCallback(() => {
        setLogoutRedirectPending(false)
    }, [])

    async function login(credentials: LoginRequest): Promise<void> {
        const tokens = await loginUser(credentials)
        const user = await getCurrentUser(tokens.access_token)
        setLogoutRedirectPending(false)
        setSession({ user, tokens })
    }


    async function logout(): Promise<void> {
        if (!session) {
            return
        }

        await logoutUser(session.tokens.access_token)

        setLogoutRedirectPending(true)
        setSession(null)
    }


    return (
        <AuthContext.Provider
            value={{
                user: session?.user ?? null,
                isAuthenticated: session !== null,
                logoutRedirectPending,
                acknowledgeLogoutRedirect,
                login,
                logout,
            }}
        >
            {children}
        </AuthContext.Provider>
    )
}
