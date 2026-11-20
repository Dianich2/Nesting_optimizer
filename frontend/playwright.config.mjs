import { defineConfig } from '@playwright/test'

if (!process.env.NESTING_API_TARGET) {
    throw new Error('Set NESTING_API_TARGET to the API connected to a separate E2E database.')
}

const port = Number(process.env.NESTING_E2E_PORT || 5178)
if (!Number.isInteger(port) || port < 1024 || port > 65535) throw new Error('Invalid E2E port')

export default defineConfig({
    testDir: './e2e',
    timeout: 90_000,
    expect: { timeout: 5_000 },

    use: {
        baseURL: `http://localhost:${port}`,
        launchOptions: { executablePath: process.env.E2E_CHROME || undefined },
        actionTimeout: 10_000,
        trace: 'retain-on-failure',
        viewport: { width: 1440, height: 1000 },
    },
    webServer: {
        command: `npm run dev -- --host localhost --port ${port} --strictPort`,
        url: `http://localhost:${port}`,
        reuseExistingServer: false,
    },
})
