import type { LibraryGeometry, LibraryPoint } from './types'

const EPS = 1e-9
const MAX_POINTS = 500
const MAX_HOLES = 20

export function rectangleGeometry(width: number, height: number): LibraryGeometry {
    return {
        exterior: {
            points: [
                { x: 0, y: 0 },
                { x: width, y: 0 },
                { x: width, y: height },
                { x: 0, y: height },
            ],
        },
        holes: [],
    }
}

export function signedArea(points: LibraryPoint[]): number {
    if (points.length < 3) return 0
    return points.reduce((area, point, index) => {
        const next = points[(index + 1) % points.length]
        return area + point.x * next.y - next.x * point.y
    }, 0) / 2
}

function orientation(a: LibraryPoint, b: LibraryPoint, c: LibraryPoint): number {
    return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x)
}

function onSegment(a: LibraryPoint, b: LibraryPoint, p: LibraryPoint): boolean {
    return Math.abs(orientation(a, b, p)) <= EPS &&
        p.x >= Math.min(a.x, b.x) - EPS &&
        p.x <= Math.max(a.x, b.x) + EPS &&
        p.y >= Math.min(a.y, b.y) - EPS &&
        p.y <= Math.max(a.y, b.y) + EPS
}

function segmentsIntersect(
    a: LibraryPoint, b: LibraryPoint, c: LibraryPoint, d: LibraryPoint,
): boolean {
    const o1 = orientation(a, b, c)
    const o2 = orientation(a, b, d)
    const o3 = orientation(c, d, a)
    const o4 = orientation(c, d, b)

    if (((o1 > EPS && o2 < -EPS) || (o1 < -EPS && o2 > EPS)) &&
        ((o3 > EPS && o4 < -EPS) || (o3 < -EPS && o4 > EPS))) {
        return true
    }
    return onSegment(a, b, c) || onSegment(a, b, d) ||
        onSegment(c, d, a) || onSegment(c, d, b)
}

function ringEdges(points: LibraryPoint[], closed: boolean): Array<[LibraryPoint, LibraryPoint]> {
    const result: Array<[LibraryPoint, LibraryPoint]> = []
    for (let i = 0; i + 1 < points.length; i++) result.push([points[i], points[i + 1]])
    if (closed && points.length >= 3) result.push([points[points.length - 1], points[0]])
    return result
}

function edgesIntersect(
    a: LibraryPoint[], aClosed: boolean, b: LibraryPoint[], bClosed: boolean,
): boolean {
    const aEdges = ringEdges(a, aClosed)
    const bEdges = ringEdges(b, bClosed)
    return aEdges.some(([a1, a2]) => bEdges.some(([b1, b2]) =>
        segmentsIntersect(a1, a2, b1, b2)))
}

/** Strict point-in-polygon: points on the boundary are NOT inside. */
export function insideRing(point: LibraryPoint, ring: LibraryPoint[]): boolean {
    if (ring.length < 3) return false
    let inside = false
    for (let i = 0, j = ring.length - 1; i < ring.length; j = i++) {
        const a = ring[i]
        const b = ring[j]
        if (onSegment(a, b, point)) return false
        const crosses = (a.y > point.y) !== (b.y > point.y) &&
            point.x < ((b.x - a.x) * (point.y - a.y)) / (b.y - a.y) + a.x
        if (crosses) inside = !inside
    }
    return inside
}

/**
 * Validate a ring during construction. Open paths with 1–2 points are permitted.
 * From the third vertex the closing edge must also be valid, so a click/drag
 * cannot introduce an invalid polygon that later surprises the user at save time.
 */
function ringProblem(points: LibraryPoint[], label: string, requireComplete: boolean): string | null {
    if (points.length > MAX_POINTS) return `${label}: максимум ${MAX_POINTS} точек.`
    if (points.some(point => !Number.isFinite(point.x) || !Number.isFinite(point.y))) {
        return `${label}: координаты должны быть конечными числами.`
    }
    if (requireComplete && points.length < 3) return `${label}: нужно минимум 3 точки.`

    const closed = points.length >= 3
    const edges = ringEdges(points, closed)
    for (const [a, b] of edges) {
        if (Math.hypot(a.x - b.x, a.y - b.y) <= EPS) {
            return `${label}: соседние точки не должны совпадать.`
        }
    }

    // Even adjacent edges are invalid if they retrace one another in reverse.
    if (points.length >= 3) {
        for (let i = 0; i < points.length; i++) {
            const a = points[(i - 1 + points.length) % points.length]
            const b = points[i]
            const c = points[(i + 1) % points.length]
            const dot = (a.x - b.x) * (c.x - b.x) + (a.y - b.y) * (c.y - b.y)
            if (Math.abs(orientation(a, b, c)) <= EPS && dot > EPS) {
                return `${label}: рёбра не должны накладываться друг на друга.`
            }
        }
    }

    for (let i = 0; i < edges.length; i++) {
        for (let j = i + 1; j < edges.length; j++) {
            // Shared endpoints of adjacent edges are expected.
            if (j === i + 1 || (closed && i === 0 && j === edges.length - 1)) continue
            if (segmentsIntersect(...edges[i], ...edges[j])) {
                return `${label}: рёбра пересекаются. Точка не добавлена или не перемещена.`
            }
        }
    }
    if (closed && Math.abs(signedArea(points)) <= EPS) {
        return `${label}: площадь контура должна быть больше нуля.`
    }
    return null
}

/** Validates both complete polygons and incomplete editor drafts. */
export function validatePolygonConstruction(
    geometry: LibraryGeometry,
    allowHoles: boolean,
): string | null {
    const exterior = geometry.exterior.points
    const exteriorProblem = ringProblem(exterior, 'Внешний контур', false)
    if (exteriorProblem) return exteriorProblem
    if (!allowHoles && geometry.holes.length > 0) {
        return 'У лекала в этой версии отверстия не предусмотрены.'
    }
    if (geometry.holes.length > MAX_HOLES) return `Максимум ${MAX_HOLES} отверстий.`
    if (geometry.holes.length > 0 && exterior.length < 3) {
        return 'Перед добавлением отверстия завершите внешний контур.'
    }

    for (const [index, hole] of geometry.holes.entries()) {
        const points = hole.points
        const label = `Отверстие ${index + 1}`
        const problem = ringProblem(points, label, false)
        if (problem) return problem

        // Check even the FIRST point and every OPEN segment of a partial hole.
        // This prevents drawing a hole outside, across a concavity, or on the edge.
        if (!points.every(point => insideRing(point, exterior)) ||
            edgesIntersect(points, points.length >= 3, exterior, true)) {
            return `${label}: все точки и рёбра должны быть строго внутри поверхности.`
        }

        for (let previous = 0; previous < index; previous++) {
            const other = geometry.holes[previous].points
            if (edgesIntersect(points, points.length >= 3, other, other.length >= 3) ||
                (other.length >= 3 && points.some(point => insideRing(point, other))) ||
                (points.length >= 3 && other.some(point => insideRing(point, points)))) {
                return `${label}: отверстия не должны пересекаться или содержать друг друга.`
            }
        }
    }
    return null
}

/** Final validation before sending geometry to the authoritative Go backend. */
export function validatePolygonDraft(geometry: LibraryGeometry, allowHoles: boolean): string | null {
    const constructionProblem = validatePolygonConstruction(geometry, allowHoles)
    if (constructionProblem) return constructionProblem
    const exteriorProblem = ringProblem(geometry.exterior.points, 'Внешний контур', true)
    if (exteriorProblem) return exteriorProblem
    for (const [index, hole] of geometry.holes.entries()) {
        const problem = ringProblem(hole.points, `Отверстие ${index + 1}`, true)
        if (problem) return problem
    }
    return null
}

export interface DrawingViewBox { x: number; y: number; width: number; height: number }

export function fitDrawingView(geometry: LibraryGeometry): DrawingViewBox {
    const points = [geometry.exterior, ...geometry.holes].flatMap(ring => ring.points)
        .filter(p => Number.isFinite(p.x) && Number.isFinite(p.y))
    if (points.length === 0) return { x: 0, y: 0, width: 1000, height: 650 }
    const minX = Math.min(...points.map(point => point.x))
    const minY = Math.min(...points.map(point => point.y))
    const maxX = Math.max(...points.map(point => point.x))
    const maxY = Math.max(...points.map(point => point.y))
    const w = Math.max(maxX - minX, 1)
    const h = Math.max(maxY - minY, 1)
    const width = Math.max(w * 1.3, h * 1.3 * (640 / 420))
    const height = width * (420 / 640)
    return {
        x: minX + (maxX - minX) / 2 - width / 2,
        y: minY + (maxY - minY) / 2 - height / 2,
        width,
        height,
    }
}
