import { useEffect, useState } from 'react'

import { useAuth } from '../auth/useAuth'
import { getApiErrorMessage } from '../../shared/api/errorMessages'

import { listProjectPatterns, listProjectSurfaces } from './api'

import type { PaginatedResponse, ProjectResource } from './types'

import './ProjectResources.css'

const PAGE_SIZE = 20

type ResourceKind = 'surfaces' | 'patterns'

interface ResourceListProps {
    projectId: number
    kind: ResourceKind
    title: string
    emptyMessage: string
}

function ResourceList({ projectId, kind, title, emptyMessage }: ResourceListProps) {
    const { authorizedRequest } = useAuth()

    const [page, setPage] = useState(1)
    const [reloadKey, setReloadKey] = useState(0)

    const [data, setData] = useState<PaginatedResponse<ProjectResource> | null>(null)

    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let active = true

        async function loadResources() {
            setIsLoading(true)
            setError(null)
            setData(null)

            try {
                const result =
                    kind === 'surfaces'
                        ? await listProjectSurfaces(authorizedRequest, projectId, page, PAGE_SIZE)
                        : await listProjectPatterns(authorizedRequest, projectId, page, PAGE_SIZE)

                if (active) {
                    setData(result)
                }
            } catch (caughtError) {
                if (active) {
                    setError(getApiErrorMessage(caughtError))
                }
            } finally {
                if (active) {
                    setIsLoading(false)
                }
            }
        }

        void loadResources()

        return () => {
            active = false
        }
    }, [authorizedRequest, projectId, kind, page, reloadKey])

    return (
        <section className="project-resource-panel">
            <header className="project-resource-header">
                <h3>{title}</h3>

                {data && <span className="project-resource-count">{data.total}</span>}
            </header>

            {isLoading ? (
                <p className="project-resource-message" role="status">
                    Загружаем…
                </p>
            ) : error ? (
                <div className="project-resource-feedback" role="alert">
                    <p>{error}</p>

                    <button
                        type="button"
                        className="projects-button"
                        onClick={() => setReloadKey((current) => current + 1)}
                    >
                        Повторить
                    </button>
                </div>
            ) : data && data.items.length === 0 ? (
                <p className="project-resource-message">{emptyMessage}</p>
            ) : data ? (
                <>
                    <ul className="project-resource-list">
                        {data.items.map((item) => (
                            <li key={item.id} className="project-resource-item">
                                <span className="project-resource-name">{item.name}</span>

                                <span className="project-resource-id">#{item.id}</span>
                            </li>
                        ))}
                    </ul>

                    {data.total_pages > 1 && (
                        <nav
                            className="project-resource-pagination"
                            aria-label={`Страницы: ${title}`}
                        >
                            <button
                                type="button"
                                className="projects-button"
                                disabled={page <= 1}
                                onClick={() => setPage((current) => current - 1)}
                            >
                                Назад
                            </button>

                            <span>
                                {data.page} / {data.total_pages}
                            </span>

                            <button
                                type="button"
                                className="projects-button"
                                disabled={page >= data.total_pages}
                                onClick={() => setPage((current) => current + 1)}
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

interface ProjectResourcesProps {
    projectId: number
}

export function ProjectResources({ projectId }: ProjectResourcesProps) {
    return (
        <div className="project-resources">
            <ResourceList
                key={`surfaces-${projectId}`}
                projectId={projectId}
                kind="surfaces"
                title="Поверхности"
                emptyMessage="В этом проекте пока нет поверхностей."
            />

            <ResourceList
                key={`patterns-${projectId}`}
                projectId={projectId}
                kind="patterns"
                title="Лекала"
                emptyMessage="В этом проекте пока нет лекал."
            />
        </div>
    )
}
