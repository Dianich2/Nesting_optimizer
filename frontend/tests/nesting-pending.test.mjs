import assert from 'node:assert/strict'
import test from 'node:test'
import { readPendingNesting } from '../src/features/nesting/runNesting.ts'

const record = {
    key: '11111111-1111-4111-8111-111111111111',
    userId: 7,
    projectId: 1,
    surfaceId: 2,
    request: {
        algorithm: 'baseline',
        patterns: [{ project_pattern_id: 1, quantity: 2 }],
        allowed_rotations: [0, 90],
        keep_existing: true,
    },
}

function installStorage(t) {
    const original = Object.getOwnPropertyDescriptor(globalThis, 'sessionStorage')
    const storage = { getItem: () => null }
    Object.defineProperty(globalThis, 'sessionStorage', { value: storage, configurable: true })
    t.after(() => {
        if (original) Object.defineProperty(globalThis, 'sessionStorage', original)
        else Reflect.deleteProperty(globalThis, 'sessionStorage')
    })
    return storage
}

test('pending nesting is bound to the authenticated owner', (t) => {
    const storage = installStorage(t)
    storage.getItem = () => JSON.stringify(record)
    assert.deepEqual(readPendingNesting('owner-and-surface-key', 7, 1, 2), record)
    assert.equal(readPendingNesting('owner-and-surface-key', 8, 1, 2), null)
    assert.equal(readPendingNesting('owner-and-surface-key', 7, 2, 2), null)
    assert.equal(readPendingNesting('owner-and-surface-key', 7, 1, 3), null)
})

test('malformed and legacy unowned pending requests are rejected', (t) => {
    const storage = installStorage(t)
    for (const value of [
        null,
        [],
        {},
        { ...record, userId: undefined },
        { ...record, key: 'bad' },
        { ...record, request: { ...record.request, algorithm: 'unknown' } },
        {
            ...record,
            request: { ...record.request, patterns: [{ project_pattern_id: 1, quantity: 501 }] },
        },
        { ...record, request: { ...record.request, allowed_rotations: [null] } },
        { ...record, request: { ...record.request, keep_existing: 'false' } },
        { ...record, request: { ...record.request, preview_id: {} } },
    ]) {
        storage.getItem = () => JSON.stringify(value)
        assert.equal(readPendingNesting('key', 7, 1, 2), null)
    }
    storage.getItem = () => '{'
    assert.equal(readPendingNesting('key', 7, 1, 2), null)
})
