import { useEffect, useState } from 'react'

import { useAuth } from '../features/auth/useAuth'
import type { LibraryKind } from '../features/library/api'
import type { LibraryListResponse } from '../features/library/types'
import { getApiErrorMessage } from '../shared/api/errorMessages'
import { createLibraryItem, listLibraryItems } from '../features/library/api'

import { CreateLibraryItemDialog } from '../features/library/CreateLibraryItemDialog'

import type { CreateLibraryItemRequest } from '../features/library/types'

import './LibraryPage.css'

const PAGE_SIZE = 20

interface LibrarySectionProps {
    kind: LibraryKind
    title: string
    emptyMessage: string
    onAdd: () => void
}

function LibrarySection({ kind, title, emptyMessage, onAdd }: LibrarySectionProps) {
    const { authorizedRequest } = useAuth()

    const [page, setPage] = useState(1)
    const [reloadKey, setReloadKey] = useState(0)

    const [data, setData] = useState<LibraryListResponse | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let active = true

        async function load() {
            setLoading(true)
            setError(null)
            setData(null)

            try {
                const result = await listLibraryItems(authorizedRequest, kind, page, PAGE_SIZE)

                if (active) {
                    setData(result)
                }
            } catch (caughtError) {
                if (active) {
                    setError(getApiErrorMessage(caughtError))
                }
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void load()

        return () => {
            active = false
        }
    }, [authorizedRequest, kind, page, reloadKey])

    return (
        <section className="library-panel">
            <header className="library-panel-heading">
                <h2>{title}</h2>

                <div className="library-panel-actions">
                    {data && <span className="library-count">{data.total}</span>}

                    <button type="button" className="library-create-primary" onClick={onAdd}>
                        + Добавить
                    </button>
                </div>
            </header>

            {loading ? (
                <p className="library-message" role="status">
                    Загружаем…
                </p>
            ) : error ? (
                <div className="library-feedback" role="alert">
                    <p>{error}</p>

                    <button
                        type="button"
                        className="library-button"
                        onClick={() => setReloadKey((value) => value + 1)}
                    >
                        Повторить
                    </button>
                </div>
            ) : data && data.items.length === 0 ? (
                <p className="library-message">{emptyMessage}</p>
            ) : data ? (
                <>
                    <ul className="library-items">
                        {data.items.map((item) => (
                            <li key={item.id} className="library-item">
                                <div className="library-item-content">
                                    <strong>{item.name}</strong>

                                    <span className="library-item-details">
                                        Вершин контура: {item.geometry.exterior.points.length}
                                        {item.geometry.holes.length > 0 &&
                                            ` · Отверстий: ${item.geometry.holes.length}`}
                                    </span>
                                </div>

                                <span className="library-item-id">#{item.id}</span>
                            </li>
                        ))}
                    </ul>

                    {data.total_pages > 1 && (
                        <nav className="library-pagination" aria-label={`Страницы: ${title}`}>
                            <button
                                type="button"
                                className="library-button"
                                disabled={page <= 1}
                                onClick={() => setPage((value) => value - 1)}
                            >
                                Назад
                            </button>

                            <span>
                                {data.page} / {data.total_pages}
                            </span>

                            <button
                                type="button"
                                className="library-button"
                                disabled={page >= data.total_pages}
                                onClick={() => setPage((value) => value + 1)}
                            >
                                Далее
                            </button>
                        </nav>
                    )}
                </>
            ) : null}
        </section>
    )
}

export function LibraryPage() {
    const { authorizedRequest } = useAuth()

    const [creatingKind, setCreatingKind] = useState<LibraryKind | null>(null)

    const [reloadKey, setReloadKey] = useState(0)

    async function handleCreate(kind: LibraryKind, data: CreateLibraryItemRequest): Promise<void> {
        await createLibraryItem(authorizedRequest, kind, data)

        setReloadKey((current) => current + 1)
    }

    return (
        <section className="library-page">
            <header className="library-heading">
                <h1>Библиотека</h1>

                <p>
                    Здесь хранятся ваши поверхности и лекала, которые можно использовать в проектах.
                </p>
            </header>

            <div className="library-grid">
                <LibrarySection
                    key={`surfaces-${reloadKey}`}
                    kind="surfaces"
                    title="Поверхности"
                    emptyMessage="В библиотеке пока нет поверхностей."
                    onAdd={() => setCreatingKind('surfaces')}
                />

                <LibrarySection
                    key={`patterns-${reloadKey}`}
                    kind="patterns"
                    title="Лекала"
                    emptyMessage="В библиотеке пока нет лекал."
                    onAdd={() => setCreatingKind('patterns')}
                />
            </div>

            {creatingKind && (
                <CreateLibraryItemDialog
                    kind={creatingKind}
                    onClose={() => setCreatingKind(null)}
                    onSave={(data) => handleCreate(creatingKind, data)}
                />
            )}
        </section>
    )
}
