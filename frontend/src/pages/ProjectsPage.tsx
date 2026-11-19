
import { useEffect, useState } from 'react'
import { Link } from 'react-router'

import { useAuth } from '../features/auth/useAuth'
import type {
    CreateProjectRequest,
    Project,
    ProjectsListResponse,
} from '../features/projects/types'

import { getApiErrorMessage } from '../shared/api/errorMessages'
import {
    createProject,
    deleteProject,
    listProjects,
    updateProject,
} from '../features/projects/api'

import { ProjectDialog } from '../features/projects/ProjectDialog'
import { DeleteProjectDialog } from '../features/projects/DeleteProjectDialog'

import './ProjectsPage.css'

const PAGE_SIZE = 20

const dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
})

export function ProjectsPage() {
    const { authorizedRequest } = useAuth()

    const [page, setPage] = useState(1)
    const [reloadKey, setReloadKey] = useState(0)

    const [data, setData] = useState<ProjectsListResponse | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const [isCreateOpen, setIsCreateOpen] = useState(false)
    const [editingProject, setEditingProject] = useState<Project | null>(null)
    const [deletingProject, setDeletingProject] = useState<Project | null>(null)

    useEffect(() => {
        let active = true

        async function loadProjects() {
            setIsLoading(true)
            setError(null)

            try {
                const result = await listProjects(
                    authorizedRequest,
                    page,
                    PAGE_SIZE,
                )

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

        void loadProjects()

        return () => {
            active = false
        }
    }, [authorizedRequest, page, reloadKey])

    function retry() {
        setReloadKey((current) => current + 1)
    }

    async function handleCreateProject(
        input: CreateProjectRequest,
    ): Promise<void> {
        await createProject(authorizedRequest, input)

        setPage(1)
        setReloadKey((current) => current + 1)
    }

    async function handleUpdateProject(
        projectId: number,
        input: CreateProjectRequest,
    ): Promise<void> {
        await updateProject(authorizedRequest, projectId, input)

        setReloadKey((current) => current + 1)
    }

    async function handleDeleteProject(
        projectId: number,
    ): Promise<void> {
        await deleteProject(authorizedRequest, projectId)

        if (page > 1 && data?.items.length === 1) {
            setPage((current) => current - 1)
        }

        setReloadKey((current) => current + 1)
    }

    return (
        <section className="projects-page">
            <header className="projects-heading">
                <div>
                    <h1>Мои проекты</h1>

                    <p>
                        Управляйте проектами раскроя материалов.
                    </p>
                </div>

                <button
                    type="button"
                    className="projects-button projects-create-button"
                    onClick={() => setIsCreateOpen(true)}
                >
                    + Новый проект
                </button>
            </header>

            {isCreateOpen && (
                <ProjectDialog
                    onClose={() => setIsCreateOpen(false)}
                    onSave={handleCreateProject}
                />
            )}


            {editingProject && (
                <ProjectDialog
                    key={editingProject.id}
                    project={editingProject}
                    onClose={() => setEditingProject(null)}
                    onSave={(input) =>
                        handleUpdateProject(editingProject.id, input)
                    }
                />
            )}

            {deletingProject && (
                <DeleteProjectDialog
                    project={deletingProject}
                    onClose={() => setDeletingProject(null)}
                    onDelete={() => handleDeleteProject(deletingProject.id)}
                />
            )}

            {isLoading ? (
                <div className="projects-state" role="status">
                    <span className="projects-spinner" aria-hidden="true" />
                    <p>Загружаем проекты…</p>
                </div>
            ) : error ? (
                <div className="projects-state" role="alert">
                    <h2>Не удалось загрузить проекты</h2>
                    <p>{error}</p>

                    <button
                        type="button"
                        className="projects-button"
                        onClick={retry}
                    >
                        Повторить
                    </button>
                </div>
            ) : data && data.items.length === 0 ? (
                <div className="projects-state">
                    <div className="projects-empty-icon" aria-hidden="true">
                        ◇
                    </div>

                    <h2>Пока нет проектов</h2>

                    <p>
                        Здесь появятся ваши проекты.
                        Скоро вы сможете создать первый проект
                        и приступить к раскрою.
                    </p>
                </div>
            ) : data ? (
                <>
                    <div className="projects-summary">
                        Всего проектов: {data.total}
                    </div>

                    <div className="projects-grid">
                        {data.items.map((project) => (
                            <article
                                className="project-card"
                                key={project.id}
                            >
                                <div className="project-card-main">

                                    <div className="project-card-top">
                                        <span className="project-card-label">
                                            Проект № {project.id}
                                        </span>

                                        <div className="project-card-actions">
                                            <button
                                                type="button"
                                                className="project-card-edit"
                                                onClick={() => setEditingProject(project)}
                                                aria-label={`Изменить проект «${project.name}»`}
                                                title="Изменить проект"
                                            >
                                                <svg
                                                    width="18"
                                                    height="18"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    strokeWidth="1.8"
                                                    strokeLinecap="round"
                                                    strokeLinejoin="round"
                                                    aria-hidden="true"
                                                >
                                                    <path d="M12 20h9" />
                                                    <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
                                                </svg>
                                            </button>
                                            <button
                                                type="button"
                                                className="project-card-edit project-card-delete"
                                                onClick={() => setDeletingProject(project)}
                                                aria-label={`Удалить проект «${project.name}»`}
                                                title="Удалить проект"
                                            >
                                                <svg
                                                    width="18"
                                                    height="18"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    strokeWidth="1.8"
                                                    strokeLinecap="round"
                                                    strokeLinejoin="round"
                                                    aria-hidden="true"
                                                >
                                                    <path d="M3 6h18" />
                                                    <path d="M8 6V4h8v2" />
                                                    <path d="M5 6l1 14h12l1-14" />
                                                    <path d="M10 10v6" />
                                                    <path d="M14 10v6" />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>

                                    <h2>{project.name}</h2>

                                    <p className="project-card-description">
                                        {project.description ||
                                            'Описание не добавлено'}
                                    </p>
                                </div>

                                <div className="project-card-footer">
                                    <span className="project-card-date">
                                        Создан: {dateFormatter.format(
                                            new Date(project.created_at),
                                        )}
                                    </span>

                                    <Link
                                        to={`/projects/${project.id}/editor`}
                                        className="project-card-link"
                                    >
                                        Открыть проект
                                    </Link>
                                </div>
                            </article>
                        ))}
                    </div>

                    {data.total_pages > 1 && (
                        <nav
                            className="projects-pagination"
                            aria-label="Страницы проектов"
                        >
                            <button
                                type="button"
                                className="projects-button"
                                disabled={page <= 1}
                                onClick={() => setPage((value) => value - 1)}
                            >
                                Назад
                            </button>

                            <span>
                                Страница {data.page} из {data.total_pages}
                            </span>

                            <button
                                type="button"
                                className="projects-button"
                                disabled={page >= data.total_pages}
                                onClick={() => setPage((value) => value + 1)}
                            >
                                Далее
                            </button>
                        </nav>
                    )}
                </>
            ) : null
            }
        </section >
    )
}
