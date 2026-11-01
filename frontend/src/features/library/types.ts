export interface LibraryPoint {
    x: number
    y: number
}

export interface LibraryRing {
    points: LibraryPoint[]
}

export interface LibraryGeometry {
    exterior: LibraryRing
    holes: LibraryRing[]
}

export interface LibraryItem {
    id: number
    name: string
    geometry: LibraryGeometry
    created_at: string
    updated_at: string
}

export interface LibraryListResponse {
    items: LibraryItem[]
    page: number
    page_size: number
    total: number
    total_pages: number
}

export interface CreateLibraryItemRequest {
    name: string
    geometry: LibraryGeometry
}
