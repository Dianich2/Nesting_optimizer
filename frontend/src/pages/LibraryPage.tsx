export function LibraryPage() {
    return (
        <section>
            <h1>Библиотека</h1>

            <p style={{ color: 'var(--color-text-muted)' }}>
                Управляйте своими поверхностями и лекалами.
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
                <h2>Мои объекты</h2>

                <p>
                    Здесь будут доступны создание по точкам,
                    импорт SVG и просмотр геометрии.
                </p>
            </div>
        </section>
    )
}