import assert from 'node:assert/strict'
import test from 'node:test'
import { parseNestingAngles, validateNestingRequest } from '../src/features/nesting/nestingValidation.ts'

const request = {
    algorithm: 'baseline',
    patterns: [{ project_pattern_id: 1, quantity: 1 }],
    allowed_rotations: [0],
    keep_existing: true,
}
for (const count of [499, 500, 501])
    test(`quantity boundary ${count}`, () => {
        const input = {
            ...request,
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
                ...request,
                allowed_rotations: Array.from({ length: count }, (_, i) => i),
            }),
            count > 36 ? 'Укажите не более 36 допустимых углов.' : null,
        )
    })
test('empty tokens, NaN, fractional and negative counts rejected; duplicate angles allowed', () => {
    for (const raw of ['', ',', '0,', '0,,90', 'Infinity', '360', '-1', 'abc'])
        assert.ok(
            validateNestingRequest({
                ...request,
                allowed_rotations: parseNestingAngles(raw),
            }),
        )
    for (const quantity of [0, -1, 0.5, NaN])
        assert.ok(
            validateNestingRequest({
                ...request,
                patterns: [{ project_pattern_id: 1, quantity }],
            }),
        )
    assert.ok(validateNestingRequest({ ...request, patterns: [] }))
    assert.equal(
        validateNestingRequest({ ...request, allowed_rotations: [0, 0, 90] }),
        null,
    )
})
