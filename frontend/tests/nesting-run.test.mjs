import assert from 'node:assert/strict'
import test from 'node:test'
import { runNesting } from '../src/features/nesting/runNesting.ts'
import { ApiError } from '../src/shared/api/http.ts'

for (const keep_existing of [true, false]) {
    for (const syncFails of [true, false]) {
        test(`confirmed POST persists through sync (keep=${keep_existing}, failed=${syncFails})`, async () => {
            const result = { placements: [{ id: 8 }] }
            const events = []
            const request = async (path, options) => {
                events.push(path)
                if (options?.method === 'POST') {
                    assert.equal(JSON.parse(options.body).keep_existing, keep_existing)
                    return result
                }
                if (syncFails) throw new TypeError('network')
                return { items: result.placements }
            }
            const status = await runNesting(
                request,
                '/surface',
                { keep_existing },
                (value) => {
                    assert.equal(value, result)
                    events.push('confirmed')
                },
                () => events.push('synchronized'),
            )
            assert.equal(status, syncFails ? 'confirmed' : 'synchronized')
            assert.deepEqual(events, [
                '/surface/nesting',
                'confirmed',
                '/surface/placements',
                ...(!syncFails ? ['synchronized'] : []),
            ])
        })
    }
}

test('uncertain POST is sent once and requires reconciliation', async () => {
    let calls = 0
    const request = async () => {
        calls++
        throw new TypeError('network')
    }
    assert.equal(await runNesting(request, '/surface', {}, assert.fail, assert.fail), 'unknown')
    assert.equal(calls, 1)
})

test('confirmed validation error is propagated', async () => {
    const error = new ApiError(409, 'conflict_error')
    const request = async () => {
        throw error
    }
    await assert.rejects(runNesting(request, '/surface', {}, assert.fail, assert.fail), error)
})

test('a timed out calculation requires reconciliation and is sent once', async (t) => {
    t.mock.method(AbortSignal, 'timeout', () =>
        AbortSignal.abort(new DOMException('Request timed out', 'TimeoutError')),
    )
    let calls = 0
    const request = async (_path, options) => {
        calls++
        options.signal.throwIfAborted()
    }
    assert.equal(await runNesting(request, '/surface', {}, assert.fail, assert.fail), 'unknown')
    assert.equal(calls, 1)
})

test('a timed out synchronization preserves the confirmed saved result', async (t) => {
    t.mock.method(AbortSignal, 'timeout', () =>
        AbortSignal.abort(new DOMException('Request timed out', 'TimeoutError')),
    )
    const result = { placements: [{ id: 8 }] }
    let confirmed = false
    const request = async (_path, options) => {
        if (options.method === 'POST') return result
        options.signal.throwIfAborted()
    }
    const status = await runNesting(
        request,
        '/surface',
        {},
        (value) => {
            assert.equal(value, result)
            confirmed = true
        },
        assert.fail,
    )
    assert.equal(status, 'confirmed')
    assert.equal(confirmed, true)
})

for (const [status, code, terminal] of [
    [408, undefined, false],
    [408, 'nesting_timeout', true],
    [409, 'nesting_stale', true],
    [429, 'rate_limit', true],
    [503, undefined, false],
    [503, 'nesting_pending', false],
    [503, 'nesting_overloaded', true],
    [500, 'internal_error', false],
]) {
    test(`HTTP ${status} / ${code} has explicit save recovery semantics`, async () => {
        const failure = new ApiError(status, code)
        const request = async () => {
            throw failure
        }
        const result = runNesting(request, '/surface', {}, assert.fail, assert.fail)
        if (terminal) await assert.rejects(result, failure)
        else assert.equal(await result, 'unknown')
    })
}
