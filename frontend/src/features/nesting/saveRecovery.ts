import { ApiError } from '../../shared/api/http.ts'
import type { PendingNesting } from './runNesting'

export function classifySaveFailure(error: unknown, retry: boolean) {
    if (
        !(error instanceof ApiError) ||
        !error.code ||
        (error.status >= 500 && error.code === 'internal_error')
    )
        return 'uncertain' as const
    // Input.Validate runs before Reserve. Its exact application envelope can
    // reject a fresh request; it cannot resolve an earlier ambiguous request.
    if (
        !retry &&
        error.status === 400 &&
        error.code === 'validation_error' &&
        error.message === 'validation failed' &&
        error.details.length > 0
    )
        return 'definitive_rejection' as const
    // Even an application timeout may belong only to this response waiter.
    // Read SQL truth before discarding a key, especially for retries/auth.
    return 'reconcile' as const
}

export type SaveOutcome =
    | { status: 'confirmed' }
    | { status: 'synchronized' }
    | { status: 'rejected'; error: ApiError }
    | { status: 'unknown'; error?: unknown; operationStatus?: 'unknown' | 'running' }

export function isPendingSaveScope(
    current: PendingNesting,
    owner: number,
    project: number,
    surface: number,
) {
    return (
        current.kind === undefined &&
        current.userId === owner &&
        current.projectId === project &&
        current.surfaceId === surface
    )
}
