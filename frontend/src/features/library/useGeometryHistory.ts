import { useState } from 'react'

export function useGeometryHistory<T>() {
    const [history, setHistory] = useState<{ past: T[]; future: T[] }>({ past: [], future: [] })

    function record(value: T) {
        setHistory(({ past }) => ({ past: [...past.slice(-99), value], future: [] }))
    }

    function undo(current: T): T | undefined {
        const previous = history.past.at(-1)
        if (previous === undefined) return
        setHistory({ past: history.past.slice(0, -1), future: [...history.future, current] })
        return previous
    }

    function redo(current: T): T | undefined {
        const next = history.future.at(-1)
        if (next === undefined) return
        setHistory({ past: [...history.past, current], future: history.future.slice(0, -1) })
        return next
    }

    return { record, undo, redo, canUndo: !!history.past.length, canRedo: !!history.future.length }
}
