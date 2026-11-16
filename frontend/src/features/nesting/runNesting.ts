import { ApiError } from '../../shared/api/http.ts'
import type { NestingRequest, NestingResult, Placement, Request } from './api'

export async function runNesting(
    request: Request,
    base: string,
    input: NestingRequest,
    onConfirmed: (result: NestingResult) => void,
    onSynchronized: (placements: Placement[]) => void,
    operationKey?: string,
): Promise<'synchronized' | 'confirmed' | 'unknown'> {
    let result: NestingResult
    try {
        result = await request<NestingResult>(`${base}/nesting`, {
            method: 'POST',
            headers: operationKey ? { 'Idempotency-Key': operationKey } : undefined,
            body: JSON.stringify(input),
        })
    } catch (error) {
        if (error instanceof ApiError && error.status < 500) throw error
        return 'unknown'
    }

    onConfirmed(result)
    let saved: { items: Placement[] }
    try {
        saved = await request<{ items: Placement[] }>(`${base}/placements`)
    } catch {
        return 'confirmed'
    }
    onSynchronized(saved.items)
    return 'synchronized'
}

export interface PendingNesting {
    key: string
    request: NestingRequest
}

export function readPendingNesting(storageKey: string): PendingNesting | null {
    try {
        const stored = sessionStorage.getItem(storageKey)
        return stored ? (JSON.parse(stored) as PendingNesting) : null
    } catch {
        return null
    }
}
