import { expect, test } from '@playwright/test'
import { fileURLToPath } from 'node:url'
import { readFile, writeFile } from 'node:fs/promises'

const fixture = (name) => fileURLToPath(new URL(`./fixtures/${name}`, import.meta.url))
async function contourOccupancy(canvas, selector) {
    const contour = await canvas.locator(selector).boundingBox()
    const field = await canvas.boundingBox()
    return Math.max(contour.width / field.width, contour.height / field.height)
}

const password = 'OnlyTestPassword2026'

test('saved project copies, manual edits, automatic run recovery and exports', async ({
    page,
}, testInfo) => {
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
        ['Лекала', 'Прямоугольное лекало', null],
    ]) {
        await page
            .locator('.library-panel')
            .filter({ has: page.getByRole('heading', { name: section, exact: true }) })
            .getByRole('button', { name: /Добавить/ })
            .click()
        const dialog = page.getByRole('dialog')
        await dialog.getByLabel('Название', { exact: true }).fill(name)
        if (source) {
            await dialog.locator('input[type=file]').setInputFiles(fixture(source))
            await expect(dialog.getByRole('button', { name: 'По точкам' })).toHaveAttribute(
                'aria-pressed',
                'true',
            )
        } else {
            await dialog.getByLabel('Ширина', { exact: true }).fill('12')
            await dialog.getByLabel('Высота', { exact: true }).fill('8')
        }
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
            await dialog.getByRole('button', { name: 'Перемещение поля', exact: true }).click()
            await dialog.getByRole('button', { name: '+ Отверстие', exact: true }).click()
            for (const [x, y] of [
                [100, 100],
                [110, 100],
                [110, 110],
                [100, 110],
            ]) {
                await dialog
                    .getByRole('textbox', { name: 'X новой точки', exact: true })
                    .fill(String(x))
                await dialog
                    .getByRole('textbox', { name: 'Y новой точки', exact: true })
                    .fill(String(y))
                await dialog.getByRole('button', { name: 'Добавить по координатам' }).click()
            }
            await expect(
                dialog.getByRole('button', { name: 'Отверстие 2 (4)', exact: true }),
            ).toBeVisible()
            await dialog.getByRole('button', { name: 'Удалить отверстие', exact: true }).click()
            await expect(dialog.getByRole('button', { name: /Отверстие 2/ })).toHaveCount(0)
            await dialog.getByRole('button', { name: 'Отменить', exact: true }).click()
            await expect(
                dialog.getByRole('button', { name: 'Отверстие 2 (4)', exact: true }),
            ).toBeVisible()
            await dialog.getByRole('button', { name: 'Повторить', exact: true }).click()
            await expect(dialog.getByRole('button', { name: /Отверстие 2/ })).toHaveCount(0)
        }
        await dialog.getByRole('button', { name: 'Создать', exact: true }).click()
        await expect(dialog).toBeHidden()
        await expect(page.getByText(name, { exact: true })).toBeVisible()
    }

    const patternRow = page.locator('.library-item').filter({ hasText: 'Угловое лекало' })
    await patternRow.getByRole('button', { name: 'Редактировать' }).click()
    const patternEditor = page.getByRole('dialog')
    await patternEditor.getByRole('combobox', { name: /Вершина контура/ }).selectOption('0')
    const patternY = patternEditor.getByLabel('Y точки 1', { exact: true })
    await patternY.fill('-1')
    await patternY.press('Tab')
    await patternEditor.getByRole('button', { name: 'Сохранить изменения' }).click()
    await expect(patternEditor).toBeHidden()
    await patternRow.getByRole('button', { name: 'Редактировать' }).click()
    await expect
        .poll(() => patternEditor.locator('.geometry-vertex').last().getAttribute('cy'))
        .toBe('15')
    await patternEditor.getByRole('button', { name: /Назад/ }).click()

    await page.goto('/projects')
    await page.getByRole('button', { name: /Новый проект/ }).click()
    await page.getByPlaceholder('Название проекта').fill('Проект E2E')
    await page.getByRole('button', { name: 'Создать проект' }).click()
    const card = page.locator('article').filter({ hasText: 'Проект E2E' })
    await card.getByRole('link').first().click()
    await expect(page.getByRole('heading', { name: 'Проект E2E' })).toBeVisible()

    for (const [kind, copies] of [
        ['Поверхности', ['Рабочий лист', 'Рабочий лист', 'Рабочий лист']],
        ['Лекала', ['Угловое лекало', 'Прямоугольное лекало']],
    ]) {
        await page
            .locator('.project-resource-panel')
            .filter({ has: page.getByRole('heading', { name: kind, exact: true }) })
            .getByRole('button', { name: 'Добавить из библиотеки' })
            .click()
        const picker = page.getByRole('dialog')
        for (const name of copies) {
            const added = page.waitForResponse(
                (res) => res.request().method() === 'POST' && res.status() === 201,
            )
            await picker
                .getByRole('listitem')
                .filter({ hasText: name })
                .getByRole('button', { name: 'Добавить', exact: true })
                .click()
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
    const geometryCanvas = editor.locator('svg.geometry-preview')
    const toolsToggle = editor.getByRole('button', { name: 'Инструменты', exact: true })
    const propertiesToggle = editor.getByRole('button', { name: 'Свойства', exact: true })
    const viewBox = async () =>
        (await geometryCanvas.getAttribute('viewBox')).split(' ').map(Number)
    async function waitForCanvasSize() {
        await expect
            .poll(async () => {
                const [, , width, height] = await viewBox()
                const box = await geometryCanvas.boundingBox()
                return Math.abs(width / height - box.width / box.height)
            })
            .toBeLessThan(0.002)
    }
    await waitForCanvasSize()
    expect(await contourOccupancy(geometryCanvas, '.geometry-outline')).toBeGreaterThan(0.8)
    const [vx, vy, vw, vh] = await viewBox()
    const expanded = await geometryCanvas.boundingBox()
    await toolsToggle.click()
    await expect
        .poll(async () => (await geometryCanvas.boundingBox()).width)
        .toBeGreaterThan(expanded.width + 150)
    await propertiesToggle.click()
    await expect
        .poll(async () => (await geometryCanvas.boundingBox()).width)
        .toBeGreaterThan(1440 * 0.9)
    const [x, y, w, h] = await viewBox()
    expect(x + w / 2).toBeCloseTo(vx + vw / 2)
    expect(y + h / 2).toBeCloseTo(vy + vh / 2)
    expect(h).toBeCloseTo(vh)
    await propertiesToggle.click()
    await toolsToggle.click()

    const screenSizes = [
        [1920, 1080],
        [1440, 900],
        [1366, 768],
        [1024, 768],
        [390, 900],
    ]
    const geometryDimensions = []
    for (const [width, height] of screenSizes) {
        await page.setViewportSize({ width, height })
        await expect
            .poll(async () => (await geometryCanvas.boundingBox()).height)
            .toBeGreaterThan(height * (width < 700 ? 0.65 : 0.77))
        expect(await editor.evaluate((element) => element.scrollWidth)).toBeLessThanOrEqual(width)
        await waitForCanvasSize()
        await editor.getByRole('button', { name: 'Уместить в поле' }).click()
        const contour = await contourOccupancy(geometryCanvas, '.geometry-outline')
        expect(contour).toBeGreaterThan(0.8)
        const box = await geometryCanvas.boundingBox()
        geometryDimensions.push({
            contour,
            width,
            height,
            canvasWidth: box.width,
            canvasHeight: box.height,
        })
        await expect(editor.getByRole('button', { name: 'Сохранить изменения' })).toBeInViewport()
        if (width === 390) {
            await expect(toolsToggle).toHaveAttribute('aria-expanded', 'false')
            await propertiesToggle.click()
            await expect(propertiesToggle).toHaveAttribute('aria-expanded', 'true')
            await toolsToggle.click()
            await expect(propertiesToggle).toHaveAttribute('aria-expanded', 'false')
            await toolsToggle.click()
            await page.screenshot({ path: testInfo.outputPath('geometria-mobile.png') })
        }
    }
    await writeFile(
        testInfo.outputPath('geometry-layout.json'),
        JSON.stringify(geometryDimensions, null, 2),
    )
    await page.setViewportSize({ width: 1440, height: 1000 })
    await toolsToggle.click()
    await propertiesToggle.click()
    await editor.getByRole('button', { name: 'Уместить в поле' }).click()
    const firstVertex = geometryCanvas.locator('.geometry-vertex').first()
    const screen = await firstVertex.evaluate((circle) => {
        const point = new DOMPoint(
            circle.cx.baseVal.value,
            circle.cy.baseVal.value,
        ).matrixTransform(circle.getScreenCTM())
        return { x: point.x, y: point.y }
    })
    const oldX = Number(await firstVertex.getAttribute('cx'))
    await page.mouse.move(screen.x, screen.y)
    await page.mouse.down()
    await page.mouse.move(screen.x - 12, screen.y, { steps: 8 })
    await page.mouse.up()
    await expect.poll(async () => Number(await firstVertex.getAttribute('cx'))).toBeLessThan(oldX)
    await editor.getByRole('button', { name: 'Отменить', exact: true }).click()
    await expect(firstVertex).toHaveAttribute('cx', String(oldX))
    await editor.getByRole('button', { name: 'Повторить', exact: true }).click()
    await expect.poll(async () => Number(await firstVertex.getAttribute('cx'))).toBeLessThan(oldX)

    await editor.getByRole('button', { name: /Отверстие 1/ }).click()
    await expect(editor.getByRole('combobox', { name: /Вершина контура/ })).toBeVisible()
    await editor.getByRole('button', { name: /Внешний контур/ }).click()
    await editor.getByRole('combobox', { name: /Вершина контура/ }).selectOption('0')
    const vertex = editor.getByLabel('X точки 1', { exact: true })
    await vertex.fill('-5')
    await vertex.press('Tab')
    await editor.getByRole('button', { name: 'Сохранить изменения' }).click()
    await expect(editor).toBeHidden()
    await expect(page.getByText('Изменённый лист', { exact: true })).toHaveCount(1)
    await expect(page.getByText('Рабочий лист', { exact: true })).toHaveCount(2)
    await page.getByRole('button', { name: 'Включить тёмную тему' }).click()
    const accentAction = page.locator(`a[href="${urls[0]}"]`)
    await expect(accentAction).toHaveCSS('background-color', 'rgb(191, 48, 56)')
    const editedCopy = page
        .locator('li.project-resource-item')
        .filter({ hasText: 'Изменённый лист' })
    await editedCopy.getByRole('button', { name: 'Редактировать' }).click()
    await editor.getByRole('combobox', { name: /Вершина контура/ }).selectOption('0')
    const savedGeometry = await geometryCanvas
        .locator('.geometry-vertex')
        .evaluateAll((vertices) =>
            vertices.map((point) => [
                Number(point.getAttribute('cx')),
                Number(point.getAttribute('cy')),
            ]),
        )
    // The server normalizes the contour to its bounding-box origin.
    expect(Math.max(...savedGeometry.slice(0, 4).map(([x]) => x))).toBeGreaterThan(180)
    await waitForCanvasSize()
    await page.screenshot({ path: testInfo.outputPath('geometria-dark.png') })
    const darkGeometryDimensions = []
    for (const [width, height] of screenSizes.filter(([width]) => width >= 1024)) {
        await page.setViewportSize({ width, height })
        await waitForCanvasSize()
        await editor.getByRole('button', { name: 'Уместить в поле' }).click()
        const box = await geometryCanvas.boundingBox()
        const contour = await contourOccupancy(geometryCanvas, '.geometry-outline')
        expect(contour).toBeGreaterThan(0.8)
        expect(box.height).toBeGreaterThan(height * 0.77)
        expect(await editor.evaluate((element) => element.scrollWidth)).toBeLessThanOrEqual(width)
        darkGeometryDimensions.push({
            width,
            height,
            contour,
            canvasWidth: box.width,
            canvasHeight: box.height,
        })
        await page.screenshot({ path: testInfo.outputPath(`geometria-dark-${width}.png`) })
    }
    await writeFile(
        testInfo.outputPath('geometry-dark-layout.json'),
        JSON.stringify(darkGeometryDimensions, null, 2),
    )
    await page.setViewportSize({ width: 1440, height: 1000 })
    await editor.getByRole('button', { name: /Назад/ }).click()
    await page.getByRole('button', { name: 'Включить светлую тему' }).click()
    await expect(accentAction).toHaveCSS('background-color', 'rgb(191, 48, 56)')
    await editedCopy.getByRole('button', { name: 'Редактировать' }).click()
    await waitForCanvasSize()
    await page.screenshot({ path: testInfo.outputPath('geometria-light.png') })
    await editor.getByRole('button', { name: /Назад/ }).click()
    await accentAction.click()
    await expect(page).toHaveURL(new RegExp(urls[0] + '$'))
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
        await page.getByRole('button', { name: /Добавить на поверхность.*Угловое лекало/ }).click()
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
    await page.locator('.nesting-page summary').filter({ hasText: 'Ещё' }).click()
    await page.getByRole('button', { name: 'Отменить действие' }).click()
    await expect(page.getByLabel('Поворот, °')).toHaveValue('0')
    await page.getByRole('button', { name: 'Повторить действие' }).click()
    await expect(page.getByLabel('Поворот, °')).toHaveValue('90')
    await page.locator('.nesting-page summary').filter({ hasText: 'Ещё' }).click()

    await addAt(20, 20)
    await expect(page.getByRole('alert')).toContainText('пересекает')
    await page.getByRole('button', { name: /Отменить добавление/ }).click()
    await addAt(80, 65)
    await expect(page.getByRole('alert')).toContainText('отверстие')
    await expect(placement).toHaveCount(1)
    await page.getByRole('button', { name: /Отменить добавление/ }).click()

    await addAt(1, 1)
    await expect(page.getByRole('alert')).toContainText(/за границ/)
    const corrected = await world(40, 25)
    await page.mouse.click(corrected.x, corrected.y)
    await expect(placement).toHaveCount(2)
    await page.reload()
    await expect(placement).toHaveCount(2)

    const automatic = page.getByRole('button', { name: 'Автоматический раскрой', exact: true })
    await automatic.click()
    await page.getByRole('spinbutton', { name: 'Угловое лекало' }).fill('1')
    await page.getByRole('spinbutton', { name: 'Прямоугольное лекало' }).fill('0')
    await page.getByRole('button', { name: 'Предпросмотр', exact: true }).click()
    await expect(page.getByText(/Предпросмотр · размещено 1 из 1/)).toBeVisible()
    await expect(placement).toHaveCount(2)
    await page.screenshot({ path: testInfo.outputPath('raskroj-light-preview.png') })
    await page.getByRole('button', { name: 'Закрыть предпросмотр' }).click()
    await expect(placement).toHaveCount(2)
    await page.getByRole('combobox', { name: 'Алгоритм', exact: true }).selectOption('hybrid')
    await page.getByRole('spinbutton', { name: 'Прямоугольное лекало' }).fill('60')
    await page.getByRole('button', { name: 'Предпросмотр', exact: true }).click()
    const running = page.getByText('Расчёт предпросмотра…', { exact: true })
    await expect(running).toBeVisible()
    const beforeZoom = await canvas.getAttribute('viewBox')
    await canvas.hover()
    await page.mouse.wheel(0, -130)
    await expect(canvas).not.toHaveAttribute('viewBox', beforeZoom)
    const cancelled = page.waitForResponse(
        (res) =>
            res.request().method() === 'DELETE' &&
            res.url().includes('/nesting/operations/') &&
            res.status() === 204,
    )
    await page.getByRole('button', { name: 'Отменить расчёт' }).click()
    await cancelled
    await expect(page.getByRole('alert')).toContainText('Расчёт отменён')
    await expect(placement).toHaveCount(2)
    await page.getByRole('button', { name: 'Уместить поверхность' }).click()
    await page.getByRole('combobox', { name: 'Алгоритм', exact: true }).selectOption('baseline')
    await page.getByRole('spinbutton', { name: 'Прямоугольное лекало' }).fill('0')
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
    await page.getByRole('button', { name: 'Рассчитать и сохранить' }).click()
    await expect(page.getByText(/Связь с сервером прервалась/)).toBeVisible({ timeout: 20000 })
    await page.reload()
    await page.getByRole('button', { name: 'Проверить статус' }).click()
    await expect(placement).toHaveCount(3)
    await page.getByRole('button', { name: 'История', exact: true }).click()
    await expect(page.getByRole('table').getByRole('row')).toHaveCount(2)
    expect(
        await page
            .getByRole('table')
            .getByRole('columnheader')
            .first()
            .evaluate((cell) => cell.getBoundingClientRect().height),
    ).toBeLessThan(60)
    await page.getByRole('button', { name: 'Включить тёмную тему' }).click()
    await page.screenshot({ path: testInfo.outputPath('raskroj-dark.png') })
    await page.getByRole('button', { name: 'Включить светлую тему' }).click()

    const dimensions = []
    for (const [width, height] of screenSizes) {
        await page.setViewportSize({ width, height })
        await expect
            .poll(async () => {
                const box = await canvas.boundingBox()
                const [, , viewWidth, viewHeight] = (await canvas.getAttribute('viewBox'))
                    .split(' ')
                    .map(Number)
                return Math.abs(viewWidth / viewHeight - box.width / box.height)
            })
            .toBeLessThan(0.002)
        await page.getByRole('button', { name: 'Уместить поверхность' }).click()
        const contour = await contourOccupancy(canvas, '.nesting-surface-outline')
        expect(contour).toBeGreaterThan(0.8)
        const box = await canvas.boundingBox()
        expect(box.height).toBeGreaterThan(height * (width < 700 ? 0.6 : 0.67))
        expect(
            await page.locator('.nesting-page').evaluate((element) => element.scrollWidth),
        ).toBeLessThanOrEqual(width)
        dimensions.push({
            contour,
            width,
            height,
            canvasWidth: box.width,
            canvasHeight: box.height,
        })
        if (width === 390)
            await page.screenshot({ path: testInfo.outputPath('raskroj-mobile.png') })
    }
    await writeFile(testInfo.outputPath('nesting-layout.json'), JSON.stringify(dimensions, null, 2))
    await page.setViewportSize({ width: 1440, height: 900 })
    await expect.poll(async () => (await canvas.boundingBox()).width).toBeGreaterThan(1440 * 0.9)
    await page.getByRole('button', { name: 'Лекала', exact: true }).click()
    await page.getByRole('button', { name: 'Панель', exact: true }).click()
    await page.getByRole('button', { name: 'Включить тёмную тему' }).click()
    const darkNestingDimensions = []
    for (const [width, height] of screenSizes.filter(([width]) => width >= 1024)) {
        await page.setViewportSize({ width, height })
        await page.getByRole('button', { name: 'Уместить поверхность' }).click()
        await expect
            .poll(async () => {
                const box = await canvas.boundingBox()
                const [, , viewWidth, viewHeight] = (await canvas.getAttribute('viewBox'))
                    .split(' ')
                    .map(Number)
                return Math.abs(viewWidth / viewHeight - box.width / box.height)
            })
            .toBeLessThan(0.002)
        const box = await canvas.boundingBox()
        const contour = await contourOccupancy(canvas, '.nesting-surface-outline')
        expect(contour).toBeGreaterThan(0.8)
        expect(box.height).toBeGreaterThan(height * 0.67)
        expect(
            await page.locator('.nesting-page').evaluate((element) => element.scrollWidth),
        ).toBeLessThanOrEqual(width)
        darkNestingDimensions.push({
            width,
            height,
            contour,
            canvasWidth: box.width,
            canvasHeight: box.height,
        })
        await page.screenshot({ path: testInfo.outputPath(`raskroj-dark-${width}.png`) })
    }
    await writeFile(
        testInfo.outputPath('nesting-dark-layout.json'),
        JSON.stringify(darkNestingDimensions, null, 2),
    )
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.getByRole('button', { name: 'Включить светлую тему' }).click()

    const downloaded = page.waitForEvent('download')
    await page.locator('.nesting-page summary').filter({ hasText: 'Ещё' }).click()
    await page.getByRole('button', { name: 'Экспорт SVG' }).click()
    const file = await downloaded
    expect(file.suggestedFilename()).toMatch(/\.svg$/)
    const exported = await readFile(await file.path(), 'utf8')
    expect(exported).toContain('fill-rule="evenodd"')
    expect(exported.match(/<path /g)).toHaveLength(4)

    await page
        .getByRole('combobox', { name: 'Поверхность', exact: true })
        .selectOption(urls[1].split('/').at(-2))
    await expect(placement).toHaveCount(0)
    await page.goto(urls[0])
    await expect(placement).toHaveCount(3)
    await placement.first().focus()
    await placement.first().press('Enter')
    await page.getByRole('button', { name: 'Удалить размещение', exact: true }).click()
    await page.getByRole('dialog').getByRole('button', { name: 'Отмена' }).click()
    await expect(placement).toHaveCount(3)
    for (const count of [2, 1, 0]) {
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
