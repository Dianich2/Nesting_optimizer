import type { AuthContextValue } from '../auth/AuthContext'
import type { LibraryListResponse } from './types'
import type { CreateLibraryItemRequest, LibraryItem } from './types'

type AuthorizedRequest = AuthContextValue['authorizedRequest']

export type LibraryKind = 'surfaces' | 'patterns'

export function listLibraryItems(
    request: AuthorizedRequest,
    kind: LibraryKind,
    page = 1,
    pageSize = 20,
): Promise<LibraryListResponse> {
    const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
    })

    return request<LibraryListResponse>(`/${kind}?${params.toString()}`)
}

export function createLibraryItem(
    request: AuthorizedRequest,
    kind: LibraryKind,
    data: CreateLibraryItemRequest,
): Promise<LibraryItem> {
    return request<LibraryItem>(`/${kind}`, {
        method: 'POST',
        body: JSON.stringify(data),
    })
}
