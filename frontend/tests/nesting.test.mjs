import assert from 'node:assert/strict'
import test from 'node:test'
import { polygonArea, polygonPath, layoutMetrics } from '../src/features/nesting/geometry.ts'
const ring = (size) => ({ points: [{ x: 0, y: 0 }, { x: size, y: 0 }, { x: size, y: size }, { x: 0, y: size }] })
const polygon = (size) => ({ exterior: ring(size), holes: [] })
test('area subtracts holes regardless of winding', () => {
    const hole = ring(2)
    hole.points.reverse()
    assert.equal(polygonArea({ exterior: ring(10), holes: [hole] }), 96)
})
test('SVG closes outer and inner contours', () => {
    const path = polygonPath({ exterior: ring(10), holes: [ring(2)] })
    assert.equal(path.split(' Z').length - 1, 2)
})
test('layout metrics count all saved placements and start fresh for another surface', () => {
    const items = [1, 1, 2].map((id) => ({ project_pattern_id: id, geometry: polygon(2) }))
    const metrics = layoutMetrics(polygon(10), items)
    assert.equal(metrics.placedArea, 12)
    assert.equal(metrics.freeArea, 88)
    assert.equal(metrics.utilization, 0.12)
    assert.deepEqual(metrics.counts, { 1: 2, 2: 1 })
    assert.equal(layoutMetrics(polygon(20), []).placedArea, 0)
    assert.equal(layoutMetrics({ exterior: { points: [] }, holes: [] }, []).utilization, 0)
})
