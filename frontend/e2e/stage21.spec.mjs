import { test, expect } from '@playwright/test'
import { session, dataset, rect, polygon, sqlStats } from './fixtures.mjs'
const ghost = (p) => p.locator('[data-placement-ghost="true"]')
async function world(page, x, y, click = false) {
    const p = await page.getByLabel('Рабочая поверхность раскроя', { exact: true }).evaluate(
        (s, p) => {
            const q = new DOMPoint(p.x, p.y).matrixTransform(s.getScreenCTM())
            return { x: q.x, y: q.y }
        },
        { x, y },
    )
    await page.mouse.move(p.x, p.y)
    if (click) await page.mouse.click(p.x, p.y)
}
test('stage21 rotated draft real409 exact pose undo redo reload', async ({ page }) => {
    const { call } = await session(page, 'draftpose')
    const d = await dataset(
        call,
        'Narrow',
        [polygon(200, 300, 300, 300, 300, 340, 200, 340)],
        rect(50, 120),
    )
    await page.goto(d.base + '/nesting')
    await page.getByRole('button', { name: /Добавить на поверхность.*Part 0/ }).click()
    await world(page, 25, 60)
    await expect(ghost(page)).toBeVisible()
    let waiting = page.waitForResponse(
        (r) => r.url().endsWith('/placements') && r.request().method() === 'POST',
    )
    await world(page, 25, 60, true)
    expect((await waiting).status()).toBe(409)
    expect((await sqlStats(page, d.surface.id)).placements).toBe(0)
    await expect(
        page.getByText('Лекало выходит за границу поверхности.', { exact: true }),
    ).toBeVisible()
    await expect(ghost(page)).toBeVisible()
    await page.getByLabel('Рабочая поверхность раскроя', { exact: true }).focus()
    await page.keyboard.press('r')
    await expect(page.getByRole('spinbutton', { name: 'Поворот до постановки, °' })).toHaveValue(
        '90',
    )
    await world(page, 25, 60)
    const expected = await ghost(page).evaluate((p, points) => {
        const m = p.transform.baseVal.consolidate().matrix
        return points.map((p) => {
            const q = new DOMPoint(p.x, p.y).matrixTransform(m)
            return { x: q.x, y: q.y }
        })
    }, d.patterns[0].geometry.exterior.points)
    waiting = page.waitForResponse(
        (r) => r.url().endsWith('/placements') && r.request().method() === 'POST',
    )
    await world(page, 25, 60, true)
    const response = await waiting
    expect(response.status()).toBe(201)
    const posted = response.request().postDataJSON()
    expect(posted.rotation).toBe(90)
    expect(posted.x).toBeCloseTo(25, 3)
    expect(posted.y).toBeCloseTo(60, 3)
    const item = await response.json()
    expect(item).toMatchObject({ x: posted.x, y: posted.y, rotation: 90 })
    item.geometry.exterior.points.forEach((p, i) => {
        expect(p.x).toBeCloseTo(expected[i].x, 4)
        expect(p.y).toBeCloseTo(expected[i].y, 4)
        const src = d.patterns[0].geometry.exterior.points
        const cx = (Math.min(...src.map((p) => p.x)) + Math.max(...src.map((p) => p.x))) / 2,
            cy = (Math.min(...src.map((p) => p.y)) + Math.max(...src.map((p) => p.y))) / 2
        expect(p.x).toBeCloseTo(posted.x - (src[i].y - cy), 8)
        expect(p.y).toBeCloseTo(posted.y + (src[i].x - cx), 8)
    })
    await expect(ghost(page)).toHaveCount(0)
    await page.getByText('Ещё', { exact: true }).click()
    await page.getByRole('button', { name: 'Отменить действие', exact: true }).click()
    await expect.poll(async () => (await sqlStats(page, d.surface.id)).placements).toBe(0)
    await page.getByRole('button', { name: 'Повторить действие', exact: true }).click()
    await expect.poll(async () => (await sqlStats(page, d.surface.id)).placements).toBe(1)
    await page.reload()
    await expect(page.getByLabel('Рабочая поверхность раскроя', { exact: true })).toBeVisible()
    const rows = (await call('GET', d.base + '/placements')).items
    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject({ x: posted.x, y: posted.y, rotation: 90 })
})
test('stage21 cancel and pattern switch never POST', async ({ page }) => {
    const { call } = await session(page, 'draftcancel')
    const d = await dataset(call, 'Cancel', [rect(100, 40), rect(20, 20)], rect(300, 200))
    const another = await call(
        'POST',
        `/projects/${d.project.id}/surfaces`,
        { source_surface_id: d.original.id, scale: 1 },
        201,
    )
    await page.goto(d.base + '/nesting')
    let posts = 0
    page.on('request', (r) => {
        if (r.method() === 'POST' && r.url().endsWith('/placements')) posts++
    })
    const first = page.getByRole('button', { name: /Добавить на поверхность.*Part 0/ }),
        second = page.getByRole('button', { name: /Добавить на поверхность.*Part 1/ }),
        canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true }),
        angle = page.getByRole('spinbutton', { name: 'Поворот до постановки, °' })
    await first.click()
    await world(page, 150, 100)
    await canvas.focus()
    await page.keyboard.press('Shift+R')
    await expect(angle).toHaveValue('270')
    await first.click()
    await expect(page.getByRole('toolbar', { name: 'Постановка лекала' })).toHaveCount(0)
    await first.click()
    await expect(angle).toHaveValue('0')
    await angle.fill('45')
    await second.click()
    await expect(angle).toHaveValue('0')
    await canvas.focus()
    await page.keyboard.press('Escape')
    await expect(ghost(page)).toHaveCount(0)
    await first.click()
    await angle.fill('')
    await expect(angle).toHaveValue('')
    await angle.press('r')
    await expect(angle).not.toHaveValue('90')
    await page.getByRole('button', { name: 'Отменить добавление', exact: true }).click()
    await expect(page.getByRole('toolbar', { name: 'Постановка лекала' })).toHaveCount(0)
    await first.click()
    await page
        .getByRole('combobox', { name: 'Поверхность', exact: true })
        .selectOption(String(another.id))
    await expect(page.getByRole('toolbar', { name: 'Постановка лекала' })).toHaveCount(0)
    expect(posts).toBe(0)
    expect(await sqlStats(page, d.surface.id)).toMatchObject({ placements: 0, nesting_runs: 0 })
})
