import { useId } from 'react'
import type { ComponentPropsWithRef, PointerEvent } from 'react'
import type { DrawingViewBox } from './geometryDraft'
import type { LibraryGeometry, LibraryPoint } from './types'

function contourPath(points: LibraryPoint[], closed: boolean): string {
    if (points.length === 0) return ''
    return `M ${points.map((point) => `${point.x} ${point.y}`).join(' L ')}${closed ? ' Z' : ''}`
}

interface Props {
    geometry: LibraryGeometry
    visibleView: DrawingViewBox
    activeRing: number
    selectedVertex: number | null
    adding: boolean
    panMode: boolean
    showGrid: boolean
    gridStep: number
    canvasProps: ComponentPropsWithRef<'svg'>
    startDragging: (event: PointerEvent<SVGCircleElement>, ring: number, index: number) => void
}

export function GeometryCanvas({
    geometry,
    visibleView,
    activeRing,
    selectedVertex,
    adding,
    panMode,
    showGrid,
    gridStep,
    canvasProps,
    startDragging,
}: Props) {
    const gridId = useId()
    const rings = [geometry.exterior, ...geometry.holes]
    const areaPath = rings
        .filter((ring) => ring.points.length >= 3)
        .map((ring) => contourPath(ring.points, true))
        .join(' ')
    return (
        <svg
            {...canvasProps}
            className={`geometry-preview ${panMode ? 'geometry-preview-pan' : adding ? 'geometry-preview-clickable' : ''}`}
            viewBox={`${visibleView.x} ${visibleView.y} ${visibleView.width} ${visibleView.height}`}
            preserveAspectRatio="xMidYMid meet"
            role="img"
            aria-label="Предпросмотр многоугольника. В режиме точек нажмите для добавления вершины."
        >
            <defs>
                <pattern
                    id={gridId}
                    width={gridStep}
                    height={gridStep}
                    patternUnits="userSpaceOnUse"
                >
                    <path
                        d={`M ${gridStep} 0 H 0 V ${gridStep}`}
                        fill="none"
                        stroke="currentColor"
                        strokeOpacity="0.15"
                        strokeWidth={0.8}
                        vectorEffect="non-scaling-stroke"
                    />
                </pattern>
            </defs>
            <rect
                x={visibleView.x}
                y={visibleView.y}
                width={visibleView.width}
                height={visibleView.height}
                fill={showGrid ? `url(#${gridId})` : 'transparent'}
            />
            {areaPath && (
                <path
                    className="geometry-outline"
                    d={areaPath}
                    fill="#e9434b"
                    fillOpacity="0.18"
                    fillRule="evenodd"
                />
            )}
            {rings.map((ring, index) => (
                <g key={index}>
                    <path
                        d={contourPath(ring.points, ring.points.length >= 3)}
                        fill="none"
                        stroke={index === 0 ? '#e9434b' : '#60a5fa'}
                        strokeWidth={index - 1 === activeRing ? 2 : 1.25}
                        vectorEffect="non-scaling-stroke"
                        strokeLinejoin="round"
                    />
                    {ring.points.map((p, i) => (
                        <g key={i}>
                            <circle
                                cx={p.x}
                                cy={p.y}
                                r={visibleView.height / 52}
                                className="geometry-vertex"
                                fill="transparent"
                                onPointerDown={(event) => startDragging(event, index - 1, i)}
                            >
                                <title>
                                    Вершина {i + 1}: ({p.x}, {p.y}). Перетащите для изменения.
                                </title>
                            </circle>
                            <circle
                                cx={p.x}
                                cy={p.y}
                                r={
                                    index - 1 === activeRing && selectedVertex === i
                                        ? visibleView.height / 115
                                        : visibleView.height / 160
                                }
                                pointerEvents="none"
                                fill={index === 0 ? '#e9434b' : '#60a5fa'}
                                stroke={
                                    index - 1 === activeRing && selectedVertex === i
                                        ? 'var(--color-text)'
                                        : 'var(--color-surface)'
                                }
                                strokeWidth={
                                    index - 1 === activeRing && selectedVertex === i ? 2 : 1
                                }
                                vectorEffect="non-scaling-stroke"
                            />
                            {index - 1 === activeRing && (
                                <text
                                    x={p.x + visibleView.height / 60}
                                    y={p.y - visibleView.height / 85}
                                    fontSize={visibleView.height / 62}
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
    )
}
