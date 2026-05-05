import { test as setup, expect } from '@playwright/test';
import fs from 'node:fs';
import { seedPersonas } from './lib/seed';

const AUTH_FILE = 'tests/e2e/.auth/teacher.json';

setup('authenticate as teacher', async ({ page }) => {
	const personas = await seedPersonas();
	fs.mkdirSync('tests/e2e/.auth', { recursive: true });

	await page.goto('/login');
	await page.locator('input[name="email"]').fill(personas.teacher.email);
	await page.locator('input[name="password"]').fill(personas.teacher.password);

	await Promise.all([
		page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 10_000 }),
		page.locator('form button[type="submit"]').click()
	]);

	await expect(page).not.toHaveURL(/\/login/);
	await page.context().storageState({ path: AUTH_FILE });
});
