import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../auth/useAuth'
import { listLibraryItems } from '../library/api'
import type { LibraryKind } from '../library/api'
import type { LibraryListResponse } from '../library/types'
import { getApiErrorMessage } from '../../shared/api/errorMessages'
import { createProjectResource } from './api'
import '../projects/ProjectDialog.css'

interface Props {
    projectId: number
    kind: LibraryKind
    onClose: () => void
    onAdded: () => void
}

export function LibraryPicker({ projectId, kind, onClose, onAdded }: Props) {
    const { authorizedRequest } = useAuth()
    const dialog = useRef<HTMLDialogElement>(null)
    const pending = useRef(false)
    const [page, setPage] = useState(1)
    const [data, setData] = useState<LibraryListResponse | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [busy, setBusy] = useState(false)
    const [reload, setReload] = useState(0)
    const [scale, setScale] = useState('1')
    const [message, setMessage] = useState('')
    useEffect(() => {
        dialog.current?.showModal()
    }, [])
    useEffect(() => {
        let active = true
        listLibraryItems(authorizedRequest, kind, page, 20)
            .then((result) => {
                if (active) setData(result)
            })
            .catch((caught) => {
                if (active) setError(getApiErrorMessage(caught))
            })
        return () => {
            active = false
        }
    }, [authorizedRequest, kind, page, reload])

    async function add(id: number, name: string) {
        if (pending.current) return
        const value = Number(scale)
        if (!scale.trim() || !Number.isFinite(value) || value <= 0) {
            setError('Укажите положительный масштаб геометрии.')
            return
        }
        pending.current = true
        setBusy(true)
        setError(null)
        try {
            const item = await createProjectResource(authorizedRequest, projectId, kind, id, value)
            setMessage(`Добавлен экземпляр «${name}» №${item.id}. Можно добавить ещё один.`)
            onAdded()
        } catch (caught) {
            setError(getApiErrorMessage(caught))
        } finally {
            pending.current = false
            setBusy(false)
        }
    }
    return (
        <dialog
            ref={dialog}
            className="project-dialog"
            aria-labelledby="library-picker-title"
            onClose={onClose}
            onCancel={(event) => {
                if (pending.current) event.preventDefault()
            }}
        >
            <div className="project-delete-content">
                <h2 id="library-picker-title">
                    {kind === 'surfaces' ? 'Поверхности библиотеки' : 'Лекала библиотеки'}
                </h2>
                <p>Каждое добавление создаёт независимый экземпляр в проекте.</p>
                <label>
                    Масштаб геометрии{' '}
                    <input
                        type="number"
                        step="any"
                        min="0"
                        value={scale}
                        disabled={busy}
                        onChange={(event) => setScale(event.target.value)}
                    />
                </label>
                {error && (
                    <div role="alert">
                        <p>{error}</p>
                        <button type="button" disabled={busy} onClick={() => setReload(reload + 1)}>
                            Повторить загрузку
                        </button>
                    </div>
                )}
                {message && <p role="status">{message}</p>}
                {!data && !error && <p role="status">Загружаем библиотеку…</p>}
                {data && (
                    <>
                        {data.items.length === 0 && <p>Библиотека пуста.</p>}
                        <ul className="project-resource-list">
                            {data.items.map((item) => (
                                <li key={item.id} className="project-resource-item">
                                    <span>
                                        {item.name} · №{item.id}
                                    </span>
                                    <button
                                        type="button"
                                        className="projects-button"
                                        disabled={busy}
                                        onClick={() => void add(item.id, item.name)}
                                    >
                                        Добавить
                                    </button>
                                </li>
                            ))}
                        </ul>
                        {data.total_pages > 1 && (
                            <nav aria-label="Страницы библиотеки">
                                <button
                                    type="button"
                                    disabled={busy || page <= 1}
                                    onClick={() => {
                                        setData(null)
                                        setError(null)
                                        setPage(page - 1)
                                    }}
                                >
                                    Назад
                                </button>
                                <span>
                                    {' '}
                                    {data.page} / {data.total_pages}{' '}
                                </span>
                                <button
                                    type="button"
                                    disabled={busy || page >= data.total_pages}
                                    onClick={() => {
                                        setData(null)
                                        setError(null)
                                        setPage(page + 1)
                                    }}
                                >
                                    Далее
                                </button>
                            </nav>
                        )}
                    </>
                )}
                <button
                    type="button"
                    className="projects-button"
                    disabled={busy}
                    onClick={() => dialog.current?.close()}
                >
                    Закрыть
                </button>
            </div>
        </dialog>
    )
}
