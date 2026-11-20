import assert from 'node:assert/strict'
import test from 'node:test'
import { ApiError } from '../src/shared/api/http.ts'
import {
    classifyPreviewFailure,
    mayClosePreview,
    parseNestingAngles,
    previewStatusOutcome,
    validateNestingRequest,
} from '../src/features/nesting/previewRecovery.ts'

const current = {
    kind: 'preview',
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
for (const [http, code, expected] of [
    [400, 'validation_error', 'rejected'],
    [401, 'unauthorized_error', 'rejected'],
    [403, 'forbidden_error', 'rejected'],
    [404, 'not_found_error', 'rejected'],
    [409, 'conflict_error', 'check_status'],
    [409, 'nesting_stale', 'rejected'],
    [409, 'nesting_cancelled', 'rejected'],
    [408, 'nesting_timeout', 'rejected'],
    [404, 'preview_expired', 'rejected'],
    [503, 'nesting_overloaded', 'rejected'],
    [503, 'nesting_pending', 'uncertain'],
    [500, 'internal_error', 'uncertain'],
    [502, undefined, 'uncertain'],
    [503, undefined, 'uncertain'],
    [504, undefined, 'uncertain'],
    [400, undefined, 'uncertain'],
    [408, undefined, 'uncertain'],
    [500, 'validation_error', 'uncertain'],
])
    test(`preview HTTP ${http}/${code} → ${expected}`, () =>
        assert.equal(classifyPreviewFailure(new ApiError(http, code)), expected))

test('authentication rejection of a retry must reconcile the earlier request', () => {
    assert.equal(
        classifyPreviewFailure(new ApiError(401, 'unauthorized_error'), true),
        'check_status',
    )
    assert.equal(classifyPreviewFailure(new ApiError(403, 'forbidden_error'), true), 'check_status')
})
test('network errors preserve original pending and never allocate a new UUID', () => {
    for (const error of [
        new TypeError('network'),
        new DOMException('aborted', 'AbortError'),
        new DOMException('timeout', 'TimeoutError'),
    ]) {
        assert.equal(classifyPreviewFailure(error), 'uncertain')
        const outcome = previewStatusOutcome(current, { status: 'unknown' })
        assert.equal(outcome.kind, 'unresolved')
        assert.equal(current.key, '11111111-1111-4111-8111-111111111111')
    }
})
test('completed restores original parameters, preview ID and mode; failed is terminal', () => {
    const result = { preview_id: 'ready', placements: [], metrics: {} }
    const outcome = previewStatusOutcome(current, { status: 'completed', result })
    assert.equal(outcome.kind, 'completed')
    assert.equal(outcome.proposal.request, current.request)
    assert.equal(outcome.proposal.result, result)
    assert.equal(outcome.proposal.includeSaved, true)
    assert.equal(
        previewStatusOutcome(current, {
            status: 'failed',
            failure: { code: 'validation_error', message: 'bad' },
        }).kind,
        'failed',
    )
    for (const status of ['unknown', 'running'])
        assert.equal(previewStatusOutcome(current, { status }).kind, 'unresolved')
    assert.equal(
        previewStatusOutcome(current, { status: 'completed', result: {} }).kind,
        'unresolved',
    )
    assert.equal(
        previewStatusOutcome({ ...current, kind: undefined }, { status: 'completed', result }).kind,
        'unresolved',
    )
})
test('closing a preview is scope-bound and never permits closing save pending', () => {
    assert.equal(mayClosePreview(current, 1, 2, 3), true)
    for (const ids of [
        [2, 2, 3],
        [1, 3, 3],
        [1, 2, 4],
    ])
        assert.equal(mayClosePreview(current, ...ids), false)
    assert.equal(mayClosePreview({ ...current, kind: undefined }, 1, 2, 3), false)
})
for (const count of [499, 500, 501])
    test(`quantity boundary ${count}`, () => {
        const input = {
            ...current.request,
            patterns: [
                { project_pattern_id: 1, quantity: count - 1 },
                { project_pattern_id: 2, quantity: 1 },
            ],
        }
        assert.equal(
            validateNestingRequest(input),
            count > 500 ? 'За один запуск можно разместить не более 500 деталей.' : null,
        )
    })
for (const count of [35, 36, 37])
    test(`rotation boundary ${count}`, () => {
        assert.equal(
            validateNestingRequest({
                ...current.request,
                allowed_rotations: Array.from({ length: count }, (_, i) => i),
            }),
            count > 36 ? 'Укажите не более 36 допустимых углов.' : null,
        )
    })
test('empty tokens, NaN, fractional and negative counts rejected; duplicate angles allowed', () => {
    for (const raw of ['', ',', '0,', '0,,90', 'Infinity', '360', '-1', 'abc'])
        assert.ok(
            validateNestingRequest({
                ...current.request,
                allowed_rotations: parseNestingAngles(raw),
            }),
        )
    for (const quantity of [0, -1, 0.5, NaN])
        assert.ok(
            validateNestingRequest({
                ...current.request,
                patterns: [{ project_pattern_id: 1, quantity }],
            }),
        )
    assert.ok(validateNestingRequest({ ...current.request, patterns: [] }))
    assert.equal(
        validateNestingRequest({ ...current.request, allowed_rotations: [0, 0, 90] }),
        null,
    )
})
