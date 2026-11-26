import { parseQuantityDrafts } from '../features/nesting/quantityDraft'
import { draftKeyAction, parseDraftRotation } from '../features/nesting/placementDraft'
import { placementIssue, transformPattern } from '../features/nesting/manualPlacement'
import { ManualDraftDialog } from '../features/nesting/ManualDraftDialog'
import { isPendingSaveScope } from '../features/nesting/saveRecovery'
import type { SaveOutcome } from '../features/nesting/saveRecovery'
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
    reconcileSavedNesting,
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

function ExecutionProgress({ request, saving }: { request?: NestingRequest; saving: boolean }) {
    const [started] = useState(() => Date.now())
    const [now, setNow] = useState(started)
    useEffect(() => {
        const timer = setInterval(() => setNow(Date.now()), 1000)
        return () => clearInterval(timer)
    }, [])
    return (
        <span>
            {request
                ? `${algorithms[request.algorithm as keyof typeof algorithms] ?? request.algorithm} · запрошено ${request.patterns.reduce((n, p) => n + p.quantity, 0)} · `
                : ''}
            {saving
                ? 'Сохраняется показанный результат без повторного расчёта'
                : `Прошло ${Math.floor((now - started) / 1000)} с · предел расчёта 120 с`}
        </span>
    )
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

type ManualDraft = {
    id: number // -1 denotes a new, not-yet-created placement
    kind: 'new' | 'existing'
    pose: PlacementPose
    notice?: string
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
    const [manualDraft, setManualDraft] = useState<ManualDraft | null>(null)
    const [manualFormRevision, setManualFormRevision] = useState(0)
    const [showDraftDecision, setShowDraftDecision] = useState(false)
    const deferredManualAction = useRef<(() => void) | null>(null)
    function requestManualTransition(action: () => void) {
        if (!manualDraft) { action(); return }
        deferredManualAction.current = action
        setShowDraftDecision(true)
    }
    function finishManualTransition() {
        const action = deferredManualAction.current
        deferredManualAction.current = null
        setManualDraft(null)
        setShowDraftDecision(false)
        setSelected(null)
        action?.()
    }
    useEffect(() => {
        if (!manualDraft) return
        const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = '' }
        window.addEventListener('beforeunload', warn)
        return () => window.removeEventListener('beforeunload', warn)
    }, [manualDraft])
    const [patternId, setPatternId] = useState<number | null>(null)
    const [adding, setAdding] = useState(false)
    const [draftRotation, setDraftRotation] = useState(0)
    const [draftRotationText, setDraftRotationText] = useState('0')
    function changeDraftRotation(raw: string) {
        setDraftRotationText(raw)
        const angle = parseDraftRotation(raw)
        if (angle !== null) setDraftRotation(angle)
    }
    const [deleting, setDeleting] = useState(false)
    const [algorithm, setAlgorithm] = useState('baseline')
    const [quantities, setQuantities] = useState<Record<number, string>>({})
    const [rotations, setRotations] = useState('0, 90, 180, 270')
    const [keepExisting, setKeepExisting] = useState(true)
    const [closePreview, setClosePreview] = useState<PendingNesting | null>(null)
    const [replacement, setReplacement] = useState<NestingRequest | null>(null)
    const [historyRevision, setHistoryRevision] = useState(0)
    const [result, setResult] = useState<NestingResult | null>(null)
    const [resultAlgorithm, setResultAlgorithm] = useState('')
    const [responseSeconds, setResponseSeconds] = useState<number | null>(null)
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
                        authorizedRequest<{ items: Placement[] }>(`${base}/placements`).then(
                            (value) => ({ ok: true as const, items: value.items }),
                            (error: unknown) => ({ ok: false as const, error }),
                        ),
                        allProjectSurfaces(authorizedRequest, projectId),
                        authorizedRequest<{ name: string }>(`/projects/${projectId}`),
                    ])
                if (active) {
                    setSurface(loadedSurface)
                    setSurfaceList(loadedSurfaces)
                    setProjectName(project.name)
                    setPatterns(loadedPatterns)
                    setHistory({ undo: [], redo: [] })
                    // A failed layout read must not discard successfully loaded
                    // surface/pattern context. Keep writes blocked until recovery.
                    if (!loadedPlacements.ok) {
                        setPlacements([])
                        setRefreshRequired(true)
                        throw loadedPlacements.error
                    }
                    setPlacements(loadedPlacements.items)
                    setSyncError(null)
                    setRefreshRequired(false)
                }
            } catch (caught) {
                if (active) {
                    setRefreshRequired(true)
                    setError(getApiErrorMessage(caught))
                }
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
    function issueForPose(id: number, pose: PlacementPose) {
        if (!surface) return 'Поверхность ещё не загружена.'
        const pattern = patterns.find((item) => item.id === pose.project_pattern_id)
        if (!pattern) return 'Лекало проекта недоступно.'
        const shape = transformPattern(pattern.geometry, pose)
        return placementIssue(surface.geometry, shape,
            placements.filter((item) => item.id !== id).map((item) => item.geometry))
    }
    async function commitManualDraft(next: ManualDraft, forceServer = false) {
        if (pending.current) return
        const problem = issueForPose(next.id, next.pose)
        setManualDraft({ ...next, notice: undefined })
        setSelected(next.id)
        setAdding(false)
        openInspector('properties')
        if (problem && !forceServer) return
        await mutate(async () => {
            try {
                if (next.kind === 'new') {
                    const item = await requestCreate(next.pose)
                    setPlacements((items) => [...items, item])
                    setHistory((value) => recordPlacement(value, {
                        id: item.id, before: null, after: poseOf(item),
                    }))
                    setSelected(item.id)
                } else {
                    const previous = placements.find((item) => item.id === next.id)
                    if (!previous) throw new Error('placement disappeared')
                    const item = await requestPatch(previous, next.pose)
                    setPlacements((items) => items.map((current) =>
                        current.id === next.id ? item : current))
                    setHistory((value) => recordPlacement(value, {
                        id: next.id, before: poseOf(previous), after: poseOf(item),
                    }))
                }
                setManualDraft(null)
            } catch (caught) {
                // The server remains authoritative: failed validation, stale data
                // and network errors must never discard the local draft.
                setManualDraft({ ...next, notice: getApiErrorMessage(caught) })
            }
        })
    }
    function editManualDraft(id: number, x: number, y: number, rotation: number) {
        const previous = manualDraft?.id === id
            ? manualDraft.pose
            : placements.find((item) => item.id === id)
        if (!previous) return null
        const next: ManualDraft = {
            id, kind: id === -1 ? 'new' : 'existing',
            pose: { project_pattern_id: previous.project_pattern_id, x, y,
                rotation: parseDraftRotation(String(rotation)) ?? rotation },
        }
        setManualDraft(next)
        return next
    }
    async function add(x: number, y: number, rotation: number) {
        if (!patternId || !adding || editingDisabled || pending.current) return
        if (parseDraftRotation(draftRotationText) === null) {
            setError('Укажите конечный угол поворота.')
            return
        }
        await commitManualDraft({ id: -1, kind: 'new',
            pose: { project_pattern_id: patternId, x, y, rotation } })
    }
    async function update(id: number, x: number, y: number, rotation?: number) {
        const previous = manualDraft?.id === id
            ? manualDraft.pose : placements.find((item) => item.id === id)
        if (!previous) return
        const nextRotation = rotation ?? previous.rotation
        if (x === previous.x && y === previous.y && nextRotation === previous.rotation) return
        const next = editManualDraft(id, x, y, nextRotation)
        if (next) {
            setManualFormRevision((value) => value + 1)
            await commitManualDraft(next)
        }
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
    function finishSavedOutcome(outcome: SaveOutcome, clear: () => void) {
        if (outcome.status === 'rejected') {
            clear()
            setRefreshRequired(false)
            setSyncError(null)
            setError(getApiErrorMessage(outcome.error))
            return
        }
        if (outcome.status === 'synchronized') {
            clear()
            setRefreshRequired(false)
            setSyncError(null)
            return
        }
        setRefreshRequired(true)
        setReplacement(null)
        if (outcome.status === 'confirmed') {
            setSyncError('Раскрой сохранён. Не удалось загрузить полный список размещений.')
            return
        }
        setSyncError(
            operation?.confirmed
                ? 'Раскрой сохранён. Не удалось проверить статус синхронизации. Повторите проверку позже.'
                : 'Подтверждение сохранения пока не получено. Новый сохраняющий запуск заблокирован: проверьте статус или повторите тот же запуск с сохранённым ключом.',
        )
        if (outcome.error) setError(getApiErrorMessage(outcome.error))
    }
    async function run(request: NestingRequest, rethrow = false, retry?: PendingNesting) {
        await mutate(async () => {
            if ((refreshRequired || operation) && !retry) return
            if (retry && !isPendingSaveScope(retry, userId, projectId, surfaceId)) return
            const began = performance.now()
            setResult(null)
            setResponseSeconds(null)
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
                kind: current.request.preview_id ? 'saving_preview' : 'calculating_saved',
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
                current.request,
                (response) => {
                    sessionStorage.setItem(
                        storageKey,
                        JSON.stringify({ ...current, confirmed: true }),
                    )
                    setOperation({ ...current, confirmed: true })
                    setActivity({ kind: 'syncing' })
                    setResult(response)
                    setResponseSeconds((performance.now() - began) / 1000)
                    setHistory({ undo: [], redo: [] })
                    setHistoryRevision((value) => value + 1)
                    setResultAlgorithm(current.request.algorithm)
                    setSelected(null)
                },
                setPlacements,
                current.key,
                !!retry,
            )
            finishSavedOutcome(status, clearOperation)
        }, rethrow)
    }
    async function checkOperation(current: PendingNesting) {
        await mutate(
            async () => {
                if (current.kind === undefined) {
                    if (!isPendingSaveScope(current, userId, projectId, surfaceId)) return
                    const outcome = await reconcileSavedNesting(
                        authorizedRequest,
                        base,
                        current.key,
                        (response) => {
                            sessionStorage.setItem(
                                storageKey,
                                JSON.stringify({ ...current, confirmed: true }),
                            )
                            setOperation({ ...current, confirmed: true })
                            setActivity({ kind: 'syncing' })
                            setResult(response)
                            setResultAlgorithm(current.request.algorithm)
                            setHistory({ undo: [], redo: [] })
                            setHistoryRevision((value) => value + 1)
                            setSelected(null)
                        },
                        setPlacements,
                    )
                    finishSavedOutcome(outcome, () => {
                        sessionStorage.removeItem(storageKey)
                        setOperation(null)
                    })
                    return
                }
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
    }
    async function previewRun(request: NestingRequest, retry?: PendingNesting) {
        await mutate(async () => {
            const validation = validateNestingRequest(retry?.request ?? request)
            if (validation) {
                setError(validation)
                return
            }
            const began = performance.now()
            setResult(null)
            setResponseSeconds(null)
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
                setResponseSeconds((performance.now() - began) / 1000)
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
        const parsed = parseQuantityDrafts(quantities)
        if (parsed.error) {
            setError(parsed.error)
            return
        }
        const requested = parsed.patterns
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
    const manualDraftIssue = manualDraft ? issueForPose(manualDraft.id, manualDraft.pose) : null
    const visualPlacements = useMemo(() => {
        if (!manualDraft) return placements
        const pattern = patterns.find((item) => item.id === manualDraft.pose.project_pattern_id)
        if (!pattern) return placements
        const rendered: Placement = {
            id: manualDraft.id,
            project_surface_id: surfaceId,
            updated_at: '',
            ...manualDraft.pose,
            geometry: transformPattern(pattern.geometry, manualDraft.pose),
        }
        if (manualDraft.kind === 'new') return [...placements, rendered]
        return placements.map((item) => item.id === manualDraft.id
            ? { ...item, ...manualDraft.pose, geometry: rendered.geometry } : item)
    }, [manualDraft, patterns, placements, surfaceId])
    const selectedPlacement = visualPlacements.find((item) => item.id === selected)
    function removeSelectedPlacement() {
        if (editingDisabled) return
        if (selectedPlacement?.id === -1) {
            // A new unsaved placement has never been written to PostgreSQL.
            setManualDraft(null)
            setSelected(null)
        } else if (selectedPlacement) {
            // Existing placements are removed only through explicit confirmation.
            setDeleting(true)
        } else if (addingPattern) {
            setAdding(false)
        }
    }
    const metrics = useMemo(
        () => (surface ? layoutMetrics(surface.geometry, placements) : null),
        [surface, placements],
    )
    return (
        <section
            className="nesting-page"
            onKeyDown={(event) => {
                const element = event.target instanceof HTMLElement ? event.target : null
                const action = draftKeyAction(event.key, {
                    adding,
                    blocked: editingDisabled,
                    editable: !!element?.closest('input,textarea,select,[contenteditable="true"]'),
                    modal: !!document.querySelector('dialog[open]'),
                    shift: event.shiftKey,
                    ctrl: event.ctrlKey,
                    alt: event.altKey,
                    meta: event.metaKey,
                })
                if (action === null) return
                event.preventDefault()
                if (action === 'cancel') setAdding(false)
                else changeDraftRotation(String(parseDraftRotation(String(draftRotation + action)) ?? 0))
            }}
        >
            <header>
                <Link to={`/projects/${projectId}/editor`} onClick={(event) => {
                    if (manualDraft) {
                        event.preventDefault()
                        requestManualTransition(() => navigate(`/projects/${projectId}/editor`))
                    }
                }}>← Состав проекта</Link>
                {projectName && <p>{projectName}</p>}
                {surfaceList.length > 1 && (
                    <label>
                        <span>Поверхность</span>{' '}
                        <select
                            value={surfaceId}
                            disabled={busy}
                            onChange={(event) => {
                                const destination = Number(event.target.value)
                                requestManualTransition(() => navigate(
                                    `/projects/${projectId}/surfaces/${destination}/nesting`,
                                ))
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
                            onClick={() => requestManualTransition(() => {
                                openInspector('automatic')
                                setAdding(false)
                            })}
                        >
                            Автораскрой
                        </button>
                        <details className="nesting-menu">
                            <summary>Ещё</summary>
                            <div className="nesting-menu-content">
                                <button
                                    disabled={editingDisabled || !history.undo.length}
                                    onClick={() => requestManualTransition(() => { void replay('undo') })}
                                >
                                    Отменить действие
                                </button>
                                <button
                                    disabled={editingDisabled || !history.redo.length}
                                    onClick={() => requestManualTransition(() => { void replay('redo') })}
                                >
                                    Повторить действие
                                </button>
                                <button disabled={busy} onClick={() => requestManualTransition(() => exportSaved('svg'))}>
                                    Экспорт SVG
                                </button>
                                <button disabled={busy} onClick={() => requestManualTransition(() => exportSaved('json'))}>
                                    Экспорт отчёта JSON
                                </button>
                            </div>
                        </details>
                    </div>
                )}
            </header>
            {operation && !execution && (
                <div role="status">
                    <p>
                        {operation.confirmed
                            ? 'Раскрой сохранён. Ожидается синхронизация размещений.'
                            : 'Ожидается подтверждение запуска раскроя.'}
                    </p>
                    <button disabled={busy} onClick={() => void checkOperation(operation)}>
                        Проверить статус
                    </button>
                    {operation.kind === 'preview' && (
                        <button disabled={busy} onClick={() => setClosePreview(operation)}>
                            Закрыть ожидание предпросмотра
                        </button>
                    )}
                    {!operation.confirmed && (
                        <button
                            disabled={busy}
                            onClick={() =>
                                void mutate(async () => {
                                    await authorizedRequest<void>(
                                        `${base}/nesting/operations/${operation.key}${operation.kind === 'preview' ? '?kind=preview' : ''}`,
                                        { method: 'DELETE', signal: AbortSignal.timeout(10_000) },
                                    )
                                    setSyncError(
                                        'Запрос отмены принят. Проверьте итоговый статус операции.',
                                    )
                                })
                            }
                        >
                            {operation.kind === 'preview'
                                ? 'Запросить отмену предпросмотра'
                                : 'Запросить отмену сохранения'}
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
            {result && !busy && !operation && !proposal && (
                <div role="status" className="nesting-result-summary">
                    <strong>
                        Сохранено · размещено {result.metrics.placed_count} из{' '}
                        {result.metrics.requested_count}
                    </strong>
                    <span>
                        Не размещено {result.metrics.requested_count - result.metrics.placed_count}.
                        {responseSeconds !== null
                            ? ` Время ответа ${format(responseSeconds)} с.`
                            : ''}
                    </span>
                    {result.metrics.placed_count === 0 && (
                        <p>
                            Расчёт завершён успешно, но подходящих свободных позиций не найдено.
                            Проверьте размеры, углы и занятые места.
                        </p>
                    )}
                </div>
            )}
            {!surface && !error && <p role="status">Загружаем рабочую поверхность…</p>}
            <div className="nesting-status">
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
                        <ExecutionProgress
                            key={execution.key}
                            request={operation?.request}
                            saving={execution.kind === 'saving_preview'}
                        />
                    </>
                ) : proposal ? (
                    <>
                        <span role="status">
                            Предпросмотр · размещено {proposal.result.metrics.placed_count} из{' '}
                            {proposal.result.metrics.requested_count}. Сохранённая раскладка
                            прежняя. Не размещено{' '}
                            {proposal.result.metrics.requested_count -
                                proposal.result.metrics.placed_count}
                            .
                            {responseSeconds !== null
                                ? ` Время ответа ${format(responseSeconds)} с.`
                                : ''}
                        </span>
                        <button className="nesting-button" onClick={() => setProposal(null)}>
                            Закрыть предпросмотр
                        </button>
                    </>
                ) : (
                    <span role="status">
                        {busy
                            ? 'Синхронизация с сервером…'
                            : manualDraft
                              ? manualDraft.notice || (manualDraftIssue ?? 'Положение изменено. Завершите ввод — корректная позиция сохранится автоматически.')
                            : addingPattern
                              ? `Щёлкните по поверхности, чтобы добавить «${addingPattern.name}»`
                              : 'Корректные ручные перемещения сохраняются автоматически'}
                    </span>
                )}
                {!editingDisabled && (selectedPlacement || addingPattern) && (
                    <button
                        type="button"
                        className="nesting-button nesting-delete-action"
                        onClick={removeSelectedPlacement}
                    >
                        {selectedPlacement?.id === -1
                            ? 'Убрать несохранённую деталь'
                            : selectedPlacement
                              ? `Удалить размещение №${selectedPlacement.id}`
                              : 'Отменить добавление'}
                    </button>
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
                                    onClick={() => requestManualTransition(() => {
                                        if (adding && patternId === item.id) {
                                            setAdding(false)
                                            return
                                        }
                                        setPatternId(item.id)
                                        changeDraftRotation('0')
                                        setAdding(true)
                                        if (window.matchMedia('(max-width: 700px)').matches)
                                            setPatternsOpen(false)
                                        setSelected(null)
                                        setError(null)
                                    })}
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
                                    if (manualDraft?.id === item.id) return
                                    requestManualTransition(() => {
                                        setSelected(item.id)
                                        setAdding(false)
                                    })
                                }}
                            >
                                №{item.id} · лекало №{item.project_pattern_id}
                            </button>
                        ))}
                    </aside>
                    <NestingCanvas
                        surface={surface.geometry}
                        placements={visualPlacements}
                        selected={selected}
                        busy={editingDisabled}
                        adding={!!addingPattern}
                        draftPattern={addingPattern?.geometry}
                        draftPatternId={addingPattern?.id}
                        draftRotation={draftRotation}
                        manualDraftId={manualDraft?.id}
                        manualDraftIssue={manualDraftIssue || manualDraft?.notice}
                        validatePose={issueForPose}
                        canEdit={(id) => !manualDraft || manualDraft.id === id}
                        draftControls={
                            addingPattern ? (
                                <div
                                    className="nesting-draft-tools"
                                    role="toolbar"
                                    aria-label="Постановка лекала"
                                >
                                    <strong>На курсоре: {addingPattern.name}</strong>
                                    <label>
                                        Поворот до постановки, °
                                        <input
                                            type="number"
                                            step="any"
                                            value={draftRotationText}
                                            disabled={editingDisabled}
                                            onChange={(event) =>
                                                changeDraftRotation(event.target.value)
                                            }
                                            onBlur={() => {
                                                const angle = parseDraftRotation(draftRotationText)
                                                if (angle !== null)
                                                    changeDraftRotation(String(angle))
                                            }}
                                            aria-invalid={
                                                parseDraftRotation(draftRotationText) === null
                                            }
                                        />
                                    </label>
                                    <div className="nesting-tools">
                                        {[0, 90, 180, 270].map((angle) => (
                                            <button
                                                type="button"
                                                key={angle}
                                                aria-pressed={draftRotation === angle}
                                                disabled={editingDisabled}
                                                onClick={() => changeDraftRotation(String(angle))}
                                            >
                                                {angle}°
                                            </button>
                                        ))}
                                    </div>
                                    <button type="button" onClick={() => setAdding(false)}>
                                        Отменить добавление
                                    </button>
                                    <small>
                                        ←/→: ±1° · Shift+←/→: ±15° · R: +90° · Shift+R: −90° · Esc: отмена.
                                        Для точного угла контур останется видимым при переходе в поле ввода.
                                    </small>
                                    {parseDraftRotation(draftRotationText) === null && (
                                        <span role="alert">Укажите конечный угол.</span>
                                    )}
                                </div>
                            ) : null
                        }
                        proposedPlacements={proposal?.result.placements}
                        showSaved={!proposal || proposal.includeSaved}
                        onSelect={(id) => {
                            if (id === selected) return
                            requestManualTransition(() => {
                                setSelected(id)
                                if (id !== null) setPanel('properties')
                            })
                        }}
                        onAdd={(x, y, rotation) => void add(x, y, rotation)}
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
                                onClick={() => requestManualTransition(() => {
                                    setPanel('automatic')
                                    setAdding(false)
                                })}
                            >
                                Автораскрой
                            </button>
                            <button
                                type="button"
                                aria-pressed={panel === 'history'}
                                onClick={() => requestManualTransition(() => setPanel('history'))}
                            >
                                История
                            </button>
                        </nav>
                        {panel === 'properties' && !selectedPlacement && (
                            <p>Выберите размещение на поверхности, чтобы изменить его положение.</p>
                        )}
                        {panel === 'properties' && selectedPlacement && (
                            <PlacementProperties
                                key={`${selectedPlacement.id}:${manualFormRevision}`}
                                placement={selectedPlacement}
                                busy={editingDisabled}
                                newPlacement={selectedPlacement.id === -1}
                                onDraftChange={(x, y, angle) => {
                                    editManualDraft(selectedPlacement.id, x, y, angle)
                                }}
                                onSave={(x, y, angle) => {
                                    const next = editManualDraft(selectedPlacement.id, x, y, angle)
                                    if (next) void commitManualDraft(next)
                                }}
                            />
                        )}
                        {panel === 'automatic' && (
                            <section>
                                <h2>Автоматический раскрой</h2>
                                <fieldset className="nesting-quantities">
                                    <legend>Количество лекал</legend>
                                    <small>
                                        Пусто или 0 — не выбирать лекало. Всего не более 500.
                                    </small>
                                    {patterns.map((item) => (
                                        <label key={item.id}>
                                            {item.name}
                                            <input
                                                type="number"
                                                min="0"
                                                max="500"
                                                step="1"
                                                value={quantities[item.id] ?? ''}
                                                placeholder="0"
                                                disabled={busy}
                                                onChange={(event) =>
                                                    setQuantities({
                                                        ...quantities,
                                                        [item.id]: event.target.value,
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
                                <p>
                                    Предпросмотр: рассчитать и показать, не изменяя сохранённую
                                    раскладку. Расчёт и сохранение сразу записывают результат в
                                    проект. Предел расчёта — 120 секунд.
                                </p>
                                {keepExisting && (
                                    <p>
                                        Новые лекала рассчитываются с учётом всех текущих
                                        размещений: {placements.length}.
                                    </p>
                                )}
                                <div className="nesting-auto-actions">
                                    <button
                                        disabled={
                                            busy ||
                                            !!operation ||
                                            refreshRequired ||
                                            !patterns.length
                                        }
                                        onClick={() => prepareRun(true)}
                                    >
                                        Предпросмотр
                                    </button>
                                    <button
                                        className="nesting-primary"
                                        disabled={
                                            busy ||
                                            refreshRequired ||
                                            !!operation ||
                                            !patterns.length
                                        }
                                        onClick={() => (proposal ? saveProposal() : prepareRun())}
                                    >
                                        {proposal
                                            ? 'Сохранить этот предпросмотр'
                                            : 'Рассчитать и сохранить'}
                                    </button>
                                </div>
                                {proposal && (
                                    <p>
                                        Не размещено:{' '}
                                        {proposal.result.metrics.requested_count -
                                            proposal.result.metrics.placed_count}
                                        .{' '}
                                        {proposal.result.metrics.placed_count === 0 &&
                                            'Подходящих свободных позиций не найдено; проверьте размеры, углы и занятые места. '}
                                        Можно сохранить показанную раскладку без повторного расчёта.
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
            {deleting && selectedPlacement && selectedPlacement.id !== -1 && (
                <ConfirmDialog
                    title="Удалить размещение?"
                    onClose={() => setDeleting(false)}
                    onConfirm={() =>
                        mutate(async () => {
                            // Use persisted data, not the possibly invalid visual draft.
                            const savedPlacement = placements.find(
                                (item) => item.id === selectedPlacement.id,
                            )
                            if (!savedPlacement) throw new Error('Размещение уже недоступно.')
                            await requestDelete(savedPlacement)
                            setHistory((value) =>
                                recordPlacement(value, {
                                    id: savedPlacement.id,
                                    before: poseOf(savedPlacement),
                                    after: null,
                                }),
                            )
                            setPlacements((items) =>
                                items.filter((item) => item.id !== savedPlacement.id),
                            )
                            setSelected(null)
                            setManualDraft(null)
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
            {showDraftDecision && manualDraft && (
                <ManualDraftDialog
                    newPlacement={manualDraft.kind === 'new'}
                    onContinue={() => {
                        setShowDraftDecision(false)
                        deferredManualAction.current = null
                    }}
                    onDiscard={finishManualTransition}
                    onDelete={async () => {
                        const previous = placements.find((item) => item.id === manualDraft.id)
                        if (!previous) throw new Error('Сохранённое размещение не найдено.')
                        // Explicit, version-checked deletion only after user confirmation.
                        await requestDelete(previous)
                        setPlacements((items) => items.filter((item) => item.id !== previous.id))
                        setHistory((value) => recordPlacement(value, {
                            id: previous.id, before: poseOf(previous), after: null,
                        }))
                        finishManualTransition()
                    }}
                />
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
