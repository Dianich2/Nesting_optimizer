import test from 'node:test'
import assert from 'node:assert/strict'
import {
    parseDraftRotation,
    draftGhostTransform,
    draftKeyAction,
} from '../src/features/nesting/placementDraft.ts'
const active = { adding: true, blocked: false, editable: false, modal: false }
test('draft normalization and invalid raw input', () => {
    for (const [raw, want] of [
        ['0', 0],
        ['-90', 270],
        ['450', 90],
        ['720', 0],
        ['27.5', 27.5],
        ['', null],
        [' ', null],
        ['NaN', null],
        ['Infinity', null],
    ])
        assert.equal(parseDraftRotation(raw), want)
})
test('R Shift+R Esc and unsafe/editable contexts', () => {
    assert.equal(draftKeyAction('r', active), 90)
    assert.equal(draftKeyAction('R', { ...active, shift: true }), -90)
    assert.equal(draftKeyAction('Escape', active), 'cancel')
    for (const flag of ['editable', 'modal', 'ctrl', 'alt', 'meta']) {
        assert.equal(draftKeyAction('R', { ...active, [flag]: true }), null)
        assert.equal(draftKeyAction('Escape', { ...active, [flag]: true }), null)
    }
    assert.equal(draftKeyAction('r', { ...active, blocked: true }), null)
    assert.equal(draftKeyAction('Escape', { ...active, blocked: true }), 'cancel')
    assert.equal(draftKeyAction('r', { ...active, adding: false }), null)
})
test('ghost uses original bbox centre before rotating and positioning', () =>
    assert.equal(
        draftGhostTransform({ x: 250, y: 320 }, { x: 25, y: 60 }, 90),
        'translate(25 60) rotate(90) translate(-250 -320)',
    ))
