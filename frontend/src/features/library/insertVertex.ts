import type { LibraryPoint } from './types'

/** Squared distance from a point to a finite segment (not its infinite line). */
function segmentDistanceSquared(p: LibraryPoint, a: LibraryPoint, b: LibraryPoint): number {
    const dx = b.x - a.x
    const dy = b.y - a.y
    const lengthSquared = dx * dx + dy * dy
    if (lengthSquared === 0) return (p.x - a.x) ** 2 + (p.y - a.y) ** 2
    const t = Math.max(0, Math.min(1, ((p.x - a.x) * dx + (p.y - a.y) * dy) / lengthSquared))
    const x = a.x + t * dx
    const y = a.y + t * dy
    return (p.x - x) ** 2 + (p.y - y) ** 2
}

/**
 * In an incomplete ring, append points normally. In a closed ring, insert on the
 * nearest suitable edge (including the closing edge from last to first).
 *
 * The callback validates the *whole* candidate geometry: crossings, holes,
 * containment, etc. We inspect only the 12 closest edges to keep large
 * polygons responsive, and return null if no candidate is admissible.
 */
export function insertVertex(
    points: LibraryPoint[],
    point: LibraryPoint,
    isValid: (candidate: LibraryPoint[]) => boolean,
): LibraryPoint[] | null {
    if (points.length < 3) {
        const candidate = [...points, point]
        return isValid(candidate) ? candidate : null
    }

    const nearest = points
        .map((start, index) => ({
            index,
            distance: segmentDistanceSquared(point, start, points[(index + 1) % points.length]),
        }))
        .sort((a, b) => a.distance - b.distance)
        .slice(0, 12)

    for (const { index } of nearest) {
        const candidate = [
            ...points.slice(0, index + 1),
            point,
            ...points.slice(index + 1),
        ]
        if (isValid(candidate)) return candidate
    }
    return null
}
