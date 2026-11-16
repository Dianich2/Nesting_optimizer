import assert from 'node:assert/strict'
import test from 'node:test'
import { polygonArea } from '../src/features/nesting/geometry.ts'
import { exportLayoutSVG } from '../src/features/nesting/exportLayout.ts'
import { confirmReplay, recordPlacement } from '../src/features/nesting/placementHistory.ts'
import { normalizeAngle, angleAround, snapCoordinate } from '../src/shared/geometry/viewport.ts'

test('export uses the world-space bounds, hole and transformed placement', () => {
    const polygon = {
        exterior: {
            points: [
                { x: -10, y: -5 },
                { x: 40, y: -5 },
                { x: 40, y: 25 },
            ],
        },
        holes: [
            {
                points: [
                    { x: 10, y: 0 },
                    { x: 12, y: 0 },
                    { x: 12, y: 1 },
                ],
            },
        ],
    }
    const item = {
        geometry: {
            exterior: {
                points: [
                    { x: 30, y: 0 },
                    { x: 31, y: 0 },
                    { x: 31, y: 1 },
                ],
            },
            holes: [],
        },
    }
    const svg = exportLayoutSVG(polygon, [item])
    assert.match(svg, /viewBox="-10 -5 50 30"/)
    assert.match(svg, /fill-rule="evenodd"/)
    assert.match(svg, /M 10 0 L 12 0 L 12 1 Z/)
    assert.match(svg, /M 30 0 L 31 0 L 31 1 Z/)
})

test('recreating a deleted placement remaps earlier commands and a new edit clears redo', () => {
    const initial = { project_pattern_id: 2, x: 20, y: 0, rotation: 0 }
    const moved = { ...initial, x: 30, rotation: 15 }
    let history = { undo: [], redo: [] }
    history = recordPlacement(history, { id: 8, before: null, after: initial })
    history = recordPlacement(history, { id: 8, before: initial, after: moved })
    history = recordPlacement(history, { id: 8, before: moved, after: null })
    history = confirmReplay(history, 'undo', 10)
    assert.deepEqual(
        history.undo.map(({ id }) => id),
        [10, 10],
    )
    assert.equal(history.redo[0].id, 10)
    history = confirmReplay(history, 'undo', 10)
    history = confirmReplay(history, 'redo', 10)
    assert.deepEqual(history.undo.at(-1).after, moved)
    assert.deepEqual(recordPlacement(history, { id: 10, before: moved, after: initial }).redo, [])
})

test('disabled snap preserves coordinates and rotation follows the pointer around the server center', () => {
    assert.equal(snapCoordinate(-13.5, 10, false), -13.5)
    assert.equal(snapCoordinate(-13.5, 10, true), -10)
    assert.equal(normalizeAngle(angleAround({ x: 12, y: 20 }, { x: 12, y: 10 })), 270)
})

test('translated geometry keeps its area at large world coordinates', () => {
    const shift = 1e12
    const points = [
        { x: shift, y: shift },
        { x: shift + 10, y: shift },
        { x: shift + 10, y: shift + 4 },
        { x: shift, y: shift + 4 },
    ]
    assert.equal(polygonArea({ exterior: { points }, holes: [] }), 40)
})
