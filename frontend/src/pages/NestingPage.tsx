import {
    classifyPreviewFailure,
    mayClosePreview,
    parseNestingAngles,
    previewStatusOutcome,
    validateNestingRequest,
} from '../features/nesting/previewRecovery'
import type { NestingOperationStatus } from '../features/nesting/previewRecovery'
import { confirmReplay, recordPlacement } from '../features/nesting/placementHistory'
import type { PlacementHistory, PlacementPose } from '../features/nesting/placementHistory'
import { NestingHistory } from '../features/nesting/NestingHistory'
import { downloadLayout, exportLayoutSVG } from '../features/nesting/exportLayout'
import {
    nestingRequestTimeout,
    readPendingNesting,
    runNesting,
} from '../features/nesting/runNesting'
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
import { ApiError } from '../shared/api/http'
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
    const { user } = useAuth()
    const { projectId = '', surfaceId = '' } = useParams()
    if (!user) return <p role="status">Проверяем сеанс…</p>
    if (
        ![projectId, surfaceId].every(
            (value) => /^[1-9]\d*$/.test(value) && Number.isSafeInteger(Number(value)),
        )
    ) {
        return <p role="alert">Некорректный адрес раскроя.</p>
    }
    return (
        <Workspace
            key={`${user.id}:${projectId}:${surfaceId}`}
            userId={user.id}
            projectId={Number(projectId)}
            surfaceId={Number(surfaceId)}
        />
    )
}

type WorkspaceActivity =
    | { kind: 'idle' | 'loading' | 'manual' | 'syncing' | 'recovering' }
    | {
          kind: 'calculating_preview' | 'calculating_saved' | 'saving_preview'
          key: string
          cancelling: boolean
      }

function Workspace({
    projectId,
    surfaceId,
    userId,
}: {
    projectId: number
    surfaceId: number
    userId: number
}) {
    const { authorizedRequest } = useAuth()
    const navigate = useNavigate()
    const base = `/projects/${projectId}/surfaces/${surfaceId}`
    const storageKey = `nesting.pending:${userId}:${base}`
    const [operation, setOperation] = useState(() =>
        readPendingNesting(storageKey, userId, projectId, surfaceId),
    )
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
    const [activity, setActivity] = useState<WorkspaceActivity>({ kind: 'loading' })
    const busy = activity.kind !== 'idle'
    const execution = 'key' in activity ? activity : null
    const cancelling = execution?.cancelling ?? false
    const [proposal, setProposal] = useState<{
        result: NestingResult
        includeSaved: boolean
        request: NestingRequest
    } | null>(null)
    const [patternsOpen, setPatternsOpen] = useState(() => window.innerWidth > 700)
    const [inspectorOpen, setInspectorOpen] = useState(() => window.innerWidth > 700)
    useEffect(() => {
        const compact = window.matchMedia('(max-width: 700px)')
        const closePanels = () => {
            if (compact.matches) {
                setPatternsOpen(false)
                setInspectorOpen(false)
            }
        }
        compact.addEventListener('change', closePanels)
        return () => compact.removeEventListener('change', closePanels)
    }, [])
    function openInspector(section: 'properties' | 'automatic' | 'history') {
        if (window.matchMedia('(max-width: 700px)').matches) setPatternsOpen(false)
        setInspectorOpen(true)
        setPanel(section)
    }
    const [panel, setPanel] = useState<'properties' | 'automatic' | 'history'>('properties')

    const pending = useRef(false)
    const [selected, setSelected] = useState<number | null>(null)
    const [patternId, setPatternId] = useState<number | null>(null)
    const [adding, setAdding] = useState(false)
    const [deleting, setDeleting] = useState(false)
    const [algorithm, setAlgorithm] = useState('baseline')
    const [quantities, setQuantities] = useState<Record<number, number>>({})
    const [rotations, setRotations] = useState('0, 90, 180, 270')
    const [keepExisting, setKeepExisting] = useState(true)
    const [closePreview, setClosePreview] = useState<PendingNesting | null>(null)
    const [replacement, setReplacement] = useState<NestingRequest | null>(null)
    const [historyRevision, setHistoryRevision] = useState(0)
    const [result, setResult] = useState<NestingResult | null>(null)
    const [resultAlgorithm, setResultAlgorithm] = useState('')
    useEffect(() => {
        let active = true
        async function load() {
            pending.current = true
            setActivity({ kind: 'loading' })
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
                    setActivity({ kind: 'idle' })
                }
            }
        }
        void load()
        return () => {
            active = false
        }
    }, [authorizedRequest, base, projectId, reload])

    async function mutate(
        action: () => Promise<void>,
        rethrow = false,
        phase: WorkspaceActivity = { kind: 'manual' },
    ) {
        if (pending.current) return
        pending.current = true
        setActivity(phase)
        setError(null)
        try {
            await action()
        } catch (caught) {
            setError(getApiErrorMessage(caught))
            if (rethrow) throw caught
        } finally {
            pending.current = false
            setActivity({ kind: 'idle' })
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
            const current = retry ?? {
                key: crypto.randomUUID(),
                request,
                userId,
                projectId,
                surfaceId,
            }
            sessionStorage.setItem(storageKey, JSON.stringify(current))
            setOperation(current)
            setProposal(null)
            setActivity({
                kind: request.preview_id ? 'saving_preview' : 'calculating_saved',
                key: current.key,
                cancelling: false,
            })
            const clearOperation = () => {
                sessionStorage.removeItem(storageKey)
                setOperation(null)
            }
            const status = await runNesting(
                authorizedRequest,
                base,
                request,
                (response) => {
                    setActivity({ kind: 'syncing' })
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
                    : 'Подтверждение сохранения пока не получено. Проверьте статус или повторите тот же запуск с сохранённым ключом.',
            )
        }, rethrow)
    }
    async function checkOperation(current: PendingNesting) {
        let reloadConfirmed = false
        await mutate(
            async () => {
                const status = await authorizedRequest<NestingOperationStatus>(
                    `${base}/nesting/operations/${current.key}${current.kind === 'preview' ? '?kind=preview' : ''}`,
                    {
                        signal: AbortSignal.timeout(10_000),
                    },
                )
                if (status.status === 'completed' && status.result) {
                    if (current.kind === 'preview') {
                        openInspector('automatic')
                        const recovered = previewStatusOutcome(current, status)
                        if (recovered.kind !== 'completed') return
                        setProposal(recovered.proposal)
                        sessionStorage.removeItem(storageKey)
                        setOperation(null)
                        setSyncError(null)
                        return
                    }
                    setResult(status.result)
                    setHistoryRevision((value) => value + 1)
                    setResultAlgorithm(current.request.algorithm)
                    setSelected(null)
                    sessionStorage.removeItem(storageKey)
                    setOperation(null)
                    reloadConfirmed = true
                } else if (status.status === 'failed') {
                    sessionStorage.removeItem(storageKey)
                    setOperation(null)
                    setRefreshRequired(false)
                    setSyncError(
                        status.failure
                            ? getApiErrorMessage(
                                  new ApiError(500, status.failure.code, status.failure.message),
                              )
                            : 'Сервер отклонил запуск. Раскладка сохранена в прежнем состоянии.',
                    )
                } else {
                    setSyncError(
                        status.status === 'running'
                            ? 'Сервер выполняет этот запуск.'
                            : 'Статус пока неизвестен: запрос ещё может быть в пути. Повторите тот же запуск с прежним ключом; новый расчёт автоматически не запускается.',
                    )
                }
            },
            false,
            { kind: 'recovering' },
        )
        if (reloadConfirmed) setReload((value) => value + 1)
    }
    async function previewRun(request: NestingRequest, retry?: PendingNesting) {
        await mutate(async () => {
            const validation = validateNestingRequest(retry?.request ?? request)
            if (validation) {
                setError(validation)
                return
            }
            const current: PendingNesting = retry ?? {
                kind: 'preview',
                key: crypto.randomUUID(),
                userId,
                projectId,
                surfaceId,
                request,
            }
            sessionStorage.setItem(storageKey, JSON.stringify(current))
            setOperation(current)
            setProposal(null)
            setActivity({ kind: 'calculating_preview', key: current.key, cancelling: false })
            try {
                const result = await authorizedRequest<NestingResult>(`${base}/nesting/preview`, {
                    method: 'POST',
                    signal: AbortSignal.timeout(nestingRequestTimeout),
                    headers: { 'Idempotency-Key': current.key },
                    body: JSON.stringify(current.request),
                })
                setProposal({
                    result,
                    request: current.request,
                    includeSaved: current.request.keep_existing,
                })
                sessionStorage.removeItem(storageKey)
                setOperation(null)
                setSyncError(null)
            } catch (caught) {
                const category = classifyPreviewFailure(caught, !!retry)
                if (category === 'rejected') {
                    sessionStorage.removeItem(storageKey)
                    setOperation(null)
                    setError(getApiErrorMessage(caught))
                    setSyncError(null)
                    return
                }
                if (category === 'check_status') {
                    try {
                        const status = await authorizedRequest<NestingOperationStatus>(
                            `${base}/nesting/operations/${current.key}?kind=preview`,
                            { signal: AbortSignal.timeout(10_000) },
                        )
                        const outcome = previewStatusOutcome(current, status)
                        if (outcome.kind === 'completed') {
                            setProposal(outcome.proposal)
                            openInspector('automatic')
                            sessionStorage.removeItem(storageKey)
                            setOperation(null)
                            setSyncError(null)
                            return
                        }
                        if (outcome.kind === 'failed') {
                            sessionStorage.removeItem(storageKey)
                            setOperation(null)
                            setError(
                                outcome.failure
                                    ? getApiErrorMessage(
                                          new ApiError(
                                              500,
                                              outcome.failure.code,
                                              outcome.failure.message,
                                          ),
                                      )
                                    : getApiErrorMessage(caught),
                            )
                            return
                        }
                    } catch {
                        /* Retain the key when reconciliation is unavailable. */
                    }
                    setError(getApiErrorMessage(caught))
                }
                setSyncError(
                    category === 'check_status'
                        ? 'Этот повтор отклонён сервером. Исход предыдущего запроса ещё не подтверждён. Проверьте статус или закройте ожидание предпросмотра.'
                        : caught instanceof ApiError
                          ? 'Получен ответ об ошибке, но итог операции пока неизвестен. Проверьте статус или повторите запрос с тем же ключом.'
                          : 'Ответ предпросмотра не получен. Проверьте статус или повторите запрос с тем же ключом.',
                )
            }
        })
    }

    async function cancelExecution() {
        if (!execution || cancelling) return
        setActivity({ ...execution, cancelling: true })
        try {
            await authorizedRequest<void>(
                `${base}/nesting/operations/${execution.key}${execution.kind === 'calculating_preview' ? '?kind=preview' : ''}`,
                {
                    method: 'DELETE',
                    signal: AbortSignal.timeout(10_000),
                },
            )
        } catch (caught) {
            setError(getApiErrorMessage(caught))
            setActivity((current) =>
                'key' in current && current.key === execution.key
                    ? { ...current, cancelling: false }
                    : current,
            )
        }
    }
    function saveProposal() {
        if (!proposal?.result.preview_id) return
        const request = { ...proposal.request, preview_id: proposal.result.preview_id }
        if (!request.keep_existing) setReplacement(request)
        else void run(request)
    }
    function prepareRun(preview = false) {
        const angles = parseNestingAngles(rotations)
        const requested = Object.entries(quantities)
            .filter(([, quantity]) => quantity !== 0)
            .map(([id, quantity]) => ({ project_pattern_id: Number(id), quantity }))
        const request = {
            algorithm,
            patterns: requested,
            allowed_rotations: angles,
            keep_existing: keepExisting,
        }
        const validation = validateNestingRequest(request)
        if (validation) {
            setError(validation)
            return
        }
        if (preview) void previewRun(request)
        else if (!keepExisting) setReplacement(request)
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
    const addingPattern = adding ? patterns.find((item) => item.id === patternId) : undefined
    const editingDisabled = busy || !!proposal || refreshRequired || !!operation
    const selectedPlacement = placements.find((item) => item.id === selected)
    const metrics = useMemo(
        () => (surface ? layoutMetrics(surface.geometry, placements) : null),
        [surface, placements],
    )
    return (
        <section
            className="nesting-page"
            onKeyDown={(event) => {
                if (event.key === 'Escape') setAdding(false)
            }}
        >
            <header>
                <Link to={`/projects/${projectId}/editor`}>← Состав проекта</Link>
                {projectName && <p>{projectName}</p>}
                {surfaceList.length > 1 && (
                    <label>
                        <span>Поверхность</span>{' '}
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
                            aria-expanded={patternsOpen}
                            onClick={() => {
                                if (
                                    !patternsOpen &&
                                    window.matchMedia('(max-width: 700px)').matches
                                )
                                    setInspectorOpen(false)
                                setPatternsOpen(!patternsOpen)
                            }}
                        >
                            Лекала
                        </button>
                        <button
                            aria-expanded={inspectorOpen}
                            onClick={() => {
                                if (inspectorOpen) setInspectorOpen(false)
                                else openInspector(panel)
                            }}
                        >
                            Панель
                        </button>
                        <button
                            className="nesting-primary"
                            aria-label="Автоматический раскрой"
                            onClick={() => {
                                openInspector('automatic')
                                setAdding(false)
                            }}
                        >
                            Автораскрой
                        </button>
                        <details className="nesting-menu">
                            <summary>Ещё</summary>
                            <div className="nesting-menu-content">
                                <button
                                    disabled={editingDisabled || !history.undo.length}
                                    onClick={() => void replay('undo')}
                                >
                                    Отменить действие
                                </button>
                                <button
                                    disabled={editingDisabled || !history.redo.length}
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
                        </details>
                    </div>
                )}
            </header>
            {operation && !execution && (
                <div role="status">
                    <p>Ожидается подтверждение запуска раскроя.</p>
                    <button disabled={busy} onClick={() => void checkOperation(operation)}>
                        Проверить статус
                    </button>
                    {operation.kind === 'preview' && (
                        <button disabled={busy} onClick={() => setClosePreview(operation)}>
                            Закрыть ожидание предпросмотра
                        </button>
                    )}
                    {operation.kind === 'preview' && (
                        <button
                            disabled={busy}
                            onClick={() =>
                                void mutate(async () => {
                                    await authorizedRequest<void>(
                                        `${base}/nesting/operations/${operation.key}?kind=preview`,
                                        { method: 'DELETE', signal: AbortSignal.timeout(10_000) },
                                    )
                                    setSyncError(
                                        'Запрос отмены принят. Проверьте итоговый статус операции.',
                                    )
                                })
                            }
                        >
                            Запросить отмену предпросмотра
                        </button>
                    )}
                    <button
                        disabled={busy}
                        onClick={() =>
                            void (operation.kind === 'preview'
                                ? previewRun(operation.request, operation)
                                : run(operation.request, false, operation))
                        }
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
            <div className="nesting-status">
                {adding && (
                    <button disabled={busy} onClick={() => setAdding(false)}>
                        Отменить добавление
                    </button>
                )}
                {execution ? (
                    <>
                        <span role="status">
                            {cancelling
                                ? 'Останавливаем расчёт…'
                                : execution.kind === 'calculating_preview'
                                  ? 'Расчёт предпросмотра…'
                                  : execution.kind === 'saving_preview'
                                    ? 'Сохранение предпросмотра…'
                                    : 'Расчёт с сохранением…'}
                        </span>
                        <button
                            className="nesting-button"
                            disabled={cancelling}
                            onClick={() => void cancelExecution()}
                        >
                            Отменить расчёт
                        </button>
                    </>
                ) : proposal ? (
                    <>
                        <span role="status">
                            Предпросмотр · размещено {proposal.result.metrics.placed_count} из{' '}
                            {proposal.result.metrics.requested_count}. Сохранённая раскладка
                            прежняя.
                        </span>
                        <button className="nesting-button" onClick={() => setProposal(null)}>
                            Закрыть предпросмотр
                        </button>
                    </>
                ) : (
                    <span role="status">
                        {busy
                            ? 'Синхронизация с сервером…'
                            : addingPattern
                              ? `Щёлкните по поверхности, чтобы добавить «${addingPattern.name}»`
                              : 'Ручные изменения сохраняются автоматически'}
                    </span>
                )}
            </div>
            {surface && (
                <div className="nesting-workspace">
                    <aside hidden={!patternsOpen} aria-label="Лекала проекта">
                        <h2>Лекала проекта</h2>
                        {!patterns.length && <p>Добавьте лекала на странице проекта.</p>}
                        {patterns.map((item) => (
                            <div className="nesting-pattern" key={item.id}>
                                <strong>{item.name}</strong>
                                <button
                                    className="nesting-button"
                                    disabled={editingDisabled}
                                    aria-pressed={adding && item.id === patternId}
                                    onClick={() => {
                                        setPatternId(item.id)
                                        setAdding(true)
                                        if (window.matchMedia('(max-width: 700px)').matches)
                                            setPatternsOpen(false)
                                        setSelected(null)
                                        setError(null)
                                    }}
                                >
                                    Добавить на поверхность
                                    <span className="visually-hidden"> · {item.name}</span>
                                </button>
                            </div>
                        ))}
                        <h3>Размещения</h3>
                        {placements.map((item) => (
                            <button
                                key={item.id}
                                disabled={editingDisabled}
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
                        busy={editingDisabled}
                        adding={!!addingPattern}
                        draftPattern={addingPattern?.geometry}
                        proposedPlacements={proposal?.result.placements}
                        showSaved={!proposal || proposal.includeSaved}
                        onSelect={(id) => {
                            setSelected(id)
                            if (id !== null) setPanel('properties')
                        }}
                        onAdd={(x, y) => void add(x, y)}
                        onMove={(id, x, y, rotation) => void update(id, x, y, rotation)}
                    />
                    <aside hidden={!inspectorOpen} aria-label="Свойства раскроя">
                        <nav className="nesting-tabs" aria-label="Панель раскроя">
                            <button
                                type="button"
                                aria-pressed={panel === 'properties'}
                                onClick={() => setPanel('properties')}
                            >
                                Свойства
                            </button>
                            <button
                                type="button"
                                aria-pressed={panel === 'automatic'}
                                onClick={() => {
                                    setPanel('automatic')
                                    setAdding(false)
                                }}
                            >
                                Автораскрой
                            </button>
                            <button
                                type="button"
                                aria-pressed={panel === 'history'}
                                onClick={() => setPanel('history')}
                            >
                                История
                            </button>
                        </nav>
                        {panel === 'properties' && !selectedPlacement && (
                            <p>Выберите размещение на поверхности, чтобы изменить его положение.</p>
                        )}
                        {panel === 'properties' && selectedPlacement && (
                            <PlacementProperties
                                key={`${selectedPlacement.id}:${selectedPlacement.x}:${selectedPlacement.y}:${selectedPlacement.rotation}`}
                                placement={selectedPlacement}
                                busy={editingDisabled}
                                onSave={(x, y, angle) =>
                                    void update(selectedPlacement.id, x, y, angle)
                                }
                                onDelete={() => setDeleting(true)}
                            />
                        )}
                        {panel === 'automatic' && (
                            <section>
                                <h2>Автоматический раскрой</h2>
                                <fieldset className="nesting-quantities">
                                    <legend>Количество лекал</legend>
                                    {patterns.map((item) => (
                                        <label key={item.id}>
                                            {item.name}
                                            <input
                                                type="number"
                                                min="0"
                                                max="500"
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
                                    ))}
                                </fieldset>
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
                                    Учитывать сохранённые размещения
                                </label>
                                <button
                                    disabled={
                                        busy || !!operation || refreshRequired || !patterns.length
                                    }
                                    onClick={() => prepareRun(true)}
                                >
                                    Предпросмотр
                                </button>
                                <button
                                    className="nesting-primary"
                                    disabled={
                                        busy || refreshRequired || !!operation || !patterns.length
                                    }
                                    onClick={() => (proposal ? saveProposal() : prepareRun())}
                                >
                                    {proposal ? 'Сохранить предпросмотр' : 'Рассчитать и сохранить'}
                                </button>
                                {proposal && (
                                    <p>
                                        Не размещено:{' '}
                                        {proposal.result.metrics.requested_count -
                                            proposal.result.metrics.placed_count}
                                        . Можно сохранить показанную раскладку без повторного
                                        расчёта.
                                    </p>
                                )}
                            </section>
                        )}
                        {panel === 'history' && (
                            <NestingHistory
                                request={authorizedRequest}
                                base={base}
                                labels={algorithms}
                                revision={historyRevision}
                            />
                        )}
                        {panel === 'properties' && metrics && (
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
            {closePreview && (
                <ConfirmDialog
                    title="Закрыть ожидание предпросмотра?"
                    confirmLabel="Закрыть ожидание"
                    onClose={() => setClosePreview(null)}
                    onConfirm={async () => {
                        if (
                            !pending.current &&
                            operation?.key === closePreview.key &&
                            mayClosePreview(operation, userId, projectId, surfaceId)
                        ) {
                            sessionStorage.removeItem(storageKey)
                            setOperation(null)
                            setSyncError(null)
                            setError(null)
                        }
                        setClosePreview(null)
                    }}
                >
                    Это удалит только ожидание в этой вкладке. Серверный расчёт может продолжаться
                    до своего deadline. Если расчёт выполняется, сначала запросите отмену и
                    проверьте статус. Сохранённая раскладка не изменится. Новый расчёт запускается
                    только вручную.
                </ConfirmDialog>
            )}
            {replacement && (
                <ConfirmDialog
                    title="Заменить текущую раскладку?"
                    confirmLabel="Запустить с заменой"
                    onClose={() => setReplacement(null)}
                    onConfirm={async () => {
                        const request = replacement
                        setReplacement(null)
                        await run(request)
                    }}
                >
                    <p>
                        Размещения на поверхности №{surfaceId} будут заменены результатом алгоритма.
                    </p>
                </ConfirmDialog>
            )}
        </section>
    )
}
