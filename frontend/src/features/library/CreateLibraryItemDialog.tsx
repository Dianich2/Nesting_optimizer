import { useWheelZoom } from '../../shared/geometry/useWheelZoom'
import { ConfirmDialog } from '../../components/dialogs/ConfirmDialog'
import { fitAspect, zoomView } from '../../shared/geometry/viewport'
import { useEffect, useId, useRef, useState } from 'react'
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

function contourPath(points: LibraryPoint[], closed: boolean): string {
    if (points.length === 0) return ''
    return `M ${points.map((point) => `${point.x} ${point.y}`).join(' L ')}${closed ? ' Z' : ''}`
}

export function CreateLibraryItemDialog({ kind, onClose, onSave, initialItem }: Props) {
    const dialogRef = useRef<HTMLDialogElement>(null)
    const fileRef = useRef<HTMLInputElement>(null)
    const svgRef = useRef<SVGSVGElement>(null)
    const [initialDraft] = useState<LibraryGeometry>(() =>
        cloneGeometry(initialItem?.geometry ?? EMPTY),
    )
    const draftRef = useRef<LibraryGeometry>(initialDraft)
    const dragRef = useRef<{ ring: number; index: number; pointerId: number } | null>(null)
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
    const [newX, setNewX] = useState('0')
    const [newY, setNewY] = useState('0')
    const [view, setView] = useState<DrawingViewBox>(() =>
        fitDrawingView(
            initialItem
                ? initialDraft
                : rectangleGeometry(kind === 'surfaces' ? 1000 : 100, kind === 'surfaces' ? 600 : 60),
        ),
    )
    const gridId = useId()
    const [gridStep, setGridStep] = useState(10)
    const [showGrid, setShowGrid] = useState(true)
    const [panMode, setPanMode] = useState(false)
    const [confirmRectangle, setConfirmRectangle] = useState(false)
    const [aspect, setAspect] = useState(640 / 420)
    const panRef = useRef<{ id: number; x: number; y: number; view: DrawingViewBox; scale: number } | null>(null)
    const visibleView = fitAspect(view, aspect)
    const initialView = fitAspect(fitDrawingView(initialItem?.geometry ?? rectangleGeometry(kind === 'surfaces' ? 1000 : 100, kind === 'surfaces' ? 600 : 60)), aspect)
    useWheelZoom(svgRef, visibleView, setView, () => !dragRef.current && !panRef.current)
    useEffect(() => {
        const svg = svgRef.current
        if (!svg) return
        const observer = new ResizeObserver(([entry]) => {
            if (entry.contentRect.height > 0) setAspect(entry.contentRect.width / entry.contentRect.height)
        })
        observer.observe(svg)
        return () => observer.disconnect()
    }, [])
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
    const visibleRings = rings.filter((ring) => ring.points.length >= 3)
    const areaPath = visibleRings.map((ring) => contourPath(ring.points, true)).join(' ')

    /** Guard ALL input methods: clicks, typed coordinates, vertex dragging and deletion. */
    function tryApplyDraft(candidate: LibraryGeometry): boolean {
        const next =
            mode === 'rectangle' ? { ...candidate, exterior: geometry.exterior } : candidate
        const problem = validatePolygonConstruction(next, allowHoles)
        if (problem) {
            setError(problem)
            return false
        }
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
        updateActive(() => nextPoints)
    }

    function pointFromPointer(event: PointerEvent<SVGSVGElement>): LibraryPoint | null {
        const matrix = event.currentTarget.getScreenCTM()
        if (!matrix) return null
        const point = new DOMPoint(event.clientX, event.clientY).matrixTransform(matrix.inverse())
        return { x: Math.round(point.x * 100) / 100, y: Math.round(point.y * 100) / 100 }
    }

    function handleCanvasPointerDown(event: PointerEvent<SVGSVGElement>) {
        if (saving || !event.isPrimary || dragRef.current || panRef.current) return
        if (panMode || event.button === 1) {
            event.preventDefault()
            const bounds = event.currentTarget.getBoundingClientRect()
            panRef.current = { id: event.pointerId, x: event.clientX, y: event.clientY, view: visibleView, scale: visibleView.width / bounds.width }
            event.currentTarget.setPointerCapture(event.pointerId)
            return
        }
        if (event.button !== 0) return
        if (mode === 'rectangle' && activeRing === -1) return
        const point = pointFromPointer(event)
        if (point) addPoint(point)
    }

    function startDragging(event: PointerEvent<SVGCircleElement>, ring: number, index: number) {
        if (panMode || saving || !event.isPrimary || dragRef.current || panRef.current || event.button !== 0) return
        event.preventDefault()
        event.stopPropagation() // A vertex is not a new point.

        if (mode === 'rectangle' && ring < 0) {
            if (!rectangleValid) return
            // Dragging a rectangle corner switches seamlessly to free-point editing.
            const converted: LibraryGeometry = {
                exterior: rectangleGeometry(w, h).exterior,
                holes: draftRef.current.holes,
            }
            draftRef.current = converted
            setDraft(converted)
            setMode('points')
        }
        setActiveRing(ring)
        dragRef.current = { ring, index, pointerId: event.pointerId }
        svgRef.current?.setPointerCapture(event.pointerId)
    }

    function handlePointerMove(event: PointerEvent<SVGSVGElement>) {
        const pan = panRef.current
        if (pan && pan.id === event.pointerId) {
            setView({ ...pan.view, x: pan.view.x - (event.clientX - pan.x) * pan.scale, y: pan.view.y - (event.clientY - pan.y) * pan.scale })
            return
        }
        const drag = dragRef.current
        if (!drag || drag.pointerId !== event.pointerId) return
        const point = pointFromPointer(event)
        if (!point) return
        updateRing(drag.ring, (points) => points.map((p, i) => (i === drag.index ? point : p)))
    }

    function stopDragging(event: PointerEvent<SVGSVGElement>) {
        if (panRef.current?.id === event.pointerId) panRef.current = null
        if (dragRef.current?.pointerId !== event.pointerId) {
            if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
            return
        }
        dragRef.current = null
        if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId)
        }
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

    function UseRectangleAsPoints() {
        if (!rectangleValid) {
            setError('Сначала укажите положительные размеры прямоугольника.')
            return
        }
        const next: LibraryGeometry = {
            exterior: rectangleGeometry(w, h).exterior,
            holes: draftRef.current.holes,
        }
        draftRef.current = next
        setDraft(next)
        setMode('points')
        setActiveRing(-1)
        setView(fitDrawingView(next))
        setError(null)
    }

    function changeMode(next: EditorMode) {
        if (next === mode) return
        if (next === 'points') {
            UseRectangleAsPoints()
            return
        }
        setConfirmRectangle(true)
    }

    function applyRectangle() {
        setMode('rectangle')
        setActiveRing(-1)
        setView(
            fitDrawingView({
                exterior: rectangleGeometry(w > 0 ? w : 1, h > 0 ? h : 1).exterior,
                holes: draftRef.current.holes,
            }),
        )
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
        const index = draftRef.current.holes.length
        const next: LibraryGeometry = {
            ...draftRef.current,
            holes: [...draftRef.current.holes, { points: [] }],
        }
        draftRef.current = next
        setDraft(next)
        setActiveRing(index)
        setError(null)
    }

    function removeHole(index: number) {
        const next: LibraryGeometry = {
            ...draftRef.current,
            holes: draftRef.current.holes.filter((_, i) => i !== index),
        }
        draftRef.current = next
        setDraft(next)
        setActiveRing(-1)
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
            draftRef.current = next
            setDraft(next)
            setMode('points')
            setActiveRing(-1)
            setView(fitDrawingView(next))
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
                <header className="geometry-dialog-heading">
                    <div>
                        <h2 id="geometry-dialog-title">
                            {initialItem
                                ? allowHoles ? 'Редактирование поверхности' : 'Редактирование лекала'
                                : allowHoles ? 'Новая поверхность' : 'Новое лекало'}
                        </h2>
                        <p>
                            Контур можно построить точками, задать размерами или импортировать из
                            SVG.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="geometry-secondary"
                        disabled={saving}
                        onClick={() => dialogRef.current?.close()}
                        aria-label="Закрыть"
                    >
                        ✕
                    </button>
                </header>

                <label className="geometry-field">
                    Название
                    <input
                        type="text"
                        value={name}
                        required
                        maxLength={150}
                        autoFocus
                        onChange={(event) => setName(event.target.value)}
                        disabled={saving}
                        placeholder={allowHoles ? 'Например, Лист материала' : 'Например, Деталь A'}
                    />
                </label>

                <div className="geometry-mode" role="group" aria-label="Способ создания">
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
                                onChange={(event) => setWidth(event.target.value)}
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
                                onChange={(event) => setHeight(event.target.value)}
                            />
                        </label>
                        <button
                            className="geometry-secondary"
                            type="button"
                            onClick={UseRectangleAsPoints}
                        >
                            Редактировать вершины
                        </button>
                    </div>
                )}

                {(mode === 'points' || allowHoles) && (
                    <div className="geometry-rings">
                        <div className="geometry-ring-tabs" role="group" aria-label="Выбор контура">
                            <button
                                type="button"
                                aria-pressed={activeRing === -1}
                                className={activeRing === -1 ? 'geometry-tab-active' : ''}
                                onClick={() => setActiveRing(-1)}
                            >
                                Внешний ({geometry.exterior.points.length})
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
                                        onClick={() => setActiveRing(index)}
                                    >
                                        Отверстие {index + 1} ({hole.points.length})
                                    </button>
                                ))}
                            {allowHoles && (
                                <button
                                    type="button"
                                    className="geometry-secondary"
                                    onClick={addHole}
                                >
                                    + Отверстие
                                </button>
                            )}
                        </div>
                        {mode === 'rectangle' && activeRing < 0 ? (
                            <p className="geometry-help">
                                Внешний контур задаётся шириной и высотой. Добавьте отверстие или
                                перетащите угол, чтобы перейти к редактированию точек.
                            </p>
                        ) : (
                            <p className="geometry-help">
                                Добавляйте точки кликами или по X/Y, перетаскивайте вершины.
                                После замыкания новая точка вставляется в ближайшее допустимое ребро.
                                Номера точек показывают порядок обхода контура.
                            </p>
                        )}
                        {(mode === 'points' || activeRing >= 0) && (
                            <>
                                <div className="geometry-point-add">
                                    <label>
                                        X{' '}
                                        <input
                                            type="text"
                                            inputMode="decimal"
                                            value={newX}
                                            onChange={(event) => setNewX(event.target.value)}
                                            aria-label="X новой точки"
                                        />
                                    </label>
                                    <label>
                                        Y{' '}
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
                                        + Точка
                                    </button>
                                    <button
                                        type="button"
                                        className="geometry-secondary"
                                        onClick={() => updateActive(() => [])}
                                    >
                                        Очистить контур
                                    </button>
                                    {activeRing >= 0 && allowHoles && (
                                        <button
                                            type="button"
                                            className="geometry-secondary"
                                            onClick={() => removeHole(activeRing)}
                                        >
                                            Удалить отверстие
                                        </button>
                                    )}
                                </div>
                                <div className="geometry-point-list" aria-label="Координаты вершин">
                                    {activePoints.length === 0 && (
                                        <p>В этом контуре ещё нет точек.</p>
                                    )}
                                    {activePoints.map((point, index) => (
                                        <div className="geometry-point-row" key={index}>
                                            <span>#{index + 1}</span>
                                            <CoordinateInput
                                                key={`x-${activeRing}-${index}-${point.x}-${point.y}`}
                                                value={point.x}
                                                label={`X точки ${index + 1}`}
                                                onCommit={(x) =>
                                                    updateActive((points) =>
                                                        points.map((p, i) =>
                                                            i === index ? { ...p, x } : p,
                                                        ),
                                                    )
                                                }
                                            />
                                            <CoordinateInput
                                                key={`y-${activeRing}-${index}-${point.x}-${point.y}`}
                                                value={point.y}
                                                label={`Y точки ${index + 1}`}
                                                onCommit={(y) =>
                                                    updateActive((points) =>
                                                        points.map((p, i) =>
                                                            i === index ? { ...p, y } : p,
                                                        ),
                                                    )
                                                }
                                            />
                                            <button
                                                type="button"
                                                className="geometry-secondary"
                                                aria-label={`Удалить точку ${index + 1}`}
                                                onClick={() =>
                                                    updateActive((points) =>
                                                        points.filter((_, i) => i !== index),
                                                    )
                                                }
                                            >
                                                ✕
                                            </button>
                                        </div>
                                    ))}
                                </div>
                            </>
                        )}
                    </div>
                )}

                <div className="geometry-preview-heading">
                    <strong>Предпросмотр геометрии</strong>
                    <button
                        type="button"
                        className="geometry-secondary"
                        onClick={() => setView(fitDrawingView(geometry))}
                    >
                        Уместить в поле
                    </button>
                </div>
                <div className="geometry-mode" role="group" aria-label="Вид и сетка">
                    <button type="button" onClick={() => setView(zoomView(visibleView, 0.8))}>+</button>
                    <span>Масштаб {Math.round(initialView.width / visibleView.width * 100)}%</span>
                    <button type="button" onClick={() => setView(zoomView(visibleView, 1.25))}>−</button>
                    <button type="button" onClick={() => setView(initialView)}>100%</button>
                    <button type="button" aria-pressed={panMode} onClick={() => setPanMode(!panMode)}>Перемещение поля</button>
                    <label><input type="checkbox" checked={showGrid} onChange={(event) => setShowGrid(event.target.checked)} /> Сетка</label>
                    <label>Шаг сетки <input type="number" min="0.01" step="any" value={gridStep}
                        onChange={(event) => { const value = Number(event.target.value); if (Number.isFinite(value) && value > 0) setGridStep(value) }} /></label>
                </div>
                <svg
                    ref={svgRef}
                    className={`geometry-preview ${mode === 'points' || activeRing >= 0 ? 'geometry-preview-clickable' : ''}`}
                    viewBox={`${visibleView.x} ${visibleView.y} ${visibleView.width} ${visibleView.height}`}
                    preserveAspectRatio="xMidYMid meet"
                    role="img"
                    aria-label="Предпросмотр многоугольника. В режиме точек нажмите для добавления вершины."
                    onPointerDown={handleCanvasPointerDown}
                    onPointerMove={handlePointerMove}
                    onPointerUp={stopDragging}
                    onPointerCancel={stopDragging}
                    onLostPointerCapture={(event) => {
                        if (dragRef.current?.pointerId === event.pointerId) dragRef.current = null
                        if (panRef.current?.id === event.pointerId) panRef.current = null
                    }}
                >
                    <defs>
                        <pattern id={gridId} width={gridStep} height={gridStep} patternUnits="userSpaceOnUse">
                            <path d={`M ${gridStep} 0 H 0 V ${gridStep}`} fill="none" stroke="currentColor" strokeOpacity="0.15" strokeWidth={visibleView.width / 1200} />
                        </pattern>
                    </defs>
                    <rect x={visibleView.x} y={visibleView.y} width={visibleView.width} height={visibleView.height}
                        fill={showGrid ? `url(#${gridId})` : 'transparent'} />
                    {areaPath && (
                        <path d={areaPath} fill="#e9434b" fillOpacity="0.18" fillRule="evenodd" />
                    )}
                    {rings.map((ring, index) => (
                        <g key={index}>
                            <path
                                d={contourPath(ring.points, ring.points.length >= 3)}
                                fill="none"
                                stroke={index === 0 ? '#e9434b' : '#60a5fa'}
                                strokeWidth={view.width / 300}
                                strokeLinejoin="round"
                            />
                            {ring.points.map((p, i) => (
                                <g key={i}>
                                    <circle
                                        cx={p.x}
                                        cy={p.y}
                                        r={view.width / 65}
                                        className="geometry-vertex"
                                        fill="transparent"
                                        onPointerDown={(event) =>
                                            startDragging(event, index - 1, i)
                                        }
                                    >
                                        <title>
                                            Вершина {i + 1}: ({p.x}, {p.y}). Перетащите для
                                            изменения.
                                        </title>
                                    </circle>
                                    <circle
                                        cx={p.x}
                                        cy={p.y}
                                        r={view.width / 125}
                                        pointerEvents="none"
                                        fill={index === 0 ? '#e9434b' : '#60a5fa'}
                                        stroke="white"
                                        strokeWidth={view.width / 650}
                                    />
                                    {index - 1 === activeRing && (
                                        <text
                                            x={p.x + view.width / 65}
                                            y={p.y - view.width / 80}
                                            fontSize={view.width / 75}
                                            fill="currentColor"
                                            pointerEvents="none"
                                        >
                                            {i + 1}
                                        </text>
                                    )}
                                </g>
                            ))}
                        </g>
                    ))}
                </svg>
                <p className="geometry-help">
                    В поле Y увеличивается вниз. Для SVG поддерживаются polygon, rect и
                    прямолинейные path (M/L/H/V/Z). Кривые и transform пока не импортируем.
                </p>
                {validationError && <p className="geometry-validation">{validationError}</p>}
                {error && (
                    <p className="geometry-error" role="alert">
                        {error}
                    </p>
                )}
                <footer className="geometry-dialog-actions">
                    <button
                        type="button"
                        className="geometry-secondary"
                        disabled={saving}
                        onClick={() => dialogRef.current?.close()}
                    >
                        Отмена
                    </button>
                    <button
                        type="submit"
                        className="geometry-primary"
                        disabled={saving || !name.trim() || !!validationError}
                    >
                        {saving ? 'Сохраняем…' : initialItem ? 'Сохранить изменения' : 'Создать'}
                    </button>
                </footer>
            </form>
            {confirmRectangle && <ConfirmDialog title="Заменить внешний контур?" confirmLabel="Заменить"
                onClose={() => setConfirmRectangle(false)} onConfirm={async () => applyRectangle()}>
                <p>Текущий внешний контур будет заменён прямоугольником с указанными размерами.</p>
            </ConfirmDialog>}
        </dialog>
    )
}
