import assert from 'node:assert/strict'
import test from 'node:test'
import { ApiError } from '../src/shared/api/http.ts'
import { classifySaveFailure, isPendingSaveScope } from '../src/features/nesting/saveRecovery.ts'
import { runNesting, readPendingNesting } from '../src/features/nesting/runNesting.ts'
const pending = {
    key: '11111111-1111-4111-8111-111111111111',
    userId: 1,
    projectId: 2,
    surfaceId: 3,
    request: {
        algorithm: 'baseline',
        patterns: [{ project_pattern_id: 1, quantity: 1 }],
        allowed_rotations: [0],
        keep_existing: true,
    },
}
const validation = new ApiError(400, 'validation_error', 'validation failed', [
    { field: 'quantity', code: 'invalid', message: 'at most 500' },
])
const result = { placements: [{ id: 7 }] }
for (const [http, code] of [
    [401, 'unauthorized_error'],
    [403, 'forbidden_error'],
    [409, 'conflict_error'],
    [409, 'nesting_stale'],
    [409, 'nesting_cancelled'],
    [409, 'nesting_interrupted'],
    [408, 'nesting_timeout'],
    [503, 'nesting_overloaded'],
    [503, 'nesting_pending'],
    [500, undefined],
    [500, 'internal_error'],
    [502, undefined],
    [503, undefined],
    [504, undefined],
]) {
    test(`save retry ${http}/${code} retains key until SQL terminal truth`, async () => {
        const failure = new ApiError(http, code)
        const uncertain = !code || code === 'internal_error'
        assert.equal(classifySaveFailure(failure, true), uncertain ? 'uncertain' : 'reconcile')
        for (const state of ['unknown', 'running', 'unavailable', 'completed', 'failed']) {
            let stored = pending
            let confirmed = false
            let posts = 0
            const request = async (path, options) => {
                if (options?.method === 'POST') {
                    posts++
                    assert.equal(options.headers['Idempotency-Key'], pending.key)
                    assert.deepEqual(JSON.parse(options.body), pending.request)
                    throw failure
                }
                if (path.endsWith('/placements')) return { items: result.placements }
                if (state === 'unavailable') throw new TypeError('offline')
                return {
                    status: state,
                    result: state === 'completed' ? result : undefined,
                    failure:
                        state === 'failed'
                            ? { code: 'nesting_stale', message: 'stale' }
                            : undefined,
                }
            }
            const outcome = await runNesting(
                request,
                '/surface',
                stored.request,
                () => {
                    confirmed = true
                },
                () => {},
                stored.key,
                true,
            )
            if (outcome.status !== 'unknown') stored = null
            assert.equal(posts, 1)
            assert.equal(stored === pending, ['unknown', 'running', 'unavailable'].includes(state))
            assert.equal(confirmed, state === 'completed')
        }
    })
}
test('fresh pre-reservation rejection releases UI but identical rejection of retry reconciles', async () => {
    assert.equal(classifySaveFailure(validation, false), 'definitive_rejection')
    assert.equal(classifySaveFailure(validation, true), 'reconcile')
    let stored = pending
    const outcome = await runNesting(
        async () => {
            throw validation
        },
        '/s',
        stored.request,
        assert.fail,
        assert.fail,
        stored.key,
        false,
    )
    if (outcome.status === 'rejected') stored = null
    assert.equal(stored, null)
    assert.equal(outcome.error, validation)
    assert.equal(
        classifySaveFailure(new ApiError(400, 'validation_error', 'other stage'), false),
        'reconcile',
    )
})
test('network/abort/timeout preserves key and never automatically sends a second POST', async () => {
    for (const failure of [
        new TypeError('network'),
        new DOMException('aborted', 'AbortError'),
        new DOMException('timeout', 'TimeoutError'),
    ]) {
        let posts = 0
        const stored = pending
        const outcome = await runNesting(
            async (_path, options) => {
                if (options?.method === 'POST') posts++
                throw failure
            },
            '/s',
            stored.request,
            assert.fail,
            assert.fail,
            stored.key,
            true,
        )
        assert.equal(outcome.status, 'unknown')
        assert.equal(stored.key, pending.key)
        assert.equal(posts, 1)
    }
})
test('completed via status remains confirmed when placements fetch fails', async () => {
    let confirmed = false
    let stored = pending
    const outcome = await runNesting(
        async (path, options) => {
            if (options?.method === 'POST') throw new ApiError(409, 'conflict_error')
            if (path.endsWith('/placements')) throw new TypeError('offline')
            return { status: 'completed', result }
        },
        '/s',
        pending.request,
        () => {
            confirmed = true
            stored = { ...stored, confirmed: true }
        },
        assert.fail,
        pending.key,
        true,
    )
    assert.equal(outcome.status, 'confirmed')
    assert.equal(confirmed, true)
    assert.equal(stored.confirmed, true)
    assert.equal(stored.key, pending.key)
})
test('legacy save scope survives storage; preview and foreign scopes cannot apply save results', () => {
    for (const ids of [
        [2, 2, 3],
        [1, 3, 3],
        [1, 2, 4],
    ])
        assert.equal(isPendingSaveScope(pending, ...ids), false)
    assert.equal(isPendingSaveScope({ ...pending, kind: 'preview' }, 1, 2, 3), false)

    const old = Object.getOwnPropertyDescriptor(globalThis, 'sessionStorage')
    try {
        Object.defineProperty(globalThis, 'sessionStorage', {
            configurable: true,
            value: { getItem: () => JSON.stringify(pending) },
        })
        assert.deepEqual(readPendingNesting('k', 1, 2, 3), pending)
        assert.equal(readPendingNesting('k', 2, 2, 3), null)
    } finally {
        if (old) Object.defineProperty(globalThis, 'sessionStorage', old)
        else Reflect.deleteProperty(globalThis, 'sessionStorage')
    }
})
