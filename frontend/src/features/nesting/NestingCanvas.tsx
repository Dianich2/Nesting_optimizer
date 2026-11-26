import { draftGhostTransform } from './placementDraft'
import { useWheelZoom } from '../../shared/geometry/useWheelZoom'
import type { PlacementPose } from './placementHistory'
import { useEffect, useId, useMemo, useRef, useState } from 'react'
import type { PointerEvent, ReactNode } from 'react'
import type { PolygonGeometry } from '../project-resources/types'
import type { Placement } from './api'
import { polygonPath } from './geometry'
import { fitDrawingView } from '../library/geometryDraft'
import {
    angleAround,
    resizeView,
    normalizeAngle,
    screenToWorld,
    snapCoordinate,
    zoomView,
} from '../../shared/geometry/viewport'
import type { Viewport } from '../../shared/geometry/viewport'

interface Props {
    surface: PolygonGeometry
    placements: Placement[]
    selected: number | null
    busy: boolean
    adding: boolean
    draftPattern?: PolygonGeometry
    draftPatternId?: number
    draftRotation?: number
    draftControls?: ReactNode
    manualDraftId?: number | null
    manualDraftIssue?: string | null
    validatePose?: (id: number, pose: PlacementPose) => string | null
    canEdit?: (id: number) => boolean
    onSelect: (id: number | null) => void
    onAdd: (x: number, y: number, rotation: number) => void
    onMove: (id: number, x: number, y: number, rotation?: number) => void
}
export function NestingCanvas({
    surface,
    placements,
    selected,
    busy,
    adding,
    draftPattern,
    draftPatternId,
    draftRotation = 0,
    draftControls,
    manualDraftId = null,
    manualDraftIssue = null,
    validatePose,
    canEdit,
    onSelect,
    onAdd,
    onMove,
}: Props) {
    const svg = useRef<SVGSVGElement>(null)
    const gridId = useId()
    const [view, setView] = useState(() => fitDrawingView(surface))
    const [aspect, setAspect] = useState(1)
    const fitted = useRef(false)
    const [cursor, setCursor] = useState<{ x: number; y: number } | null>(null)
    const ghost = useMemo(() => {
        if (!draftPattern?.exterior.points.length) return null
        const xs = draftPattern.exterior.points.map((point) => point.x)
        const ys = draftPattern.exterior.points.map((point) => point.y)
        return {
            path: polygonPath(draftPattern),
            x: (Math.min(...xs) + Math.max(...xs)) / 2,
            y: (Math.min(...ys) + Math.max(...ys)) / 2,
        }
    }, [draftPattern])
    const [pan, setPan] = useState(false)
    const [grid, setGrid] = useState(true)
    const [snap, setSnap] = useState(false)
    const [step, setStep] = useState(10)
    const [stepText, setStepText] = useState('10')
    const [preview, setPreview] = useState<{
        id: number
        dx?: number
        dy?: number
        rotation?: number
        issue?: string | null
    } | null>(null)
    const drag = useRef<{
        pointer: number
        x: number
        y: number
        view: Viewport
        item?: Placement
        initialAngle?: number
        scale: number
    } | null>(null)
    const visible = resizeView(view, aspect)
    useWheelZoom(svg, visible, setView, () => !drag.current)
    const selection = placements.find((item) => item.id === selected)
    const fit = fitDrawingView(surface, aspect)
    const ghostPosition = cursor ?? { x: visible.x + visible.width / 2, y: visible.y + visible.height / 2 }
    function snapped(x: number, y: number) {
        return { x: snapCoordinate(x, step, snap), y: snapCoordinate(y, step, snap) }
    }
    useEffect(() => {
        const element = svg.current
        if (!element) return
        const observer = new ResizeObserver(([entry]) => {
            const { width, height } = entry.contentRect
            if (width <= 0 || height <= 0) return
            const next = width / height
            setAspect(next)
            if (!fitted.current) {
                fitted.current = true
                setView(fitDrawingView(surface, next))
            }
        })
        observer.observe(element)
        return () => observer.disconnect()
    }, [surface])
    function start(event: PointerEvent<SVGSVGElement>) {
        if (drag.current || !event.isPrimary || event.button > 1) return
        if (pan || event.button === 1) {
            event.preventDefault()
            drag.current = {
                pointer: event.pointerId,
                x: event.clientX,
                y: event.clientY,
                view: visible,
                scale: visible.width / event.currentTarget.getBoundingClientRect().width,
            }
            event.currentTarget.setPointerCapture(event.pointerId)
            return
        }
        if (busy) return
        if (adding) {
            const point = screenToWorld(event.currentTarget, event.clientX, event.clientY)
            if (!point) return
            const position = snapped(point.x, point.y)
            event.currentTarget.focus()
            onAdd(position.x, position.y, draftRotation)
        } else onSelect(null)
    }
    function rotationAt(event: PointerEvent<SVGSVGElement>, item: Placement, original: number) {
        const point = screenToWorld(event.currentTarget, event.clientX, event.clientY)
        if (!point) return item.rotation
        const angle = item.rotation + angleAround(item, point) - original
        return normalizeAngle(event.shiftKey ? Math.round(angle / 15) * 15 : angle)
    }
    function move(event: PointerEvent<SVGSVGElement>) {
        const current = drag.current
        if (!current) {
            if (adding && !pan && !busy) {
                const point = screenToWorld(event.currentTarget, event.clientX, event.clientY)
                if (point) setCursor(snapped(point.x, point.y))
            }
            return
        }
        if (current.pointer !== event.pointerId) return
        const dx = (event.clientX - current.x) * current.scale
        const dy = (event.clientY - current.y) * current.scale
        if (current.item && current.initialAngle !== undefined) {
            const rotation = rotationAt(event, current.item, current.initialAngle)
            setPreview({
                id: current.item.id,
                rotation,
                issue: validatePose?.(current.item.id, {
                    project_pattern_id: current.item.project_pattern_id,
                    x: current.item.x, y: current.item.y, rotation,
                }),
            })
        } else if (current.item) {
            const target = snapped(current.item.x + dx, current.item.y + dy)
            setPreview({
                id: current.item.id,
                dx: target.x - current.item.x,
                dy: target.y - current.item.y,
                issue: validatePose?.(current.item.id, {
                    project_pattern_id: current.item.project_pattern_id,
                    x: target.x, y: target.y, rotation: current.item.rotation,
                }),
            })
        } else setView({ ...current.view, x: current.view.x - dx, y: current.view.y - dy })
    }
    function stop(event: PointerEvent<SVGSVGElement>, cancelled = false) {
        const current = drag.current
        if (!current || current.pointer !== event.pointerId) return
        drag.current = null
        setPreview(null)
        if (current?.item && !cancelled) {
            const dx = (event.clientX - current.x) * current.scale
            const dy = (event.clientY - current.y) * current.scale
            if (current.initialAngle !== undefined) {
                const rotation = rotationAt(event, current.item, current.initialAngle)
                if (rotation !== current.item.rotation)
                    onMove(current.item.id, current.item.x, current.item.y, rotation)
            } else if (dx !== 0 || dy !== 0) {
                const target = snapped(current.item.x + dx, current.item.y + dy)
                onMove(current.item.id, target.x, target.y)
            }
        }
        if (event.currentTarget.hasPointerCapture(event.pointerId))
            event.currentTarget.releasePointerCapture(event.pointerId)
    }
    return (
        <div className="nesting-canvas">
            {draftControls}
            <div className="nesting-tools">
                <button
                    type="button"
                    aria-label="Приблизить"
                    title="Приблизить"
                    onClick={() => setView(zoomView(visible, 0.8))}
                >
                    +
                </button>
                <button
                    type="button"
                    aria-label="Отдалить"
                    title="Отдалить"
                    onClick={() => setView(zoomView(visible, 1.25))}
                >
                    −
                </button>
                <output aria-label="Масштаб полотна">
                    {Math.round((fit.width / visible.width) * 100)}%
                </output>
                <button type="button" onClick={() => setView(fitDrawingView(surface, aspect))}>
                    Уместить поверхность
                </button>
                <button type="button" aria-pressed={pan} onClick={() => setPan(!pan)}>
                    Перемещение поля
                </button>
                <details className="nesting-menu">
                    <summary>Настройки сетки</summary>
                    <div className="nesting-menu-content">
                        <label>
                            <input
                                type="checkbox"
                                checked={grid}
                                onChange={(event) => setGrid(event.target.checked)}
                            />
                            Сетка
                        </label>
                        <label>
                            <input
                                type="checkbox"
                                checked={snap}
                                onChange={(event) => setSnap(event.target.checked)}
                            />
                            Привязка к сетке
                        </label>
                        <label>
                            Шаг{' '}
                            <input
                                type="number"
                                value={stepText}
                                min="0.01"
                                step="any"
                                onChange={(event) => {
                                    setStepText(event.target.value)
                                    const value = Number(event.target.value)
                                    if (
                                        event.target.value.trim() &&
                                        Number.isFinite(value) &&
                                        value >= 0.01
                                    )
                                        setStep(value)
                                }}
                            />
                        </label>
                    </div>
                </details>
                {(!stepText.trim() ||
                    !Number.isFinite(Number(stepText)) ||
                    Number(stepText) < 0.01) && (
                    <span role="alert">
                        Шаг сетки должен быть не меньше 0,01. Пока используется {step}.
                    </span>
                )}
                {preview?.rotation !== undefined && <output>{preview.rotation.toFixed(1)}°</output>}
            </div>
            <svg
                ref={svg}
                viewBox={`${visible.x} ${visible.y} ${visible.width} ${visible.height}`}
                aria-label="Рабочая поверхность раскроя"
                tabIndex={0}
                onPointerDown={start}
                onPointerMove={move}
                onPointerLeave={() => {
                    // Preserve ghost position while the user edits the angle field.
                    if (!adding) setCursor(null)
                }}
                onPointerUp={(event) => stop(event)}
                onPointerCancel={(event) => stop(event, true)}
                onLostPointerCapture={(event) => {
                    if (drag.current?.pointer !== event.pointerId) return
                    drag.current = null
                    setPreview(null)
                }}
            >
                <defs>
                    <pattern id={gridId} width={step} height={step} patternUnits="userSpaceOnUse">
                        <path
                            d={`M ${step} 0 H 0 V ${step}`}
                            fill="none"
                            stroke="currentColor"
                            strokeOpacity="0.12"
                            strokeWidth={visible.width / 1200}
                        />
                    </pattern>
                </defs>
                <rect
                    x={visible.x}
                    y={visible.y}
                    width={visible.width}
                    height={visible.height}
                    fill={grid ? `url(#${gridId})` : 'transparent'}
                />
                <path
                    className="nesting-surface-outline"
                    d={polygonPath(surface)}
                    fill="var(--color-bg-secondary, #64748b)"
                    fillOpacity="0.3"
                    stroke="currentColor"
                    fillRule="evenodd"
                    strokeWidth={visible.width / 800}
                />
                {placements.map((item) => (
                        <path
                            key={item.id}
                            d={polygonPath(item.geometry)}
                            transform={
                                preview?.id === item.id
                                    ? preview.rotation !== undefined
                                        ? `rotate(${preview.rotation - item.rotation} ${item.x} ${item.y})`
                                        : `translate(${preview.dx} ${preview.dy})`
                                    : undefined
                            }
                            fill={preview?.id === item.id && preview.issue
                                ? '#dc2626' : manualDraftId === item.id
                                  ? manualDraftIssue ? '#dc2626' : '#f59e0b'
                                  : selected === item.id ? '#e9434b' : '#60a5fa'}
                            fillOpacity="0.65"
                            fillRule="evenodd"
                            stroke={preview?.id === item.id && preview.issue || manualDraftId === item.id && manualDraftIssue
                                ? '#dc2626' : selected === item.id ? '#e9434b' : 'currentColor'}
                            strokeWidth={visible.width / 600}
                            tabIndex={busy ? -1 : 0}
                            role="button"
                            aria-label={`Размещение №${item.id}`}
                            onKeyDown={(event) => {
                                if (!busy && (event.key === 'Enter' || event.key === ' ')) {
                                    event.preventDefault()
                                    onSelect(item.id)
                                }
                            }}
                            onPointerDown={(event) => {
                                if (pan || event.button === 1 || adding) return
                                event.stopPropagation()
                                if (
                                    busy ||
                                    drag.current ||
                                    !event.isPrimary ||
                                    event.button !== 0 ||
                                    !svg.current
                                )
                                    return
                                if (canEdit && !canEdit(item.id)) {
                                    onSelect(item.id)
                                    return
                                }
                                onSelect(item.id)
                                drag.current = {
                                    pointer: event.pointerId,
                                    x: event.clientX,
                                    y: event.clientY,
                                    view: visible,
                                    item,
                                    scale:
                                        visible.width / svg.current.getBoundingClientRect().width,
                                }
                                svg.current.setPointerCapture(event.pointerId)
                            }}
                        />
                    ))}
                {adding && !pan && !busy && ghost && (
                    <path
                        d={ghost.path}
                        transform={draftGhostTransform(ghost, ghostPosition, draftRotation)}
                        data-placement-ghost="true"
                        fill={draftPatternId && validatePose?.(-1, {
                            project_pattern_id: draftPatternId, x: ghostPosition.x, y: ghostPosition.y, rotation: draftRotation,
                        }) ? '#dc2626' : '#6da5ee'}
                        fillOpacity="0.35"
                        stroke={draftPatternId && validatePose?.(-1, {
                            project_pattern_id: draftPatternId, x: ghostPosition.x, y: ghostPosition.y, rotation: draftRotation,
                        }) ? '#dc2626' : '#4089d6'}
                        strokeDasharray="4 3"
                        strokeWidth="2"
                        vectorEffect="non-scaling-stroke"
                        pointerEvents="none"
                    />
                )}
                {!busy && selection && !pan && !adding && (
                    <circle
                        cx={selection.x}
                        cy={selection.y - visible.width / 20}
                        r={visible.width / 95}
                        fill="#e9434b"
                        stroke="white"
                        strokeWidth={visible.width / 600}
                        aria-label="Ручка поворота"
                        onPointerDown={(event) => {
                            event.stopPropagation()
                            if (
                                busy ||
                                (canEdit && !canEdit(selection.id)) ||
                                drag.current ||
                                !event.isPrimary ||
                                event.button !== 0 ||
                                !svg.current
                            )
                                return
                            const point = screenToWorld(svg.current, event.clientX, event.clientY)
                            if (!point) return
                            drag.current = {
                                pointer: event.pointerId,
                                x: event.clientX,
                                y: event.clientY,
                                view: visible,
                                item: selection,
                                scale: visible.width / svg.current.getBoundingClientRect().width,
                                initialAngle: angleAround(selection, point),
                            }
                            svg.current.setPointerCapture(event.pointerId)
                        }}
                    >
                        <title>Перетащите для поворота. Shift привязывает к шагу 15°.</title>
                    </circle>
                )}
            </svg>
            {(manualDraftId !== null || preview?.issue) && (
                <div className={preview?.issue || manualDraftIssue
                    ? 'nesting-manual-feedback nesting-manual-feedback-invalid'
                    : 'nesting-manual-feedback'}
                    role={preview?.issue || manualDraftIssue ? 'alert' : 'status'}>
                    {preview?.issue || manualDraftIssue
                        ? `Некорректное размещение — не сохранено. ${preview?.issue || manualDraftIssue} Переместите или поверните лекало.`
                        : 'Положение пока не сохранено. После завершения редактирования корректная позиция сохраняется автоматически.'}
                </div>
            )}
        </div>
    )
}
