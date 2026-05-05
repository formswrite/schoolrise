import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:3001';

export default defineConfig({
	testDir: './tests/e2e',
	timeout: 30_000,
	expect: { timeout: 5_000 },
	fullyParallel: false,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	workers: 1,
	reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],
	use: {
		baseURL,
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
		viewport: { width: 1440, height: 900 }
	},
	projects: [
		{
			name: 'setup',
			testMatch: /global\.setup\.ts/
		},
		{
			name: 'setup-teacher',
			testMatch: /global\.setup\.teacher\.ts/
		},
		{
			name: 'setup-personas',
			testMatch: /global\.setup\.personas\.ts/
		},
		{
			name: 'chromium',
			use: {
				...devices['Desktop Chrome'],
				viewport: { width: 1440, height: 900 },
				storageState: 'tests/e2e/.auth/admin.json'
			},
			dependencies: ['setup'],
			testIgnore: /(04-teacher-grade-entry|05-personas)\.spec\.ts/
		},
		{
			name: 'teacher',
			use: {
				...devices['Desktop Chrome'],
				viewport: { width: 1440, height: 900 },
				storageState: 'tests/e2e/.auth/teacher.json'
			},
			dependencies: ['setup-teacher'],
			testMatch: /04-teacher-grade-entry\.spec\.ts/
		},
		{
			name: 'personas',
			use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
			dependencies: ['setup-personas'],
			testMatch: /05-personas\.spec\.ts/
		}
	]
});
