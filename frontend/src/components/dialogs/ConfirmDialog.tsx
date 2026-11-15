import { useEffect, useId, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { getApiErrorMessage } from '../../shared/api/errorMessages'
import '../../features/projects/ProjectDialog.css'

interface Props {
    title: string
    children: ReactNode
    confirmLabel?: string
    onConfirm: () => Promise<void>
    onClose: () => void
}

export function ConfirmDialog({ title, children, confirmLabel = 'Удалить', onConfirm, onClose }: Props) {
    const dialog = useRef<HTMLDialogElement>(null)
    const pending = useRef(false)
    const titleId = useId()
    const descriptionId = useId()
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        dialog.current?.showModal()
    }, [])

    async function confirm() {
        if (pending.current) return
        pending.current = true
        setBusy(true)
        setError(null)
        try {
            await onConfirm()
            dialog.current?.close()
        } catch (caught) {
            setError(getApiErrorMessage(caught))
        } finally {
            pending.current = false
            setBusy(false)
        }
    }

    return (
        <dialog ref={dialog} className="project-dialog" aria-labelledby={titleId}
            aria-describedby={descriptionId} onClose={(event) => { event.stopPropagation(); onClose() }}
            onCancel={(event) => { event.stopPropagation(); if (pending.current) event.preventDefault() }}>
            <div className="project-delete-content">
                <header className="project-dialog-header">
                    <h2 id={titleId}>{title}</h2>
                    <div id={descriptionId}>{children}</div>
                </header>
                {error && <p className="project-dialog-error" role="alert">{error}</p>}
                <footer className="project-dialog-actions">
                    <button type="button" className="projects-button" autoFocus disabled={busy}
                        onClick={() => dialog.current?.close()}>Отмена</button>
                    <button type="button" className="projects-button projects-create-button" disabled={busy}
                        onClick={() => void confirm()}>{busy ? 'Выполняем…' : confirmLabel}</button>
                </footer>
            </div>
        </dialog>
    )
}
