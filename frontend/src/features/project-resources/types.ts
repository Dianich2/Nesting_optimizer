
export interface GeometryPoint {
    x: number
    y: number
}

export interface GeometryRing {
    points: GeometryPoint[]
}

export interface PolygonGeometry {
    exterior: GeometryRing
    holes: GeometryRing[]
}

export interface ProjectResource {
    id: number
    project_id: number
    name: string
    geometry: PolygonGeometry
    created_at: string
    updated_at: string
}

export interface ProjectSurface extends ProjectResource {
    source_surface_id: number | null
}

export interface ProjectPattern extends ProjectResource {
    source_pattern_id: number | null
}

export interface PaginatedResponse<T> {
    items: T[]
    page: number
    page_size: number
    total: number
    total_pages: number
}
