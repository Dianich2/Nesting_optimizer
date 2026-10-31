import { Navigate, Outlet, useLocation } from 'react-router'

import { useAuth } from '../../features/auth/useAuth'

export function ProtectedRoute() {
    const {
        isAuthenticated,
        isInitializing,
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
