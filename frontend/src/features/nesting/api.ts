import type { AuthContextValue } from '../auth/AuthContext'
import type { PolygonGeometry, ProjectPattern, PaginatedResponse } from '../project-resources/types'

export interface Placement {
    id: number
    project_surface_id: number
    project_pattern_id: number
    x: number
    y: number
    rotation: number
    geometry: PolygonGeometry
}
export interface NestingRequest {
    algorithm: string
    patterns: { project_pattern_id: number; quantity: number }[]
    allowed_rotations: number[]
    keep_existing: boolean
}
export interface NestingResult {
    placements: Placement[]
    unplaced: { project_pattern_id: number; quantity: number }[]
    metrics: {
        requested_count: number
        placed_count: number
        surface_area: number
        placed_area: number
        utilization: number
    }
}
export type Request = AuthContextValue['authorizedRequest']

export async function allProjectPatterns(
    request: Request,
    projectId: number,
): Promise<ProjectPattern[]> {
    const result: ProjectPattern[] = []
    let page = 1
    for (;;) {
        const response = await request<PaginatedResponse<ProjectPattern>>(
            `/projects/${projectId}/patterns?page=${page}&page_size=100`,
        )
        result.push(...response.items)
        if (page >= response.total_pages) return result
        page++
    }
}
