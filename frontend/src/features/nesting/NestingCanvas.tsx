import { useWheelZoom } from '../../shared/geometry/useWheelZoom'
import { useEffect, useId, useMemo, useRef, useState } from 'react'
import type { PointerEvent } from 'react'
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
    proposedPlacements?: Placement[]
    showSaved: boolean
    onSelect: (id: number | null) => void
    onAdd: (x: number, y: number) => void
    onMove: (id: number, x: number, y: number, rotation?: number) => void
}
export function NestingCanvas({
    surface,
    placements,
    selected,
    busy,
    adding,
    draftPattern,
    proposedPlacements,
    showSaved,
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
    const [preview, setPreview] = useState<{
        id: number
        dx?: number
        dy?: number
        rotation?: number
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
            onAdd(position.x, position.y)
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
            setPreview({
                id: current.item.id,
                rotation: rotationAt(event, current.item, current.initialAngle),
            })
        } else if (current.item) {
            const target = snapped(current.item.x + dx, current.item.y + dy)
            setPreview({
                id: current.item.id,
                dx: target.x - current.item.x,
                dy: target.y - current.item.y,
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
                                value={step}
                                min="0.01"
                                step="any"
                                onChange={(event) => {
                                    const value = Number(event.target.value)
                                    if (Number.isFinite(value) && value > 0) setStep(value)
                                }}
                            />
                        </label>
                    </div>
                </details>
                {preview?.rotation !== undefined && <output>{preview.rotation.toFixed(1)}°</output>}
            </div>
            <svg
                ref={svg}
                viewBox={`${visible.x} ${visible.y} ${visible.width} ${visible.height}`}
                aria-label="Рабочая поверхность раскроя"
                onPointerDown={start}
                onPointerMove={move}
                onPointerLeave={() => setCursor(null)}
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
                {showSaved &&
                    placements.map((item) => (
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
                            fill={selected === item.id ? '#e9434b' : '#60a5fa'}
                            fillOpacity="0.65"
                            fillRule="evenodd"
                            stroke={selected === item.id ? '#e9434b' : 'currentColor'}
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
                {proposedPlacements?.map((item, index) => (
                    <path
                        key={index}
                        d={polygonPath(item.geometry)}
                        fill="#9e77ed"
                        fillOpacity="0.42"
                        fillRule="evenodd"
                        stroke="#7b51bd"
                        strokeWidth="2"
                        vectorEffect="non-scaling-stroke"
                        pointerEvents="none"
                    />
                ))}
                {adding && !pan && !busy && ghost && cursor && (
                    <path
                        d={ghost.path}
                        transform={`translate(${cursor.x - ghost.x} ${cursor.y - ghost.y})`}
                        fill="#6da5ee"
                        fillOpacity="0.35"
                        stroke="#4089d6"
                        strokeDasharray="4 3"
                        strokeWidth="2"
                        vectorEffect="non-scaling-stroke"
                        pointerEvents="none"
                    />
                )}
                {showSaved && !busy && selection && !pan && !adding && (
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
        </div>
    )
}
