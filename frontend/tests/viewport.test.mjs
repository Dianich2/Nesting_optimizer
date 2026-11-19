import assert from 'node:assert/strict'
import test from 'node:test'
import { fitAspect, resizeView, zoomView } from '../src/shared/geometry/viewport.ts'

test('fitting a wide container preserves center and includes original bounds', () => {
    const source = { x: -10, y: 20, width: 100, height: 100 }
    const result = fitAspect(source, 2)
    assert.deepEqual(result, { x: -60, y: 20, width: 200, height: 100 })
    assert.deepEqual(source, { x: -10, y: 20, width: 100, height: 100 })
})

test('fitting a tall container preserves proportions', () => {
    assert.deepEqual(fitAspect({ x: 0, y: 0, width: 100, height: 50 }, 0.5), {
        x: 0,
        y: -75,
        width: 100,
        height: 200,
    })
})

test('zoom keeps the anchor at the same relative screen position', () => {
    const source = { x: -30, y: 15, width: 120, height: 80 }
    const anchor = { x: 10, y: 30 }
    const result = zoomView(source, 0.5, anchor)
    assert.equal((anchor.x - source.x) / source.width, (anchor.x - result.x) / result.width)
    assert.equal((anchor.y - source.y) / source.height, (anchor.y - result.y) / result.height)
    assert.deepEqual(zoomView(result, 2, anchor), source)
})

test('invalid and extreme zoom leave the view usable', () => {
    const source = { x: 0, y: 0, width: 100, height: 50 }
    for (const factor of [0, -1, Infinity, NaN, 1e20])
        assert.deepEqual(zoomView(source, factor), source)
    assert.deepEqual(fitAspect(source, 0), source)
})

test('resizing the workspace keeps its center and vertical zoom', () => {
    const original = { x: -120, y: 35, width: 800, height: 600 }
    const wider = resizeView(original, 2)
    assert.equal(wider.x + wider.width / 2, original.x + original.width / 2)
    assert.equal(wider.y + wider.height / 2, original.y + original.height / 2)
    assert.equal(wider.height, original.height)
    assert.deepEqual(resizeView(wider, original.width / original.height), original)
})
