import { expect } from '@playwright/test'

const apiURL = `${process.env.NESTING_API_TARGET}/api/v1`
export const polygon = (...xy) => ({
    exterior: {
        points: Array.from({ length: xy.length / 2 }, (_, i) => ({
            x: xy[2 * i],
            y: xy[2 * i + 1],
        })),
    },
    holes: [],
})
export const rect = (w, h) => polygon(0, 0, w, 0, w, h, 0, h)
export const poses = (items) =>
    items.map(({ project_pattern_id, x, y, rotation, geometry }) => ({
        project_pattern_id,
        x,
        y,
        rotation,
        geometry,
    }))

export async function session(page, prefix) {
    let authorization
    page.on('request', (r) => {
        if (r.headers().authorization) authorization = r.headers().authorization
    })
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
        const response = await page.request.fetch(apiURL + path, {
            method,
            data,
            headers: { Authorization: authorization, ...extra },
        })
        expect(response.status(), `${method} ${path}`).toBe(status)
        return status === 204 ? null : response.json()
    }
    return { call }
}
export async function dataset(call, name, geometries = [rect(10, 10)], sheet = rect(100, 100)) {
    const project = await call('POST', '/projects', { name }, 201)
    const original = await call('POST', '/surfaces', { name: 'Sheet', geometry: sheet }, 201)
    const surface = await call(
        'POST',
        `/projects/${project.id}/surfaces`,
        { source_surface_id: original.id, scale: 1 },
        201,
    )
    const patterns = []
    for (let i = 0; i < geometries.length; i++) {
        const pattern = await call(
            'POST',
            '/patterns',
            { name: `Part ${i}`, geometry: geometries[i] },
            201,
        )
        patterns.push(
            await call(
                'POST',
                `/projects/${project.id}/patterns`,
                { source_pattern_id: pattern.id, scale: 1 },
                201,
            ),
        )
    }
    return {
        project,
        original,
        surface,
        patterns,
        base: `/projects/${project.id}/surfaces/${surface.id}`,
    }
}
export async function sqlStats(page, id) {
    const response = await page.request.get(`${process.env.NESTING_E2E_STATS}?surface_id=${id}`)
    expect(response.status()).toBe(200)
    return response.json()
}
