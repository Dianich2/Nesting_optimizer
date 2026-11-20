
import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router'

import { AuthLayout } from '../components/auth/AuthLayout'
import { registerUser } from '../features/auth/api'
import { getApiErrorMessage } from '../shared/api/errorMessages'

export function RegisterPage() {
    const navigate = useNavigate()

    const [login, setLogin] = useState('')
    const [email, setEmail] = useState('')
    const [firstName, setFirstName] = useState('')
    const [lastName, setLastName] = useState('')
    const [password, setPassword] = useState('')
    const [repeatPassword, setRepeatPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)

    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        if (isSubmitting) {
            return
        }

        setError(null)

        if (password !== repeatPassword) {
            setError('Пароли не совпадают.')
            return
        }

        if (password.length < 8) {
            setError('Пароль должен содержать не менее 8 символов.')
            return
        }

        if (new TextEncoder().encode(password).length > 72) {
            setError('Пароль слишком длинный: допускается не более 72 байт UTF-8.')
            return
        }

        setIsSubmitting(true)

        try {
            await registerUser({
                login: login.trim().toLowerCase(),
                email: email.trim().toLowerCase(),
                first_name: firstName.trim(),
                last_name: lastName.trim(),
                password,
            })

            navigate('/login?registered=1', { replace: true })
        } catch (caughtError) {
            setError(getApiErrorMessage(caughtError, 'register'))
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <AuthLayout>
            <div className="auth-card">
                <h1>Создание аккаунта</h1>

                <p className="auth-description">
                    Зарегистрируйтесь, чтобы создавать проекты и работать
                    с поверхностями и лекалами.
                </p>

                <form className="auth-form" onSubmit={handleSubmit}>
                    <div className="auth-field">
                        <label htmlFor="register-login">Логин</label>
                        <input
                            id="register-login"
                            className="auth-input"
                            type="text"
                            autoComplete="username"
                            maxLength={100}
                            pattern="[a-z0-9](?:[a-z0-9._\-]*[a-z0-9])?"
                            title="Используйте латинские строчные буквы, цифры, точку, дефис или подчёркивание. Первый и последний символ — буква или цифра."
                            value={login}
                            onChange={(event) => setLogin(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
                        <small className="auth-hint">
                            Латинские буквы, цифры, точка, дефис и подчёркивание.
                        </small>
                    </div>

                    <div className="auth-field">
                        <label htmlFor="register-email">Электронная почта</label>
                        <input
                            id="register-email"
                            className="auth-input"
                            type="email"
                            autoComplete="email"
                            maxLength={254}
                            value={email}
                            onChange={(event) => setEmail(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
                    </div>

                    <div className="auth-field">
                        <label htmlFor="register-first-name">Имя</label>
                        <input
                            id="register-first-name"
                            className="auth-input"
                            type="text"
                            autoComplete="given-name"
                            maxLength={50}
                            value={firstName}
                            onChange={(event) => setFirstName(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
                    </div>

                    <div className="auth-field">
                        <label htmlFor="register-last-name">Фамилия</label>
                        <input
                            id="register-last-name"
                            className="auth-input"
                            type="text"
                            autoComplete="family-name"
                            maxLength={50}
                            value={lastName}
                            onChange={(event) => setLastName(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
                    </div>

                    <div className="auth-field">
                        <label htmlFor="register-password">Пароль</label>

                        <div className="auth-password-field">
                            <input
                                id="register-password"
                                className="auth-input"
                                type={showPassword ? 'text' : 'password'}
                                autoComplete="new-password"
                                minLength={8}
                                value={password}
                                onChange={(event) => setPassword(event.target.value)}
                                required
                                disabled={isSubmitting}
                            />

                            <button
                                type="button"
                                className="auth-password-toggle"
                                onClick={() => setShowPassword((value) => !value)}
                                aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                                aria-pressed={showPassword}
                            >
                                {showPassword ? 'Скрыть' : 'Показать'}
                            </button>
                        </div>

                        <small className="auth-hint">
                            Не менее 8 символов, не более 72 байт UTF-8.
                        </small>
                    </div>

                    <div className="auth-field">
                        <label htmlFor="register-repeat-password">
                            Повторите пароль
                        </label>
                        <input
                            id="register-repeat-password"
                            className="auth-input"
                            type={showPassword ? 'text' : 'password'}
                            autoComplete="new-password"
                            value={repeatPassword}
                            onChange={(event) => setRepeatPassword(event.target.value)}
                            required
                            disabled={isSubmitting}
                        />
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
                        {isSubmitting ? 'Создаём аккаунт…' : 'Зарегистрироваться'}
                    </button>
                </form>

                <p className="auth-footer">
                    Уже есть аккаунт? <Link to="/login">Войти</Link>
                </p>
            </div>
        </AuthLayout>
    )
}
