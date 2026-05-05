import { test as setup, expect } from '@playwright/test';

const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@local.test';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'ChangeMe123!';
const AUTH_FILE = 'tests/e2e/.auth/admin.json';

setup('authenticate as admin', async ({ page }) => {
	await page.goto('/login');
	await expect(page.locator('h1')).toContainText('Sign in');

	await page.locator('input[name="email"]').fill(ADMIN_EMAIL);
	await page.locator('input[name="password"]').fill(ADMIN_PASSWORD);

	await Promise.all([
		page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 10_000 }),
		page.locator('form button[type="submit"]').click()
	]);

	await expect(page).not.toHaveURL(/\/login/);
	await page.context().storageState({ path: AUTH_FILE });
});
