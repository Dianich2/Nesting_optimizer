
import { useCallback, useEffect, useRef, useState } from 'react'
import type { ReactNode } from 'react'

import {
    getCurrentUser,
    loginUser,
    refreshBrowserSession,
} from './api'

import { AuthContext } from './AuthContext'
import { ApiError, apiRequest } from '../../shared/api/http'

import type { LoginRequest, User } from './types'

interface AuthProviderProps {
    children: ReactNode
}

interface Session {
    user: User
    accessToken: string
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [session, setSession] = useState<Session | null>(null)
    const sessionRef = useRef<Session | null>(null)

    const [isInitializing, setIsInitializing] = useState(true)
    const [logoutRedirectPending, setLogoutRedirectPending] = useState(false)

    const updateSession = useCallback((next: Session | null) => {
        sessionRef.current = next
        setSession(next)
    }, [])

    const acknowledgeLogoutRedirect = useCallback(() => {
        setLogoutRedirectPending(false)
    }, [])

    useEffect(() => {
        let active = true

        async function restoreSession() {
            try {
                const tokens = await refreshBrowserSession()

                if (!active) {
                    return
                }

                const user = await getCurrentUser(tokens.access_token)

                if (active) {
                    updateSession({
                        user,
                        accessToken: tokens.access_token,
                    })
                }
            } catch {
                // Нет доступной сессии или произошла ошибка сети.
            } finally {
                if (active) {
                    setIsInitializing(false)
                }
            }
        }

        void restoreSession()

        return () => {
            active = false
        }
    }, [updateSession])

    async function login(credentials: LoginRequest): Promise<void> {
        const tokens = await loginUser(credentials)
        const user = await getCurrentUser(tokens.access_token)

        setLogoutRedirectPending(false)

        updateSession({
            user,
            accessToken: tokens.access_token,
        })
    }

    async function authorizedRequest<T>(
        path: string,
        options: RequestInit = {},
    ): Promise<T> {
        const initialSession = sessionRef.current

        if (!initialSession) {
            throw new ApiError(401, 'unauthorized_error')
        }

        function send(token: string): Promise<T> {
            const headers = new Headers(options.headers)
            headers.set('Authorization', `Bearer ${token}`)

            return apiRequest<T>(path, {
                ...options,
                headers,
            })
        }

        try {
            return await send(initialSession.accessToken)
        } catch (error) {
            if (!(error instanceof ApiError) || error.status !== 401) {
                throw error
            }
        }

        let currentSession = sessionRef.current

        if (
            !currentSession ||
            currentSession.user.id !== initialSession.user.id
        ) {
            throw new ApiError(401, 'unauthorized_error')
        }

        let token = currentSession.accessToken

        if (token === initialSession.accessToken) {
            try {
                const refreshed = await refreshBrowserSession()

                currentSession = sessionRef.current

                if (
                    !currentSession ||
                    currentSession.user.id !== initialSession.user.id
                ) {
                    throw new ApiError(401, 'unauthorized_error')
                }

                token = refreshed.access_token

                updateSession({
                    ...currentSession,
                    accessToken: token,
                })
            } catch (error) {
                if (
                    error instanceof ApiError &&
                    (error.status === 400 || error.status === 401)
                ) {
                    const latestSession = sessionRef.current

                    if (
                        latestSession &&
                        latestSession.user.id === initialSession.user.id
                    ) {
                        updateSession(null)
                    }
                }

                throw error
            }
        }

        // Только один повтор исходного запроса.
        return send(token)
    }

    async function logout(): Promise<void> {
        if (!sessionRef.current) {
            return
        }

        await authorizedRequest<void>('/auth/browser/logout', {
            method: 'POST',
            credentials: 'same-origin',
        })

        setLogoutRedirectPending(true)
        updateSession(null)
    }

    return (
        <AuthContext.Provider
            value={{
                user: session?.user ?? null,
                isAuthenticated: session !== null,
                isInitializing,
                logoutRedirectPending,
                acknowledgeLogoutRedirect,
                login,
                logout,
                authorizedRequest,
            }}
        >
            {children}
        </AuthContext.Provider>
    )
}
