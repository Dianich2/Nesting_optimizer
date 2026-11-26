import test from 'node:test'
import assert from 'node:assert/strict'
import { transformPattern, placementIssue } from '../src/features/nesting/manualPlacement.ts'
const poly = (xmin, ymin, xmax, ymax, holes=[]) => ({exterior:{points:[{x:xmin,y:ymin},{x:xmax,y:ymin},{x:xmax,y:ymax},{x:xmin,y:ymax}]},holes})
const surface=poly(0,0,100,100)
const pattern=poly(0,0,10,10)
test('translation and rotation around original bbox centre',()=>{
  let p=transformPattern(pattern,{x:30,y:40,rotation:90})
  assert.ok(Math.abs(p.exterior.points[0].x-35)<1e-8)
  assert.ok(Math.abs(p.exterior.points[0].y-35)<1e-8)
})
test('valid placement, bounds and touch permitted',()=>{
  const a=transformPattern(pattern,{x:30,y:30,rotation:0})
  const b=transformPattern(pattern,{x:40,y:30,rotation:0})
  assert.equal(placementIssue(surface,a,[]),null)
  assert.equal(placementIssue(surface,b,[a]),null)
  assert.match(placementIssue(surface,transformPattern(pattern,{x:99,y:30,rotation:0}),[]),/границу/)
})
test('overlap, nesting and identical locations prohibited',()=>{
  const a=transformPattern(pattern,{x:30,y:30,rotation:0})
  assert.match(placementIssue(surface,a,[a]),/пересекает другое/)
  assert.match(placementIssue(surface,transformPattern(pattern,{x:34,y:30,rotation:0}),[a]),/пересекает другое/)
  assert.match(placementIssue(surface,poly(20,20,40,40),[a]),/пересекает другое/)
})
test('hole and enclosing hole prohibited',()=>{
  const h=poly(0,0,100,100,[poly(40,40,60,60).exterior])
  assert.match(placementIssue(h,poly(35,35,65,65),[]),/отверстие/)
  assert.match(placementIssue(h,poly(45,45,55,55),[]),/отверстие/)
})

test('rotated pattern outside bounds is flagged', () => {
    const tall = poly(0,0,8,25)
    assert.match(placementIssue(surface, transformPattern(tall,{x:95,y:95,rotation:45}),[]), /границу/)
})
test('concave surface does not accept bridge across exterior notch', () => {
    const concave = {exterior:{points:[{x:0,y:0},{x:100,y:0},{x:100,y:35},{x:35,y:35},{x:35,y:100},{x:0,y:100}]},holes:[]}
    assert.match(placementIssue(concave,poly(20,20,48,48),[]), /границу|область/)
})
