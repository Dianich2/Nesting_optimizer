
import type { ReactNode } from 'react'
import { useTheme } from '../../app/useTheme'

import './AuthLayout.css'

interface AuthLayoutProps {
    children: ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
    const { theme, toggleTheme } = useTheme()

    return (
        <div className="auth-layout">
            <header className="auth-header">
                <div className="auth-brand">
                    <span className="auth-brand-mark">N</span>
                    <span>Nesting Optimizer</span>
                </div>

                <button
                    type="button"
                    className="theme-toggle"
                    onClick={toggleTheme}
                    aria-label={
                        theme === 'dark'
                            ? 'Включить светлую тему'
                            : 'Включить тёмную тему'
                    }
                    title="Переключить тему"
                >
                    {theme === 'dark' ? '☀' : '☾'}
                </button>
            </header>

            <main className="auth-main">
                {children}
            </main>
        </div>
    )
}
