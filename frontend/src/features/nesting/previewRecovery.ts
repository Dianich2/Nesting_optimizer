import { ApiError } from '../../shared/api/http.ts'
import type { NestingRequest, NestingResult } from './api'
import type { PendingNesting } from './runNesting'

export type PreviewFailure = 'rejected' | 'check_status' | 'uncertain'

// Match application codes AND HTTP statuses. A proxy's HTTP status alone is
// insufficient evidence. Save operations deliberately do not use this policy.
export function classifyPreviewFailure(error: unknown, retry = false): PreviewFailure {
    if (!(error instanceof ApiError)) return 'uncertain'
    if (error.status === 409 && error.code === 'conflict_error') return 'check_status'
    const authentication =
        (error.status === 401 && error.code === 'unauthorized_error') ||
        (error.status === 403 && error.code === 'forbidden_error')
    // A rejected retry does not prove that the earlier request stopped.
    if (retry && authentication) return 'check_status'
    if (authentication) return 'rejected'
    const rejections: Record<string, number> = {
        validation_error: 400,
        not_found_error: 404,
        nesting_stale: 409,
        nesting_cancelled: 409,
        nesting_timeout: 408,
        nesting_overloaded: 503,
        preview_expired: 404,
        nesting_interrupted: 409,
    }
    return error.code && rejections[error.code] === error.status ? 'rejected' : 'uncertain'
}

export interface NestingOperationStatus {
    status: 'unknown' | 'running' | 'completed' | 'failed'
    result?: NestingResult
    failure?: { code: string; message: string }
}

export function previewStatusOutcome(current: PendingNesting, status: NestingOperationStatus) {
    if (current.kind !== 'preview') return { kind: 'unresolved' as const }
    if (status.status === 'completed' && status.result?.preview_id) {
        return {
            kind: 'completed' as const,
            proposal: {
                result: status.result,
                request: current.request,
                includeSaved: current.request.keep_existing,
            },
        }
    }
    if (status.status === 'failed') return { kind: 'failed' as const, failure: status.failure }
    return { kind: 'unresolved' as const }
}

export function mayClosePreview(
    current: PendingNesting,
    owner: number,
    project: number,
    surface: number,
) {
    return (
        current.kind === 'preview' &&
        current.userId === owner &&
        current.projectId === project &&
        current.surfaceId === surface
    )
}

export function validateNestingRequest(input: NestingRequest): string | null {
    if (
        !input.patterns.length ||
        input.patterns.some((p) => !Number.isSafeInteger(p.quantity) || p.quantity <= 0)
    )
        return 'Укажите целое положительное количество хотя бы одного лекала.'
    if (input.patterns.reduce((sum, p) => sum + p.quantity, 0) > 500)
        return 'За один запуск можно разместить не более 500 деталей.'
    // Validate the submitted list before deduplication, as Backend.Validate does.
    if (input.allowed_rotations.length > 36) return 'Укажите не более 36 допустимых углов.'
    if (
        !input.allowed_rotations.length ||
        input.allowed_rotations.some(
            (angle) => !Number.isFinite(angle) || angle < 0 || angle >= 360,
        )
    )
        return 'Укажите непустой список конечных углов от 0 до 360° (не включая 360).'
    return null
}

export function parseNestingAngles(raw: string): number[] {
    // Empty tokens must not silently become zero through Number('').
    return raw.split(',').map((token) => (token.trim() ? Number(token.trim()) : NaN))
}
