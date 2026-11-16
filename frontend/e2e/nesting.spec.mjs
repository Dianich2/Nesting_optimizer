import { expect, test } from '@playwright/test'
import { fileURLToPath } from 'node:url'
import { readFile } from 'node:fs/promises'

const fixture = (name) => fileURLToPath(new URL(`./fixtures/${name}`, import.meta.url))
const password = 'OnlyTestPassword2026'

test('saved project copies, manual edits, automatic run recovery and exports', async ({ page }) => {
    const login = `e2e${Date.now().toString(36)}`
    await page.goto('/register')
    await page.locator('#register-login').fill(login)
    await page.locator('#register-email').fill(`${login}@example.test`)
    await page.locator('#register-first-name').fill('Тест')
    await page.locator('#register-last-name').fill('Раскрой')
    await page.locator('#register-password').fill(password)
    await page.locator('#register-repeat-password').fill(password)
    await page.getByRole('button', { name: 'Зарегистрироваться' }).click()
    await expect(page).toHaveURL(/\/login/)
    await page.getByLabel('Логин или электронная почта').fill(login)
    await page.locator('#login-password').fill(password)
    await page.getByRole('button', { name: 'Войти', exact: true }).click()
    await expect(page).toHaveURL(/\/projects/)
    await page.goto('/library')

    for (const [section, name, source] of [
        ['Поверхности', 'Рабочий лист', 'offset-surface.svg'],
        ['Лекала', 'Угловое лекало', 'concave-pattern.svg'],
    ]) {
        await page
            .locator('.library-panel')
            .filter({ has: page.getByRole('heading', { name: section, exact: true }) })
            .getByRole('button', { name: /Добавить/ })
            .click()
        const dialog = page.getByRole('dialog')
        await dialog.getByLabel('Название', { exact: true }).fill(name)
        await dialog.locator('input[type=file]').setInputFiles(fixture(source))
        await expect(dialog.getByRole('button', { name: 'По точкам' })).toHaveAttribute(
            'aria-pressed',
            'true',
        )
        if (section === 'Поверхности') {
            await expect(dialog.getByRole('button', { name: /Отверстие 1/ })).toBeVisible()
            const canvas = dialog.locator('svg.geometry-preview')
            const before = await canvas.getAttribute('viewBox')
            await canvas.hover()
            await page.mouse.wheel(0, -240)
            await expect(canvas).not.toHaveAttribute('viewBox', before)
            const zoomed = await canvas.getAttribute('viewBox')
            await dialog.getByRole('button', { name: /Панорамирование|Перемещение поля/ }).click()
            const box = await canvas.boundingBox()
            await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
            await page.mouse.down()
            await page.mouse.move(box.x + box.width / 2 + 40, box.y + box.height / 2 + 25, {
                steps: 6,
            })
            await page.mouse.up()
            await expect(canvas).not.toHaveAttribute('viewBox', zoomed)
        }
        await dialog.getByRole('button', { name: 'Создать', exact: true }).click()
        await expect(dialog).toBeHidden()
        await expect(page.getByText(name, { exact: true })).toBeVisible()
    }

    await page.goto('/projects')
    await page.getByRole('button', { name: /Новый проект/ }).click()
    await page.getByPlaceholder('Название проекта').fill('Проект E2E')
    await page.getByRole('button', { name: 'Создать проект' }).click()
    const card = page.locator('article').filter({ hasText: 'Проект E2E' })
    await card.getByRole('link').first().click()
    await expect(page.getByRole('heading', { name: 'Проект E2E' })).toBeVisible()

    for (const [kind, copies] of [
        ['Поверхности', 3],
        ['Лекала', 1],
    ]) {
        await page
            .locator('.project-resource-panel')
            .filter({ has: page.getByRole('heading', { name: kind, exact: true }) })
            .getByRole('button', { name: 'Добавить из библиотеки' })
            .click()
        const picker = page.getByRole('dialog')
        for (let i = 0; i < copies; i++) {
            const added = page.waitForResponse(
                (res) => res.request().method() === 'POST' && res.status() === 201,
            )
            await picker.getByRole('button', { name: 'Добавить', exact: true }).click()
            await added
            await expect(picker.getByRole('status')).toContainText('Добавлен экземпляр')
        }
        await picker.getByRole('button', { name: 'Закрыть', exact: true }).click()
    }
    const addresses = page.getByRole('link', { name: 'Открыть раскрой' })
    await expect(addresses).toHaveCount(3)
    const urls = await addresses.evaluateAll((links) =>
        links.map((link) => link.getAttribute('href')),
    )
    expect(new Set(urls).size).toBe(3)
    const secondCopy = page.locator('li.project-resource-item').filter({
        has: page.locator(`a[href="${urls[1]}"]`),
    })
    await secondCopy.getByRole('button', { name: 'Редактировать' }).click()
    const editor = page.getByRole('dialog')
    await editor.getByLabel('Название', { exact: true }).fill('Изменённый лист')
    const vertex = editor.getByLabel('X точки 1', { exact: true })
    await vertex.fill('-5')
    await vertex.press('Tab')
    await editor.getByRole('button', { name: 'Сохранить изменения' }).click()
    await expect(editor).toBeHidden()
    await expect(page.getByText('Изменённый лист', { exact: true })).toHaveCount(1)
    await expect(page.getByText('Рабочий лист', { exact: true })).toHaveCount(2)
    await page.goto(urls[0])
    const canvas = page.getByLabel('Рабочая поверхность раскроя', { exact: true })
    await expect(canvas).toBeVisible()

    async function world(x, y) {
        return canvas.evaluate(
            (svg, { x, y }) => {
                const point = new DOMPoint(x, y).matrixTransform(svg.getScreenCTM())
                return { x: point.x, y: point.y }
            },
            { x, y },
        )
    }
    async function addAt(x, y) {
        await page.getByRole('button', { name: /Угловое лекало/ }).click()
        await page.getByRole('button', { name: /Разместить кликом/ }).click()
        const point = await world(x, y)
        await page.mouse.click(point.x, point.y)
    }

    await addAt(10, 20)
    const placement = canvas.getByRole('button', { name: /Размещение №/ })
    await expect(placement).toHaveCount(1)
    const from = await world(8, 18)
    const to = await world(18, 18)
    await page.mouse.move(from.x, from.y)
    await page.mouse.down()
    await page.mouse.move(to.x, to.y, { steps: 8 })
    await page.mouse.up()
    await expect
        .poll(async () => Number(await page.getByLabel('X центра').inputValue()))
        .toBeCloseTo(20, 1)
    await page.getByRole('button', { name: '90°' }).click()
    await page.getByRole('button', { name: 'Сохранить положение' }).click()
    await expect(page.getByLabel('Поворот, °')).toHaveValue('90')
    await page.getByRole('button', { name: 'Отменить действие' }).click()
    await expect(page.getByLabel('Поворот, °')).toHaveValue('0')
    await page.getByRole('button', { name: 'Повторить действие' }).click()
    await expect(page.getByLabel('Поворот, °')).toHaveValue('90')

    await addAt(20, 20)
    await expect(page.getByRole('alert')).toContainText('пересекает')
    await page.getByRole('button', { name: /Отменить добавление/ }).click()
    await addAt(80, 65)
    await expect(page.getByRole('alert')).toContainText('отверстие')
    await expect(placement).toHaveCount(1)
    await page.getByRole('button', { name: /Отменить добавление/ }).click()

    await page.getByRole('spinbutton', { name: /Количество для алгоритма/ }).fill('1')
    await page.route(
        '**/nesting',
        async (route) => {
            if (route.request().method() !== 'POST') return route.continue()
            const original = route.request()
            const status = await page.evaluate(
                async ({ url, headers, body }) => {
                    const result = await fetch(url, { method: 'POST', headers, body })
                    return result.status
                },
                {
                    url: original.url(),
                    headers: {
                        Authorization: original.headers().authorization,
                        'Content-Type': 'application/json',
                        'Idempotency-Key': original.headers()['idempotency-key'],
                    },
                    body: original.postData(),
                },
            )
            expect(status).toBe(200)
            await route.abort('failed')
        },
        { times: 1 },
    )
    await page.getByRole('button', { name: 'Запустить раскрой' }).click()
    await expect(page.getByText(/Связь с сервером прервалась/)).toBeVisible({ timeout: 20000 })
    await page.reload()
    await page.getByRole('button', { name: 'Проверить статус' }).click()
    await expect(page.getByText('Последний запуск в этой сессии')).toBeVisible()
    await expect(placement).toHaveCount(2)
    await expect(page.getByRole('table').getByRole('row')).toHaveCount(2)

    const downloaded = page.waitForEvent('download')
    await page.getByRole('button', { name: 'Экспорт SVG' }).click()
    const file = await downloaded
    expect(file.suggestedFilename()).toMatch(/\.svg$/)
    const exported = await readFile(await file.path(), 'utf8')
    expect(exported).toContain('fill-rule="evenodd"')
    expect(exported.match(/<path /g)).toHaveLength(3)

    await page
        .getByRole('combobox', { name: 'Поверхность', exact: true })
        .selectOption(urls[1].split('/').at(-2))
    await expect(placement).toHaveCount(0)
    await page.goto(urls[0])
    await expect(placement).toHaveCount(2)
    await placement.first().focus()
    await placement.first().press('Enter')
    await page.getByRole('button', { name: 'Удалить размещение', exact: true }).click()
    await page.getByRole('dialog').getByRole('button', { name: 'Отмена' }).click()
    await expect(placement).toHaveCount(2)
    for (const count of [1, 0]) {
        await placement.first().focus()
        await placement.first().press('Enter')
        await page.getByRole('button', { name: 'Удалить размещение', exact: true }).click()
        await page.getByRole('dialog').getByRole('button', { name: 'Удалить', exact: true }).click()
        await expect(placement).toHaveCount(count)
    }
    await page.getByRole('link', { name: /Состав проекта/ }).click()
    await page
        .locator('li.project-resource-item')
        .filter({ has: page.locator(`a[href="${urls[0]}"]`) })
        .getByRole('button', { name: 'Удалить', exact: true })
        .click()
    await page.getByRole('dialog').getByRole('button', { name: 'Удалить', exact: true }).click()
    await expect(page.getByRole('link', { name: 'Открыть раскрой' })).toHaveCount(2)
    await page.goto('/library')
    await page
        .locator('.library-item')
        .filter({ hasText: 'Рабочий лист' })
        .getByRole('button', { name: 'Удалить', exact: true })
        .click()
    await page.getByRole('dialog').getByRole('button', { name: 'Удалить', exact: true }).click()
    await expect(page.getByText('Рабочий лист', { exact: true })).toHaveCount(0)
    await page.goto(urls[1])
    await expect(page.getByRole('heading', { name: /Изменённый лист/ })).toBeVisible()
    await page.goto('/projects')
    await page.getByRole('button', { name: /Удалить проект.*Проект E2E/ }).click()
    await page
        .getByRole('dialog')
        .getByRole('button', { name: 'Удалить проект', exact: true })
        .click()
    await expect(card).toHaveCount(0)
})
