import { GeometryCanvas } from './GeometryCanvas'
import { GeometryWorkspace } from './GeometryWorkspace'
import { useGeometryHistory } from './useGeometryHistory'
import { useWheelZoom } from '../../shared/geometry/useWheelZoom'
import { ConfirmDialog } from '../../components/dialogs/ConfirmDialog'
import { resizeView, screenToWorld, zoomView } from '../../shared/geometry/viewport'
import { useEffect, useRef, useState } from 'react'
import type { FormEvent, PointerEvent } from 'react'

import { getApiErrorMessage } from '../../shared/api/errorMessages'
import type { LibraryKind } from './api'
import type { CreateLibraryItemRequest, LibraryGeometry, LibraryItem, LibraryPoint } from './types'
import {
    fitDrawingView,
    rectangleGeometry,
    validatePolygonConstruction,
    validatePolygonDraft,
} from './geometryDraft'
import type { DrawingViewBox } from './geometryDraft'
import { insertVertex } from './insertVertex'
import { importSVGGeometry } from './svgImport'

import './CreateLibraryItemDialog.css'

interface Props {
    kind: LibraryKind
    onClose: () => void
    initialItem?: LibraryItem
    onSave: (data: CreateLibraryItemRequest) => Promise<void>
}

type EditorMode = 'rectangle' | 'points'
interface GeometrySnapshot {
    draft: LibraryGeometry
    mode: EditorMode
    width: string
    height: string
    ring: number
    vertex: number | null
}

function CoordinateInput({
    value,
    label,
    onCommit,
}: {
    value: number
    label: string
    onCommit: (value: number) => boolean
}) {
    const [raw, setRaw] = useState(String(value))

    function commit() {
        const number = Number(raw.trim().replace(',', '.'))
        if (raw.trim() && Number.isFinite(number)) {
            const accepted = onCommit(number)
            setRaw(accepted ? String(number) : String(value))
        } else {
            setRaw(String(value))
        }
    }

    return (
        <input
            aria-label={label}
            inputMode="decimal"
            type="text"
            value={raw}
            onChange={(event) => setRaw(event.target.value)}
            onBlur={commit}
            onKeyDown={(event) => {
                if (event.key === 'Enter') {
                    event.preventDefault()
                    event.currentTarget.blur()
                }
            }}
        />
    )
}

const EMPTY: LibraryGeometry = {
    exterior: { points: [] },
    holes: [],
}

function cloneGeometry(value: LibraryGeometry): LibraryGeometry {
    return {
        exterior: { points: value.exterior.points.map((point) => ({ ...point })) },
        holes: value.holes.map((hole) => ({
            points: hole.points.map((point) => ({ ...point })),
        })),
    }
}

export function CreateLibraryItemDialog({ kind, onClose, onSave, initialItem }: Props) {
    const dialogRef = useRef<HTMLDialogElement>(null)
    const fileRef = useRef<HTMLInputElement>(null)
    const svgRef = useRef<SVGSVGElement>(null)
    const [initialDraft] = useState<LibraryGeometry>(() =>
        cloneGeometry(initialItem?.geometry ?? EMPTY),
    )
    const draftRef = useRef<LibraryGeometry>(initialDraft)
    const dragRef = useRef<{
        ring: number
        index: number
        pointerId: number
        offset: LibraryPoint
        before: GeometrySnapshot
        changed: boolean
    } | null>(null)
    const allowHoles = kind === 'surfaces'

    const [name, setName] = useState(initialItem?.name ?? '')
    // Editing starts in free-point mode to preserve arbitrary existing geometry.
    const [mode, setMode] = useState<EditorMode>(initialItem ? 'points' : 'rectangle')
    const [width, setWidth] = useState(() => {
        const points = initialItem?.geometry.exterior.points
        if (!points?.length) return kind === 'surfaces' ? '1000' : '100'
        return String(Math.max(...points.map((p) => p.x)) - Math.min(...points.map((p) => p.x)))
    })
    const [height, setHeight] = useState(() => {
        const points = initialItem?.geometry.exterior.points
        if (!points?.length) return kind === 'surfaces' ? '600' : '60'
        return String(Math.max(...points.map((p) => p.y)) - Math.min(...points.map((p) => p.y)))
    })
    const [draft, setDraft] = useState<LibraryGeometry>(initialDraft)
    const [activeRing, setActiveRing] = useState(-1) // -1 = exterior, 0+ = holes
    const [selectedVertex, setSelectedVertex] = useState<number | null>(null)
    const [selectMode, setSelectMode] = useState(false)
    const history = useGeometryHistory<GeometrySnapshot>()
    const sizeBeforeEdit = useRef<GeometrySnapshot | null>(null)
    const [newX, setNewX] = useState('0')
    const [newY, setNewY] = useState('0')
    const [view, setView] = useState<DrawingViewBox>(() =>
        fitDrawingView(
            initialItem
                ? initialDraft
                : rectangleGeometry(
                      kind === 'surfaces' ? 1000 : 100,
                      kind === 'surfaces' ? 600 : 60,
                  ),
        ),
    )
    const [gridStep, setGridStep] = useState(10)
    const [gridStepText, setGridStepText] = useState('10')
    const [showGrid, setShowGrid] = useState(true)
    const [panMode, setPanMode] = useState(false)
    const [confirmRectangle, setConfirmRectangle] = useState(false)
    const [aspect, setAspect] = useState(640 / 420)
    const panRef = useRef<{
        id: number
        x: number
        y: number
        view: DrawingViewBox
        scale: number
    } | null>(null)
    const visibleView = resizeView(view, aspect)
    const [initialView, setInitialView] = useState(view)
    function fitToCanvas(value: LibraryGeometry) {
        const bounds = fitDrawingView(value, aspect)
        setView(bounds)
    }
    const initialFitDone = useRef(false)
    useWheelZoom(svgRef, visibleView, setView, () => !dragRef.current && !panRef.current)
    useEffect(() => {
        const svg = svgRef.current
        if (!svg) return
        const observer = new ResizeObserver(([entry]) => {
            const { width, height } = entry.contentRect
            if (width <= 0 || height <= 0) return
            const nextAspect = width / height
            setAspect(nextAspect)
            if (!initialFitDone.current) {
                initialFitDone.current = true
                const fitted = fitDrawingView(
                    initialItem
                        ? initialDraft
                        : rectangleGeometry(
                              kind === 'surfaces' ? 1000 : 100,
                              kind === 'surfaces' ? 600 : 60,
                          ),
                    nextAspect,
                )
                setView(fitted)
                setInitialView(fitted)
            }
        })
        observer.observe(svg)
        return () => observer.disconnect()
    }, [initialDraft, initialItem, kind])
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const dialog = dialogRef.current
        if (dialog && !dialog.open) dialog.showModal()
    }, [])

    const w = Number(width)
    const h = Number(height)
    const rectangleValid =
        width.trim() !== '' &&
        height.trim() !== '' &&
        Number.isFinite(w) &&
        Number.isFinite(h) &&
        w > 0 &&
        h > 0
    // The rectangle defines the exterior; holes are preserved and editable in BOTH modes.
    const geometry: LibraryGeometry =
        mode === 'rectangle'
            ? {
                  exterior: (rectangleValid ? rectangleGeometry(w, h) : EMPTY).exterior,
                  holes: draft.holes,
              }
            : draft
    const validationError = validatePolygonDraft(geometry, allowHoles)
    const activePoints =
        activeRing < 0 ? geometry.exterior.points : (geometry.holes[activeRing]?.points ?? [])
    const rings = [geometry.exterior, ...geometry.holes]

    function snapshot(): GeometrySnapshot {
        return {
            draft: draftRef.current,
            mode,
            width,
            height,
            ring: activeRing,
            vertex: selectedVertex,
        }
    }

    function restore(value: GeometrySnapshot) {
        draftRef.current = value.draft
        setDraft(value.draft)
        setMode(value.mode)
        setWidth(value.width)
        setHeight(value.height)
        setActiveRing(value.ring)
        setSelectedVertex(value.vertex)
        setError(null)
    }

    function replay(direction: 'undo' | 'redo') {
        if (dragRef.current || panRef.current || saving) return
        const value = history[direction](snapshot())
        if (value) restore(value)
    }

    function selectRing(ring: number) {
        setActiveRing(ring)
        setSelectedVertex(null)
    }

    function editSize(set: (value: string) => void, value: string) {
        if (!sizeBeforeEdit.current) {
            const before = snapshot()
            sizeBeforeEdit.current = before
            history.record(before)
        }
        set(value)
        setError(null)
    }

    /** Guard ALL input methods: clicks, typed coordinates, vertex dragging and deletion. */
    function tryApplyDraft(candidate: LibraryGeometry): boolean {
        const next =
            mode === 'rectangle' ? { ...candidate, exterior: geometry.exterior } : candidate
        const problem = validatePolygonConstruction(next, allowHoles)
        if (problem) {
            setError(problem)
            return false
        }
        if (!dragRef.current) history.record(snapshot())
        draftRef.current = candidate
        setDraft(candidate)
        setError(null)
        return true
    }

    function updateRing(
        ringIndex: number,
        updater: (points: LibraryPoint[]) => LibraryPoint[],
    ): boolean {
        if (ringIndex < 0 && mode === 'rectangle') {
            setError('Для изменения внешнего контура переключитесь в режим «По точкам».')
            return false
        }
        const previous = draftRef.current
        if (ringIndex >= previous.holes.length) return false
        const next: LibraryGeometry =
            ringIndex < 0
                ? { ...previous, exterior: { points: updater(previous.exterior.points) } }
                : {
                      ...previous,
                      holes: previous.holes.map((hole, index) =>
                          index === ringIndex ? { points: updater(hole.points) } : hole,
                      ),
                  }
        return tryApplyDraft(next)
    }

    function updateActive(updater: (points: LibraryPoint[]) => LibraryPoint[]): boolean {
        return updateRing(activeRing, updater)
    }

    function addPoint(point: LibraryPoint) {
        // After the first three points, choose an edge rather than always appending
        // after the last vertex. This includes the closing edge (last -> first).
        const nextPoints = insertVertex(activePoints, point, (candidatePoints) => {
            const candidateGeometry: LibraryGeometry =
                activeRing < 0
                    ? { ...geometry, exterior: { points: candidatePoints } }
                    : {
                          ...geometry,
                          holes: geometry.holes.map((hole, index) =>
                              index === activeRing ? { points: candidatePoints } : hole,
                          ),
                      }
            return validatePolygonConstruction(candidateGeometry, allowHoles) === null
        })

        if (!nextPoints) {
            setError(
                'Не удалось вставить вершину без пересечений. Попробуйте точку ближе к нужному ребру.',
            )
            return
        }
        if (updateActive(() => nextPoints)) setSelectedVertex(nextPoints.indexOf(point))
    }

    function pointFromPointer(event: PointerEvent<SVGSVGElement>): LibraryPoint | null {
        return screenToWorld(event.currentTarget, event.clientX, event.clientY)
    }

    function handleCanvasPointerDown(event: PointerEvent<SVGSVGElement>) {
        if (saving || !event.isPrimary || dragRef.current || panRef.current) return
        if (panMode || event.button === 1) {
            event.preventDefault()
            const bounds = event.currentTarget.getBoundingClientRect()
            panRef.current = {
                id: event.pointerId,
                x: event.clientX,
                y: event.clientY,
                view: visibleView,
                scale: visibleView.width / bounds.width,
            }
            event.currentTarget.setPointerCapture(event.pointerId)
            return
        }
        if (event.button !== 0) return
        if (selectMode) {
            setSelectedVertex(null)
            return
        }
        if (mode === 'rectangle' && activeRing === -1) return
        const point = pointFromPointer(event)
        if (point)
            addPoint({ x: Math.round(point.x * 100) / 100, y: Math.round(point.y * 100) / 100 })
    }

    function startDragging(event: PointerEvent<SVGCircleElement>, ring: number, index: number) {
        if (
            panMode ||
            saving ||
            !event.isPrimary ||
            dragRef.current ||
            panRef.current ||
            event.button !== 0
        )
            return
        event.preventDefault()
        event.stopPropagation() // A vertex is not a new point.

        const svg = svgRef.current
        if (!svg) return
        const pointer = screenToWorld(svg, event.clientX, event.clientY)
        const vertex = rings[ring + 1].points[index]
        if (!pointer) return
        dragRef.current = {
            ring,
            index,
            pointerId: event.pointerId,
            before: snapshot(),
            changed: false,
            offset: { x: vertex.x - pointer.x, y: vertex.y - pointer.y },
        }
        setActiveRing(ring)
        setSelectedVertex(index)
        svg.setPointerCapture(event.pointerId)
    }

    function handlePointerMove(event: PointerEvent<SVGSVGElement>) {
        const pan = panRef.current
        if (pan && pan.id === event.pointerId) {
            setView({
                ...pan.view,
                x: pan.view.x - (event.clientX - pan.x) * pan.scale,
                y: pan.view.y - (event.clientY - pan.y) * pan.scale,
            })
            return
        }
        const drag = dragRef.current
        if (!drag || drag.pointerId !== event.pointerId) return
        const point = pointFromPointer(event)
        if (!point) return
        const next = {
            x: Math.round((point.x + drag.offset.x) * 100) / 100,
            y: Math.round((point.y + drag.offset.y) * 100) / 100,
        }
        const current = rings[drag.ring + 1]?.points[drag.index]
        if (
            !current ||
            (Math.abs(current.x - next.x) < 0.006 && Math.abs(current.y - next.y) < 0.006)
        )
            return
        if (mode === 'rectangle' && drag.ring < 0) {
            const candidate = {
                ...geometry,
                exterior: {
                    points: geometry.exterior.points.map((value, i) =>
                        i === drag.index ? next : value,
                    ),
                },
            }
            const problem = validatePolygonConstruction(candidate, allowHoles)
            if (problem) {
                setError(problem)
                return
            }
            draftRef.current = candidate
            setDraft(candidate)
            setMode('points')
            setError(null)
            drag.changed = true
            return
        }
        if (
            updateRing(drag.ring, (points) => points.map((p, i) => (i === drag.index ? next : p)))
        ) {
            drag.changed = true
        }
    }

    function finishDrag(pointerId: number, cancelled: boolean) {
        if (panRef.current?.id === pointerId) panRef.current = null
        const drag = dragRef.current
        if (drag?.pointerId !== pointerId) return
        dragRef.current = null
        if (drag.changed) {
            if (cancelled) restore(drag.before)
            else history.record(drag.before)
        }
    }

    function stopDragging(event: PointerEvent<SVGSVGElement>) {
        finishDrag(event.pointerId, event.type === 'pointercancel')
        if (event.currentTarget.hasPointerCapture(event.pointerId))
            event.currentTarget.releasePointerCapture(event.pointerId)
    }

    function handleManualPoint() {
        if (!newX.trim() || !newY.trim()) {
            setError('Введите обе координаты точки.')
            return
        }
        const x = Number(newX.trim().replace(',', '.'))
        const y = Number(newY.trim().replace(',', '.'))
        if (!Number.isFinite(x) || !Number.isFinite(y)) {
            setError('Координаты должны быть числами.')
            return
        }
        addPoint({ x, y })
    }

    function switchRectangleToPoints() {
        if (!rectangleValid) {
            setError('Сначала укажите положительные размеры прямоугольника.')
            return
        }
        history.record(snapshot())
        const next: LibraryGeometry = {
            exterior: rectangleGeometry(w, h).exterior,
            holes: draftRef.current.holes,
        }
        draftRef.current = next
        setDraft(next)
        setMode('points')
        selectRing(-1)
        fitToCanvas(next)
        setError(null)
    }

    function changeMode(next: EditorMode) {
        if (next === mode) return
        if (next === 'points') {
            switchRectangleToPoints()
            return
        }
        setConfirmRectangle(true)
    }

    function applyRectangle() {
        history.record(snapshot())
        setMode('rectangle')
        selectRing(-1)
        fitToCanvas({
            exterior: rectangleGeometry(w > 0 ? w : 1, h > 0 ? h : 1).exterior,
            holes: draftRef.current.holes,
        })
        setError(null)
    }

    function addHole() {
        if (!allowHoles || draftRef.current.holes.length >= 20) return
        // Avoid confusing partially drawn holes and drawing against an invalid exterior.
        const current: LibraryGeometry = {
            exterior: geometry.exterior,
            holes: draftRef.current.holes,
        }
        const problem = validatePolygonDraft(current, true)
        if (problem) {
            setError(`Перед новым отверстием завершите предыдущие контуры. ${problem}`)
            return
        }
        history.record(snapshot())
        const index = draftRef.current.holes.length
        const next: LibraryGeometry = {
            ...draftRef.current,
            holes: [...draftRef.current.holes, { points: [] }],
        }
        draftRef.current = next
        setDraft(next)
        selectRing(index)
        setError(null)
    }

    function removeHole(index: number) {
        history.record(snapshot())
        const next: LibraryGeometry = {
            ...draftRef.current,
            holes: draftRef.current.holes.filter((_, i) => i !== index),
        }
        draftRef.current = next
        setDraft(next)
        selectRing(-1)
        setError(null)
    }

    async function importFile(file: File | undefined) {
        if (!file) return
        if (file.size > 1024 * 1024) {
            setError('SVG слишком большой. Максимальный размер — 1 МБ.')
            return
        }
        try {
            const text = await file.text()
            const next = importSVGGeometry(text, allowHoles)
            history.record(snapshot())
            draftRef.current = next
            setDraft(next)
            setMode('points')
            selectRing(-1)
            fitToCanvas(next)
            setError(null)
        } catch (caught) {
            setError(caught instanceof Error ? caught.message : 'Не удалось импортировать SVG.')
        }
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        if (saving) return
        const trimmed = name.trim()
        if (!trimmed) {
            setError('Введите название.')
            return
        }
        if (validationError) {
            setError(validationError)
            return
        }
        setSaving(true)
        setError(null)
        try {
            const request: CreateLibraryItemRequest = { name: trimmed, geometry }
            await onSave(request)
            dialogRef.current?.close()
        } catch (caught) {
            setError(getApiErrorMessage(caught))
        } finally {
            setSaving(false)
        }
    }

    return (
        <dialog
            ref={dialogRef}
            className="geometry-dialog"
            aria-labelledby="geometry-dialog-title"
            onClose={onClose}
            onCancel={(event) => {
                if (saving) event.preventDefault()
            }}
            onKeyDown={(event) => {
                if (event.key === 'Escape' && !saving && !confirmRectangle) {
                    event.preventDefault()
                    dialogRef.current?.close()
                }
            }}
        >
            <form onSubmit={(event) => void handleSubmit(event)}>
                <GeometryWorkspace
                    toolbar={
                        <>
                            <div className="geometry-toolbar-title">
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    disabled={saving}
                                    onClick={() => dialogRef.current?.close()}
                                >
                                    ← Назад
                                </button>
                                <h2 id="geometry-dialog-title" title={name}>
                                    {name.trim() ||
                                        (allowHoles ? 'Новая поверхность' : 'Новое лекало')}
                                </h2>
                            </div>
                            <div
                                className="geometry-tool-group"
                                role="group"
                                aria-label="Вид холста"
                            >
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    aria-label="Уменьшить масштаб"
                                    title="Уменьшить масштаб"
                                    onClick={() => setView(zoomView(visibleView, 1.25))}
                                >
                                    −
                                </button>
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    title="Вернуть исходный масштаб"
                                    onClick={() => setView(initialView)}
                                >
                                    {Math.round((initialView.height / view.height) * 100)}%
                                </button>
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    aria-label="Увеличить масштаб"
                                    title="Увеличить масштаб"
                                    onClick={() => setView(zoomView(visibleView, 0.8))}
                                >
                                    +
                                </button>
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    onClick={() => fitToCanvas(geometry)}
                                >
                                    Уместить в поле
                                </button>
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    aria-pressed={panMode}
                                    onClick={() => setPanMode(!panMode)}
                                >
                                    Перемещение поля
                                </button>
                            </div>
                            <div
                                className="geometry-tool-group"
                                role="group"
                                aria-label="История геометрии"
                            >
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    disabled={saving || !history.canUndo}
                                    onClick={() => replay('undo')}
                                >
                                    Отменить
                                </button>
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    disabled={saving || !history.canRedo}
                                    onClick={() => replay('redo')}
                                >
                                    Повторить
                                </button>
                            </div>
                            <button
                                type="submit"
                                className="geometry-primary"
                                disabled={saving || !name.trim() || !!validationError}
                            >
                                {saving
                                    ? 'Сохраняем…'
                                    : initialItem
                                      ? 'Сохранить изменения'
                                      : 'Создать'}
                            </button>
                        </>
                    }
                    tools={
                        <fieldset disabled={saving}>
                            <legend>Построение</legend>
                            <div
                                className="geometry-mode"
                                role="group"
                                aria-label="Способ создания"
                            >
                                <button
                                    type="button"
                                    aria-pressed={mode === 'rectangle'}
                                    className={mode === 'rectangle' ? 'geometry-mode-active' : ''}
                                    onClick={() => changeMode('rectangle')}
                                    disabled={saving}
                                >
                                    Прямоугольник
                                </button>
                                <button
                                    type="button"
                                    aria-pressed={mode === 'points'}
                                    className={mode === 'points' ? 'geometry-mode-active' : ''}
                                    onClick={() => changeMode('points')}
                                    disabled={saving}
                                >
                                    По точкам
                                </button>
                                <button
                                    type="button"
                                    disabled={saving}
                                    onClick={() => fileRef.current?.click()}
                                >
                                    Импорт SVG…
                                </button>
                                <input
                                    ref={fileRef}
                                    type="file"
                                    accept=".svg,image/svg+xml"
                                    hidden
                                    onChange={(event) => {
                                        void importFile(event.target.files?.[0])
                                        event.target.value = ''
                                    }}
                                />
                            </div>

                            <div
                                className="geometry-ring-tabs"
                                role="group"
                                aria-label="Выбор контура"
                            >
                                <button
                                    type="button"
                                    aria-pressed={activeRing === -1}
                                    className={activeRing === -1 ? 'geometry-tab-active' : ''}
                                    onClick={() => selectRing(-1)}
                                >
                                    Внешний контур ({geometry.exterior.points.length})
                                </button>
                                {allowHoles &&
                                    draft.holes.map((hole, index) => (
                                        <button
                                            key={index}
                                            type="button"
                                            aria-pressed={activeRing === index}
                                            className={
                                                activeRing === index ? 'geometry-tab-active' : ''
                                            }
                                            onClick={() => selectRing(index)}
                                        >
                                            Отверстие {index + 1} ({hole.points.length})
                                        </button>
                                    ))}
                                {allowHoles && (
                                    <button
                                        type="button"
                                        className="geometry-secondary"
                                        onClick={addHole}
                                        disabled={
                                            saving || !!validationError || draft.holes.length >= 20
                                        }
                                    >
                                        + Отверстие
                                    </button>
                                )}
                            </div>

                            <div
                                className="geometry-mode"
                                role="group"
                                aria-label="Инструмент холста"
                            >
                                <button
                                    type="button"
                                    aria-pressed={selectMode && !panMode}
                                    onClick={() => {
                                        setSelectMode(true)
                                        setPanMode(false)
                                    }}
                                >
                                    Выбор вершин
                                </button>
                                <button
                                    type="button"
                                    disabled={mode === 'rectangle' && activeRing < 0}
                                    aria-pressed={
                                        !selectMode &&
                                        !panMode &&
                                        (mode === 'points' || activeRing >= 0)
                                    }
                                    onClick={() => {
                                        setSelectMode(false)
                                        setPanMode(false)
                                    }}
                                >
                                    Добавление вершин
                                </button>
                            </div>
                            {(mode === 'points' || activeRing >= 0) && (
                                <div className="geometry-mode">
                                    <button
                                        type="button"
                                        className="geometry-danger"
                                        disabled={!activePoints.length}
                                        onClick={() => {
                                            if (updateActive(() => [])) setSelectedVertex(null)
                                        }}
                                    >
                                        Очистить контур
                                    </button>
                                    {activeRing >= 0 && (
                                        <button
                                            type="button"
                                            className="geometry-danger"
                                            onClick={() => removeHole(activeRing)}
                                        >
                                            Удалить отверстие
                                        </button>
                                    )}
                                </div>
                            )}
                            <p className="geometry-help">
                                Импорт SVG: polygon, rect и прямолинейные path (M/L/H/V/Z).
                            </p>
                        </fieldset>
                    }
                    properties={
                        <fieldset disabled={saving}>
                            <legend>Точное редактирование</legend>
                            <label className="geometry-field">
                                Название
                                <input
                                    type="text"
                                    value={name}
                                    required
                                    maxLength={150}
                                    onChange={(event) => setName(event.target.value)}
                                    disabled={saving}
                                    placeholder={
                                        allowHoles
                                            ? 'Например, Лист материала'
                                            : 'Например, Деталь A'
                                    }
                                />
                            </label>

                            {mode === 'rectangle' && (
                                <div className="geometry-rectangle">
                                    <label className="geometry-field">
                                        Ширина
                                        <input
                                            type="number"
                                            min="0"
                                            step="any"
                                            value={width}
                                            disabled={saving}
                                            onChange={(event) =>
                                                editSize(setWidth, event.target.value)
                                            }
                                            onBlur={() => {
                                                sizeBeforeEdit.current = null
                                            }}
                                        />
                                    </label>
                                    <label className="geometry-field">
                                        Высота
                                        <input
                                            type="number"
                                            min="0"
                                            step="any"
                                            value={height}
                                            disabled={saving}
                                            onChange={(event) =>
                                                editSize(setHeight, event.target.value)
                                            }
                                            onBlur={() => {
                                                sizeBeforeEdit.current = null
                                            }}
                                        />
                                    </label>
                                    <button
                                        className="geometry-secondary"
                                        type="button"
                                        onClick={switchRectangleToPoints}
                                        disabled={saving || !rectangleValid}
                                    >
                                        Редактировать вершины
                                    </button>
                                </div>
                            )}

                            {(mode === 'points' || activeRing >= 0) && (
                                <>
                                    <label className="geometry-field">
                                        Вершина контура
                                        <select
                                            value={selectedVertex ?? ''}
                                            onChange={(event) =>
                                                setSelectedVertex(
                                                    event.target.value === ''
                                                        ? null
                                                        : Number(event.target.value),
                                                )
                                            }
                                        >
                                            <option value="">Выберите вершину</option>
                                            {activePoints.map((_, index) => (
                                                <option key={index} value={index}>
                                                    Вершина № {index + 1}
                                                </option>
                                            ))}
                                        </select>
                                    </label>
                                    {selectedVertex !== null && activePoints[selectedVertex] ? (
                                        <div className="geometry-point-add">
                                            {(['x', 'y'] as const).map((axis) => {
                                                const point = activePoints[selectedVertex]
                                                return (
                                                    <label key={axis} className="geometry-field">
                                                        {axis.toUpperCase()}
                                                        <CoordinateInput
                                                            key={`${activeRing}-${selectedVertex}-${point[axis]}`}
                                                            value={point[axis]}
                                                            label={`${axis.toUpperCase()} точки ${selectedVertex + 1}`}
                                                            onCommit={(coordinate) =>
                                                                coordinate === point[axis] ||
                                                                updateActive((points) =>
                                                                    points.map((entry, index) =>
                                                                        index === selectedVertex
                                                                            ? {
                                                                                  ...entry,
                                                                                  [axis]: coordinate,
                                                                              }
                                                                            : entry,
                                                                    ),
                                                                )
                                                            }
                                                        />
                                                    </label>
                                                )
                                            })}
                                            <button
                                                type="button"
                                                className="geometry-secondary geometry-danger"
                                                onClick={() => {
                                                    if (
                                                        updateActive((points) =>
                                                            points.filter(
                                                                (_, index) =>
                                                                    index !== selectedVertex,
                                                            ),
                                                        )
                                                    )
                                                        setSelectedVertex(null)
                                                }}
                                            >
                                                Удалить вершину
                                            </button>
                                        </div>
                                    ) : (
                                        <p className="geometry-help">
                                            Выберите вершину на холсте или в списке.
                                        </p>
                                    )}
                                    <div className="geometry-point-add">
                                        <label className="geometry-field">
                                            Новая X
                                            <input
                                                type="text"
                                                inputMode="decimal"
                                                value={newX}
                                                onChange={(event) => setNewX(event.target.value)}
                                                aria-label="X новой точки"
                                            />
                                        </label>
                                        <label className="geometry-field">
                                            Новая Y
                                            <input
                                                type="text"
                                                inputMode="decimal"
                                                value={newY}
                                                onChange={(event) => setNewY(event.target.value)}
                                                aria-label="Y новой точки"
                                            />
                                        </label>
                                        <button
                                            type="button"
                                            className="geometry-secondary"
                                            onClick={handleManualPoint}
                                        >
                                            Добавить по координатам
                                        </button>
                                    </div>
                                </>
                            )}
                            <div className="geometry-grid-settings">
                                <label>
                                    <input
                                        type="checkbox"
                                        checked={showGrid}
                                        onChange={(event) => setShowGrid(event.target.checked)}
                                    />{' '}
                                    Сетка
                                </label>
                                <label className="geometry-field">
                                    Шаг сетки
                                    <input
                                        type="number"
                                        min="0.01"
                                        step="any"
                                        value={gridStepText}
                                        onChange={(event) => {
                                            setGridStepText(event.target.value)
                                            const value = Number(event.target.value)
                                            if (
                                                event.target.value.trim() &&
                                                Number.isFinite(value) &&
                                                value >= 0.01
                                            )
                                                setGridStep(value)
                                        }}
                                    />
                                </label>
                            </div>
                        </fieldset>
                    }
                    status={
                        <>
                            <span>
                                {activeRing < 0
                                    ? 'Внешний контур'
                                    : `Отверстие № ${activeRing + 1}`}{' '}
                                · вершин: {activePoints.length}
                            </span>
                            <span className="geometry-status-axis">Ось Y направлена вниз</span>
                            {error ? (
                                <span className="geometry-error" role="alert">
                                    {error}
                                </span>
                            ) : (
                                validationError && (
                                    <span className="geometry-validation">{validationError}</span>
                                )
                            )}
                        </>
                    }
                >
                    <GeometryCanvas
                        geometry={geometry}
                        visibleView={visibleView}
                        activeRing={activeRing}
                        selectedVertex={selectedVertex}
                        adding={!selectMode && (mode === 'points' || activeRing >= 0)}
                        panMode={panMode}
                        showGrid={showGrid}
                        gridStep={gridStep}
                        startDragging={startDragging}
                        canvasProps={{
                            ref: svgRef,
                            onPointerDown: handleCanvasPointerDown,
                            onPointerMove: handlePointerMove,
                            onPointerUp: stopDragging,
                            onPointerCancel: stopDragging,
                            onLostPointerCapture: (event) => finishDrag(event.pointerId, true),
                        }}
                    />
                </GeometryWorkspace>
            </form>
            {confirmRectangle && (
                <ConfirmDialog
                    title="Заменить внешний контур?"
                    confirmLabel="Заменить"
                    onClose={() => setConfirmRectangle(false)}
                    onConfirm={async () => applyRectangle()}
                >
                    <p>
                        Текущий внешний контур будет заменён прямоугольником с указанными размерами.
                    </p>
                </ConfirmDialog>
            )}
        </dialog>
    )
}
