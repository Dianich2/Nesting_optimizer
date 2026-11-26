import { useRef, useState } from 'react'
import type { Placement } from './api'

interface Props {
    placement: Placement
    busy: boolean
    onSave: (x: number, y: number, rotation: number) => void
    onDraftChange: (x: number, y: number, rotation: number) => void
    newPlacement?: boolean
}
export function PlacementProperties({ placement, busy, onSave, onDraftChange, newPlacement = false }: Props) {
    const [x, setX] = useState(String(placement.x))
    const [y, setY] = useState(String(placement.y))
    const [rotation, setRotation] = useState(String(placement.rotation))
    const dirty = useRef(false)
    function validPose(nextX: string, nextY: string, nextRotation: string) {
        if (![nextX, nextY, nextRotation].every((value) =>
            value.trim() && Number.isFinite(Number(value)),
        )) return null
        return [Number(nextX), Number(nextY), Number(nextRotation)] as const
    }
    function updateDraft(nextX: string, nextY: string, nextRotation: string) {
        dirty.current = true
        const pose = validPose(nextX, nextY, nextRotation)
        if (pose) onDraftChange(...pose)
    }
    function finishEditing(nextRotation = rotation) {
        if (!dirty.current || busy) return
        const pose = validPose(x, y, nextRotation)
        if (!pose) return // Incomplete numeric input must never be persisted.
        dirty.current = false
        onSave(...pose)
    }
    return (
        <form
            className="nesting-placement-properties"
            onBlur={(event) => {
                // Moving between properties and preset buttons is still one edit.
                if (event.currentTarget.contains(event.relatedTarget)) return
                // Explicit deletion must not accidentally save an unsaved draft first.
                if (event.relatedTarget instanceof Element &&
                    event.relatedTarget.closest('.nesting-delete-action')) return
                finishEditing()
            }}
            onSubmit={(event) => {
                event.preventDefault()
                if (
                    [x, y, rotation].every(
                        (value) => value.trim() && Number.isFinite(Number(value)),
                    )
                )
                    finishEditing()
            }}
        >
            <h3>{newPlacement ? 'Новое лекало (не сохранено)' : `Размещение №${placement.id}`}</h3>
            <label>
                X центра
                <input
                    type="number"
                    step="any"
                    required
                    value={x}
                    disabled={busy}
                    onChange={(event) => {
                        setX(event.target.value)
                        updateDraft(event.target.value, y, rotation)
                    }}
                />
            </label>
            <label>
                Y центра
                <input
                    type="number"
                    step="any"
                    required
                    value={y}
                    disabled={busy}
                    onChange={(event) => {
                        setY(event.target.value)
                        updateDraft(x, event.target.value, rotation)
                    }}
                />
            </label>
            <label>
                Поворот, °
                <input
                    type="number"
                    step="any"
                    min="0"
                    max="359.999999"
                    required
                    value={rotation}
                    disabled={busy}
                    onChange={(event) => {
                        setRotation(event.target.value)
                        updateDraft(x, y, event.target.value)
                    }}
                />
            </label>
            <div className="nesting-tools">
                {[0, 90, 180, 270].map((angle) => (
                    <button
                        key={angle}
                        type="button"
                        disabled={busy}
                        onClick={() => {
                            setRotation(String(angle))
                            updateDraft(x, y, String(angle))
                            // Preset buttons apply immediately, without a Save button.
                            const pose = validPose(x, y, String(angle))
                            if (pose && !busy) {
                                dirty.current = false
                                onSave(...pose)
                            }
                        }}
                    >
                        {angle}°
                    </button>
                ))}
            </div>
            <small>Координаты и угол применяются после завершения ввода. Некорректное положение не сохраняется.</small>
        </form>
    )
}
