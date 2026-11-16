import type { PolygonGeometry } from '../project-resources/types'
import { polygonPath } from './geometry.ts'

export function exportLayoutSVG(
    surface: PolygonGeometry,
    placements: { geometry: PolygonGeometry }[],
): string {
    const points = surface.exterior.points
    if (!points.length) throw new Error('Поверхность не содержит контура.')
    let left = points[0].x
    let right = left
    let top = points[0].y
    let bottom = top
    for (const { x, y } of points) {
        left = Math.min(left, x)
        right = Math.max(right, x)
        top = Math.min(top, y)
        bottom = Math.max(bottom, y)
    }
    const shapes = placements.map(
        ({ geometry }) =>
            `<path d="${polygonPath(geometry)}" fill="#8ab4f8" stroke="#2456a6" vector-effect="non-scaling-stroke"/>`,
    )
    return [
        `<svg xmlns="http://www.w3.org/2000/svg" viewBox="${left} ${top} ${right - left} ${bottom - top}" fill-rule="evenodd">`,
        `<path d="${polygonPath(surface)}" fill="#eef1f5"/>`,
        ...shapes,
        '</svg>',
    ].join('\n')
}

export function downloadLayout(name: string, data: string, mime: string) {
    const url = URL.createObjectURL(new Blob([data], { type: mime }))
    const link = document.createElement('a')
    link.href = url
    link.download = name
    link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
}
