export function parseDraftRotation(raw: string): number | null {
    if (!raw.trim()) return null
    const value = Number(raw)
    return Number.isFinite(value) ? ((value % 360) + 360) % 360 : null
}

export function draftKeyAction(
    key: string,
    options: {
        adding: boolean
        blocked: boolean
        editable: boolean
        modal: boolean
        shift?: boolean
        ctrl?: boolean
        alt?: boolean
        meta?: boolean
    },
): 'cancel' | 90 | -90 | 15 | -15 | 1 | -1 | null {
    if (
        !options.adding ||
        options.editable ||
        options.modal ||
        options.ctrl ||
        options.alt ||
        options.meta
    )
        return null
    if (key === 'Escape') return 'cancel'
    if (options.blocked) return null
    if (key === 'ArrowLeft') return options.shift ? -15 : -1
    if (key === 'ArrowRight') return options.shift ? 15 : 1
    if (key.toLowerCase() !== 'r') return null
    return options.shift ? -90 : 90
}

// SVG applies these transforms right-to-left, matching Engine.Transform:
// subtract original bbox centre, rotate, position the centre at the cursor.
export function draftGhostTransform(
    center: { x: number; y: number },
    cursor: { x: number; y: number },
    rotation: number,
) {
    return `translate(${cursor.x} ${cursor.y}) rotate(${rotation}) translate(${-center.x} ${-center.y})`
}
