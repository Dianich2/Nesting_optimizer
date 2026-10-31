
export interface Project {
    id: number
    user_id: number
    name: string
    description: string
    created_at: string
    updated_at: string
}

export interface ProjectsListResponse {
    items: Project[]
    page: number
    page_size: number
    total: number
    total_pages: number
}

export interface CreateProjectRequest {
    name: string
    description: string
}