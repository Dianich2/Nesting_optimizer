
import { useState, useEffect } from 'react'
import type { FormEvent } from 'react'
import {
    Link,
    useLocation,
    useNavigate,
    useSearchParams,
} from 'react-router'

import { AuthLayout } from '../components/auth/AuthLayout'
import { useAuth } from '../features/auth/useAuth'

import { getApiErrorMessage } from '../shared/api/errorMessages'

export function LoginPage() {
    const navigate = useNavigate()

    const location = useLocation()

    const requestedPath = (
        location.state as { from?: unknown } | null
    )?.from

    const redirectTo =
        typeof requestedPath === 'string' &&
            requestedPath.startsWith('/') &&
            !requestedPath.startsWith('//')
            ? requestedPath
            : '/projects'

    const [searchParams] = useSearchParams()
    const registered = searchParams.get('registered') === '1'
    const {
        login,
        logoutRedirectPending,
        acknowledgeLogoutRedirect,
    } = useAuth()

    useEffect(() => {
        if (logoutRedirectPending) {
            acknowledgeLogoutRedirect()
        }
    }, [logoutRedirectPending, acknowledgeLogoutRedirect])

    const [identifier, setIdentifier] = useState('')
    const [password, setPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        if (isSubmitting) {
            return
        }

        setError(null)
        setIsSubmitting(true)

        try {
            await login({
                identifier: identifier.trim(),
                password,
            })

            navigate(redirectTo, { replace: true })

        } catch (caughtError) {
            setError(getApiErrorMessage(caughtError, 'login'))
        } finally {
            setIsSubmitting(false)
        }

    }

    return (
        <AuthLayout>
            <div className="auth-card">
                <h1>Добро пожаловать</h1>

                <p className="auth-description">
                    Войдите, чтобы продолжить работу с проектами раскроя.
                </p>

                {registered && (
                    <div className="auth-success" role="status">
                        Аккаунт создан. Теперь войдите в систему.
                    </div>
                )}

                <form className="auth-form" onSubmit={handleSubmit}>
                    <div className="auth-field">
                        <label htmlFor="login-identifier">
                            Логин или электронная почта
                        </label>

                        <input
                            id="login-identifier"
                            className="auth-input"
                            type="text"
                            autoComplete="username"
                            value={identifier}
                            onChange={(event) => setIdentifier(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
                    </div>

                    <div className="auth-field">
                        <label htmlFor="login-password">Пароль</label>

                        <div className="auth-password-field">
                            <input
                                id="login-password"
                                className="auth-input"
                                type={showPassword ? 'text' : 'password'}
                                autoComplete="current-password"
                                value={password}
                                onChange={(event) => setPassword(event.target.value)}
                                required
                                disabled={isSubmitting}
                            />

                            <button
                                type="button"
                                className="auth-password-toggle"
                                onClick={() => setShowPassword((value) => !value)}
                                aria-label={
                                    showPassword ? 'Скрыть пароль' : 'Показать пароль'
                                }
                                aria-pressed={showPassword}
                            >
                                {showPassword ? 'Скрыть' : 'Показать'}
                            </button>
                        </div>
                    </div>

                    {error && (
                        <div className="auth-error" role="alert">
                            {error}
                        </div>
                    )}

                    <button
                        type="submit"
                        className="auth-submit"
                        disabled={isSubmitting}
                    >
                        {isSubmitting ? 'Выполняется вход…' : 'Войти'}
                    </button>
                </form>

                <p className="auth-footer">
                    Ещё нет аккаунта?{' '}
                    <Link to="/register">Зарегистрироваться</Link>
                </p>
            </div>
        </AuthLayout>
    )
}
