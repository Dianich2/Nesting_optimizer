import { expect, test } from '@playwright/test'
import { installDiagnostics } from './diagnostics.mjs'
installDiagnostics(test)
import { fileURLToPath } from 'node:url'
import { readFile } from 'node:fs/promises'

const radial = fileURLToPath(new URL('./fixtures/radial-pattern.svg', import.meta.url))
const api = process.env.NESTING_API_TARGET + '/api/v1'
const poses = (items) => items.map(({ x, y, rotation, geometry }) => ({ x, y, rotation, geometry }))

for (const width of [1440, 1366, 768, 700, 390]) {
    for (const theme of ['light', 'dark']) {
        test(`radial direct save, exports and layout ${width} ${theme}`, async ({ page, request }, info) => {
            const errors = []
            const anonymousRefresh = []
            let signedIn = false
            page.on('pageerror', () => errors.push('unhandled page error'))
            page.on('console', (message) => {
                if (message.type() !== 'error') return
                const location = message.location().url
                if (!signedIn && location.endsWith('/api/v1/auth/browser/refresh') && message.text().includes('status of 400')) {
                    anonymousRefresh.push('expected missing-cookie 400 before login')
                    return
                }
                errors.push(`console error: ${message.text().replace(/[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]+/g, '[redacted]')} @ ${location}`)
            })
            page.on('response', (response) => { if (response.status() >= 500) errors.push(`HTTP ${response.status()}`) })
            let authorization
            page.on('request', (r) => { if (r.headers().authorization) authorization = r.headers().authorization })
            await page.setViewportSize({ width, height: 1000 })
            const login = `rad${width}${Date.now().toString(36)}`
            const password = 'BrowserOnlyPassword2026'
            await page.goto('/register')
            await page.locator('#register-login').fill('bad!')
            expect(await page.locator('#register-login').evaluate((input) => input.checkValidity())).toBe(false)
            await page.locator('#register-login').fill(login)
            expect(await page.locator('#register-login').evaluate((input) => input.checkValidity())).toBe(true)
            await page.locator('#register-email').fill(`${login}@example.test`)
            await page.locator('#register-first-name').fill('Browser')
            await page.locator('#register-last-name').fill('Test')
            await page.locator('#register-password').fill(password)
            await page.locator('#register-repeat-password').fill(password)
            await page.getByRole('button', { name: 'Зарегистрироваться' }).click()
            await expect(page).toHaveURL(/\/login/)
            await page.getByLabel('Логин или электронная почта').fill(login)
            await page.locator('#login-password').fill(password)
            await page.getByRole('button', { name: 'Войти', exact: true }).click()
            await expect(page).toHaveURL(/\/projects/)
            signedIn = true
            const toggle = page.getByRole('button', { name: theme === 'dark' ? 'Включить тёмную тему' : 'Включить светлую тему' })
            if (await toggle.count()) await toggle.click()
            await expect(page.locator('html')).toHaveAttribute('data-theme', theme)
            await page.goto('/library')
            const patternName = 'Радиальное вогнутое лекало с длинным проверяемым названием'
            for (const [section, name] of [['Поверхности', 'Лист bootstrap'], ['Лекала', patternName]]) {
                await page.locator('.library-panel').filter({ has: page.getByRole('heading', { name: section, exact: true }) }).getByRole('button', { name: /Добавить/ }).click()
                const dialog = page.getByRole('dialog')
                if (!(await dialog.getByLabel('Название', { exact: true }).isVisible())) {
                    await dialog.getByRole('button', { name: 'Свойства', exact: true }).click()
                }
                await dialog.getByLabel('Название', { exact: true }).fill(name)
                if (section === 'Лекала') await dialog.locator('input[type=file]').setInputFiles(radial)
                else {
                    await dialog.getByLabel('Ширина', { exact: true }).fill('100')
                    await dialog.getByLabel('Высота', { exact: true }).fill('100')
                }
                await dialog.getByRole('button', { name: 'Создать', exact: true }).click()
                await expect(dialog).toBeHidden()
            }
            await page.goto('/projects')
            await page.getByRole('button', { name: /Новый проект/ }).click()
            await page.getByPlaceholder('Название проекта').fill('Проект bootstrap')
            await page.getByRole('button', { name: 'Создать проект' }).click()
            await page.locator('article').filter({ hasText: 'Проект bootstrap' }).getByRole('link').first().click()
            for (const [kind, name] of [['Поверхности', 'Лист bootstrap'], ['Лекала', patternName]]) {
                await page.locator('.project-resource-panel').filter({ has: page.getByRole('heading', { name: kind, exact: true }) }).getByRole('button', { name: 'Добавить из библиотеки' }).click()
                const dialog = page.getByRole('dialog')
                await dialog.getByRole('listitem').filter({ hasText: name }).getByRole('button', { name: 'Добавить', exact: true }).click()
                await expect(dialog.getByRole('status')).toContainText('Добавлен экземпляр')
                await dialog.getByRole('button', { name: 'Закрыть', exact: true }).click()
            }
            // Read/edit project geometry through the actual constructor before use.
            const row = page.locator('li.project-resource-item').filter({ hasText: patternName })
            await row.getByRole('button', { name: 'Редактировать' }).click()
            const editor = page.getByRole('dialog')
            if (!(await editor.getByRole('combobox', { name: /Вершина контура/ }).isVisible())) {
                await editor.getByRole('button', { name: 'Свойства', exact: true }).click()
            }
            await editor.getByRole('combobox', { name: /Вершина контура/ }).selectOption('0')
            const x = editor.getByLabel('X точки 1', { exact: true })
            const originalX = Number(await x.inputValue())
            expect(Number.isFinite(originalX)).toBe(true)
            await x.fill(String(originalX + 0.25)); await x.press('Tab')
            await editor.getByRole('button', { name: 'Сохранить изменения' }).click()
            await expect(editor).toBeHidden()
            await page.getByRole('link', { name: 'Открыть раскрой' }).click()
            const canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true })
            const placement = canvas.getByRole('button', { name: /Размещение №/ })
            const base = page.url().match(/\/projects\/\d+\/surfaces\/\d+/)[0]
            const get = async (path) => {
                const response = await request.get(api + path, { headers: { Authorization: authorization } })
                expect(response.status()).toBe(200)
                return response.json()
            }
            const count = async () => (await (await request.get(process.env.NESTING_E2E_STATS)).json()).optimize_calls
            const addButton = page.getByRole('button', { name: new RegExp(`Добавить на поверхность.*${patternName}`) })
            const patternsToggle = page.getByRole('button', { name: 'Лекала', exact: true })
            await expect(patternsToggle).toBeVisible()
            if ((await patternsToggle.getAttribute('aria-expanded')) === 'false') await patternsToggle.click()
            await expect(addButton).toBeVisible()
            await addButton.click()
            const screen = await canvas.evaluate((svg) => { const p = new DOMPoint(50, 50).matrixTransform(svg.getScreenCTM()); return { x: p.x, y: p.y } })
            await page.mouse.click(screen.x, screen.y)
            await expect(placement).toHaveCount(1)
            const more = page.locator('.nesting-page summary').filter({ hasText: 'Ещё' })
            await more.click(); await page.getByRole('button', { name: 'Отменить действие' }).click()
            await expect(placement).toHaveCount(0)
            await page.getByRole('button', { name: 'Повторить действие' }).click()
            await expect(placement).toHaveCount(1)
            await page.getByRole('button', { name: 'Отменить действие' }).click()
            await expect(placement).toHaveCount(0)
            await more.click()
            expect((await get(base + '/placements')).items).toHaveLength(0)
            const before = await count()
            await page.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
            await page.getByRole('spinbutton', { name: patternName }).fill('1')
            const saving = page.waitForResponse((r) => r.url().endsWith('/nesting') && r.request().method() === 'POST')
            await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
            const savedResponse = await saving
            expect(savedResponse.status()).toBe(200)
            const savedBody = savedResponse.request().postDataJSON()
            const savedKey = savedResponse.request().headers()['idempotency-key']
            expect(savedBody).not.toHaveProperty('preview_id')
            expect(savedKey).toBeTruthy()
            const saved = await savedResponse.json()
            expect(saved.placements).toHaveLength(1)
            expect(saved.placements[0].id).toBeGreaterThan(0)
            expect(await count()).toBe(before + 1)
            await expect(placement).toHaveCount(1)
            await page.reload(); await expect(placement).toHaveCount(1)
            expect(poses((await get(base + '/placements')).items)).toEqual(poses(saved.placements))
            if (!(await page.getByRole('button', { name: 'История', exact: true }).isVisible())) {
                await page.getByRole('button', { name: 'Панель', exact: true }).click()
            }
            await page.getByRole('button', { name: 'История', exact: true }).click()
            await expect(page.getByRole('table').getByRole('row')).toHaveCount(2)
            for (const format of ['SVG', 'отчёта JSON']) {
                const download = page.waitForEvent('download')
                const exportButton = page.getByRole('button', { name: `Экспорт ${format}`, exact: true })
                if (!(await exportButton.isVisible())) await more.click()
                await exportButton.click()
                const file = await download
                const outputPath = info.outputPath(format === 'SVG' ? 'radial-layout.svg' : 'radial-layout.json')
                await file.saveAs(outputPath)
                const text = await readFile(outputPath, 'utf8')
                if (format === 'SVG') {
                    expect(text.match(/<path /g)).toHaveLength(2)
                    const points = saved.placements[0].geometry.exterior.points
                    for (const p of points) expect(text).toContain(`${p.x} ${p.y}`)
                } else {
                    const output = JSON.parse(text)
                    expect(poses(output.placements)).toEqual(poses(saved.placements))
                    expect(output.placements).toHaveLength(1)
                    for (const p of output.placements[0].geometry.exterior.points) {
                        expect(Number.isFinite(p.x) && Number.isFinite(p.y)).toBe(true)
                        expect(p.x).toBeGreaterThanOrEqual(0); expect(p.x).toBeLessThanOrEqual(100)
                        expect(p.y).toBeGreaterThanOrEqual(0); expect(p.y).toBeLessThanOrEqual(100)
                    }
                }
            }
            if (width < 700) await page.getByRole('button', { name: 'Панель', exact: true }).click()
            const box = await canvas.boundingBox()
            expect(box.width).toBeGreaterThan(50); expect(box.height).toBeGreaterThan(50)
            expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 2)).toBe(true)
            await expect(page.getByRole('button', { name: 'Автоматический раскрой', exact: true })).toBeVisible()
            await page.screenshot({ path: info.outputPath(`radial-${width}-${theme}.png`), fullPage: true })
            await info.attach("expected-anonymous-refresh", { body: JSON.stringify(anonymousRefresh), contentType: "application/json" })
            expect(errors).toEqual([])
            if (width === 1440 && theme === 'light') {
                // A real second-tab edit must survive replay of the completed operation.
                const startCalls = await count()
                const second = await page.context().newPage()
                try {
                    second.on('pageerror', () => errors.push('second-tab pageerror'))
                    await second.goto(page.url())
                    const secondCanvas = second.getByLabel('Рабочая поверхность раскроя', { exact: true })
                    const secondToggle = second.getByRole('button', { name: 'Лекала', exact: true })
                    await expect(secondCanvas).toBeVisible()
                    if ((await secondToggle.getAttribute('aria-expanded')) === 'false') await secondToggle.click()
                    await second.getByRole('button', { name: new RegExp(`Добавить на поверхность.*${patternName}`) }).click()
                    const target = await secondCanvas.evaluate((svg) => { const p = new DOMPoint(75, 75).matrixTransform(svg.getScreenCTM()); return { x: p.x, y: p.y } })
                    await second.mouse.click(target.x, target.y)
                    await expect.poll(async () => (await get(base + '/placements')).items.length).toBe(2)
                    const afterEdit = (await get(base + '/placements')).items
                    const replay = await request.post(api + base + '/nesting', {
                        data: savedBody,
                        headers: { Authorization: authorization, 'Idempotency-Key': savedKey },
                    })
                    expect(replay.status()).toBe(200)
                    expect(poses((await replay.json()).placements)).toEqual(poses(saved.placements))
                    expect((await get(base + '/placements')).items).toEqual(afterEdit)
                    expect(await count()).toBe(startCalls)
                    await page.reload()
                    await expect(placement).toHaveCount(2)
                } finally {
                    await second.close()
                }
                expect(errors).toEqual([])
            }
        })
    }
}
