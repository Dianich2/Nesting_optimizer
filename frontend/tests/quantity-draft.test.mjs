import test from 'node:test'
import assert from 'node:assert/strict'
import { parseQuantityDrafts } from '../src/features/nesting/quantityDraft.ts'
test('empty and zero drafts remain unselected; normalize only on submit', () => {
    assert.deepEqual(parseQuantityDrafts({ 1: '', 2: '0', 3: '002' }), {
        patterns: [{ project_pattern_id: 3, quantity: 2 }],
        error: null,
    })
    assert.deepEqual(parseQuantityDrafts({ 1: '' }).patterns, [])
})
test('reject fractions negatives NaN unsafe integers and total overflow', () => {
    for (const raw of ['-1', '1.5', 'NaN', '501', '1e999', '9007199254740993'])
        assert.ok(parseQuantityDrafts({ 1: raw }).error, raw)
    assert.ok(parseQuantityDrafts({ 1: '250', 2: '251' }).error)
    assert.equal(parseQuantityDrafts({ 1: '250', 2: '250' }).error, null)
})
