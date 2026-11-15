import { useState } from 'react'
import type { Placement } from './api'

interface Props {
    placement: Placement
    busy: boolean
    onSave: (x: number, y: number, rotation: number) => void
    onDelete: () => void
}
export function PlacementProperties({ placement, busy, onSave, onDelete }: Props) {
    const [x, setX] = useState(String(placement.x))
    const [y, setY] = useState(String(placement.y))
    const [rotation, setRotation] = useState(String(placement.rotation))
    return (
        <form
            onSubmit={(event) => {
                event.preventDefault()
                if (
                    [x, y, rotation].every(
                        (value) => value.trim() && Number.isFinite(Number(value)),
                    )
                )
                    onSave(Number(x), Number(y), Number(rotation))
            }}
        >
            <h3>Размещение №{placement.id}</h3>
            <label>
                X центра
                <input
                    type="number"
                    step="any"
                    required
                    value={x}
                    disabled={busy}
                    onChange={(event) => setX(event.target.value)}
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
                    onChange={(event) => setY(event.target.value)}
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
                    onChange={(event) => setRotation(event.target.value)}
                />
            </label>
            <div className="nesting-tools">
                {[0, 90, 180, 270].map((angle) => (
                    <button
                        key={angle}
                        type="button"
                        disabled={busy}
                        onClick={() => setRotation(String(angle))}
                    >
                        {angle}°
                    </button>
                ))}
            </div>
            <button disabled={busy}>Сохранить положение</button>
            <button type="button" disabled={busy} onClick={onDelete}>
                Удалить размещение
            </button>
        </form>
    )
}
