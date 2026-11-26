import { useEffect, useState } from 'react'
import { getApiErrorMessage } from '../../shared/api/errorMessages'
import type { Request } from './api'

interface Run {
    id: number
    algorithm: string
    keep_existing: boolean
    requested_count: number
    placed_count: number
    utilization: number
    duration_ms: number
    created_at: string
}

interface HistoryPage {
    items: Run[]
    has_next: boolean
}

const PAGE_SIZE = 10

export function NestingHistory({
    request,
    base,
    revision,
    labels,
}: {
    request: Request
    base: string
    revision: number
    labels: Record<string, string>
}) {
    const [page, setPage] = useState(1)
    const [retry, setRetry] = useState(0)
    const query = `${revision}:${page}:${retry}`
    const [state, setState] = useState<{ query: string; data?: HistoryPage; error?: string }>()
    useEffect(() => {
        let active = true
        request<HistoryPage>(`${base}/nesting/history?page=${page}&page_size=${PAGE_SIZE}`)
            .then((data) => {
                if (active) setState({ query, data })
            })
            .catch((error: unknown) => {
                if (active) setState({ query, error: getApiErrorMessage(error) })
            })
        return () => {
            active = false
        }
    }, [base, page, query, request])
    const current = state?.query === query ? state : undefined
    return (
        <section aria-label="История запусков" className="nesting-history">
            <h2>История сохранённых раскладок</h2>
            <p>Здесь показываются только успешно сохранённые раскладки. Предпросмотры и отменённые расчёты пока в этот список не входят.</p>
            {!current && <p role="status">Загружаем историю…</p>}
            {current?.error && (
                <p role="alert">
                    {current.error}{' '}
                    <button onClick={() => setRetry((value) => value + 1)}>Повторить</button>
                </p>
            )}
            {current?.data && (
                <>
                    {current.data.items.length === 0 ? (
                        <p>На этой странице нет запусков.</p>
                    ) : (
                        <>
                            <p>
                                Сравнивайте показатели запусков с одинаковой геометрией поверхности,
                                набором лекал, поворотами и исходной раскладкой.
                            </p>
                            <div className="nesting-history-scroll">
                                <table>
                                    <thead>
                                        <tr>
                                            <th scope="col">Запуск</th>
                                            <th scope="col">Дата</th>
                                            <th scope="col">Алгоритм</th>
                                            <th scope="col">Размещено / запрошено</th>
                                            <th scope="col">Использование</th>
                                            <th scope="col">Время</th>
                                            <th scope="col">Режим</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {current.data.items.map((run) => (
                                            <tr key={run.id}>
                                                <td>№{run.id}</td>
                                                <td>
                                                    {new Date(run.created_at).toLocaleString(
                                                        'ru-RU',
                                                    )}
                                                </td>
                                                <td>{labels[run.algorithm] ?? run.algorithm}</td>
                                                <td>
                                                    {run.placed_count} / {run.requested_count}
                                                </td>
                                                <td>
                                                    {(run.utilization * 100).toLocaleString(
                                                        'ru-RU',
                                                        { maximumFractionDigits: 2 },
                                                    )}
                                                    %
                                                </td>
                                                <td>
                                                    {(run.duration_ms / 1000).toLocaleString(
                                                        'ru-RU',
                                                        { maximumFractionDigits: 3 },
                                                    )}{' '}
                                                    с
                                                </td>
                                                <td>
                                                    {run.keep_existing ? 'Сохранение' : 'Замена'}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </>
                    )}
                    <button disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>
                        Назад
                    </button>
                    <span> Страница {page} </span>
                    <button
                        disabled={!current.data.has_next}
                        onClick={() => setPage((value) => value + 1)}
                    >
                        Вперёд
                    </button>
                </>
            )}
        </section>
    )
}
