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
    // The API must still reject invalid geometry, even though the UI now
    // keeps invalid manual placements locally without making a POST.
    const before = await sqlStats(page, d.surface.id)
    const rejected = await call('POST', d.base + '/placements', {
        project_pattern_id: d.patterns[0].id, x: 25, y: 60, rotation: 0,
    }, 409)
    expect(rejected.error.details).toEqual(expect.arrayContaining([
        expect.objectContaining({ field: 'placement', code: 'placement_outside_surface' }),
    ]))
    expect(await sqlStats(page, d.surface.id)).toEqual(before)

    await page.goto(d.base + '/nesting')
    const canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true })
    const angle = page.getByLabel('Угол поворота до постановки', { exact: true })
    let posts = 0
    page.on('request', (r) => {
        if (r.method() === 'POST' && r.url().endsWith('/placements')) posts++
    })
    await page.getByRole('button', { name: /Добавить на поверхность.*Part 0/ }).click()
    await world(page, 25, 60)
    await expect(ghost(page)).toHaveAttribute('fill', '#dc2626')
    await canvas.focus()
    await page.keyboard.press('r')
    await expect(angle).toHaveText('Угол: 90°')
    await world(page, 25, 60)
    await expect(ghost(page)).toHaveAttribute('fill', '#6da5ee')
    const expected = await ghost(page).evaluate((p, points) => {
        const m = p.transform.baseVal.consolidate().matrix
        return points.map((p) => {
            const q = new DOMPoint(p.x, p.y).matrixTransform(m)
            return { x: q.x, y: q.y }
        })
    }, d.patterns[0].geometry.exterior.points)
    await page.keyboard.press('Shift+R')
    await expect(angle).toHaveText('Угол: 0°')
    await world(page, 25, 60, true)
    const draft = canvas.getByRole('button', { name: 'Размещение №-1', exact: true })
    await expect(draft).toHaveAttribute('fill', '#dc2626')
    await expect(page.locator('.nesting-status')).toContainText('Деталь выходит за границу поверхности.')
    await expect(ghost(page)).toHaveCount(0)
    expect(posts).toBe(0)
    expect(await sqlStats(page, d.surface.id)).toEqual(before)

    // Correct the actual unsaved draft; the preset applies immediately.
    const waiting = page.waitForResponse(
        (r) => r.url().endsWith('/placements') && r.request().method() === 'POST',
    )
    await page.getByRole('button', { name: '90°', exact: true }).click()
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
    await expect(draft).toHaveCount(0)
    await expect(canvas.getByRole('button', { name: `Размещение №${item.id}`, exact: true })).toBeVisible()
    expect(posts).toBe(1)
    expect((await sqlStats(page, d.surface.id)).placements).toBe(1)
    await page.getByText('Ещё', { exact: true }).click()
    await page.getByRole('button', { name: 'Отменить действие', exact: true }).click()
    await expect.poll(async () => (await sqlStats(page, d.surface.id)).placements).toBe(0)
    await page.getByRole('button', { name: 'Повторить действие', exact: true }).click()
    await expect.poll(async () => (await sqlStats(page, d.surface.id)).placements).toBe(1)
    await page.reload()
    await expect(canvas.getByRole('button', { name: /Размещение №/ })).toHaveCount(1)
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
        angle = page.getByLabel('Угол поворота до постановки', { exact: true }),
        controls = page.getByRole('status', { name: 'Постановка лекала', exact: true })
    await first.click()
    await world(page, 150, 100)
    await canvas.focus()
    await page.keyboard.press('Shift+R')
    await expect(angle).toHaveText('Угол: 270°')
    await first.click()
    await expect(controls).toHaveCount(0)
    await first.click()
    await expect(angle).toHaveText('Угол: 0°')
    await canvas.focus()
    await page.keyboard.press('Shift+ArrowRight')
    await page.keyboard.press('ArrowRight')
    await expect(angle).toHaveText('Угол: 16°')
    await page.keyboard.press('ArrowLeft')
    await expect(angle).toHaveText('Угол: 15°')
    await second.click()
    await expect(angle).toHaveText('Угол: 0°')
    await canvas.focus()
    await page.keyboard.press('Escape')
    await expect(ghost(page)).toHaveCount(0)
    await first.click()
    // Hotkeys must not rotate a draft while editing another numeric field.
    await page.getByText('Настройки сетки', { exact: true }).click()
    const step = page.getByRole('spinbutton', { name: 'Шаг', exact: true })
    await step.fill('')
    await expect(step).toHaveValue('')
    await step.press('r')
    await expect(angle).toHaveText('Угол: 0°')
    await step.fill('10')
    await page.getByText('Настройки сетки', { exact: true }).click()
    await page.getByRole('button', { name: 'Отменить добавление', exact: true }).click()
    await expect(controls).toHaveCount(0)
    await first.click()
    await page
        .getByRole('combobox', { name: 'Поверхность', exact: true })
        .selectOption(String(another.id))
    await expect(controls).toHaveCount(0)
    await expect(canvas).toBeVisible()
    expect(posts).toBe(0)
    expect(await sqlStats(page, d.surface.id)).toMatchObject({ placements: 0, nesting_runs: 0 })
    expect(await sqlStats(page, another.id)).toMatchObject({ placements: 0, nesting_runs: 0 })
})
