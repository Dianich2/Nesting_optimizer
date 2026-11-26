import { test, expect } from '@playwright/test'
import { session, dataset, rect, sqlStats, poses } from './fixtures.mjs'

test('stage20 real two-part direct saves and zero result', async ({ page }) => {
    test.setTimeout(180000)
    const { call } = await session(page, 'twoparts')
    for (const occupied of [false, true])
        for (const algorithm of ['baseline', 'nfp_greedy']) {
            const d = await dataset(call, `Two parts ${algorithm}`, [rect(40, 30)], rect(300, 200))
            if (occupied)
                await call(
                    'POST',
                    d.base + '/placements',
                    { project_pattern_id: d.patterns[0].id, x: 20, y: 15, rotation: 0 },
                    201,
                )
            await page.goto(d.base + '/nesting')
            await page.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
            await page
                .getByRole('combobox', { name: 'Алгоритм', exact: true })
                .selectOption(algorithm)
            await page.getByRole('spinbutton', { name: 'Part 0', exact: true }).fill('2')
            await page.getByRole('textbox', { name: 'Углы через запятую' }).fill('0,90')
            const before = await sqlStats(page, d.surface.id)
            await expect(page.getByRole('button', { name: 'Предпросмотр', exact: true })).toHaveCount(0)
            const save = page.waitForResponse(
                (r) => r.url().endsWith('/nesting') && r.request().method() === 'POST',
            )
            await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
            const sr = await save
            expect(sr.status()).toBe(200)
            expect(sr.request().postDataJSON()).not.toHaveProperty('preview_id')
            const saved = await sr.json()
            expect(saved.metrics).toMatchObject({ placed_count: 2, requested_count: 2 })
            expect(saved.placements).toHaveLength(2)
            await expect(
                page.getByText('Сохранено · размещено 2 из 2', { exact: true }),
            ).toBeVisible()
            expect(await sqlStats(page, d.surface.id)).toMatchObject({
                placements: before.placements + 2,
                nesting_runs: 1,
                optimize_calls: before.optimize_calls + 1,
            })
            const persisted = (await call('GET', d.base + '/placements')).items
            const savedIds = new Set(saved.placements.map((item) => item.id))
            expect(poses(persisted.filter((item) => savedIds.has(item.id)))).toEqual(
                poses(saved.placements),
            )
            const direct = page.waitForResponse(
                (r) => r.url().endsWith('/nesting') && r.request().method() === 'POST',
            )
            await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
            expect((await direct).status()).toBe(200)
            await expect(page.getByText('Сохранено · размещено 2 из 2', { exact: true })).toBeVisible()
            await expect.poll(async () => (await sqlStats(page, d.surface.id)).nesting_runs).toBe(2)
            expect(await sqlStats(page, d.surface.id)).toMatchObject({
                placements: before.placements + 4,
                nesting_runs: 2,
                optimize_calls: before.optimize_calls + 2,
            })
            expect((await call('GET', d.base + '/placements')).items).toHaveLength(
                before.placements + 4,
            )
        }
    const d = await dataset(call, 'Zero is success', [rect(40, 30)], rect(10, 10))
    await page.goto(d.base + '/nesting')
    await page.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
    await page.getByRole('spinbutton', { name: 'Part 0', exact: true }).fill('2')
    await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
    await expect(page.getByText('Сохранено · размещено 0 из 2', { exact: true })).toBeVisible()
    await expect(
        page.getByText(/Расчёт завершён успешно, но подходящих свободных позиций/),
    ).toBeVisible()
    expect(await sqlStats(page, d.surface.id)).toMatchObject({ placements: 0, nesting_runs: 1 })
})

test('pattern geometry dialog opens repeatedly and never writes nesting data', async ({ page }) => {
    const { call } = await session(page, 'patternview')
    const d = await dataset(call, 'Pattern preview is read only', [rect(40, 30)])
    await page.goto(d.base + '/nesting')
    const card = page.locator('.nesting-pattern').filter({ hasText: 'Part 0' })
    const thumbnail = card.locator('svg.nesting-pattern-thumbnail path')
    await expect(thumbnail).toBeVisible()
    const path = await thumbnail.getAttribute('d')
    expect(path).toBeTruthy()
    const before = await sqlStats(page, d.surface.id)
    const open = card.getByRole('button', { name: 'Предпросмотр лекала «Part 0»', exact: true })
    const dialog = page.getByRole('dialog', { name: 'Part 0', exact: true })
    await open.click()
    await expect(dialog).toBeVisible()
    await expect(dialog.locator('svg.nesting-pattern-large-svg path')).toHaveAttribute('d', path)
    await expect(dialog.getByText('40 ед.', { exact: true })).toBeVisible()
    await expect(dialog.getByText('30 ед.', { exact: true })).toBeVisible()
    await expect(dialog.locator('dt').filter({ hasText: /^Площадь$/ })).toBeVisible()
    await page.keyboard.press('Escape')
    await expect(dialog).toBeHidden()
    await open.click()
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: 'Закрыть', exact: true }).click()
    await expect(dialog).toBeHidden()
    expect(await sqlStats(page, d.surface.id)).toEqual(before)
})
