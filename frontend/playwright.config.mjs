import { defineConfig } from '@playwright/test'

if (!process.env.NESTING_API_TARGET) {
    throw new Error('Set NESTING_API_TARGET to the API connected to a separate E2E database.')
}

export default defineConfig({
    testDir: './e2e',
    timeout: 90_000,
    use: {
        baseURL: 'http://localhost:5178',
        launchOptions: { executablePath: process.env.E2E_CHROME || undefined },
        viewport: { width: 1440, height: 1000 },
    },
    webServer: {
        command: 'npm run dev -- --port 5178 --strictPort',
        url: 'http://localhost:5178',
        reuseExistingServer: false,
    },
})
