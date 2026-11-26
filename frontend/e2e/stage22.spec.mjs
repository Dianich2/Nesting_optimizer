import { test, expect } from '@playwright/test'
import { installDiagnostics } from './diagnostics.mjs'
installDiagnostics(test)
import { session, dataset, rect, sqlStats } from './fixtures.mjs'

async function replace(input, text) {
    await input.focus()
    await input.press('ControlOrMeta+A')
    await input.press('Backspace')
    await expect(input).toHaveValue('')
    await input.pressSequentially(text)
    await expect(input).toHaveValue(text)
}

test('stage22 editable numeric drafts reject invalid counts before POST', async ({ page }) => {
    const { call } = await session(page, 'numeric')
    const d = await dataset(call, 'Numeric drafts', [rect(10, 10), rect(20, 10)], rect(300, 200))
    await page.goto(d.base + '/nesting')
    await page.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
    const quantity = page.getByRole('spinbutton', { name: 'Part 0', exact: true })
    let posts = 0
    page.on('request', (r) => {
        if (r.method() === 'POST' && /\/nesting(?:\/preview)?$/.test(r.url())) posts++
    })
    for (const raw of ['-1', '1.5', '501']) {
        await replace(quantity, raw)
        await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
        await expect(
            page.getByText('Количество: укажите целое число от 0 до 500. 0 — не выбирать лекало.', {
                exact: true,
            }),
        ).toBeVisible()
        expect(posts).toBe(0)
    }
    await replace(quantity, '500')
    await page.getByRole('spinbutton', { name: 'Part 1', exact: true }).fill('1')
    await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
    await expect(page.getByText(/Суммарно за один запуск/)).toBeVisible()
    expect(posts).toBe(0)
    await page.getByRole('spinbutton', { name: 'Part 1', exact: true }).fill('0')
    await replace(quantity, '2')
    const response = page.waitForResponse(
        (r) => r.request().method() === 'POST' && r.url().endsWith('/nesting'),
    )
    await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
    const saved = await response
    expect(saved.status()).toBe(200)
    expect(saved.request().postDataJSON().patterns).toEqual([
        { project_pattern_id: d.patterns[0].id, quantity: 2 },
    ])
    expect(saved.request().postDataJSON()).not.toHaveProperty('preview_id')
    await expect.poll(async () => (await sqlStats(page, d.surface.id)).placements).toBe(2)
    const placement = (await call('GET', d.base + '/placements')).items[0]
    await page
        .getByRole('button', { name: `Размещение №${placement.id}`, exact: true })
        .press('Enter')
    await page.getByRole('button', { name: 'Свойства', exact: true }).click()
    await replace(page.getByLabel('X центра', { exact: true }), String(placement.x))
    await replace(page.getByLabel('Y центра', { exact: true }), String(placement.y))
    await replace(page.getByLabel('Поворот, °', { exact: true }), '90')
    const updated = page.waitForResponse(
        (r) => r.request().method() === 'PATCH' && r.url().endsWith(`/placements/${placement.id}`),
    )
    await page.getByLabel('Рабочая поверхность раскроя', { exact: true }).focus()
    const patched = await updated
    expect(patched.status()).toBe(200)
    expect(patched.request().postDataJSON()).toMatchObject({
        x: placement.x, y: placement.y, rotation: 90,
    })
    expect(await patched.json()).toMatchObject({
        id: placement.id, x: placement.x, y: placement.y, rotation: 90,
    })
    await page.getByText('Настройки сетки', { exact: true }).click()
    await replace(page.getByRole('spinbutton', { name: 'Шаг', exact: true }), '5')
    await page.goto(`/projects/${d.project.id}/editor`)
    await page
        .locator('.project-resource-panel')
        .filter({ has: page.getByRole('heading', { name: 'Лекала', exact: true }) })
        .getByRole('button', { name: 'Добавить из библиотеки' })
        .click()
    await replace(page.getByRole('dialog').getByLabel('Масштаб геометрии'), '2')
    await page.getByRole('dialog').getByRole('button', { name: 'Закрыть', exact: true }).click()
    await page.goto('/library')
    await page
        .locator('.library-panel')
        .filter({ has: page.getByRole('heading', { name: 'Лекала', exact: true }) })
        .getByRole('button', { name: /Добавить/ })
        .click()
    const dialog = page.getByRole('dialog')
    await replace(dialog.getByLabel('Ширина', { exact: true }), '25')
    await replace(dialog.getByLabel('Высота', { exact: true }), '15')
    await replace(dialog.getByRole('spinbutton', { name: /Шаг/ }), '5')
    await dialog.getByRole('button', { name: '← Назад', exact: true }).click()
})

test('stage22 48 viewport theme zoom-equivalent layouts have spaced usable buttons', async ({
    page,
    browser,
}, testInfo) => {
    test.setTimeout(240000)
    const { call } = await session(page, 'layouts')
    const d = await dataset(call, 'Adaptive controls', [rect(10, 10), rect(20, 10)], rect(300, 200))
    for (const part of d.patterns)
        await call('PATCH', `/projects/${d.project.id}/patterns/${part.id}`, {
            name: 'Очень длинное название лекала LongEnglishPatternWithoutBreaks012345678901234567890123456789',
        })
    let storageState = await page.context().storageState()
    const measurements = []
    for (const width of [320, 390, 700, 768, 1024, 1366, 1440, 1920])
        for (const theme of ['light', 'dark'])
            for (const zoom of [1, 1.5, 2]) {
                // Desktop page zoom changes CSS viewport and DPR together. This is that
                // rendering equivalence, not automation of Chrome's native zoom menu.
                const context = await browser.newContext({
                    baseURL: testInfo.project.use.baseURL,
                    storageState,
                    viewport: { width: Math.floor(width / zoom), height: Math.floor(1000 / zoom) },
                    deviceScaleFactor: zoom,
                })
                await context.addInitScript(
                    (value) => localStorage.setItem('nesting-optimizer-theme', value),
                    theme,
                )
                const p = await context.newPage()
                await p.goto(d.base + '/nesting')
                await p.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
                const actions = p.locator('.nesting-auto-actions')
                const save = actions.getByRole('button', { name: 'Рассчитать и сохранить', exact: true })
                await expect(actions.getByRole('button')).toHaveCount(1)
                await save.scrollIntoViewIfNeeded()
                await expect(save).toBeEnabled()
                const result = await p.evaluate(() => {
                    const panel = document.querySelector('aside[aria-label="Свойства раскроя"]')
                    const buttons = [
                        ...document.querySelectorAll('.nesting-auto-actions > button'),
                    ].map((x) => {
                        const r = x.getBoundingClientRect()
                        return {
                            x: r.x,
                            y: r.y,
                            right: r.right,
                            bottom: r.bottom,
                            width: r.width,
                            height: r.height,
                        }
                    })
                    return {
                        actionsWidth: document.querySelector('.nesting-auto-actions').clientWidth,
                        buttons,
                        documentWidth: document.documentElement.scrollWidth,
                        viewportWidth: innerWidth,
                        panel: { width: panel.clientWidth, scrollWidth: panel.scrollWidth },
                        overflow: [
                            ...document.querySelectorAll(
                                '.nesting-page, .nesting-canvas, .nesting-workspace aside',
                            ),
                        ].map((x) => ({
                            class: x.className,
                            width: x.clientWidth,
                            scrollWidth: x.scrollWidth,
                        })),
                    }
                })
                measurements.push({ width, theme, zoom, ...result })
                // Only one automatic action remains. Check its full width,
                // minimum height, and actual clickability in every layout.
                expect(result.buttons).toHaveLength(1)
                expect(result.buttons[0].width).toBeGreaterThan(0)
                expect(result.buttons[0].width).toBeCloseTo(result.actionsWidth, 0)
                expect(result.buttons[0].height).toBeGreaterThanOrEqual(40)
                expect(
                    result.documentWidth,
                    `${width}/${theme}/${zoom} document`,
                ).toBeLessThanOrEqual(result.viewportWidth + 1)
                expect(result.panel.scrollWidth).toBeLessThanOrEqual(result.panel.width + 1)
                for (const box of result.overflow)
                    expect(
                        box.scrollWidth,
                        `${width}/${theme}/${zoom} ${box.class}`,
                    ).toBeLessThanOrEqual(box.width + 1)
                await save.click({ trial: true })
                await p.screenshot({
                    path: testInfo.outputPath(`layout-${width}-${theme}-${zoom}.png`),
                })
                storageState = await context.storageState()
                await context.close()
            }
    await testInfo.attach('layout-measurements.json', {
        body: JSON.stringify(measurements, null, 2),
        contentType: 'application/json',
    })
})
