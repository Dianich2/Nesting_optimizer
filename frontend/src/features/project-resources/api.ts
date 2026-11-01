
import type { AuthContextValue } from '../auth/AuthContext'

import type {
    PaginatedResponse,
    ProjectPattern,
    ProjectSurface,
} from './types'

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
