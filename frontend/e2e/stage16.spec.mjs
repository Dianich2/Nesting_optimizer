import { expect, test } from '@playwright/test'

const apiURL = `${process.env.NESTING_API_TARGET}/api/v1`
const polygon = (...xy) => ({ exterior: { points: Array.from({ length: xy.length / 2 }, (_, i) => ({ x: xy[2 * i], y: xy[2 * i + 1] })) }, holes: [] })
const rect = (w, h) => polygon(0, 0, w, 0, w, h, 0, h)
const poses = (items) => items.map(({ project_pattern_id, x, y, rotation, geometry }) => ({ project_pattern_id, x, y, rotation, geometry }))

async function session(page, prefix) {
    let authorization
    page.on('request', (r) => { if (r.headers().authorization) authorization = r.headers().authorization })
    const login = `${prefix}${Date.now().toString(36)}`
    const password = 'OnlySyntheticBrowser2026'
    await page.goto('/register')
    await page.locator('#register-login').fill(login)
    await page.locator('#register-email').fill(`${login}@example.test`)
    await page.locator('#register-first-name').fill('Browser')
    await page.locator('#register-last-name').fill('Acceptance')
    await page.locator('#register-password').fill(password)
    await page.locator('#register-repeat-password').fill(password)
    await page.getByRole('button', { name: 'Зарегистрироваться' }).click()
    await expect(page).toHaveURL(/\/login/)
    await page.getByLabel('Логин или электронная почта').fill(login)
    await page.locator('#login-password').fill(password)
    await page.getByRole('button', { name: 'Войти', exact: true }).click()
    await expect(page).toHaveURL(/\/projects/)
    await expect.poll(() => Boolean(authorization)).toBe(true)
    const call = async (method, path, data, status = 200, extra = {}) => {
        const response = await page.request.fetch(apiURL + path, { method, data, headers: { Authorization: authorization, ...extra } })
        expect(response.status(), `${method} ${path}`).toBe(status)
        return status === 204 ? null : response.json()
    }
    return { call }
}
async function dataset(call, name, geometries = [rect(10, 10)]) {
    const project = await call('POST', '/projects', { name }, 201)
    const original = await call('POST', '/surfaces', { name: 'Sheet', geometry: rect(100, 100) }, 201)
    const surface = await call('POST', `/projects/${project.id}/surfaces`, { source_surface_id: original.id, scale: 1 }, 201)
    const patterns = []
    for (let i = 0; i < geometries.length; i++) {
        const pattern = await call('POST', '/patterns', { name: `Part ${i}`, geometry: geometries[i] }, 201)
        patterns.push(await call('POST', `/projects/${project.id}/patterns`, { source_pattern_id: pattern.id, scale: 1 }, 201))
    }
    return { project, original, surface, patterns, base: `/projects/${project.id}/surfaces/${surface.id}` }
}
async function sqlStats(page, id) {
    const response = await page.request.get(`${process.env.NESTING_E2E_STATS}?surface_id=${id}`)
    expect(response.status()).toBe(200)
    return response.json()
}

test('strict diagonal contact API, SQL preservation, versions, owner isolation and keyboard', async ({ page, browser, request }) => {
    const { call } = await session(page, 'strict')
    const d = await dataset(call, 'Strict API', [polygon(0, 0, 2, 0, 0, 2), polygon(2, 0, 2, 2, 0, 2)])
    const place = (pattern, x, y) => ({ project_pattern_id: pattern.id, x, y, rotation: 0 })
    const first = await call('POST', d.base + '/placements', place(d.patterns[0], 10, 10), 201)
    const before = await sqlStats(page, d.surface.id)
    const microX = 10 - Number.EPSILON * 8 // one ULP below 10, exact stored float64
    const failure = await call('POST', d.base + '/placements', place(d.patterns[1], microX, 10), 409)
    expect(failure.error.details.some((item) => item.code === 'placement_overlap')).toBe(true)
    expect(await sqlStats(page, d.surface.id)).toEqual(before)
    const contact = await call('POST', d.base + '/placements', place(d.patterns[1], 10, 10), 201)
    const touching = await sqlStats(page, d.surface.id)
    expect(touching.placements).toBe(2)
    const rejected = await call('PATCH', `/projects/${d.project.id}/placements/${contact.id}`, { x: microX }, 409)
    expect(rejected.error.details.some((item) => item.code === 'placement_overlap')).toBe(true)
    expect(await sqlStats(page, d.surface.id)).toEqual(touching)
    expect((await call('GET', d.base + '/placements')).items.map((p) => p.id)).toEqual([first.id, contact.id])
    const anonymous = await request.get(apiURL + `/projects/${d.project.id}`)
    expect(anonymous.status()).toBe(401)
    await page.goto(d.base + '/nesting')
    const canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true })
    const placement = canvas.getByRole('button', { name: `Размещение №${contact.id}`, exact: true })
    await expect(placement).toBeVisible()
    await placement.focus(); await page.keyboard.press('Space')
    const remove = page.getByRole('button', { name: `Удалить размещение №${contact.id}`, exact: true })
    await expect(remove).toBeVisible()
    await remove.focus(); await page.keyboard.press('Enter')
    const dialog = page.getByRole('dialog')
    const cancel = dialog.getByRole('button', { name: 'Отмена', exact: true })
    await expect(cancel).toBeFocused()
    await page.keyboard.press('Tab')
    await expect(dialog.getByRole('button', { name: 'Удалить', exact: true })).toBeFocused()
    await page.keyboard.press('Shift+Tab'); await expect(cancel).toBeFocused()
    await page.keyboard.press('Escape'); await expect(dialog).toBeHidden()
    await expect(remove).toBeFocused()
    await page.keyboard.press('Space'); await expect(cancel).toBeFocused()
    await page.keyboard.press('Enter'); await expect(dialog).toBeHidden()
    expect((await sqlStats(page, d.surface.id)).layout_hash).toBe(touching.layout_hash)
    const external = await call('PATCH', `/projects/${d.project.id}/placements/${contact.id}`, { x: 30, y: 30 }, 200, { 'X-Placement-Version': contact.updated_at })
    expect(external.x).toBe(30)
    const updated = await sqlStats(page, d.surface.id)
    const xInput = page.getByLabel('X центра')
    await xInput.fill('20')
    const conflict = page.waitForResponse((r) => r.request().method() === 'PATCH' && r.url().includes(`/placements/${contact.id}`))
    // Leaving the properties form commits the draft. With multiple numeric
    // inputs and no submit button, Enter does not implicitly submit it.
    await canvas.focus()
    const response = await conflict
    expect(response.status()).toBe(409)
    expect(response.request().postDataJSON().x).toBe(20)
    expect(response.request().headers()['x-placement-version']).toBe(contact.updated_at)
    expect((await response.json()).error.details.some((item) => item.code === 'placement_version_conflict')).toBe(true)
    await expect(page.getByRole('status').filter({ hasText: 'Размещение изменилось после загрузки. Обновите раскладку.' })).toBeVisible()
    expect(await sqlStats(page, d.surface.id)).toEqual(updated)
    const refresh = page.waitForResponse((r) => r.url().endsWith('/auth/browser/refresh') && r.status() === 200)
    // The rejected draft is retained; explicitly accept its loss on reload.
    const acceptReload = async (dialog) => {
        expect(dialog.type()).toBe('beforeunload')
        await dialog.accept()
    }
    page.on('dialog', acceptReload)
    try {
        await page.reload(); await refresh
    } finally {
        page.off('dialog', acceptReload)
    }
    await expect(canvas.getByRole('button', { name: /Размещение №/ })).toHaveCount(2)
    const otherContext = await browser.newContext()
    try {
        const other = await otherContext.newPage()
        const stranger = await session(other, 'stranger')
        await stranger.call('GET', `/projects/${d.project.id}`, undefined, 404)
        await stranger.call('GET', d.base + '/placements', undefined, 404)
        await stranger.call('PATCH', `/projects/${d.project.id}/placements/${first.id}`, { x: 60 }, 404)
        expect(await sqlStats(page, d.surface.id)).toEqual(updated)
    } finally { await otherContext.close() }
})

test('all five native algorithms save once with real SQL history', async ({ page }) => {
    const { call } = await session(page, 'five')
    const d = await dataset(call, 'Five algorithms')
    const initial = (await sqlStats(page, d.surface.id)).optimize_calls
    let index = 0
    for (const algorithm of ['baseline', 'nfp_greedy', 'genetic', 'simulated_annealing', 'hybrid']) {
        const surface = index === 0 ? d.surface : await call('POST', `/projects/${d.project.id}/surfaces`, { source_surface_id: d.original.id, scale: 1 }, 201)
        const base = `/projects/${d.project.id}/surfaces/${surface.id}`
        await page.goto(base + '/nesting')
        await page.getByRole('button', { name: 'Автоматический раскрой', exact: true }).click()
        await page.getByRole('combobox', { name: 'Алгоритм', exact: true }).selectOption(algorithm)
        await page.getByRole('spinbutton', { name: 'Part 0', exact: true }).fill('1')
        const saving = page.waitForResponse((r) => r.url().endsWith('/nesting') && r.request().method() === 'POST')
        await page.getByRole('button', { name: 'Рассчитать и сохранить', exact: true }).click()
        const savedResponse = await saving
        expect(savedResponse.status()).toBe(200)
        expect(savedResponse.request().postDataJSON()).not.toHaveProperty('preview_id')
        const saved = await savedResponse.json()
        expect(saved.placements).toHaveLength(1)
        expect(saved.metrics).toMatchObject({ requested_count: 1, placed_count: 1 })
        await expect(page.getByText('Сохранено · размещено 1 из 1', { exact: true })).toBeVisible()
        expect(poses((await call('GET', base + '/placements')).items)).toEqual(poses(saved.placements))
        expect(await sqlStats(page, surface.id)).toMatchObject({ optimize_calls: initial + index + 1, placements: 1, nesting_runs: 1 })
        const body = savedResponse.request().postDataJSON()
        const key = savedResponse.request().headers()['idempotency-key']
        const replay = await call('POST', base + '/nesting', body, 200, { 'Idempotency-Key': key })
        expect(poses(replay.placements)).toEqual(poses(saved.placements))
        expect(await sqlStats(page, surface.id)).toMatchObject({ optimize_calls: initial + index + 1, placements: 1, nesting_runs: 1 })
        await page.reload()
        await expect(page.getByLabel('Рабочая поверхность раскроя', { exact: true }).getByRole('button', { name: /Размещение №/ })).toHaveCount(1)
        index++
    }
})

test.describe('Chrome touch emulation', () => {
    test.use({ isMobile: true, hasTouch: true, viewport: { width: 390, height: 844 } })
    test('real emulated tap, drag, pan and zoom controls', async ({ page }, info) => {
        const { call } = await session(page, 'touch')
        const d = await dataset(call, 'Touch')
        await page.goto(d.base + '/nesting')
        const canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true })
        await expect(canvas).toBeVisible()
        await canvas.evaluate((svg) => {
            window.stage16PointerTypes = []; window.stage16ActiveTouches = new Set()
            svg.addEventListener('pointerdown', (e) => {
                window.stage16PointerTypes.push(e.pointerType); window.stage16ActiveTouches.add(e.pointerId)
                window.stage16GestureInverse = svg.getScreenCTM().inverse()
                const p = new DOMPoint(e.clientX, e.clientY).matrixTransform(window.stage16GestureInverse)
                window.stage16LastGesture = { start: { x: p.x, y: p.y } }
            })
            svg.addEventListener('pointerup', (e) => {
                const p = new DOMPoint(e.clientX, e.clientY).matrixTransform(window.stage16GestureInverse)
                window.stage16LastGesture.end = { x: p.x, y: p.y }
            })
            for (const type of ['pointerup','pointercancel']) window.addEventListener(type, (e) => window.stage16ActiveTouches.delete(e.pointerId), true)
        })
        await page.getByRole('button', { name: 'Лекала', exact: true }).tap()
        await page.getByRole('button', { name: /Добавить на поверхность.*Part 0/ }).tap()
        const world = (x, y) => canvas.evaluate((svg, p) => { const point = new DOMPoint(p.x, p.y).matrixTransform(svg.getScreenCTM()); return { x: point.x, y: point.y } }, { x, y })
        const center = await world(50, 50)
        await page.touchscreen.tap(center.x, center.y)
        const placements = canvas.getByRole('button', { name: /Размещение №/ })
        await expect(placements).toHaveCount(1)
        const cancelAdding = page.getByRole('button', { name: /Отменить добавление/ })
        if (await cancelAdding.isVisible()) await cancelAdding.tap()
        // A successful draft now opens its properties, as required by stage21.
        // Close the mobile overlay before interacting with the underlying canvas.
        const propertiesPanel = page.getByRole('button', { name: 'Панель', exact: true })
        if (await propertiesPanel.getAttribute('aria-expanded') === 'true') await propertiesPanel.tap()
        const zoom = page.getByRole('button', { name: 'Приблизить', exact: true })
        await zoom.evaluate((button) => { window.stage16ZoomEvents = []; for (const type of ['pointerdown','pointerup','click']) button.addEventListener(type, (e) => window.stage16ZoomEvents.push({ type, pointerType: e.pointerType })) })
        const initialView = await canvas.getAttribute('viewBox')
        await zoom.tap()
        await expect(canvas).not.toHaveAttribute('viewBox', initialView)
        await page.getByRole('button', { name: 'Уместить поверхность', exact: true }).tap()
        const cdp = await page.context().newCDPSession(page)
        const drag = async (from, to) => {
            // Chromium generates a paced native touch gesture and waits for
            // completion. Rapid hand-written CDP events can leave its gesture
            // recognizer suppressing the next tap even after pointerup.
            await cdp.send('Input.synthesizeScrollGesture', {
                x: from.x, y: from.y, xDistance: to.x - from.x, yDistance: to.y - from.y,
                speed: 200, preventFling: true, gestureSourceType: 'touch',
            })
            await expect.poll(() => page.evaluate(() => window.stage16ActiveTouches.size)).toBe(0)
            await page.evaluate(() => new Promise((resolve) => requestAnimationFrame(resolve)))
        }
        await drag(await world(50, 50), await world(60, 50))
        // synthesizeScrollGesture includes Chromium's touch-slop distance.
        // Check the exact pointer coordinates actually delivered to the app,
        // rather than assuming its requested distance omits that slop.
        // SVG CTM projection has float32 rounding (~2.4e-6 here): this UI-only
        // 5e-5 bound is below one screen pixel, not a geometry overlap tolerance.
        const gesture = await page.evaluate(() => window.stage16LastGesture)
        expect(gesture.end.x - gesture.start.x).toBeGreaterThan(5)
        await expect.poll(async () => (await call('GET', d.base + '/placements')).items[0].x).toBeCloseTo(50 + gesture.end.x - gesture.start.x, 4)
        // Saving a drag opens the properties overlay again on mobile.
        if (await propertiesPanel.getAttribute('aria-expanded') === 'true')
            await propertiesPanel.tap()
        await expect(propertiesPanel).toHaveAttribute('aria-expanded', 'false')
        await page.getByRole('button', { name: 'Перемещение поля', exact: true }).tap()
        await expect(page.getByRole('button', { name: 'Перемещение поля', exact: true })).toHaveAttribute('aria-pressed', 'true')
        const original = await canvas.getAttribute('viewBox')
        const spot = await world(80, 80)
        await drag(spot, { x: spot.x + 25, y: spot.y + 20 })
        await expect(canvas).not.toHaveAttribute('viewBox', original)
        const panned = await canvas.getAttribute('viewBox')
        await zoom.tap()
        await expect(canvas).not.toHaveAttribute('viewBox', panned)
        expect(await page.evaluate(() => window.stage16ZoomEvents.filter((e) => e.type === 'click' && e.pointerType === 'touch').length)).toBe(2)
        await page.screenshot({ path: info.outputPath('touch-layout.png') })
        expect(await page.evaluate(() => window.stage16PointerTypes.every((type) => type === 'touch') && window.stage16PointerTypes.length >= 3)).toBe(true)
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 2)).toBe(true)
        expect(await sqlStats(page, d.surface.id)).toMatchObject({ placements: 1, nesting_runs: 0 })
        await cdp.detach()
    })
})
