export interface PlacementPose {
    project_pattern_id: number
    x: number
    y: number
    rotation: number
}

export interface PlacementChange {
    id: number
    before: PlacementPose | null
    after: PlacementPose | null
}

export interface PlacementHistory {
    undo: PlacementChange[]
    redo: PlacementChange[]
}

export function recordPlacement(
    history: PlacementHistory,
    change: PlacementChange,
): PlacementHistory {
    return { undo: [...history.undo, change], redo: [] }
}

export function confirmReplay(
    history: PlacementHistory,
    direction: 'undo' | 'redo',
    newID: number,
): PlacementHistory {
    const reversed = direction === 'undo' ? 'redo' : 'undo'
    const change = history[direction].at(-1)
    if (!change) return history
    const next = {
        ...history,
        [direction]: history[direction].slice(0, -1),
        [reversed]: [...history[reversed], change],
    }
    if (newID === change.id) return next
    const remap = (value: PlacementChange) =>
        value.id === change.id ? { ...value, id: newID } : value
    return { undo: next.undo.map(remap), redo: next.redo.map(remap) }
}
