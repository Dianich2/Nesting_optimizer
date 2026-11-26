import { useEffect, useId, useRef, useState } from 'react'
import { getApiErrorMessage } from '../../shared/api/errorMessages'
import '../projects/ProjectDialog.css'

type Props = {
    newPlacement: boolean
    onContinue: () => void
    onDiscard: () => void
    onDelete: () => Promise<void>
}

// One pending manual placement at a time. The original placement remains in
// PostgreSQL until the user explicitly chooses to delete it.
export function ManualDraftDialog({ newPlacement, onContinue, onDiscard, onDelete }: Props) {
    const dialog = useRef<HTMLDialogElement>(null)
    const titleId = useId()
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState<string | null>(null)
    useEffect(() => { dialog.current?.showModal() }, [])
    async function remove() {
        if (busy) return
        setBusy(true)
        setError(null)
        try { await onDelete() }
        catch (caught) { setError(getApiErrorMessage(caught)) }
        finally { setBusy(false) }
    }
    return (
        <dialog ref={dialog} className="project-dialog" aria-labelledby={titleId}
            onCancel={(event) => { event.preventDefault(); if (!busy) onContinue() }}>
            <div className="project-delete-content">
                <header className="project-dialog-header">
                    <h2 id={titleId}>Есть несохранённое размещение</h2>
                    <p>
                        {newPlacement
                            ? 'Новое лекало ещё не сохранено. Сначала исправьте его положение или отмените добавление.'
                            : 'Положение лекала изменено, но не сохранено. На сервере остаётся прежняя корректная позиция.'}
                    </p>
                </header>
                {error && <p role="alert" className="project-dialog-error">{error}</p>}
                <footer className="project-dialog-actions manual-draft-dialog-actions">
                    <button type="button" className="projects-button" disabled={busy} onClick={onContinue}>
                        Продолжить редактирование
                    </button>
                    <button type="button" className="projects-button" disabled={busy} onClick={onDiscard}>
                        {newPlacement ? 'Отменить добавление' : 'Вернуть прежнюю позицию'}
                    </button>
                    {!newPlacement && <button type="button" className="projects-button" disabled={busy}
                        onClick={() => void remove()}>
                        {busy ? 'Удаляем…' : 'Удалить размещение'}
                    </button>}
                </footer>
            </div>
        </dialog>
    )
}
