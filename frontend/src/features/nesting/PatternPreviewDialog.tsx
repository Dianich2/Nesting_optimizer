import { useEffect, useId, useRef } from 'react'
import type { PolygonGeometry, ProjectPattern } from '../project-resources/types'
import { fitDrawingView } from '../library/geometryDraft'
import { polygonArea, polygonDimensions, polygonPath } from './geometry'
import '../projects/ProjectDialog.css'

const format = (value: number) => value.toLocaleString('ru-RU', { maximumFractionDigits: 2 })

function PatternShape({ geometry, large = false, label }: {
    geometry: PolygonGeometry
    large?: boolean
    label?: string
}) {
    const aspect = large ? 16 / 9 : 1
    const view = fitDrawingView(geometry, aspect)
    const viewBox = `${view.x} ${view.y} ${view.width} ${view.height}`
    return (
        <svg
            className={large ? 'nesting-pattern-large-svg' : 'nesting-pattern-thumbnail'}
            viewBox={viewBox}
            role={large ? 'img' : undefined}
            aria-label={large ? label : undefined}
            aria-hidden={large ? undefined : true}
            focusable="false"
        >
            <path
                d={polygonPath(geometry)}
                fill="#60a5fa"
                fillOpacity="0.35"
                stroke="#3b82f6"
                strokeWidth={large ? 2.5 : 1.6}
                vectorEffect="non-scaling-stroke"
                fillRule="evenodd"
            />
        </svg>
    )
}

export function PatternThumbnail({ geometry }: { geometry: PolygonGeometry }) {
    return <PatternShape geometry={geometry} />
}

export function PatternPreviewDialog({ pattern, onClose }: {
    pattern: ProjectPattern
    onClose: () => void
}) {
    const dialog = useRef<HTMLDialogElement>(null)
    const titleId = useId()
    const { width, height } = polygonDimensions(pattern.geometry)
    const area = polygonArea(pattern.geometry)
    const vertices = pattern.geometry.exterior.points.length
    const holeVertices = pattern.geometry.holes.reduce((sum, ring) => sum + ring.points.length, 0)

    useEffect(() => {
        const node = dialog.current
        // StrictMode replays effects in development. Closing here would emit
        // a close event and discard the preview on the second setup.
        // React removes the dialog on unmount; explicit close and Esc remain native.
        if (node && !node.open) node.showModal()
    }, [])

    return (
        <dialog
            ref={dialog}
            className="project-dialog nesting-pattern-dialog"
            aria-labelledby={titleId}
            onClose={(event) => { event.stopPropagation(); onClose() }}
        >
            <div className="project-delete-content">
                <header className="project-dialog-header">
                    <h2 id={titleId}>{pattern.name}</h2>
                    <p>Лекало проекта №{pattern.id}. Контур до размещения и поворота.</p>
                </header>
                <PatternShape
                    geometry={pattern.geometry}
                    large
                    label={`Контур лекала «${pattern.name}»`}
                />
                <dl className="nesting-pattern-preview-stats">
                    <div><dt>Ширина</dt><dd>{format(width)} ед.</dd></div>
                    <div><dt>Высота</dt><dd>{format(height)} ед.</dd></div>
                    <div><dt>Площадь</dt><dd>{format(area)} кв. ед.</dd></div>
                    <div><dt>Вершин внешнего контура</dt><dd>{vertices}</dd></div>
                    <div><dt>Отверстий</dt><dd>{pattern.geometry.holes.length}</dd></div>
                    <div><dt>Вершин отверстий</dt><dd>{holeVertices}</dd></div>
                </dl>
                <footer className="project-dialog-actions">
                    <button
                        type="button"
                        className="projects-button"
                        onClick={() => dialog.current?.close()}
                    >
                        Закрыть
                    </button>
                </footer>
            </div>
        </dialog>
    )
}
