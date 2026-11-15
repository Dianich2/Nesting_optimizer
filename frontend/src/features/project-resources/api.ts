import type { AuthContextValue } from '../auth/AuthContext'

import type { PaginatedResponse, ProjectPattern, ProjectSurface } from './types'

type AuthorizedRequest = AuthContextValue['authorizedRequest']

export function listProjectSurfaces(
    request: AuthorizedRequest,
    projectId: number,
    page = 1,
    pageSize = 20,
): Promise<PaginatedResponse<ProjectSurface>> {
    const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
    })

    return request<PaginatedResponse<ProjectSurface>>(
        `/projects/${projectId}/surfaces?${params.toString()}`,
    )
}

export function listProjectPatterns(
    request: AuthorizedRequest,
    projectId: number,
    page = 1,
    pageSize = 20,
): Promise<PaginatedResponse<ProjectPattern>> {
    const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
    })

    return request<PaginatedResponse<ProjectPattern>>(
        `/projects/${projectId}/patterns?${params.toString()}`,
    )
}

export function createProjectResource(
    request: AuthorizedRequest,
    projectId: number,
    kind: 'surfaces' | 'patterns',
    sourceId: number,
    scale: number,
): Promise<ProjectSurface | ProjectPattern> {
    const sourceKey = kind === 'surfaces' ? 'source_surface_id' : 'source_pattern_id'
    return request(`/projects/${projectId}/${kind}`, {
        method: 'POST',
        body: JSON.stringify({ [sourceKey]: sourceId, scale }),
    })
}

export function updateProjectResource(
    request: AuthorizedRequest,
    projectId: number,
    kind: 'surfaces' | 'patterns',
    id: number,
    data: { name?: string; geometry?: ProjectSurface['geometry'] },
): Promise<ProjectSurface | ProjectPattern> {
    return request(`/projects/${projectId}/${kind}/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
    })
}

export function deleteProjectResource(
    request: AuthorizedRequest,
    projectId: number,
    kind: 'surfaces' | 'patterns',
    id: number,
): Promise<void> {
    return request(`/projects/${projectId}/${kind}/${id}`, { method: 'DELETE' })
}
