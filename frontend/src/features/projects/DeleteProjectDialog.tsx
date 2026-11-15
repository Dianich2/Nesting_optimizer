import { ConfirmDialog } from '../../components/dialogs/ConfirmDialog'
import type { Project } from './types'

interface DeleteProjectDialogProps {
    project: Project
    onClose: () => void
    onDelete: () => Promise<void>
}

export function DeleteProjectDialog({ project, onClose, onDelete }: DeleteProjectDialogProps) {
    return (
        <ConfirmDialog title="Удалить проект?" confirmLabel="Удалить проект"
            onClose={onClose} onConfirm={onDelete}>
            <p>Вы собираетесь удалить проект <strong>«{project.name}»</strong>.</p>
            <p>Проект исчезнет из списка, а доступ к его данным через обычный интерфейс будет закрыт.</p>
        </ConfirmDialog>
    )
}
