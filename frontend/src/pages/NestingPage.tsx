import { confirmReplay, recordPlacement } from '../features/nesting/placementHistory'
import type { PlacementHistory, PlacementPose } from '../features/nesting/placementHistory'
import { NestingHistory } from '../features/nesting/NestingHistory'
import { downloadLayout, exportLayoutSVG } from '../features/nesting/exportLayout'
import { readPendingNesting, runNesting } from '../features/nesting/runNesting'
import type { PendingNesting } from '../features/nesting/runNesting'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { useAuth } from '../features/auth/useAuth'
import { getApiErrorMessage } from '../shared/api/errorMessages'
import type { ProjectSurface, ProjectPattern } from '../features/project-resources/types'
import { allProjectPatterns, allProjectSurfaces } from '../features/nesting/api'
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
    const navigate = useNavigate()
    const base = `/projects/${projectId}/surfaces/${surfaceId}`
    const storageKey = `nesting.pending:${base}`
    const [operation, setOperation] = useState(() => readPendingNesting(storageKey))
    const [surfaceList, setSurfaceList] = useState<ProjectSurface[]>([])
    const [projectName, setProjectName] = useState('')
    const [surface, setSurface] = useState<ProjectSurface | null>(null)
    const [patterns, setPatterns] = useState<ProjectPattern[]>([])
    const [history, setHistory] = useState<PlacementHistory>({ undo: [], redo: [] })
    const [placements, setPlacements] = useState<Placement[]>([])
    const [error, setError] = useState<string | null>(null)
    const [syncError, setSyncError] = useState<string | null>(null)
    const [refreshRequired, setRefreshRequired] = useState(false)
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
    const [historyRevision, setHistoryRevision] = useState(0)
    const [result, setResult] = useState<NestingResult | null>(null)
    const [resultAlgorithm, setResultAlgorithm] = useState('')
    useEffect(() => {
        let active = true
        async function load() {
            pending.current = true
            setBusy(true)
            setError(null)
            try {
                const [loadedSurface, loadedPatterns, loadedPlacements, loadedSurfaces, project] =
                    await Promise.all([
                        authorizedRequest<ProjectSurface>(base),
                        allProjectPatterns(authorizedRequest, projectId),
                        authorizedRequest<{ items: Placement[] }>(`${base}/placements`),
                        allProjectSurfaces(authorizedRequest, projectId),
                        authorizedRequest<{ name: string }>(`/projects/${projectId}`),
                    ])
                if (active) {
                    setSurface(loadedSurface)
                    setSurfaceList(loadedSurfaces)
                    setProjectName(project.name)
                    setPatterns(loadedPatterns)
                    setPlacements(loadedPlacements.items)
                    setHistory({ undo: [], redo: [] })
                    setSyncError(null)
                    setRefreshRequired(false)
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
    function requestCreate(pose: PlacementPose) {
        return authorizedRequest<Placement>(`${base}/placements`, {
            method: 'POST',
            body: JSON.stringify(pose),
        })
    }
    function requestPatch(item: Placement, { x, y, rotation }: PlacementPose) {
        return authorizedRequest<Placement>(`/projects/${projectId}/placements/${item.id}`, {
            method: 'PATCH',
            headers: { 'X-Placement-Version': item.updated_at },
            body: JSON.stringify({ x, y, rotation }),
        })
    }
    function requestDelete(item: Placement) {
        return authorizedRequest<void>(`/projects/${projectId}/placements/${item.id}`, {
            method: 'DELETE',
            headers: { 'X-Placement-Version': item.updated_at },
        })
    }
    function poseOf(item: Placement): PlacementPose {
        const { project_pattern_id, x, y, rotation } = item
        return { project_pattern_id, x, y, rotation }
    }
    async function add(x: number, y: number) {
        if (!patternId) return
        await mutate(async () => {
            const item = await requestCreate({ project_pattern_id: patternId, x, y, rotation: 0 })
            setPlacements((items) => [...items, item])
            setHistory((value) =>
                recordPlacement(value, { id: item.id, before: null, after: poseOf(item) }),
            )
            setSelected(item.id)
            setAdding(false)
        })
    }
    async function update(id: number, x: number, y: number, rotation?: number) {
        const previous = placements.find((item) => item.id === id)
        if (!previous) return
        const target = { ...poseOf(previous), x, y, rotation: rotation ?? previous.rotation }
        if (
            target.x === previous.x &&
            target.y === previous.y &&
            target.rotation === previous.rotation
        )
            return
        await mutate(async () => {
            const item = await requestPatch(previous, target)
            setPlacements((items) => items.map((current) => (current.id === id ? item : current)))
            setHistory((value) =>
                recordPlacement(value, { id, before: poseOf(previous), after: poseOf(item) }),
            )
        })
    }
    async function replay(direction: 'undo' | 'redo') {
        const change = history[direction].at(-1)
        if (!change) return
        await mutate(async () => {
            const previous = direction === 'undo' ? change.after : change.before
            const target = direction === 'undo' ? change.before : change.after
            const item = placements.find((placement) => placement.id === change.id)
            if (
                previous &&
                (!item ||
                    previous.x !== item.x ||
                    previous.y !== item.y ||
                    previous.rotation !== item.rotation)
            ) {
                setError('Раскладка изменилась. Обновите размещения перед продолжением.')
                return
            }
            if (item && !target) {
                await requestDelete(item)
                setPlacements((items) => items.filter((placement) => placement.id !== item.id))
                setSelected(null)
                setHistory((value) => confirmReplay(value, direction, item.id))
            } else if (target) {
                const saved = item ? await requestPatch(item, target) : await requestCreate(target)
                setPlacements((items) =>
                    item
                        ? items.map((placement) => (placement.id === item.id ? saved : placement))
                        : [...items, saved],
                )
                setSelected(saved.id)
                setHistory((value) => confirmReplay(value, direction, saved.id))
            }
        })
    }
    async function run(request: NestingRequest, rethrow = false, retry?: PendingNesting) {
        await mutate(async () => {
            if ((refreshRequired || operation) && !retry) return
            const current = retry ?? { key: crypto.randomUUID(), request }
            sessionStorage.setItem(storageKey, JSON.stringify(current))
            setOperation(current)
            const clearOperation = () => {
                sessionStorage.removeItem(storageKey)
                setOperation(null)
            }
            const status = await runNesting(
                authorizedRequest,
                base,
                request,
                (response) => {
                    setResult(response)
                    setHistory({ undo: [], redo: [] })
                    setHistoryRevision((value) => value + 1)
                    setResultAlgorithm(request.algorithm)
                    setSelected(null)
                    if (!retry)
                        setPlacements((items) =>
                            Array.from(
                                new Map(
                                    [
                                        ...(request.keep_existing ? items : []),
                                        ...response.placements,
                                    ].map((item) => [item.id, item]),
                                ).values(),
                            ),
                        )
                },
                setPlacements,
                current.key,
            ).catch((caught) => {
                clearOperation()
                throw caught
            })
            if (status !== 'unknown') clearOperation()
            if (status === 'synchronized') {
                setSyncError(null)
                return
            }
            setRefreshRequired(true)
            setReplacement(null)
            setSyncError(
                status === 'confirmed'
                    ? 'Раскрой сохранён. Не удалось загрузить полный список размещений.'
                    : 'Связь с сервером прервалась. Проверьте статус или повторите тот же запуск с сохранённым ключом.',
            )
        }, rethrow)
    }
    async function checkOperation(current: PendingNesting) {
        await mutate(async () => {
            const status = await authorizedRequest<{
                status: 'unknown' | 'running' | 'completed' | 'failed'
                result?: NestingResult
            }>(`${base}/nesting/operations/${current.key}`)
            if (status.status === 'completed' && status.result) {
                setResult(status.result)
                setHistoryRevision((value) => value + 1)
                setResultAlgorithm(current.request.algorithm)
                setSelected(null)
                sessionStorage.removeItem(storageKey)
                setOperation(null)
                setReload((value) => value + 1)
            } else if (status.status === 'failed') {
                sessionStorage.removeItem(storageKey)
                setOperation(null)
                setRefreshRequired(false)
                setSyncError('Сервер отклонил запуск. Раскладка сохранена в прежнем состоянии.')
            } else {
                setSyncError(
                    status.status === 'running'
                        ? 'Сервер выполняет этот запуск.'
                        : 'Сохранённый результат пока отсутствует. Можно безопасно повторить этот запуск.',
                )
            }
        })
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
    function exportSaved(format: 'svg' | 'json') {
        if (!surface) return
        void mutate(async () => {
            const [savedSurface, saved, project] = await Promise.all([
                authorizedRequest<ProjectSurface>(base),
                authorizedRequest<{ items: Placement[] }>(`${base}/placements`),
                authorizedRequest<{ id: number; name: string }>(`/projects/${projectId}`),
            ])
            const report = {
                project: { id: project.id, name: project.name },
                surface: savedSurface,
                placements: saved.items,
                metrics: layoutMetrics(savedSurface.geometry, saved.items),
                patterns: patterns
                    .filter((pattern) =>
                        saved.items.some((item) => item.project_pattern_id === pattern.id),
                    )
                    .map(({ id, name }) => ({ id, name })),
            }
            const data =
                format === 'svg'
                    ? exportLayoutSVG(savedSurface.geometry, saved.items)
                    : JSON.stringify(report, null, 2)
            downloadLayout(
                `nesting-${projectId}-${surfaceId}.${format}`,
                data,
                format === 'svg' ? 'image/svg+xml' : 'application/json',
            )
        })
    }
    const selectedPlacement = placements.find((item) => item.id === selected)
    const metrics = useMemo(
        () => (surface ? layoutMetrics(surface.geometry, placements) : null),
        [surface, placements],
    )
    return (
        <section className="nesting-page">
            <header>
                <Link to={`/projects/${projectId}/editor`}>← Состав проекта</Link>
                {projectName && <p>{projectName}</p>}
                {surfaceList.length > 1 && (
                    <label>
                        Поверхность{' '}
                        <select
                            value={surfaceId}
                            disabled={busy}
                            onChange={(event) => {
                                navigate(
                                    `/projects/${projectId}/surfaces/${Number(event.target.value)}/nesting`,
                                )
                            }}
                        >
                            {surfaceList.map(({ id, name }) => (
                                <option key={id} value={id}>
                                    {name} · №{id}
                                </option>
                            ))}
                        </select>
                    </label>
                )}
                <h1>
                    {surface?.name ?? 'Раскрой'} · экземпляр №{surfaceId}
                </h1>
                {surface && (
                    <div className="nesting-tools">
                        <button
                            disabled={busy || !history.undo.length}
                            onClick={() => void replay('undo')}
                        >
                            Отменить действие
                        </button>
                        <button
                            disabled={busy || !history.redo.length}
                            onClick={() => void replay('redo')}
                        >
                            Повторить действие
                        </button>
                        <button disabled={busy} onClick={() => exportSaved('svg')}>
                            Экспорт SVG
                        </button>
                        <button disabled={busy} onClick={() => exportSaved('json')}>
                            Экспорт отчёта JSON
                        </button>
                    </div>
                )}
            </header>
            {operation && (
                <div role="status">
                    <p>Ожидается подтверждение запуска раскроя.</p>
                    <button disabled={busy} onClick={() => void checkOperation(operation)}>
                        Проверить статус
                    </button>
                    <button
                        disabled={busy}
                        onClick={() => void run(operation.request, false, operation)}
                    >
                        Повторить тот же запуск
                    </button>
                </div>
            )}
            {syncError && (
                <div role="status">
                    <p>{syncError}</p>
                    <button disabled={busy} onClick={() => setReload((value) => value + 1)}>
                        Загрузить размещения
                    </button>
                </div>
            )}
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
                        onMove={(id, x, y, rotation) => void update(id, x, y, rotation)}
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
                        <button
                            disabled={busy || refreshRequired || !!operation || !patterns.length}
                            onClick={prepareRun}
                        >
                            Запустить раскрой
                        </button>
                        {metrics && (
                            <section>
                                <h2>Текущая раскладка</h2>
                                <p>Доступная площадь {format(metrics.surfaceArea)}</p>
                                {metrics.holesArea > 0 && (
                                    <p>
                                        Внешний контур {format(metrics.outerArea)} · отверстия{' '}
                                        {format(metrics.holesArea)}
                                    </p>
                                )}
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
                            await requestDelete(selectedPlacement)
                            setHistory((value) =>
                                recordPlacement(value, {
                                    id: selectedPlacement.id,
                                    before: poseOf(selectedPlacement),
                                    after: null,
                                }),
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
            {surface && (
                <NestingHistory
                    request={authorizedRequest}
                    base={base}
                    labels={algorithms}
                    revision={historyRevision}
                />
            )}
        </section>
    )
}
