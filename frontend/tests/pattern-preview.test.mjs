import test from 'node:test'
import assert from 'node:assert/strict'
import { polygonArea, polygonDimensions } from '../src/features/nesting/geometry.ts'

const shape = {
    exterior: { points: [
        { x: 10, y: -5 }, { x: 40, y: -5 },
        { x: 40, y: 15 }, { x: 10, y: 15 },
    ] },
    holes: [{ points: [
        { x: 16, y: 1 }, { x: 22, y: 1 },
        { x: 22, y: 7 }, { x: 16, y: 7 },
    ] }],
}

test('pattern preview shows original exterior bounding-box dimensions', () => {
    assert.deepEqual(polygonDimensions(shape), { width: 30, height: 20 })
    assert.equal(polygonArea(shape), 600 - 36)
})

test('empty pattern bounds are safe', () => {
    assert.deepEqual(polygonDimensions({ exterior: { points: [] }, holes: [] }), { width: 0, height: 0 })
})
