import { Link } from 'react-router'

export function ProjectsPage() {
    return (
        <section>
            <h1>Мои проекты</h1>

            <p style={{ color: 'var(--color-text-muted)' }}>
                Создавайте проекты и управляйте раскроем материалов.
            </p>

            <div
                style={{
                    marginTop: 24,
                    padding: 24,
                    borderRadius: 14,
                    border: '1px solid var(--color-border)',
                    background: 'var(--color-surface)',
                }}
            >
                <h2>Начните работу</h2>

                <p>
                    Здесь появятся ваши проекты, поверхности и лекала.
                </p>

                <Link
                    to="/projects/1/editor"
                    style={{ color: 'var(--color-accent)' }}
                >
                    Открыть демонстрационный редактор →
                </Link>
            </div>
        </section>
    )
}