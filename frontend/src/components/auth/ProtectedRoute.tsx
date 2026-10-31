import { Navigate, Outlet, useLocation } from 'react-router'

import { useAuth } from '../../features/auth/useAuth'
import { AuthLayout } from './AuthLayout'

export function ProtectedRoute() {
    const {
        isAuthenticated,
        isInitializing,
        restoreError,
        retryRestoreSession,
        logoutRedirectPending,
    } = useAuth()

    const location = useLocation()

    if (isInitializing) {
        return (
            <div role="status" aria-live="polite">
                Проверяем сессию…
            </div>
        )
    }


    if (restoreError) {
        return (
            <AuthLayout>
                <section className="auth-card" role="alert">
                    <h1>Нет подключения к серверу</h1>

                    <p className="auth-description">
                        Не удалось проверить вашу сессию.
                        Проверьте подключение и попробуйте ещё раз.
                    </p>

                    <button
                        type="button"
                        className="auth-submit"
                        onClick={retryRestoreSession}
                    >
                        Повторить подключение
                    </button>
                </section>
            </AuthLayout>
        )
    }

    if (!isAuthenticated) {
        if (logoutRedirectPending) {
            return <Navigate to="/login" replace state={null} />
        }

        return (
            <Navigate
                to="/login"
                replace
                state={{
                    from: location.pathname + location.search + location.hash,
                }}
            />
        )
    }

    return <Outlet />
}
