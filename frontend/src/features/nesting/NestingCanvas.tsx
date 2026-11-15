import { useEffect, useId, useRef, useState } from 'react'
import type { PointerEvent } from 'react'
import type { PolygonGeometry } from '../project-resources/types'
import type { Placement } from './api'
import { polygonPath } from './geometry'
import { fitDrawingView } from '../library/geometryDraft'
import { fitAspect, zoomView } from '../../shared/geometry/viewport'
import type { Viewport } from '../../shared/geometry/viewport'

interface Props {
    surface: PolygonGeometry
    placements: Placement[]
    selected: number | null
    busy: boolean
    adding: boolean
    onSelect: (id: number | null) => void
    onAdd: (x: number, y: number) => void
    onMove: (id: number, x: number, y: number) => void
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
    const [step, setStep] = useState(10)
    const [preview, setPreview] = useState<{ id: number; dx: number; dy: number } | null>(null)
    const drag = useRef<{
        pointer: number
        x: number
        y: number
        view: Viewport
        item?: Placement
        scale: number
    } | null>(null)
    const visible = fitAspect(view, aspect)
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
        if (busy || event.button > 1) return
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
            const matrix = event.currentTarget.getScreenCTM()
            if (!matrix) return
            const point = new DOMPoint(event.clientX, event.clientY).matrixTransform(
                matrix.inverse(),
            )
            onAdd(point.x, point.y)
        } else onSelect(null)
    }
    function move(event: PointerEvent<SVGSVGElement>) {
        const current = drag.current
        if (!current || current.pointer !== event.pointerId) return
        const dx = (event.clientX - current.x) * current.scale
        const dy = (event.clientY - current.y) * current.scale
        if (current.item) setPreview({ id: current.item.id, dx, dy })
        else setView({ ...current.view, x: current.view.x - dx, y: current.view.y - dy })
    }
    function stop(event: PointerEvent<SVGSVGElement>, cancelled = false) {
        const current = drag.current
        drag.current = null
        setPreview(null)
        if (current?.item && !cancelled) {
            const dx = (event.clientX - current.x) * current.scale
            const dy = (event.clientY - current.y) * current.scale
            if (dx !== 0 || dy !== 0)
                onMove(current.item.id, current.item.x + dx, current.item.y + dy)
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
                onLostPointerCapture={() => {
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
                                ? `translate(${preview.dx} ${preview.dy})`
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
                            if (busy || event.button !== 0 || !svg.current) return
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
            </svg>
        </div>
    )
}
