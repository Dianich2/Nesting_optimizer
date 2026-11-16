import { useWheelZoom } from '../../shared/geometry/useWheelZoom'
import { useEffect, useId, useRef, useState } from 'react'
import type { PointerEvent } from 'react'
import type { PolygonGeometry } from '../project-resources/types'
import type { Placement } from './api'
import { polygonPath } from './geometry'
import { fitDrawingView } from '../library/geometryDraft'
import {
    angleAround,
    fitAspect,
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
    onSelect,
    onAdd,
    onMove,
}: Props) {
    const svg = useRef<SVGSVGElement>(null)
    const gridId = useId()
    const [view, setView] = useState(() => fitDrawingView(surface))
    const [aspect, setAspect] = useState(1)
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
    const visible = fitAspect(view, aspect)
    useWheelZoom(svg, visible, setView, () => !drag.current && !busy)
    const selection = placements.find((item) => item.id === selected)
    const fit = fitAspect(fitDrawingView(surface), aspect)
    function snapped(x: number, y: number) {
        return { x: snapCoordinate(x, step, snap), y: snapCoordinate(y, step, snap) }
    }
    useEffect(() => {
        const element = svg.current
        if (!element) return
        const observer = new ResizeObserver(([entry]) => {
            if (entry.contentRect.height > 0)
                setAspect(entry.contentRect.width / entry.contentRect.height)
        })
        observer.observe(element)
        return () => observer.disconnect()
    }, [])
    function start(event: PointerEvent<SVGSVGElement>) {
        if (busy || drag.current || !event.isPrimary || event.button > 1) return
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
        if (!current || current.pointer !== event.pointerId) return
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
                <button type="button" onClick={() => setView(zoomView(visible, 0.8))}>
                    +
                </button>
                <button type="button" onClick={() => setView(zoomView(visible, 1.25))}>
                    −
                </button>
                <output aria-label="Масштаб полотна">
                    {Math.round((fit.width / visible.width) * 100)}%
                </output>
                <button type="button" onClick={() => setView(fitDrawingView(surface))}>
                    Уместить поверхность
                </button>
                <button type="button" aria-pressed={pan} onClick={() => setPan(!pan)}>
                    Перемещение поля
                </button>
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
                {preview?.rotation !== undefined && <output>{preview.rotation.toFixed(1)}°</output>}
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
            <svg
                ref={svg}
                viewBox={`${visible.x} ${visible.y} ${visible.width} ${visible.height}`}
                aria-label="Рабочая поверхность раскроя"
                onPointerDown={start}
                onPointerMove={move}
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
                                scale: visible.width / svg.current.getBoundingClientRect().width,
                            }
                            svg.current.setPointerCapture(event.pointerId)
                        }}
                    />
                ))}
                {selection && !pan && !adding && (
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
