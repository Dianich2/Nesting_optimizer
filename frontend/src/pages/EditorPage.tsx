
import { useParams } from 'react-router'
import { BackLink } from '../components/navigation/BackLink'

export function EditorPage() {
    const { projectId } = useParams()

    return (
        <section>
            <BackLink to="/projects">
                К проектам
            </BackLink>
            <h2>Редактор раскроя</h2>
            <p>Проект № {projectId}</p>
            <p>Здесь появится рабочая область раскроя.</p>
        </section>
    )
}
