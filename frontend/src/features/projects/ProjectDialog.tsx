
import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'

import { getApiErrorMessage } from '../../shared/api/errorMessages'
import type { CreateProjectRequest, Project } from './types'

import './ProjectDialog.css'

interface ProjectDialogProps {
    project?: Project
    onClose: () => void
    onSave: (data: CreateProjectRequest) => Promise<void>
}

export function ProjectDialog({
    project,
    onClose,
    onSave,
}: ProjectDialogProps) {
    const dialogRef = useRef<HTMLDialogElement>(null)

    const isEditing = project !== undefined

    const [name, setName] = useState(project?.name ?? '')
    const [description, setDescription] = useState(
        project?.description ?? '',
    )
    const [isSaving, setIsSaving] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const dialog = dialogRef.current

        if (dialog && !dialog.open) {
            dialog.showModal()
        }
    }, [])

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()

        const trimmedName = name.trim()

        if (!trimmedName || isSaving) {
            return
        }

        setIsSaving(true)
        setError(null)

        try {
            await onSave({
                name: trimmedName,
                description: description.trim(),
            })

            dialogRef.current?.close()
        } catch (caughtError) {
            setError(getApiErrorMessage(caughtError))
        } finally {
            setIsSaving(false)
        }
    }

    return (
        <dialog
            ref={dialogRef}
            className="project-dialog"
            aria-labelledby="project-dialog-title"
            onClose={onClose}
            onCancel={(event) => {
                event.preventDefault()

                if (!isSaving) {
                    dialogRef.current?.close()
                }
            }}
            onKeyDown={(event) => {
                if (event.key === 'Escape') {
                    event.preventDefault()

                    if (!isSaving) {
                        dialogRef.current?.close()
                    }
                }
            }}
        >
            <form onSubmit={(event) => void handleSubmit(event)}>
                <header className="project-dialog-header">
                    <h2 id="project-dialog-title">
                        {isEditing ? 'Изменить проект' : 'Новый проект'}
                    </h2>

                    <p>
                        {isEditing
                            ? 'Измените название или описание проекта.'
                            : 'Укажите название и при необходимости добавьте описание.'}
                    </p>
                </header>

                <div className="project-dialog-field">
                    <label htmlFor="project-name">
                        Название проекта
                    </label>

                    <input
                        id="project-name"
                        type="text"
                        value={name}
                        onChange={(event) => setName(event.target.value)}
                        maxLength={150}
                        required
                        autoFocus
                        disabled={isSaving}
                        placeholder="Название проекта"
                    />
                </div>

                <div className="project-dialog-field">
                    <label htmlFor="project-description">
                        Описание
                    </label>

                    <textarea
                        id="project-description"
                        value={description}
                        onChange={(event) =>
                            setDescription(event.target.value)
                        }
                        maxLength={2000}
                        rows={4}
                        disabled={isSaving}
                        placeholder="Необязательно"
                    />
                </div>

                {error && (
                    <p className="project-dialog-error" role="alert">
                        {error}
                    </p>
                )}

                <footer className="project-dialog-actions">
                    <button
                        type="button"
                        className="projects-button"
                        onClick={() => dialogRef.current?.close()}
                        disabled={isSaving}
                    >
                        Отмена
                    </button>

                    <button
                        type="submit"
                        className="projects-button projects-create-button"
                        disabled={isSaving || !name.trim()}
                    >
                        {isSaving
                            ? 'Сохраняем…'
                            : isEditing
                                ? 'Сохранить изменения'
                                : 'Создать проект'}
                    </button>
                </footer>
            </form>
        </dialog>
    )
}
