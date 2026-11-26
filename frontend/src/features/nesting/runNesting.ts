import { classifySaveFailure } from './saveRecovery.ts'
import type { SaveOutcome } from './saveRecovery'
import { ApiError } from '../../shared/api/http.ts'
import type { NestingOperationStatus, NestingRequest, NestingResult, Placement, Request } from './api'

export const nestingRequestTimeout = 130_000

export async function runNesting(
    request: Request,
    base: string,
    input: NestingRequest,
    onConfirmed: (result: NestingResult) => void,
    onSynchronized: (placements: Placement[]) => void,
    operationKey?: string,
    retry = false,
): Promise<SaveOutcome> {
    let result: NestingResult
    try {
        result = await request<NestingResult>(`${base}/nesting`, {
            method: 'POST',
            signal: AbortSignal.timeout(nestingRequestTimeout),
            headers: operationKey ? { 'Idempotency-Key': operationKey } : undefined,
            body: JSON.stringify(input),
        })
    } catch (error) {
        if (
            classifySaveFailure(error, retry) === 'definitive_rejection' &&
            error instanceof ApiError
        )
            return { status: 'rejected', error }
        if ((!retry && classifySaveFailure(error, retry) === 'uncertain') || !operationKey)
            return { status: 'unknown', error }
        return reconcileSavedNesting(
            request,
            base,
            operationKey,
            onConfirmed,
            onSynchronized,
            error,
        )
    }
    return synchronizeSavedNesting(request, base, result, onConfirmed, onSynchronized)
}

export async function reconcileSavedNesting(
    request: Request,
    base: string,
    key: string,
    onConfirmed: (result: NestingResult) => void,
    onSynchronized: (placements: Placement[]) => void,
    postError?: unknown,
): Promise<SaveOutcome> {
    let status: NestingOperationStatus
    try {
        status = await request<NestingOperationStatus>(`${base}/nesting/operations/${key}`, {
            signal: AbortSignal.timeout(10_000),
        })
    } catch (error) {
        return { status: 'unknown', error }
    }
    if (status.status === 'completed' && status.result)
        return synchronizeSavedNesting(request, base, status.result, onConfirmed, onSynchronized)
    if (status.status === 'failed' && status.failure)
        return {
            status: 'rejected',
            error: new ApiError(500, status.failure.code, status.failure.message),
        }
    return {
        status: 'unknown',
        error: postError,
        operationStatus: status.status === 'running' ? 'running' : 'unknown',
    }
}

async function synchronizeSavedNesting(
    request: Request,
    base: string,
    result: NestingResult,
    onConfirmed: (result: NestingResult) => void,
    onSynchronized: (placements: Placement[]) => void,
): Promise<SaveOutcome> {
    onConfirmed(result)
    let saved: { items: Placement[] }
    try {
        saved = await request<{ items: Placement[] }>(`${base}/placements`, {
            signal: AbortSignal.timeout(10_000),
        })
    } catch {
        return { status: 'confirmed' }
    }
    onSynchronized(saved.items)
    return { status: 'synchronized' }
}

export interface PendingNesting {
    confirmed?: true
    kind?: 'preview'
    key: string
    userId: number
    projectId: number
    surfaceId: number
    request: NestingRequest
}

function isNestingKey(value: unknown): value is string {
    return (
        typeof value === 'string' &&
        value !== '00000000-0000-0000-0000-000000000000' &&
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(value)
    )
}

export function readPendingNesting(
    storageKey: string,
    userId: number,
    projectId: number,
    surfaceId: number,
): PendingNesting | null {
    try {
        const raw: unknown = JSON.parse(sessionStorage.getItem(storageKey) ?? 'null')
        if (!raw || typeof raw !== 'object') return null
        const value = raw as Partial<PendingNesting>
        const input = value.request
        if (
            (value.confirmed !== undefined && value.confirmed !== true) ||
            // Retired calculation-only requests cannot write placements.
            // Do not recover them as saving operations.
            value.kind !== undefined ||
            value.userId !== userId ||
            value.projectId !== projectId ||
            value.surfaceId !== surfaceId ||
            !isNestingKey(value.key) ||
            !input ||
            !['baseline', 'nfp_greedy', 'genetic', 'simulated_annealing', 'hybrid'].includes(
                input.algorithm,
            ) ||
            typeof input.keep_existing !== 'boolean' ||
            !Array.isArray(input.patterns) ||
            !input.patterns.length ||
            !input.patterns.every(
                (item) =>
                    Number.isSafeInteger(item.project_pattern_id) &&
                    item.project_pattern_id > 0 &&
                    Number.isSafeInteger(item.quantity) &&
                    item.quantity > 0,
            ) ||
            input.patterns.reduce((sum, item) => sum + item.quantity, 0) > 500 ||
            !Array.isArray(input.allowed_rotations) ||
            !input.allowed_rotations.length ||
            input.allowed_rotations.length > 36 ||
            !input.allowed_rotations.every(
                (angle) => Number.isFinite(angle) && angle >= 0 && angle < 360,
            ) ||
            (input.preview_id !== undefined && !isNestingKey(input.preview_id))
        )
            return null
        return value as PendingNesting
    } catch {
        return null
    }
}
