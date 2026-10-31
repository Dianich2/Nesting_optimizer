
import type { AuthContextValue } from '../auth/AuthContext'
import type {
    CreateProjectRequest,
    ProjectsListResponse,
    Project,
} from './types'

type AuthorizedRequest = AuthContextValue['authorizedRequest']

export function listProjects(
    request: AuthorizedRequest,
    page = 1,
    pageSize = 20,
): Promise<ProjectsListResponse> {
    const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
    })

    return request<ProjectsListResponse>(
        `/projects?${params.toString()}`,
    )
}

export function createProject(
    request: AuthorizedRequest,
    data: CreateProjectRequest,
): Promise<Project> {
    return request<Project>('/projects', {
        method: 'POST',
        body: JSON.stringify(data),
    })
}

export function updateProject(
    request: AuthorizedRequest,
    projectId: number,
    data: CreateProjectRequest,
): Promise<Project> {
    return request<Project>(`/projects/${projectId}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
    })
}

export function deleteProject(
    request: AuthorizedRequest,
    projectId: number,
): Promise<void> {
    return request<void>(`/projects/${projectId}`, {
        method: 'DELETE',
    })
}
