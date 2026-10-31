
import { useParams } from 'react-router'

export function EditorPage() {
    const { projectId } = useParams()

    return (
        <section>
            <h2>Редактор раскроя</h2>
            <p>Проект № {projectId}</p>
            <p>Здесь появится рабочая область раскроя.</p>
        </section>
    )
}
