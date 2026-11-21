import { test, expect } from '@playwright/test'
import { session, dataset, rect, sqlStats, poses } from './fixtures.mjs'

test('stage20 real two-part preview exact save direct save and zero result', async ({ page }) => {
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
            const response = page.waitForResponse(
                (r) => r.url().endsWith('/nesting/preview') && r.request().method() === 'POST',
            )
            await page.getByRole('button', { name: 'Предпросмотр', exact: true }).click()
            const p = await response
            expect(p.status()).toBe(200)
            const preview = await p.json()
            expect(preview.metrics).toMatchObject({ placed_count: 2, requested_count: 2 })
            expect(await sqlStats(page, d.surface.id)).toMatchObject({
                placements: before.placements,
                nesting_runs: before.nesting_runs,
            })
            await expect(page.getByText(/Предпросмотр · размещено 2 из 2/)).toBeVisible()
            const save = page.waitForResponse(
                (r) => r.url().endsWith('/nesting') && r.request().method() === 'POST',
            )
            await page
                .getByRole('button', { name: 'Сохранить этот предпросмотр', exact: true })
                .click()
            const sr = await save
            expect(sr.status()).toBe(200)
            expect(sr.request().postDataJSON().preview_id).toBe(preview.preview_id)
            expect(poses((await sr.json()).placements)).toEqual(poses(preview.placements))
            await expect(
                page.getByText('Сохранено · размещено 2 из 2', { exact: true }),
            ).toBeVisible()
            expect(await sqlStats(page, d.surface.id)).toMatchObject({
                placements: before.placements + 2,
                nesting_runs: 1,
                optimize_calls: before.optimize_calls + 1,
            })
            const direct = page.waitForResponse(
                (r) => r.url().endsWith('/nesting') && r.request().method() === 'POST',
            )
            await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
            expect((await direct).status()).toBe(200)
            await expect.poll(async () => (await sqlStats(page, d.surface.id)).nesting_runs).toBe(2)
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
