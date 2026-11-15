import type { PolygonGeometry } from '../project-resources/types'

export function polygonPath(polygon: PolygonGeometry): string {
    return [polygon.exterior, ...polygon.holes]
        .filter((ring) => ring.points.length > 0)
        .map((ring) => `M ${ring.points.map((p) => `${p.x} ${p.y}`).join(' L ')} Z`)
        .join(' ')
}

export function polygonArea(polygon: PolygonGeometry): number {
    const areas = [polygon.exterior, ...polygon.holes].map(
        (ring) =>
            Math.abs(
                ring.points.reduce((sum, p, i, points) => {
                    const next = points[(i + 1) % points.length]
                    return sum + p.x * next.y - next.x * p.y
                }, 0),
            ) / 2,
    )
    return Math.max(0, areas[0] - areas.slice(1).reduce((sum, area) => sum + area, 0))
}

export function layoutMetrics(
    surface: PolygonGeometry,
    placements: { geometry: PolygonGeometry; project_pattern_id: number }[],
) {
    const surfaceArea = polygonArea(surface)
    const placedArea = placements.reduce((sum, item) => sum + polygonArea(item.geometry), 0)
    const counts: Record<number, number> = {}
    for (const item of placements)
        counts[item.project_pattern_id] = (counts[item.project_pattern_id] ?? 0) + 1
    const utilization = surfaceArea > 0 ? Math.min(1, Math.max(0, placedArea / surfaceArea)) : 0
    return {
        surfaceArea,
        placedArea,
        freeArea: Math.max(0, surfaceArea - placedArea),
        utilization,
        counts,
    }
}
