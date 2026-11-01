import { useEffect, useState } from 'react'
import { useParams } from 'react-router'

import { BackLink } from '../components/navigation/BackLink'
import { useAuth } from '../features/auth/useAuth'
import { getProject } from '../features/projects/api'
import type { Project } from '../features/projects/types'

import { ApiError } from '../shared/api/http'
import { getApiErrorMessage } from '../shared/api/errorMessages'
import { ProjectResources } from '../features/project-resources/ProjectResources'

import './EditorPage.css'

const dateFormatter = new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
})

export function EditorPage() {
    const { projectId } = useParams()
    const { authorizedRequest } = useAuth()

    const [project, setProject] = useState<Project | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [reloadKey, setReloadKey] = useState(0)

    useEffect(() => {
        let active = true

        async function loadProject() {
            setIsLoading(true)
            setError(null)
            setProject(null)

            const id = Number(projectId)

            if (
                !projectId ||
                !/^[1-9]\d*$/.test(projectId) ||
                !Number.isSafeInteger(id)
            ) {
                setError('Некорректный идентификатор проекта.')
                setIsLoading(false)
                return
            }

            try {
                const result = await getProject(authorizedRequest, id)

                if (active) {
                    setProject(result)
                }
            } catch (caughtError) {
                if (active) {
                    if (
                        caughtError instanceof ApiError &&
                        caughtError.status === 404
                    ) {
                        setError('Проект не найден или недоступен.')
                    } else {
                        setError(getApiErrorMessage(caughtError))
                    }
                }
            } finally {
                if (active) {
                    setIsLoading(false)
                }
            }
        }

        void loadProject()

        return () => {
            active = false
        }
    }, [authorizedRequest, projectId, reloadKey])

    return (
        <section className="editor-page">
            <div className="editor-toolbar">
                <BackLink to="/projects">К проектам</BackLink>

                {project && (
                    <span className="editor-date">
                        Создан: {dateFormatter.format(new Date(project.created_at))}
                    </span>
                )}
            </div>

            {isLoading ? (
                <div className="editor-state" role="status">
                    <p>Загружаем проект…</p>
                </div>
            ) : error ? (
                <div className="editor-state" role="alert">
                    <h2>Не удалось открыть проект</h2>
                    <p>{error}</p>

                    <button
                        type="button"
                        className="projects-button"
                        onClick={() => setReloadKey((current) => current + 1)}
                    >
                        Повторить
                    </button>
                </div>
            ) : project ? (
                <>
                    <header className="editor-heading">
                        <span className="editor-project-id">
                            Проект #{project.id}
                        </span>

                        <h1>{project.name}</h1>

                        <p>
                            {project.description || 'Описание проекта не добавлено.'}
                        </p>
                    </header>

                    <div className="editor-workspace">
    <h2>Состав проекта</h2>

    <p>
        Поверхности и лекала, добавленные в этот проект.
    </p>

    <ProjectResources projectId={project.id} />
</div>
                </>
            ) : null}
        </section>
    )
}