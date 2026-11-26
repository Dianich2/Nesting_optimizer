import type { PolygonGeometry, GeometryPoint } from '../project-resources/types'
import type { PlacementPose } from './placementHistory'

type Point = GeometryPoint
type Class = 'inside' | 'outside' | 'edge'
const EPS = 1e-9

function ringPoints(polygon: PolygonGeometry): Point[][] {
    return [polygon.exterior, ...polygon.holes].map((ring) => ring.points)
}
function cross(a: Point, b: Point, c: Point) {
    return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x)
}
function onEdge(a: Point, b: Point, p: Point) {
    const scale = Math.max(1, Math.hypot(b.x - a.x, b.y - a.y))
    return Math.abs(cross(a, b, p)) <= EPS * scale &&
        p.x >= Math.min(a.x, b.x) - EPS && p.x <= Math.max(a.x, b.x) + EPS &&
        p.y >= Math.min(a.y, b.y) - EPS && p.y <= Math.max(a.y, b.y) + EPS
}
function inRing(p: Point, ring: Point[]): Class {
    if (ring.length < 3) return 'outside'
    let inside = false
    for (let i = 0, j = ring.length - 1; i < ring.length; j = i++) {
        const a = ring[j], b = ring[i]
        if (onEdge(a, b, p)) return 'edge'
        if ((a.y > p.y) !== (b.y > p.y) &&
            p.x < ((b.x - a.x) * (p.y - a.y)) / (b.y - a.y) + a.x) inside = !inside
    }
    return inside ? 'inside' : 'outside'
}
function inPolygon(p: Point, poly: PolygonGeometry): Class {
    const outer = inRing(p, poly.exterior.points)
    if (outer !== 'inside') return outer
    for (const hole of poly.holes) {
        const insideHole = inRing(p, hole.points)
        if (insideHole === 'inside') return 'outside'
        if (insideHole === 'edge') return 'edge'
    }
    return 'inside'
}
function edges(points: Point[]): Array<[Point, Point]> {
    return points.map((p, i): [Point, Point] => [p, points[(i + 1) % points.length]]).filter(([a, b]) => a.x !== b.x || a.y !== b.y)
}
function properCross(a: Point, b: Point, c: Point, d: Point) {
    const s1 = cross(a, b, c), s2 = cross(a, b, d)
    const s3 = cross(c, d, a), s4 = cross(c, d, b)
    return (s1 > EPS && s2 < -EPS || s1 < -EPS && s2 > EPS) &&
        (s3 > EPS && s4 < -EPS || s3 < -EPS && s4 > EPS)
}
function boundaryCrosses(a: PolygonGeometry, b: PolygonGeometry) {
    for (const ar of ringPoints(a)) for (const br of ringPoints(b)) {
        for (const [a1, a2] of edges(ar)) for (const [b1, b2] of edges(br)) {
            if (properCross(a1, a2, b1, b2)) return true
        }
    }
    return false
}
function sampleInterior(poly: PolygonGeometry): Point[] {
    const pts = poly.exterior.points
    if (pts.length < 3) return []
    let signed = 0
    for (const [a, b] of edges(pts)) signed += a.x * b.y - b.x * a.y
    const direction = signed >= 0 ? 1 : -1
    const samples: Point[] = []
    for (const [a, b] of edges(pts)) {
        const length = Math.hypot(b.x - a.x, b.y - a.y)
        if (length < EPS) continue
        const offset = Math.min(length * 0.0001, 0.0001)
        const p = {
            x: (a.x + b.x) / 2 - direction * (b.y - a.y) / length * offset,
            y: (a.y + b.y) / 2 + direction * (b.x - a.x) / length * offset,
        }
        if (inPolygon(p, poly) === 'inside') samples.push(p)
    }
    return samples
}
function hasInteriorSample(source: PolygonGeometry, other: PolygonGeometry) {
    const points = [...ringPoints(source).flat(), ...sampleInterior(source)]
    // A boundary point touching another polygon is valid; only interiors overlap.
    return points.some((point) => inPolygon(point, source) === 'inside' && inPolygon(point, other) === 'inside')
}
function interiorsOverlap(a: PolygonGeometry, b: PolygonGeometry) {
    return boundaryCrosses(a, b) || hasInteriorSample(a, b) || hasInteriorSample(b, a)
}

// Mirrors the server's pattern transform: original bounding-box centre is
// placed at (x,y) after rotation. No persistence occurs in this module.
export function transformPattern(pattern: PolygonGeometry, pose: Pick<PlacementPose, 'x' | 'y' | 'rotation'>): PolygonGeometry {
    const points = ringPoints(pattern).flat()
    const minX = Math.min(...points.map((p) => p.x))
    const maxX = Math.max(...points.map((p) => p.x))
    const minY = Math.min(...points.map((p) => p.y))
    const maxY = Math.max(...points.map((p) => p.y))
    const cx = (minX + maxX) / 2, cy = (minY + maxY) / 2
    const angle = pose.rotation * Math.PI / 180
    const cos = Math.cos(angle), sin = Math.sin(angle)
    const move = (p: Point) => ({
        x: pose.x + (p.x - cx) * cos - (p.y - cy) * sin,
        y: pose.y + (p.x - cx) * sin + (p.y - cy) * cos,
    })
    return {
        exterior: { points: pattern.exterior.points.map(move) },
        holes: pattern.holes.map((ring) => ({ points: ring.points.map(move) })),
    }
}

export function placementIssue(surface: PolygonGeometry, candidate: PolygonGeometry, others: PolygonGeometry[]): string | null {
    const boundary = surface.exterior.points
    if (boundary.length < 3 || candidate.exterior.points.length < 3) return 'Некорректная геометрия.'
    const vertices = ringPoints(candidate).flat()
    if (vertices.some((p) => !Number.isFinite(p.x) || !Number.isFinite(p.y))) return 'Некорректные координаты.'
    const totalOtherVertices = others.reduce((sum, item) => sum + ringPoints(item).flat().length, 0)
    if (vertices.length * totalOtherVertices > 250_000) {
        return 'Геометрия слишком сложна для быстрой проверки. Нажмите «Проверить и сохранить», чтобы проверить её на сервере.'
    }
    if (vertices.some((p) => inRing(p, boundary) === 'outside') || boundaryCrosses(candidate, {exterior:surface.exterior, holes:[]})) {
        return 'Деталь выходит за границу поверхности.'
    }
    if (surface.holes.some((hole) =>
        vertices.some((p) => inRing(p, hole.points) === 'inside') ||
        boundaryCrosses(candidate, {exterior:hole, holes:[]}) ||
        hole.points.some((p) => inPolygon(p, candidate) === 'inside') ||
        sampleInterior({exterior:hole, holes:[]}).some((p) => inPolygon(p, candidate) === 'inside')
    )) return 'Деталь пересекает отверстие поверхности.'
    // An edge may leave a concave surface and return without crossing an
    // edge properly (e.g. while following a vertex). Check midpoints, too.
    for (const ring of ringPoints(candidate)) for (const [a, b] of edges(ring)) {
        if (inPolygon({x:(a.x+b.x)/2, y:(a.y+b.y)/2}, surface) === 'outside') return 'Деталь выходит за допустимую область поверхности.'
    }
    if (others.some((other) => interiorsOverlap(candidate, other))) {
        return 'Деталь пересекает другое размещение.'
    }
    return null
}
