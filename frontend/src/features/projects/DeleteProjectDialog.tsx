
import { useEffect, useRef, useState } from 'react'

import { getApiErrorMessage } from '../../shared/api/errorMessages'
import type { Project } from './types'

import './ProjectDialog.css'

interface DeleteProjectDialogProps {
    project: Project
    onClose: () => void
    onDelete: () => Promise<void>
}

export function DeleteProjectDialog({
    project,
    onClose,
    onDelete,
}: DeleteProjectDialogProps) {
    const dialogRef = useRef<HTMLDialogElement>(null)

    const [isDeleting, setIsDeleting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const dialog = dialogRef.current

        if (dialog && !dialog.open) {
            dialog.showModal()
        }
    }, [])

    async function handleDelete() {
        if (isDeleting) {
            return
        }

        setIsDeleting(true)
        setError(null)

        try {
            await onDelete()
            dialogRef.current?.close()
        } catch (caughtError) {
            setError(getApiErrorMessage(caughtError))
        } finally {
            setIsDeleting(false)
        }
    }

    return (
        <dialog
            ref={dialogRef}
            className="project-dialog"
            aria-labelledby="delete-project-title"
            aria-describedby="delete-project-description"
            onClose={onClose}
            onCancel={(event) => {
                if (isDeleting) {
                    event.preventDefault()
                }
            }}
            onKeyDown={(event) => {
                if (event.key === 'Escape' && !isDeleting) {
                    event.preventDefault()
                    dialogRef.current?.close()
                }
            }}

        >
            <div className="project-delete-content">
                <header className="project-dialog-header">
                    <h2 id="delete-project-title">
                        Удалить проект?
                    </h2>

                    <p id="delete-project-description">
                        Вы собираетесь удалить проект
                        {' '}
                        <strong>«{project.name}»</strong>.
                    </p>

                    <p>
                        Проект исчезнет из списка, а доступ
                        к его данным через обычный интерфейс
                        будет закрыт.
                    </p>
                </header>

                {error && (
                    <p className="project-dialog-error" role="alert">
                        {error}
                    </p>
                )}

                <footer className="project-dialog-actions">
                    <button
                        type="button"
                        className="projects-button"
                        disabled={isDeleting}
                        onClick={() => dialogRef.current?.close()}
                    >
                        Отмена
                    </button>

                    <button
                        type="button"
                        className="projects-button projects-create-button"
                        disabled={isDeleting}
                        onClick={() => void handleDelete()}
                    >
                        {isDeleting
                            ? 'Удаляем…'
                            : 'Удалить проект'}
                    </button>
                </footer>
            </div>
        </dialog>
    )
}
