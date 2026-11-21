// Keep failures inspectable without recording headers, cookies or storage.
export function installDiagnostics(test) {
    const observations = new WeakMap()
    test.beforeEach(async ({ page }) => {
        const events = []
        observations.set(page, events)
        const scrub = (value) =>
            String(value).replace(
                /[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]+/g,
                '[redacted]',
            )
        page.on('pageerror', (error) =>
            events.push({ type: 'pageerror', message: scrub(error.message) }),
        )
        page.on('requestfailed', (request) =>
            events.push({
                type: 'requestfailed',
                path: new URL(request.url()).pathname,
                error: request.failure()?.errorText,
            }),
        )
        page.on('response', (response) => {
            if (response.status() >= 500)
                events.push({
                    type: 'response',
                    path: new URL(response.url()).pathname,
                    status: response.status(),
                })
        })
    })
    test.afterEach(async ({ page }, info) => {
        if (info.status !== info.expectedStatus)
            await info.attach('client-diagnostics.json', {
                body: JSON.stringify(
                    { events: observations.get(page), path: new URL(page.url()).pathname },
                    null,
                    2,
                ),
                contentType: 'application/json',
            })
    })
}
