import { NavLink, Outlet } from 'react-router'

import { useTheme } from '../../app/useTheme'
import './AppLayout.css'
import { useState } from 'react'

import { useAuth } from '../../features/auth/useAuth'
import { getApiErrorMessage } from '../../shared/api/errorMessages'

export function AppLayout() {
    const { theme, toggleTheme } = useTheme()

    const { user, logout } = useAuth()

    const [isLoggingOut, setIsLoggingOut] = useState(false)
    const [logoutError, setLogoutError] = useState<string | null>(null)

    async function handleLogout() {
        if (isLoggingOut) {
            return
        }

        setLogoutError(null)
        setIsLoggingOut(true)

        try {
            await logout()
        } catch (error) {
            setLogoutError(getApiErrorMessage(error))
        } finally {
            setIsLoggingOut(false)
        }
    }


    return (
        <div className="app-shell">
            <header className="app-header">
                <NavLink to="/projects" className="app-brand">
                    <span className="app-brand-mark">N</span>
                    <span>Nesting Optimizer</span>
                </NavLink>

                <nav className="app-navigation" aria-label="Основная навигация">
                    <NavLink
                        to="/projects"
                        className={({ isActive }) =>
                            `app-nav-link ${isActive ? 'app-nav-link-active' : ''}`
                        }
                    >
                        Проекты
                    </NavLink>

                    <NavLink
                        to="/library"
                        className={({ isActive }) =>
                            `app-nav-link ${isActive ? 'app-nav-link-active' : ''}`
                        }
                    >
                        Библиотека
                    </NavLink>
                </nav>

                <div className="app-header-actions">
                    <button
                        type="button"
                        className="theme-toggle"
                        onClick={toggleTheme}
                        aria-label={
                            theme === 'dark'
                                ? 'Включить светлую тему'
                                : 'Включить тёмную тему'
                        }
                    >
                        {theme === 'dark' ? '☀' : '☾'}
                    </button>
                </div>

                <div className="app-user-actions">
                    <span className="app-user-name">
                        {user?.first_name || user?.login}
                    </span>

                    <button
                        type="button"
                        className="app-logout-button"
                        onClick={handleLogout}
                        disabled={isLoggingOut}
                    >
                        {isLoggingOut ? 'Выходим…' : 'Выйти'}
                    </button>
                </div>

            </header>

            <main className="app-content">
                {logoutError && (
                    <div className="app-logout-error" role="alert">
                        {logoutError}
                    </div>
                )}
                <Outlet />
            </main>
        </div>
    )
}