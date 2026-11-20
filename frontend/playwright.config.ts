import { defineConfig } from '@playwright/test';
export default defineConfig({
    testDir: './tests', timeout: 900_000, workers: 1,
    outputDir: '../artifacts/browser',
    reporter: [['list'], ['json', {outputFile: '../artifacts/validation/browser-report.json'}]],
    use: {
        baseURL: process.env.CCR_BASE_URL || 'https://codecrashreview.laytech.com.br',
        headless: true, screenshot: 'only-on-failure',
        launchOptions: process.env.CCR_CHROMIUM ? {executablePath: process.env.CCR_CHROMIUM} : {},
    },
});
