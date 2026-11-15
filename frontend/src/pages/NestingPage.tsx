import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router'
import { useAuth } from '../features/auth/useAuth'
import { getApiErrorMessage } from '../shared/api/errorMessages'
import type { ProjectSurface, ProjectPattern } from '../features/project-resources/types'
import { allProjectPatterns } from '../features/nesting/api'
import type { Placement, NestingRequest, NestingResult } from '../features/nesting/api'
import { NestingCanvas } from '../features/nesting/NestingCanvas'
import { PlacementProperties } from '../features/nesting/PlacementProperties'
import { layoutMetrics } from '../features/nesting/geometry'
import { ConfirmDialog } from '../components/dialogs/ConfirmDialog'
import './NestingPage.css'

const format = (value: number) => value.toLocaleString('ru-RU', { maximumFractionDigits: 2 })
const algorithms = {
    baseline: 'Базовый',
    nfp_greedy: 'Жадный NFP',
    genetic: 'Генетический',
    simulated_annealing: 'Имитация отжига',
    hybrid: 'Гибридный',
}

export function NestingPage() {
    const { projectId = '', surfaceId = '' } = useParams()
    if (
        ![projectId, surfaceId].every(
            (value) => /^[1-9]\d*$/.test(value) && Number.isSafeInteger(Number(value)),
        )
    ) {
        return <p role="alert">Некорректный адрес раскроя.</p>
    }
    return (
        <Workspace
            key={`${projectId}:${surfaceId}`}
            projectId={Number(projectId)}
            surfaceId={Number(surfaceId)}
        />
    )
}

function Workspace({ projectId, surfaceId }: { projectId: number; surfaceId: number }) {
    const { authorizedRequest } = useAuth()
    const base = `/projects/${projectId}/surfaces/${surfaceId}`
    const [surface, setSurface] = useState<ProjectSurface | null>(null)
    const [patterns, setPatterns] = useState<ProjectPattern[]>([])
    const [placements, setPlacements] = useState<Placement[]>([])
    const [error, setError] = useState<string | null>(null)
    const [reload, setReload] = useState(0)
    const [busy, setBusy] = useState(false)
    const pending = useRef(false)
    const [selected, setSelected] = useState<number | null>(null)
    const [patternId, setPatternId] = useState<number | null>(null)
    const [adding, setAdding] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const [algorithm, setAlgorithm] = useState('baseline')
    const [quantities, setQuantities] = useState<Record<number, number>>({})
    const [rotations, setRotations] = useState('0, 90, 180, 270')
    const [keepExisting, setKeepExisting] = useState(true)
    const [replacement, setReplacement] = useState<NestingRequest | null>(null)
    const [result, setResult] = useState<NestingResult | null>(null)
    const [resultAlgorithm, setResultAlgorithm] = useState('')
    useEffect(() => {
        let active = true
        async function load() {
            pending.current = true
            setBusy(true)
            setError(null)
            try {
                const [loadedSurface, loadedPatterns, loadedPlacements] = await Promise.all([
                    authorizedRequest<ProjectSurface>(base),
                    allProjectPatterns(authorizedRequest, projectId),
                    authorizedRequest<{ items: Placement[] }>(`${base}/placements`),
                ])
                if (active) {
                    setSurface(loadedSurface)
                    setPatterns(loadedPatterns)
                    setPlacements(loadedPlacements.items)
                }
            } catch (caught) {
                if (active) setError(getApiErrorMessage(caught))
            } finally {
                if (active) {
                    pending.current = false
                    setBusy(false)
                }
            }
        }
        void load()
        return () => {
            active = false
        }
    }, [authorizedRequest, base, projectId, reload])

    async function mutate(action: () => Promise<void>, rethrow = false) {
        if (pending.current) return
        pending.current = true
        setBusy(true)
        setError(null)
        try {
            await action()
        } catch (caught) {
            setError(getApiErrorMessage(caught))
            if (rethrow) throw caught
        } finally {
            pending.current = false
            setBusy(false)
        }
    }
    async function add(x: number, y: number) {
        if (!patternId) return
        await mutate(async () => {
            const item = await authorizedRequest<Placement>(`${base}/placements`, {
                method: 'POST',
                body: JSON.stringify({ project_pattern_id: patternId, x, y, rotation: 0 }),
            })
            setPlacements((items) => [...items, item])
            setSelected(item.id)
            setAdding(false)
        })
    }
    async function update(id: number, x: number, y: number, rotation?: number) {
        await mutate(async () => {
            const item = await authorizedRequest<Placement>(
                `/projects/${projectId}/placements/${id}`,
                {
                    method: 'PATCH',
                    body: JSON.stringify({ x, y, rotation }),
                },
            )
            setPlacements((items) => items.map((current) => (current.id === id ? item : current)))
        })
    }
    async function run(request: NestingRequest, rethrow = false) {
        await mutate(async () => {
            const response = await authorizedRequest<NestingResult>(`${base}/nesting`, {
                method: 'POST',
                body: JSON.stringify(request),
            })
            setResult(response)
            setResultAlgorithm(request.algorithm)
            setPlacements((items) =>
                request.keep_existing ? [...items, ...response.placements] : response.placements,
            )
            const saved = await authorizedRequest<{ items: Placement[] }>(`${base}/placements`)
            setPlacements(saved.items)
            setSelected(null)
        }, rethrow)
    }
    function prepareRun() {
        const angles = rotations.split(',').map((value) => Number(value.trim()))
        const requested = Object.entries(quantities)
            .filter(([, quantity]) => quantity > 0)
            .map(([id, quantity]) => ({ project_pattern_id: Number(id), quantity }))
        if (
            !requested.length ||
            requested.some((item) => !Number.isSafeInteger(item.quantity)) ||
            rotations.split(',').some((value) => !value.trim()) ||
            angles.some((angle) => !Number.isFinite(angle) || angle < 0 || angle >= 360)
        ) {
            setError(
                'Укажите целое положительное количество хотя бы одного лекала и углы от 0 до 360° (не включая 360).',
            )
            return
        }
        const request = {
            algorithm,
            patterns: requested,
            allowed_rotations: angles,
            keep_existing: keepExisting,
        }
        if (!keepExisting) setReplacement(request)
        else void run(request)
    }
    const selectedPlacement = placements.find((item) => item.id === selected)
    const metrics = surface ? layoutMetrics(surface.geometry, placements) : null
    return (
        <section className="nesting-page">
            <header>
                <Link to={`/projects/${projectId}/editor`}>← Состав проекта</Link>
                <h1>
                    {surface?.name ?? 'Раскрой'} · экземпляр №{surfaceId}
                </h1>
            </header>
            {error && (
                <div role="alert">
                    <p>{error}</p>
                    <button disabled={busy} onClick={() => setReload(reload + 1)}>
                        Обновить данные
                    </button>
                </div>
            )}
            {!surface && !error && <p role="status">Загружаем рабочую поверхность…</p>}
            {busy && <p role="status">Выполняется операция на сервере…</p>}
            {surface && (
                <div className="nesting-workspace">
                    <aside>
                        <h2>Лекала проекта</h2>
                        {!patterns.length && <p>Добавьте лекала на странице проекта.</p>}
                        {patterns.map((item) => (
                            <div className="nesting-pattern" key={item.id}>
                                <button
                                    disabled={busy}
                                    aria-pressed={patternId === item.id}
                                    onClick={() => setPatternId(item.id)}
                                >
                                    {item.name} · №{item.id}
                                </button>
                                <label>
                                    Количество для алгоритма
                                    <input
                                        type="number"
                                        min="0"
                                        step="1"
                                        value={quantities[item.id] ?? 0}
                                        disabled={busy}
                                        onChange={(event) =>
                                            setQuantities({
                                                ...quantities,
                                                [item.id]: Number(event.target.value),
                                            })
                                        }
                                    />
                                </label>
                            </div>
                        ))}
                        <button
                            disabled={busy || !patternId}
                            aria-pressed={adding}
                            onClick={() => setAdding(!adding)}
                        >
                            {adding ? 'Отменить добавление' : 'Разместить кликом'}
                        </button>
                        {adding && <p>Выберите положение центра лекала на поверхности.</p>}
                        <h3>Размещения</h3>
                        {placements.map((item) => (
                            <button
                                key={item.id}
                                disabled={busy}
                                aria-pressed={selected === item.id}
                                onClick={() => {
                                    setSelected(item.id)
                                    setAdding(false)
                                }}
                            >
                                №{item.id} · лекало №{item.project_pattern_id}
                            </button>
                        ))}
                    </aside>
                    <NestingCanvas
                        surface={surface.geometry}
                        placements={placements}
                        selected={selected}
                        busy={busy}
                        adding={adding}
                        onSelect={setSelected}
                        onAdd={(x, y) => void add(x, y)}
                        onMove={(id, x, y) => void update(id, x, y)}
                    />
                    <aside>
                        {selectedPlacement && (
                            <PlacementProperties
                                key={`${selectedPlacement.id}:${selectedPlacement.x}:${selectedPlacement.y}:${selectedPlacement.rotation}`}
                                placement={selectedPlacement}
                                busy={busy}
                                onSave={(x, y, angle) =>
                                    void update(selectedPlacement.id, x, y, angle)
                                }
                                onDelete={() => setDeleting(true)}
                            />
                        )}
                        <h2>Автоматический раскрой</h2>
                        <label>
                            Алгоритм
                            <select
                                value={algorithm}
                                disabled={busy}
                                onChange={(event) => setAlgorithm(event.target.value)}
                            >
                                {Object.entries(algorithms).map(([value, label]) => (
                                    <option key={value} value={value}>
                                        {label}
                                    </option>
                                ))}
                            </select>
                        </label>
                        <label>
                            Углы через запятую
                            <input
                                value={rotations}
                                disabled={busy}
                                onChange={(event) => setRotations(event.target.value)}
                            />
                        </label>
                        <label>
                            <input
                                type="checkbox"
                                checked={keepExisting}
                                disabled={busy}
                                onChange={(event) => setKeepExisting(event.target.checked)}
                            />{' '}
                            Сохранить существующие размещения
                        </label>
                        <button disabled={busy || !patterns.length} onClick={prepareRun}>
                            Запустить раскрой
                        </button>
                        {metrics && (
                            <section>
                                <h2>Текущая раскладка</h2>
                                <p>Площадь поверхности {format(metrics.surfaceArea)}</p>
                                <p>Занято {format(metrics.placedArea)}</p>
                                <p>Свободно {format(metrics.freeArea)}</p>
                                <p>
                                    Коэффициент {format(metrics.utilization)} · использование{' '}
                                    {format(metrics.utilization * 100)}%
                                </p>
                                <p>Размещений {placements.length}</p>
                                {Object.entries(metrics.counts).map(([id, count]) => (
                                    <p key={id}>
                                        Лекало №{id} — {count}
                                    </p>
                                ))}
                            </section>
                        )}
                        {result && (
                            <section>
                                <h2>Последний запуск в этой сессии</h2>
                                <p>{algorithms[resultAlgorithm as keyof typeof algorithms]}</p>
                                <p>
                                    Запрошено {result.metrics.requested_count}, размещено{' '}
                                    {result.metrics.placed_count}
                                </p>
                                <p>
                                    Площадь поверхности {format(result.metrics.surface_area)},
                                    размещено {format(result.metrics.placed_area)}
                                </p>
                                <p>Использование {format(result.metrics.utilization * 100)}%</p>
                                {result.unplaced.map((item) => (
                                    <p key={item.project_pattern_id}>
                                        Не размещено лекало №{item.project_pattern_id} —{' '}
                                        {item.quantity}
                                    </p>
                                ))}
                                {!result.unplaced.length && (
                                    <p>Все запрошенные лекала размещены.</p>
                                )}
                            </section>
                        )}
                    </aside>
                </div>
            )}
            {deleting && selectedPlacement && (
                <ConfirmDialog
                    title="Удалить размещение?"
                    onClose={() => setDeleting(false)}
                    onConfirm={() =>
                        mutate(async () => {
                            await authorizedRequest<void>(
                                `/projects/${projectId}/placements/${selectedPlacement.id}`,
                                { method: 'DELETE' },
                            )
                            setPlacements((items) =>
                                items.filter((item) => item.id !== selectedPlacement.id),
                            )
                            setSelected(null)
                            setDeleting(false)
                        }, true)
                    }
                >
                    <p>Удалить размещение №{selectedPlacement.id}? Проектное лекало сохраняется.</p>
                </ConfirmDialog>
            )}
            {replacement && (
                <ConfirmDialog
                    title="Заменить текущую раскладку?"
                    confirmLabel="Запустить с заменой"
                    onClose={() => setReplacement(null)}
                    onConfirm={() => run(replacement, true)}
                >
                    <p>
                        Размещения на поверхности №{surfaceId} будут заменены результатом алгоритма.
                    </p>
                </ConfirmDialog>
            )}
        </section>
    )
}
