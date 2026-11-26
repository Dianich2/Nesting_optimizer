import type { NestingRequest } from './api'

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
